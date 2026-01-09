package entity

// VariantConfig defines how to generate an image variant.
type VariantConfig struct {
	Name    string `json:"name"`    // "thumb", "medium", "large"
	Width   int    `json:"width"`   // Target width
	Height  int    `json:"height"`  // Target height
	Method  string `json:"method"`  // "fit", "fill", "resize"
	Quality int    `json:"quality"` // JPEG quality (1-100)
}

// DefaultImageVariants returns common image variants for processing.
func DefaultImageVariants() []VariantConfig {
	return []VariantConfig{
		{Name: "thumb", Width: 150, Height: 150, Method: "fill", Quality: 80},
		{Name: "medium", Width: 600, Height: 600, Method: "fit", Quality: 85},
		{Name: "large", Width: 1200, Height: 1200, Method: "fit", Quality: 90},
	}
}

// AllowedImageMimeTypes returns the list of allowed image MIME types.
func AllowedImageMimeTypes() []string {
	return []string{
		"image/jpeg",
		"image/png",
		"image/gif",
		"image/webp",
	}
}

// AllowedDocumentMimeTypes returns the list of allowed document MIME types.
func AllowedDocumentMimeTypes() []string {
	return []string{
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"text/plain",
		"text/csv",
	}
}

// AllowedVideoMimeTypes returns the list of allowed video MIME types.
func AllowedVideoMimeTypes() []string {
	return []string{
		"video/mp4",
		"video/webm",
		"video/quicktime",
	}
}

// AllowedAudioMimeTypes returns the list of allowed audio MIME types.
func AllowedAudioMimeTypes() []string {
	return []string{
		"audio/mpeg",
		"audio/wav",
		"audio/ogg",
		"audio/webm",
	}
}

// IsAllowedMimeType checks if a MIME type is in the allowed list.
func IsAllowedMimeType(mimeType string, allowedTypes []string) bool {
	for _, allowed := range allowedTypes {
		if mimeType == allowed {
			return true
		}
	}
	return false
}

// DefaultAllowedMimeTypes returns all commonly allowed MIME types.
func DefaultAllowedMimeTypes() []string {
	allowed := make([]string, 0, 20)
	allowed = append(allowed, AllowedImageMimeTypes()...)
	allowed = append(allowed, AllowedDocumentMimeTypes()...)
	allowed = append(allowed, AllowedVideoMimeTypes()...)
	allowed = append(allowed, AllowedAudioMimeTypes()...)
	return allowed
}
