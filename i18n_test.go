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
	"testing"
	"time"
)

func TestGetDateTimeFmt(t *testing.T) {
	var tests = []struct {
		testName string
		lang     string
		expected string
	}{
		{
			"test with missing lang",
			"es_UY",
			"2006-01-02 15:04",
		},
		{
			"test with ok lang",
			"pt_BR",
			"02/01/2006 15:04",
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			r := GetDateTimeFmt(test.lang)
			if r != test.expected {
				t.Fatalf("Bad datetime fmt %s", r)
			}
		})
	}
}

func TestLocalizeTimestamp(t *testing.T) {
	ts := time.Date(2023, 10, 20, 9, 0, 0, 0, time.UTC).Unix()
	tz := "America/Sao_Paulo"
	fmt := GetDateTimeFmt("pt_BR")
	locdt, _ := LocalizeTimestamp(ts, tz, fmt)
	expected := "20/10/2023 06:00"

	if locdt != expected {
		t.Fatalf("bad localized ts %s", locdt)
	}
}
