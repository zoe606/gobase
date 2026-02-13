package permission

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	permissiondto "go-boilerplate/internal/dto/permission"
	v1 "go-boilerplate/internal/handlers/http/v1"
	permissionusecase "go-boilerplate/internal/usecase/permission"
	"go-boilerplate/pkg/response"
)

// Create godoc
// @Summary     Create permission
// @Description Create a new permission
// @ID          permission-create
// @Tags        permissions
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       request body permissiondto.CreateRequest true "Permission data"
// @Success     201 {object} response.Response[permissiondto.Response]
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     409 {object} response.ErrorResponse "Permission already exists"
// @Failure     500 {object} response.ErrorResponse
// @Router      /permissions [post]
func (h *Handler) Create(ctx *fiber.Ctx) error {
	var req permissiondto.CreateRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, "INVALID_JSON", "Invalid request body")
	}

	if err := h.v.Struct(req); err != nil {
		return response.ValidationError(ctx, v1.ParseValidationErrors(err))
	}

	result, err := h.permissionUC.Create(ctx.UserContext(), req)
	if err != nil {
		if errors.Is(err, permissionusecase.ErrPermissionExists) {
			return response.Conflict(ctx, "Permission already exists")
		}
		h.l.Error(err, "handlers - http - v1 - permission - Create")
		return response.InternalError(ctx)
	}

	return response.Created(ctx, result)
}
