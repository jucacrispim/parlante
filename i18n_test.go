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
