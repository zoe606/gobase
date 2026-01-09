// Package media implements the media use case.
package media

import (
	"fmt"
	"strings"
	"time"

	"github.com/hibiken/asynq"

	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
	"go-boilerplate/internal/repo/storage"
	"go-boilerplate/internal/worker/tasks"
	"go-boilerplate/pkg/logger"
)

// UseCase implements the media use case.
type UseCase struct {
	mediaRepo   repo.MediaRepo
	storage     storage.Provider
	asynqClient *asynq.Client
	l           logger.Interface
	disk        string
	maxSize     int64
}

// New creates a new media use case.
func New(
	mediaRepo repo.MediaRepo,
	storage storage.Provider,
	asynqClient *asynq.Client,
	l logger.Interface,
	disk string,
	maxSize int64,
) *UseCase {
	return &UseCase{
		mediaRepo:   mediaRepo,
		storage:     storage,
		asynqClient: asynqClient,
		l:           l,
		disk:        disk,
		maxSize:     maxSize,
	}
}

// queueImageProcessing queues an image for background processing.
func (uc *UseCase) queueImageProcessing(mediaID uint) {
	if uc.asynqClient == nil {
		return
	}

	task, err := tasks.NewImageProcessingTask(mediaID)
	if err != nil {
		uc.l.Error(err, "Failed to create image processing task")
		return
	}

	_, err = uc.asynqClient.Enqueue(task)
	if err != nil {
		uc.l.Error(err, "Failed to enqueue image processing task")
	}
}

// buildStoragePath creates a structured storage path.
func buildStoragePath(attachableType, collection string, attachableID uint, filename string) string {
	return fmt.Sprintf("%s/%s/%d/%s/%s",
		attachableType,
		collection,
		attachableID,
		time.Now().Format("2006/01"),
		filename,
	)
}

// detectMediaType determines the media type from MIME type.
func detectMediaType(mimeType string) entity.MediaType {
	switch {
	case strings.HasPrefix(mimeType, "image/"):
		return entity.MediaTypeImage
	case strings.HasPrefix(mimeType, "video/"):
		return entity.MediaTypeVideo
	case strings.HasPrefix(mimeType, "audio/"):
		return entity.MediaTypeAudio
	case mimeType == "application/pdf",
		strings.HasPrefix(mimeType, "application/msword"),
		strings.HasPrefix(mimeType, "application/vnd."):
		return entity.MediaTypeDocument
	default:
		return entity.MediaTypeOther
	}
}
