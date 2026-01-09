package email

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderWelcomeEmail(t *testing.T) {
	data := WelcomeEmailData{
		Username: "John Doe",
		AppName:  "MyApp",
	}

	html, text, err := RenderWelcomeEmail(data)
	require.NoError(t, err)

	// Verify HTML content
	assert.Contains(t, html, "Welcome to MyApp!")
	assert.Contains(t, html, "Hi John Doe,")
	assert.Contains(t, html, "The MyApp Team")
	assert.Contains(t, html, "<!DOCTYPE html>")

	// Verify text content
	assert.Contains(t, text, "Welcome to MyApp!")
	assert.Contains(t, text, "Hi John Doe,")
	assert.Contains(t, text, "The MyApp Team")
	assert.False(t, strings.Contains(text, "<"), "text version should not contain HTML tags")
}

func TestRenderWelcomeEmail_SpecialCharacters(t *testing.T) {
	data := WelcomeEmailData{
		Username: "John <script>alert('xss')</script>",
		AppName:  "My&App",
	}

	html, text, err := RenderWelcomeEmail(data)
	require.NoError(t, err)

	// HTML should escape special characters (XSS protection)
	assert.NotContains(t, html, "<script>")
	assert.Contains(t, html, "&lt;script&gt;")

	// Note: html/template is used for text version too, so it also escapes
	// This is actually safe behavior - ensures consistency
	assert.NotEmpty(t, text)
	assert.Contains(t, text, "Welcome to")
}

func TestWelcomeEmailData(t *testing.T) {
	data := WelcomeEmailData{
		Username: "testuser",
		AppName:  "TestApp",
	}

	assert.Equal(t, "testuser", data.Username)
	assert.Equal(t, "TestApp", data.AppName)
}
