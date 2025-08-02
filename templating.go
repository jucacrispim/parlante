package parlante

import (
	"bytes"
	"embed"
	"html/template"
)

//go:embed templates
var embeddedTemplates embed.FS

// LoadTemplates returns a template based on its path.
// Note that the templates are embedded.
func LoadTemplates() (*template.Template, error) {
	return template.ParseFS(embeddedTemplates, "templates/*.html")
}

func RenderTemplate(path string, data map[string]any) ([]byte, error) {
	tmpl, err := LoadTemplates()
	if err != nil {
		return nil, err
	}
	var buff bytes.Buffer
	err = tmpl.ExecuteTemplate(&buff, path, data)
	if err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}
