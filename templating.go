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
func LoadTemplates(funcMap template.FuncMap) (*template.Template, error) {
	tmpl := template.New("").Funcs(funcMap)
	return tmpl.ParseFS(embeddedTemplates, "templates/*.html")
}

func RenderTemplate(
	path string,
	lang string,
	timezone string,
	data map[string]any) ([]byte, error) {

	funcMap := template.FuncMap{
		"fmtTimestap": func(ts int64) string {
			fmt := GetDateTimeFmt(lang)
			dtstr, _ := LocalizeTimestamp(ts, timezone, fmt)
			return dtstr
		},
	}

	tmpl, err := LoadTemplates(funcMap)
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
