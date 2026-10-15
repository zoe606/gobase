package email

import (
	"bytes"
	"fmt"
	"html/template"
)

var authEmailTemplate = template.Must(template.New("auth-email").Parse(`<p>{{.Title}} for {{.AppName}}.</p><p><a href="{{.URL}}">{{.Title}}</a></p><p>If you did not request this email, you can ignore it.</p>`))

func RenderAuthEmail(title, appName, link string) (htmlBody, textBody string, err error) {
	data := struct{ Title, AppName, URL string }{title, appName, link}
	var html bytes.Buffer
	if err := authEmailTemplate.Execute(&html, data); err != nil {
		return "", "", err
	}
	text := fmt.Sprintf("%s for %s.\n\n%s\n\nIf you did not request this email, you can ignore it.", title, appName, link)
	return html.String(), text, nil
}
