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

import "testing"

func TestNewClient(t *testing.T) {
	c, plain_key, err := NewClient("the test client")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if plain_key == "" || plain_key == c.Key {
		t.Fatalf("Bad plan_key %s %s", plain_key, c.Key)
	}
}

func TestClientUpdateKey(t *testing.T) {
	c, plain_key, err := NewClient("the test client")
	if err != nil {
		t.Fatalf(err.Error())
	}

	new_key, err := c.UpdateKey()
	if err != nil {
		t.Fatalf(err.Error())
	}

	if plain_key == new_key {
		t.Fatalf("key not updated %s %s", plain_key, new_key)
	}
}

func TestNewComment(t *testing.T) {
	c, _, _ := NewClient("the test client")
	d := NewClientDomain(c, "bla.net")

	var tests = []struct {
		testName string
		author   string
		content  string
		page_url string
		hasError bool
	}{
		{"new without author", "", "blabla", "https:/bla.net", true},
		{"new without content", "zé", "", "https:/bla.net", true},
		{"new without page_url", "zé", "blabla", "", true},
		{"new ok", "zé", "blabla", "https:/bla.net", false},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			_, err := NewComment(c, d, test.author, test.content, test.page_url)
			if err == nil && test.hasError {
				t.Fatalf("No error")
			}

			if err != nil && !test.hasError {
				t.Fatalf("Error!")
			}
		})
	}
}
