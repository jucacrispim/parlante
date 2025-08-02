package parlante

import (
	"bytes"
	"embed"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/leonelquinteros/gotext"
)

var DEFAULT_DTFMT = "2006-01-02 15:04"
var DTFMTS = map[string]string{
	"pt_BR": "02/01/2006 15:04",
}

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

// GetDateTimeFmt returns a string for datetime formatting for a given
// language.
func GetDateTimeFmt(lang string) string {
	fmt, exists := DTFMTS[lang]
	if !exists {
		return DEFAULT_DTFMT
	}
	return fmt
}

// LocalizeTimestamp returns a formatted string for a given timestamp
// taking into account a timezone and a format string.
// TODO: months/days names are not translated.
func LocalizeTimestamp(
	timestamp int64, timezone string, fmt string) (string, error) {

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return "", err
	}
	t := time.Unix(timestamp, 0).In(loc)
	return t.Format(fmt), nil
}

// Tprintf uses the template mechanism for a named string formating.
func Tprintf(tmpl string, data map[string]any) string {
	t := template.Must(template.New("translation").Parse(tmpl))
	buf := &bytes.Buffer{}
	if err := t.Execute(buf, data); err != nil {
		// notest
		return tmpl
	}
	return buf.String()
}
