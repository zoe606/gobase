package permission

// CreateRequest represents a request to create a new permission.
type CreateRequest struct {
	Resource string `json:"resource" validate:"required,min=2,max=50"`
	Action   string `json:"action" validate:"required,min=2,max=50"`
}
