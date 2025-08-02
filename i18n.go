package parlante

import (
	"bytes"
	"embed"
	"os"
	"strings"
	"text/template"

	"github.com/leonelquinteros/gotext"
)

const localesDir = "locales"
const defaultDomain = "default"

//go:embed locales
var embeddedLocales embed.FS

// GetDefaultLocale returns the default system locatin defined
// in the LANG environment variable
func GetDefaultLocale() *gotext.Locale {
	lang := os.Getenv("LANG")
	lang = strings.Split(lang, ".")[0]
	return GetLocale(lang)
}

// GetLocale returns a locale for strings transation
func GetLocale(lang string) *gotext.Locale {
	l := gotext.NewLocaleFSWithPath(lang, embeddedLocales, localesDir)
	l.AddDomain(defaultDomain)
	return l
}

// tprintf uses the template mechanism to make named string formating.
func Tprintf(tmpl string, data map[string]any) string {
	t := template.Must(template.New("translation").Parse(tmpl))
	buf := &bytes.Buffer{}
	if err := t.Execute(buf, data); err != nil {
		// notest
		return tmpl
	}
	return buf.String()
}
