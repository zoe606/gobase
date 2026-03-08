package article

import (
	"context"
	"fmt"

	articledto "go-boilerplate/internal/dto/article"
)

// List retrieves a paginated list of articles with filters.
func (uc *UseCase) List(ctx context.Context, req articledto.ListRequest) (*articledto.ListResponse, error) {
	req.Normalize()

	articles, total, err := uc.articleRepo.List(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("article - List - articleRepo.List: %w", err)
	}

	return articledto.NewListResponse(articles, total, req.Params), nil
}
