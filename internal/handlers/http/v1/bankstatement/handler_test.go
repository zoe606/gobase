package bankstatement_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	bankstatementdto "go-boilerplate/internal/dto/bankstatement"
	"go-boilerplate/internal/handlers/http/v1/bankstatement"
	"go-boilerplate/pkg/jwt"
	"go-boilerplate/pkg/logger"
	"go-boilerplate/pkg/pagination"
)

//go:generate mockgen -source=../../../../usecase/contracts.go -destination=./mocks_test.go -package=bankstatement_test

const testSecret = "test-secret"

// generateTestToken creates a valid JWT token with bank-statement permissions for tests.
func generateTestToken(t *testing.T, jwtService jwt.Service) string {
	t.Helper()
	token, _, err := jwtService.GenerateAccessToken(1, "admin@test.com", "admin", []string{"bank-statement:read", "bank-statement:write", "bank-statement:delete"})
	require.NoError(t, err)
	return token
}

// authRequest creates an http.Request with Authorization header.
func authRequest(t *testing.T, method, url, token string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(method, url, http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func TestHandler_List(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBSUC := NewMockBankStatement(ctrl)
	jwtService := jwt.New(testSecret, 15*time.Minute, 24*time.Hour)
	l := logger.NewDevelopment()

	handler := bankstatement.New(mockBSUC, jwtService, l)

	app := fiber.New()
	handler.RegisterRoutes(app.Group("/v1"))

	token := generateTestToken(t, jwtService)

	tests := []struct {
		name       string
		setupMock  func()
		wantStatus int
	}{
		{
			name: "success",
			setupMock: func() {
				mockBSUC.EXPECT().
					List(gomock.Any(), gomock.Any()).
					Return(&bankstatementdto.ListResponse{
						Data: []*bankstatementdto.Response{
							{ID: 1, BankName: "BCA", BankCode: "BCA", Status: "completed"},
							{ID: 2, BankName: "BRI", BankCode: "BRI", Status: "pending"},
						},
						Meta: pagination.NewMeta(1, 20, 2),
					}, nil)
			},
			wantStatus: fiber.StatusOK,
		},
		{
			name: "empty list",
			setupMock: func() {
				mockBSUC.EXPECT().
					List(gomock.Any(), gomock.Any()).
					Return(&bankstatementdto.ListResponse{
						Data: []*bankstatementdto.Response{},
						Meta: pagination.NewMeta(1, 20, 0),
					}, nil)
			},
			wantStatus: fiber.StatusOK,
		},
		{
			name: "usecase error",
			setupMock: func() {
				mockBSUC.EXPECT().
					List(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("internal error"))
			},
			wantStatus: fiber.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := authRequest(t, "GET", "/v1/bank-statements/", token)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestHandler_List_Unauthorized(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBSUC := NewMockBankStatement(ctrl)
	jwtService := jwt.New(testSecret, 15*time.Minute, 24*time.Hour)
	l := logger.NewDevelopment()

	handler := bankstatement.New(mockBSUC, jwtService, l)

	app := fiber.New()
	handler.RegisterRoutes(app.Group("/v1"))

	req := httptest.NewRequest("GET", "/v1/bank-statements/", http.NoBody)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestHandler_GetByID(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBSUC := NewMockBankStatement(ctrl)
	jwtService := jwt.New(testSecret, 15*time.Minute, 24*time.Hour)
	l := logger.NewDevelopment()

	handler := bankstatement.New(mockBSUC, jwtService, l)

	app := fiber.New()
	handler.RegisterRoutes(app.Group("/v1"))

	token := generateTestToken(t, jwtService)

	tests := []struct {
		name       string
		id         string
		setupMock  func()
		wantStatus int
	}{
		{
			name: "success",
			id:   "1",
			setupMock: func() {
				mockBSUC.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(&bankstatementdto.ResponseWithItems{
						Response: bankstatementdto.Response{
							ID:       1,
							BankName: "BCA",
							BankCode: "BCA",
							Status:   "completed",
						},
						Items: []*bankstatementdto.LineItemResponse{
							{ID: 1, Date: "2025-01-01", Description: "Transaction 1", Debit: 100000},
						},
					}, nil)
			},
			wantStatus: fiber.StatusOK,
		},
		{
			name:       "invalid id",
			id:         "abc",
			setupMock:  func() {},
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name: "not found",
			id:   "99",
			setupMock: func() {
				mockBSUC.EXPECT().
					GetByID(gomock.Any(), uint(99)).
					Return(nil, errors.New("record not found"))
			},
			wantStatus: fiber.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := authRequest(t, "GET", "/v1/bank-statements/"+tt.id, token)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestHandler_Delete(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBSUC := NewMockBankStatement(ctrl)
	jwtService := jwt.New(testSecret, 15*time.Minute, 24*time.Hour)
	l := logger.NewDevelopment()

	handler := bankstatement.New(mockBSUC, jwtService, l)

	app := fiber.New()
	handler.RegisterRoutes(app.Group("/v1"))

	token := generateTestToken(t, jwtService)

	tests := []struct {
		name       string
		id         string
		setupMock  func()
		wantStatus int
	}{
		{
			name: "success",
			id:   "1",
			setupMock: func() {
				mockBSUC.EXPECT().
					Delete(gomock.Any(), uint(1)).
					Return(nil)
			},
			wantStatus: fiber.StatusNoContent,
		},
		{
			name:       "invalid id",
			id:         "abc",
			setupMock:  func() {},
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name: "not found",
			id:   "99",
			setupMock: func() {
				mockBSUC.EXPECT().
					Delete(gomock.Any(), uint(99)).
					Return(errors.New("record not found"))
			},
			wantStatus: fiber.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := authRequest(t, "DELETE", "/v1/bank-statements/"+tt.id, token)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestHandler_ListBanks(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBSUC := NewMockBankStatement(ctrl)
	jwtService := jwt.New(testSecret, 15*time.Minute, 24*time.Hour)
	l := logger.NewDevelopment()

	handler := bankstatement.New(mockBSUC, jwtService, l)

	app := fiber.New()
	handler.RegisterRoutes(app.Group("/v1"))

	token := generateTestToken(t, jwtService)

	tests := []struct {
		name       string
		setupMock  func()
		wantStatus int
	}{
		{
			name: "success",
			setupMock: func() {
				mockBSUC.EXPECT().
					ListBanks(gomock.Any()).
					Return(&bankstatementdto.BankListResponse{
						Data: []*bankstatementdto.BankResponse{
							{ID: 1, Name: "BCA", Code: "BCA"},
							{ID: 2, Name: "BRI", Code: "BRI"},
						},
					}, nil)
			},
			wantStatus: fiber.StatusOK,
		},
		{
			name: "empty list",
			setupMock: func() {
				mockBSUC.EXPECT().
					ListBanks(gomock.Any()).
					Return(&bankstatementdto.BankListResponse{
						Data: []*bankstatementdto.BankResponse{},
					}, nil)
			},
			wantStatus: fiber.StatusOK,
		},
		{
			name: "usecase error",
			setupMock: func() {
				mockBSUC.EXPECT().
					ListBanks(gomock.Any()).
					Return(nil, errors.New("internal error"))
			},
			wantStatus: fiber.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := authRequest(t, "GET", "/v1/banks/", token)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}
