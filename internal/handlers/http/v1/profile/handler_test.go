package profile_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	profiledto "go-boilerplate/internal/dto/profile"
	"go-boilerplate/internal/handlers/http/v1/profile"
	profileuc "go-boilerplate/internal/usecase/profile"
	"go-boilerplate/pkg/jwt"
)

func setupTestApp(t *testing.T, mockProfileUC *MockProfile) (app *fiber.App, token string) {
	t.Helper()

	jwtService := jwt.New("test-secret", 15*time.Minute, 24*time.Hour)
	l := NewMockLogger()

	handler := profile.New(mockProfileUC, jwtService, l)

	app = fiber.New()

	handler.RegisterRoutes(app.Group("/v1"))

	// Generate a valid test token
	var err error
	token, _, err = jwtService.GenerateAccessToken(1, "test@example.com", "user", nil)
	require.NoError(t, err)

	return app, token
}

func TestHandler_GetProfile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupMock  func(mockUC *MockProfile)
		wantStatus int
	}{
		{
			name: "success - with avatar",
			setupMock: func(mockUC *MockProfile) {
				mockUC.EXPECT().
					GetProfile(gomock.Any(), uint(1)).
					Return(&profiledto.ProfileResponse{
						UserID: 1,
						Bio:    "Test bio",
						Phone:  "+1234567890",
						Avatar: &profiledto.AvatarResponse{
							ID:  100,
							URL: "https://s3.example.com/avatar.jpg",
						},
					}, nil)
			},
			wantStatus: fiber.StatusOK,
		},
		{
			name: "success - without avatar",
			setupMock: func(mockUC *MockProfile) {
				mockUC.EXPECT().
					GetProfile(gomock.Any(), uint(1)).
					Return(&profiledto.ProfileResponse{
						UserID: 1,
						Bio:    "",
						Phone:  "",
						Avatar: nil,
					}, nil)
			},
			wantStatus: fiber.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockProfileUC := NewMockProfile(ctrl)
			app, token := setupTestApp(t, mockProfileUC)

			tt.setupMock(mockProfileUC)

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/v1/profile", http.NoBody)
			req.Header.Set("Authorization", "Bearer "+token)

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestHandler_UpdateProfile(t *testing.T) {
	t.Parallel()

	bio := "Updated bio"
	phone := "+1234567890"
	avatarMediaID := uint(100)

	tests := []struct {
		name       string
		body       string
		setupMock  func(mockUC *MockProfile)
		wantStatus int
	}{
		{
			name: "success - update bio only",
			body: `{"bio": "Updated bio"}`,
			setupMock: func(mockUC *MockProfile) {
				mockUC.EXPECT().
					UpdateProfile(gomock.Any(), uint(1), profiledto.UpdateProfileRequest{
						Bio: &bio,
					}).
					Return(&profiledto.ProfileResponse{
						UserID: 1,
						Bio:    bio,
						Phone:  "",
						Avatar: nil,
					}, nil)
			},
			wantStatus: fiber.StatusOK,
		},
		{
			name: "success - update all fields",
			body: `{"bio": "Updated bio", "phone": "+1234567890", "avatar_media_id": 100}`,
			setupMock: func(mockUC *MockProfile) {
				mockUC.EXPECT().
					UpdateProfile(gomock.Any(), uint(1), profiledto.UpdateProfileRequest{
						Bio:           &bio,
						Phone:         &phone,
						AvatarMediaID: &avatarMediaID,
					}).
					Return(&profiledto.ProfileResponse{
						UserID: 1,
						Bio:    bio,
						Phone:  phone,
						Avatar: &profiledto.AvatarResponse{
							ID:  avatarMediaID,
							URL: "https://s3.example.com/avatar.jpg",
						},
					}, nil)
			},
			wantStatus: fiber.StatusOK,
		},
		{
			name:       "error - invalid json",
			body:       `{invalid}`,
			setupMock:  func(mockUC *MockProfile) {},
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name: "error - invalid media",
			body: `{"avatar_media_id": 999}`,
			setupMock: func(mockUC *MockProfile) {
				invalidMediaID := uint(999)
				mockUC.EXPECT().
					UpdateProfile(gomock.Any(), uint(1), profiledto.UpdateProfileRequest{
						AvatarMediaID: &invalidMediaID,
					}).
					Return(nil, profileuc.ErrInvalidMedia)
			},
			wantStatus: fiber.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockProfileUC := NewMockProfile(ctrl)
			app, token := setupTestApp(t, mockProfileUC)

			tt.setupMock(mockProfileUC)

			req := httptest.NewRequestWithContext(t.Context(), http.MethodPatch, "/v1/profile", bytes.NewBufferString(tt.body))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestHandler_Unauthorized(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProfileUC := NewMockProfile(ctrl)
	app, _ := setupTestApp(t, mockProfileUC)

	// Test without token
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/v1/profile", http.NoBody)

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}
