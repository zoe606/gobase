package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	articledto "go-boilerplate/internal/dto/article"
	authdto "go-boilerplate/internal/dto/auth"
	mediadto "go-boilerplate/internal/dto/media"
	profiledto "go-boilerplate/internal/dto/profile"
	"go-boilerplate/pkg/response"
)

func engineRequest(t *testing.T, method, path, token, contentType string, body io.Reader, status int) []byte {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), method, basePathV1+path, body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", contentType)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := (&http.Client{Timeout: requestTimeout}).Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, status, resp.StatusCode, string(data))
	return data
}

func TestHTTPCRUDV1(t *testing.T) {
	unique := strconv.FormatInt(time.Now().UnixNano(), 10)
	registerBody := fmt.Sprintf("{\"email\":\"engine-%s@example.com\",\"password\":\"testpassword123\",\"name\":\"Engine Test\"}", unique)
	data := engineRequest(t, "POST", "/auth/register", "", "application/json", bytes.NewBufferString(registerBody), 201)
	var auth response.Response[authdto.LoginResponse]
	require.NoError(t, json.Unmarshal(data, &auth))
	require.NotEmpty(t, auth.Data.AccessToken)
	require.NotZero(t, auth.Data.User.ID)
	token := auth.Data.AccessToken
	profileBody := "{\"bio\":\"Engine profile\",\"phone\":\"123456\"}"
	data = engineRequest(t, "PATCH", "/profile", token, "application/json", bytes.NewBufferString(profileBody), 200)
	var profile response.Response[profiledto.ProfileResponse]
	require.NoError(t, json.Unmarshal(data, &profile))
	require.Equal(t, "Engine profile", profile.Data.Bio)
	require.Equal(t, auth.Data.User.ID, profile.Data.UserID)
	data = engineRequest(t, "GET", "/profile", token, "application/json", http.NoBody, 200)
	require.NoError(t, json.Unmarshal(data, &profile))
	require.Equal(t, "Engine profile", profile.Data.Bio)

	var upload bytes.Buffer
	form := multipart.NewWriter(&upload)
	part, err := form.CreatePart(textproto.MIMEHeader{"Content-Disposition": {"form-data; name=\"file\"; filename=\"cover.png\""}, "Content-Type": {"image/png"}})
	require.NoError(t, err)
	require.NoError(t, png.Encode(part, image.NewRGBA(image.Rect(0, 0, 1, 1))))
	require.NoError(t, form.WriteField("attachable_type", "users"))
	require.NoError(t, form.WriteField("attachable_id", strconv.FormatUint(uint64(auth.Data.User.ID), 10)))
	require.NoError(t, form.WriteField("collection", "covers"))
	require.NoError(t, form.Close())
	data = engineRequest(t, "POST", "/media/upload", token, form.FormDataContentType(), &upload, 201)
	var media response.Response[mediadto.MediaResponse]
	require.NoError(t, json.Unmarshal(data, &media))
	require.NotZero(t, media.Data.ID)
	mediaPath := fmt.Sprintf("/media/%d", media.Data.ID)
	if os.Getenv("TEST_WORKER_ENABLED") == "true" {
		deadline := time.Now().Add(time.Minute)
		for len(media.Data.Variants) != 3 && time.Now().Before(deadline) {
			time.Sleep(time.Second)
			data = engineRequest(t, "GET", mediaPath, token, "application/json", http.NoBody, 200)
			require.NoError(t, json.Unmarshal(data, &media))
		}
		require.Len(t, media.Data.Variants, 3, "worker must write all three image variants")
		require.NotNil(t, media.Data.Width)
		require.NotNil(t, media.Data.Height)
	}
	engineRequest(t, "GET", mediaPath, token, "application/json", http.NoBody, 200)
	engineRequest(t, "GET", fmt.Sprintf("/media?attachable_type=users&attachable_id=%d", auth.Data.User.ID), token, "application/json", http.NoBody, 200)

	createBody := fmt.Sprintf("{\"title\":\"Engine article\",\"slug\":\"engine-%s\",\"content\":\"Body\",\"excerpt\":\"Summary\",\"cover_media_id\":%d}", unique, media.Data.ID)
	data = engineRequest(t, "POST", "/articles", token, "application/json", bytes.NewBufferString(createBody), 201)
	var article response.Response[articledto.Response]
	require.NoError(t, json.Unmarshal(data, &article))
	require.NotZero(t, article.Data.ID)
	require.Equal(t, auth.Data.User.ID, article.Data.UserID)
	articlePath := fmt.Sprintf("/articles/%d", article.Data.ID)
	data = engineRequest(t, "GET", articlePath, "", "application/json", http.NoBody, 200)
	require.NoError(t, json.Unmarshal(data, &article))
	require.Equal(t, "Engine article", article.Data.Title)
	data = engineRequest(t, "GET", "/articles?page=1&limit=5&user_id="+strconv.FormatUint(uint64(auth.Data.User.ID), 10), "", "application/json", http.NoBody, 200)
	require.Contains(t, string(data), "Engine article")
	data = engineRequest(t, "PUT", articlePath, token, "application/json", bytes.NewBufferString("{\"title\":\"Updated article\"}"), 200)
	require.NoError(t, json.Unmarshal(data, &article))
	require.Equal(t, "Updated article", article.Data.Title)
	require.Empty(t, engineRequest(t, "DELETE", articlePath, token, "application/json", http.NoBody, 204))
	engineRequest(t, "GET", articlePath, "", "application/json", http.NoBody, 404)
	require.Contains(t, string(engineRequest(t, "DELETE", mediaPath, token, "application/json", http.NoBody, 200)), "Media deleted successfully")
	engineRequest(t, "GET", mediaPath, token, "application/json", http.NoBody, 404)
}
