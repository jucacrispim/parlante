package parlante

import (
	"strings"
	"testing"
)

func TestTemplating(t *testing.T) {
	tmpl := "comments.html"
	data := make(map[string]any)
	data["header"] = "bla"
	data["noComments"] = "no comments"
	data["comments"] = make([]Comment, 0)
	b, err := RenderTemplate(tmpl, "pt_br", "UTC", data)
	if err != nil {
		t.Fatalf("Error rendering template: %s", err.Error())
	}

	s := string(b)
	if !strings.Contains(s, "bla") {
		t.Fatalf("bad render: %s", s)
	}

}
