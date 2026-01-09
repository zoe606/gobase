package tasks

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/gif"  // GIF support
	_ "image/jpeg" // JPEG support
	_ "image/png"  // PNG support
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/hibiken/asynq"

	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
	"go-boilerplate/internal/repo/storage"
	"go-boilerplate/pkg/json"
	"go-boilerplate/pkg/logger"
)

// ImageProcessingPayload contains the payload for image processing.
type ImageProcessingPayload struct {
	MediaID uint `json:"media_id"`
}

// ImageProcessingHandler handles image processing tasks.
type ImageProcessingHandler struct {
	l         logger.Interface
	mediaRepo repo.MediaRepo
	storage   storage.Provider
	variants  []entity.VariantConfig
}

// NewImageProcessingHandler creates a new image processing handler.
func NewImageProcessingHandler(l logger.Interface, mediaRepo repo.MediaRepo, storage storage.Provider) *ImageProcessingHandler {
	return &ImageProcessingHandler{
		l:         l,
		mediaRepo: mediaRepo,
		storage:   storage,
		variants:  entity.DefaultImageVariants(),
	}
}

// ProcessTask processes an image processing task.
func (h *ImageProcessingHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload ImageProcessingPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	h.l.Info("Processing image", "media_id", payload.MediaID)

	// Get media record.
	media, err := h.mediaRepo.GetByID(ctx, payload.MediaID)
	if err != nil {
		return fmt.Errorf("get media: %w", err)
	}

	if !media.IsImage() {
		h.l.Info("Skipping non-image media", "media_id", payload.MediaID)
		return nil
	}

	// Download original.
	reader, err := h.storage.Get(ctx, media.Path)
	if err != nil {
		return fmt.Errorf("get original: %w", err)
	}
	defer reader.Close()

	// Decode image.
	img, format, err := image.Decode(reader)
	if err != nil {
		return fmt.Errorf("decode image: %w", err)
	}

	// Get dimensions.
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Generate variants.
	variants := make(map[string]interface{})
	for _, variant := range h.variants {
		variantPath, err := h.generateVariant(ctx, media, img, variant, format)
		if err != nil {
			h.l.Error(err, "generate variant", "variant", variant.Name)
			continue
		}
		variants[variant.Name] = variantPath
	}

	// Update media record.
	media.Width = &width
	media.Height = &height
	media.Variants = variants

	if err := h.mediaRepo.Update(ctx, media); err != nil {
		return fmt.Errorf("update media: %w", err)
	}

	h.l.Info("Image processed",
		"media_id", payload.MediaID,
		"width", width,
		"height", height,
		"variants", len(variants),
	)
	return nil
}

// generateVariant generates a single image variant.
func (h *ImageProcessingHandler) generateVariant(ctx context.Context, media *entity.Media, img image.Image, variant entity.VariantConfig, format string) (string, error) {
	var processed *image.NRGBA

	switch variant.Method {
	case "fill":
		processed = imaging.Fill(img, variant.Width, variant.Height, imaging.Center, imaging.Lanczos)
	case "fit":
		processed = imaging.Fit(img, variant.Width, variant.Height, imaging.Lanczos)
	default:
		processed = imaging.Resize(img, variant.Width, variant.Height, imaging.Lanczos)
	}

	// Encode to buffer.
	var buf bytes.Buffer
	ext := filepath.Ext(media.Path)
	mimeType := media.MimeType

	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		if err := imaging.Encode(&buf, processed, imaging.JPEG, imaging.JPEGQuality(variant.Quality)); err != nil {
			return "", fmt.Errorf("encode jpeg: %w", err)
		}
	case "png":
		if err := imaging.Encode(&buf, processed, imaging.PNG); err != nil {
			return "", fmt.Errorf("encode png: %w", err)
		}
	case "gif":
		if err := imaging.Encode(&buf, processed, imaging.GIF); err != nil {
			return "", fmt.Errorf("encode gif: %w", err)
		}
	default:
		// Convert unknown formats to JPEG.
		if err := imaging.Encode(&buf, processed, imaging.JPEG, imaging.JPEGQuality(variant.Quality)); err != nil {
			return "", fmt.Errorf("encode jpeg fallback: %w", err)
		}
		ext = ".jpg"
		mimeType = "image/jpeg"
	}

	// Generate variant path.
	variantPath := strings.TrimSuffix(media.Path, filepath.Ext(media.Path)) + "_" + variant.Name + ext

	// Upload variant.
	_, err := h.storage.Put(ctx, variantPath, &buf, int64(buf.Len()), mimeType)
	if err != nil {
		return "", fmt.Errorf("upload variant: %w", err)
	}

	h.l.Info("Generated variant",
		"media_id", media.ID,
		"variant", variant.Name,
		"path", variantPath,
	)

	return variantPath, nil
}

// NewImageProcessingTask creates a new image processing task.
func NewImageProcessingTask(mediaID uint) (*asynq.Task, error) {
	payload, err := json.Marshal(ImageProcessingPayload{MediaID: mediaID})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeImageProcessing, payload), nil
}
