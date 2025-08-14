// Copyright 2025 Juca Crispim <juca@poraodojuca.dev>

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
