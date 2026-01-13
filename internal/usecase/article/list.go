package article

import (
	"context"
	"fmt"

	articledto "go-boilerplate/internal/dto/article"
)

// List retrieves a paginated list of articles.
func (uc *UseCase) List(ctx context.Context, req articledto.ListRequest) (*articledto.ListResponse, error) {
	pageSize := req.GetPageSize()
	offset := req.GetOffset()
	page := req.Page
	if page <= 0 {
		page = 1
	}

	articles, total, err := uc.articleRepo.List(ctx, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("article - List - articleRepo.List: %w", err)
	}

	return articledto.NewListResponse(articles, total, page, pageSize), nil
}
