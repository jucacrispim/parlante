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
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/emersion/go-maildir"
)

func TestEmaiMessage2Maildir(t *testing.T) {

	gen := func() (string, error) {
		return "xxx", nil
	}

	from := "me@bla.net"
	to := []string{"other@ble.com"}
	subject := "Hello"
	body := "How are you doing?"
	msg, _ := NewEmailMessage(from, to, subject, body)
	msg.Timestamp = time.Date(2025, 8, 14, 16, 43, 0, 0, time.UTC).Unix()
	key, _ := gen()
	expected := []string{
		"From: me@bla.net",
		"To: other@ble.com",
		"Subject: Hello",
		"Date: Thu, 14 Aug 2025 16:43:00 +0000",
		fmt.Sprintf("Message-ID: <%d.%s@localhost>", msg.Timestamp, key),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=\"UTF-8\"",
		msg.Body,
	}
	mformat, err := EmailMessage2Maildir(msg, gen)
	if err != nil {
		t.Fatalf("error EmailMessage2Maildir %s", err.Error())
	}

	for _, str := range expected {
		if !strings.Contains(mformat, str) {
			t.Fatalf("missing from maildir msg %s\n%s", str, mformat)
		}
	}
}

func TestSendEmail(t *testing.T) {

	gen := func() (string, error) {
		return "xxx", nil
	}

	mdirPath := "/var/tmp/parlante-test-maildir"
	defer os.Remove(mdirPath)
	s := NewMaildirSender(mdirPath)
	s.keyGen = gen
	msg, _ := NewEmailMessage("a@a.com", []string{"b@b.com"}, "bla", "bleble")
	err := s.SendEmail(msg)
	if err != nil {
		t.Fatalf("error SendEmail maildir %s", err.Error())
	}

	d := maildir.Dir(mdirPath)
	msgs, err := d.Unseen()
	if err != nil {
		t.Fatalf("unseen error %s", err.Error())
	}
	if len(msgs) == 0 {
		t.Fatalf("no unseen messages")
	}
	path := msgs[0].Filename()

	_, err = os.Stat(path)
	if err != nil {
		t.Fatalf("could  not get msg file %s", err.Error())
	}

	f, _ := os.Open(path)
	defer f.Close()
	c, _ := io.ReadAll(f)
	msgStr, _ := EmailMessage2Maildir(msg, gen)
	if string(c) != msgStr {
		t.Fatalf("bad email content \n%s\n\n%s", string(c), msgStr)
	}

}
