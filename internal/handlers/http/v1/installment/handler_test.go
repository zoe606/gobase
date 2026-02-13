package installment_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	installmentdto "go-boilerplate/internal/dto/installment"
	"go-boilerplate/internal/handlers/http/v1/installment"
	"go-boilerplate/pkg/jwt"
	"go-boilerplate/pkg/logger"
	"go-boilerplate/pkg/pagination"
)

//go:generate mockgen -source=../../../../usecase/contracts.go -destination=./mocks_test.go -package=installment_test

const testSecret = "test-secret"

// generateTestToken creates a valid JWT token with installment permissions for tests.
func generateTestToken(t *testing.T, jwtService jwt.Service) string {
	t.Helper()
	token, _, err := jwtService.GenerateAccessToken(1, "admin@test.com", "admin", []string{"installment:read", "installment:write", "installment:delete"})
	require.NoError(t, err)
	return token
}

// authRequest creates an http.Request with Authorization header.
func authRequest(t *testing.T, method, url string, body *bytes.Buffer, token string) *http.Request {
	t.Helper()
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, url, body)
	} else {
		req = httptest.NewRequest(method, url, http.NoBody)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func TestHandler_Create(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockInstUC := NewMockInstallment(ctrl)
	jwtService := jwt.New(testSecret, 15*time.Minute, 24*time.Hour)
	l := logger.NewDevelopment()

	handler := installment.New(mockInstUC, jwtService, l)

	app := fiber.New()
	handler.RegisterRoutes(app.Group("/v1"))

	token := generateTestToken(t, jwtService)

	tests := []struct {
		name       string
		body       string
		setupMock  func()
		wantStatus int
	}{
		{
			name: "success",
			body: `{"name":"Home Loan","total_amount":500000000,"monthly_amount":5000000,"total_terms":120}`,
			setupMock: func() {
				mockInstUC.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(&installmentdto.Response{
						ID:            1,
						Name:          "Home Loan",
						TotalAmount:   500000000,
						MonthlyAmount: 5000000,
						TotalTerms:    120,
						Status:        "active",
						CreatedAt:     "2025-01-01T00:00:00Z",
					}, nil)
			},
			wantStatus: fiber.StatusCreated,
		},
		{
			name:       "invalid json",
			body:       `{invalid}`,
			setupMock:  func() {},
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name:       "validation error - missing name",
			body:       `{"total_amount":500000000,"monthly_amount":5000000,"total_terms":120}`,
			setupMock:  func() {},
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name:       "validation error - missing total_amount",
			body:       `{"name":"Home Loan","monthly_amount":5000000,"total_terms":120}`,
			setupMock:  func() {},
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name: "usecase error",
			body: `{"name":"Home Loan","total_amount":500000000,"monthly_amount":5000000,"total_terms":120}`,
			setupMock: func() {
				mockInstUC.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("internal error"))
			},
			wantStatus: fiber.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := authRequest(t, "POST", "/v1/installments/", bytes.NewBufferString(tt.body), token)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestHandler_Create_Unauthorized(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockInstUC := NewMockInstallment(ctrl)
	jwtService := jwt.New(testSecret, 15*time.Minute, 24*time.Hour)
	l := logger.NewDevelopment()

	handler := installment.New(mockInstUC, jwtService, l)

	app := fiber.New()
	handler.RegisterRoutes(app.Group("/v1"))

	req := httptest.NewRequest("POST", "/v1/installments/", bytes.NewBufferString(`{"name":"Test","total_amount":100,"monthly_amount":10,"total_terms":12}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestHandler_List(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockInstUC := NewMockInstallment(ctrl)
	jwtService := jwt.New(testSecret, 15*time.Minute, 24*time.Hour)
	l := logger.NewDevelopment()

	handler := installment.New(mockInstUC, jwtService, l)

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
				mockInstUC.EXPECT().
					List(gomock.Any(), gomock.Any()).
					Return(&installmentdto.ListResponse{
						Data: []*installmentdto.Response{
							{ID: 1, Name: "Home Loan", Status: "active", CreatedAt: "2025-01-01T00:00:00Z"},
							{ID: 2, Name: "Car Loan", Status: "active", CreatedAt: "2025-01-02T00:00:00Z"},
						},
						Meta: pagination.NewMeta(1, 20, 2),
					}, nil)
			},
			wantStatus: fiber.StatusOK,
		},
		{
			name: "empty list",
			setupMock: func() {
				mockInstUC.EXPECT().
					List(gomock.Any(), gomock.Any()).
					Return(&installmentdto.ListResponse{
						Data: []*installmentdto.Response{},
						Meta: pagination.NewMeta(1, 20, 0),
					}, nil)
			},
			wantStatus: fiber.StatusOK,
		},
		{
			name: "usecase error",
			setupMock: func() {
				mockInstUC.EXPECT().
					List(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("internal error"))
			},
			wantStatus: fiber.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := authRequest(t, "GET", "/v1/installments/", nil, token)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestHandler_GetByID(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockInstUC := NewMockInstallment(ctrl)
	jwtService := jwt.New(testSecret, 15*time.Minute, 24*time.Hour)
	l := logger.NewDevelopment()

	handler := installment.New(mockInstUC, jwtService, l)

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
				mockInstUC.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(&installmentdto.Response{
						ID:            1,
						Name:          "Home Loan",
						TotalAmount:   500000000,
						MonthlyAmount: 5000000,
						TotalTerms:    120,
						Status:        "active",
						CreatedAt:     "2025-01-01T00:00:00Z",
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
			name: "usecase error",
			id:   "99",
			setupMock: func() {
				mockInstUC.EXPECT().
					GetByID(gomock.Any(), uint(99)).
					Return(nil, errors.New("record not found"))
			},
			wantStatus: fiber.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := authRequest(t, "GET", "/v1/installments/"+tt.id, nil, token)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestHandler_Update(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockInstUC := NewMockInstallment(ctrl)
	jwtService := jwt.New(testSecret, 15*time.Minute, 24*time.Hour)
	l := logger.NewDevelopment()

	handler := installment.New(mockInstUC, jwtService, l)

	app := fiber.New()
	handler.RegisterRoutes(app.Group("/v1"))

	token := generateTestToken(t, jwtService)

	tests := []struct {
		name       string
		id         string
		body       string
		setupMock  func()
		wantStatus int
	}{
		{
			name: "success",
			id:   "1",
			body: `{"name":"Updated Loan"}`,
			setupMock: func() {
				mockInstUC.EXPECT().
					Update(gomock.Any(), uint(1), gomock.Any()).
					Return(&installmentdto.Response{
						ID:            1,
						Name:          "Updated Loan",
						TotalAmount:   500000000,
						MonthlyAmount: 5000000,
						TotalTerms:    120,
						Status:        "active",
						CreatedAt:     "2025-01-01T00:00:00Z",
					}, nil)
			},
			wantStatus: fiber.StatusOK,
		},
		{
			name:       "invalid id",
			id:         "abc",
			body:       `{"name":"Updated Loan"}`,
			setupMock:  func() {},
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name: "usecase error",
			id:   "1",
			body: `{"name":"Updated Loan"}`,
			setupMock: func() {
				mockInstUC.EXPECT().
					Update(gomock.Any(), uint(1), gomock.Any()).
					Return(nil, errors.New("internal error"))
			},
			wantStatus: fiber.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := authRequest(t, "PUT", "/v1/installments/"+tt.id, bytes.NewBufferString(tt.body), token)
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

	mockInstUC := NewMockInstallment(ctrl)
	jwtService := jwt.New(testSecret, 15*time.Minute, 24*time.Hour)
	l := logger.NewDevelopment()

	handler := installment.New(mockInstUC, jwtService, l)

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
				mockInstUC.EXPECT().
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
				mockInstUC.EXPECT().
					Delete(gomock.Any(), uint(99)).
					Return(errors.New("record not found"))
			},
			wantStatus: fiber.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := authRequest(t, "DELETE", "/v1/installments/"+tt.id, nil, token)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestHandler_LinkItems(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockInstUC := NewMockInstallment(ctrl)
	jwtService := jwt.New(testSecret, 15*time.Minute, 24*time.Hour)
	l := logger.NewDevelopment()

	handler := installment.New(mockInstUC, jwtService, l)

	app := fiber.New()
	handler.RegisterRoutes(app.Group("/v1"))

	token := generateTestToken(t, jwtService)

	tests := []struct {
		name       string
		id         string
		body       string
		setupMock  func()
		wantStatus int
	}{
		{
			name: "success",
			id:   "1",
			body: `{"line_item_ids":[10,20,30]}`,
			setupMock: func() {
				mockInstUC.EXPECT().
					LinkItems(gomock.Any(), uint(1), gomock.Any()).
					Return(nil)
			},
			wantStatus: fiber.StatusOK,
		},
		{
			name:       "invalid id",
			id:         "abc",
			body:       `{"line_item_ids":[10]}`,
			setupMock:  func() {},
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name:       "invalid json",
			id:         "1",
			body:       `{invalid}`,
			setupMock:  func() {},
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name: "usecase error",
			id:   "1",
			body: `{"line_item_ids":[10,20]}`,
			setupMock: func() {
				mockInstUC.EXPECT().
					LinkItems(gomock.Any(), uint(1), gomock.Any()).
					Return(errors.New("internal error"))
			},
			wantStatus: fiber.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := authRequest(t, "POST", "/v1/installments/"+tt.id+"/link", bytes.NewBufferString(tt.body), token)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}
