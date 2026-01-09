// Package email provides email templates and rendering utilities.
package email

import (
	"bytes"
	"html/template"
)

// WelcomeEmailData holds data for the welcome email template.
type WelcomeEmailData struct {
	Username string
	AppName  string
}

//nolint:lll // HTML template
const welcomeEmailHTML = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; margin: 0; padding: 0; background-color: #f5f5f5;">
    <div style="max-width: 600px; margin: 0 auto; padding: 40px 20px;">
        <div style="background-color: #ffffff; border-radius: 8px; padding: 40px; box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);">
            <h1 style="color: #1a1a1a; margin: 0 0 24px 0; font-size: 24px; font-weight: 600;">Welcome to {{.AppName}}!</h1>
            <p style="color: #4a4a4a; font-size: 16px; line-height: 24px; margin: 0 0 16px 0;">Hi {{.Username}},</p>
            <p style="color: #4a4a4a; font-size: 16px; line-height: 24px; margin: 0 0 24px 0;">Thank you for registering. Your account is now active and ready to use.</p>
            <p style="color: #4a4a4a; font-size: 16px; line-height: 24px; margin: 0;">Best regards,<br>The {{.AppName}} Team</p>
        </div>
        <p style="color: #999; font-size: 12px; text-align: center; margin-top: 24px;">This email was sent by {{.AppName}}. If you didn't create an account, you can safely ignore this email.</p>
    </div>
</body>
</html>`

const welcomeEmailText = `Welcome to {{.AppName}}!

Hi {{.Username}},

Thank you for registering. Your account is now active and ready to use.

Best regards,
The {{.AppName}} Team`

// RenderWelcomeEmail renders the welcome email template with the provided data.
func RenderWelcomeEmail(data WelcomeEmailData) (html string, text string, err error) {
	// Render HTML
	htmlTmpl, err := template.New("welcome_html").Parse(welcomeEmailHTML)
	if err != nil {
		return "", "", err
	}

	var htmlBuf bytes.Buffer
	if err := htmlTmpl.Execute(&htmlBuf, data); err != nil {
		return "", "", err
	}

	// Render plain text
	textTmpl, err := template.New("welcome_text").Parse(welcomeEmailText)
	if err != nil {
		return "", "", err
	}

	var textBuf bytes.Buffer
	if err := textTmpl.Execute(&textBuf, data); err != nil {
		return "", "", err
	}

	return htmlBuf.String(), textBuf.String(), nil
}
