// Copyright 2025 Juca Crispim <juca@poraodojuca.net>

// This file is part of parlante.

// parlante is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// parlante is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with parlante. If not, see <http://www.gnu.org/licenses/>.

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
