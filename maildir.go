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
	"strings"
	"sync"
	"time"

	"github.com/emersion/go-maildir"
)

var mu = sync.Mutex{}

const DEFAULT_MAILDIR_PATH = "/var/local/maildir/parlante"

type keyGen func() (string, error)

// MaildirSender represents a maildir delivery
type MaildirSender struct {
	MaildirPath string
	keyGen      keyGen
}

// SendEmail writes an EmailMessage to a local maildir
func (s MaildirSender) SendEmail(msg EmailMessage) error {
	var d = maildir.Dir(s.MaildirPath)
	err := initMaildir(d)
	if err != nil {
		return err
	}

	mformat, err := EmailMessage2Maildir(msg, s.keyGen)
	if err != nil {
		return err
	}

	del, err := maildir.NewDelivery(s.MaildirPath)
	if err != nil {
		return err
	}

	_, err = del.Write([]byte(mformat))
	if err != nil {
		return err
	}

	err = del.Close()
	if err != nil {
		return err
	}
	return nil
}

// NewMaildirSender returns a new NewMaildirSender instance
func NewMaildirSender(path string) MaildirSender {
	s := MaildirSender{
		MaildirPath: path,
		keyGen:      GenKey,
	}
	return s
}

// EmailMessage2Maildir converts an EmailMessage to a string in the
// maildir file format.
func EmailMessage2Maildir(msg EmailMessage, gen keyGen) (string, error) {

	dtfmt := "Mon, 2 Jan 2006 15:04:05 -0700"
	loc, err := time.LoadLocation("UTC")
	if err != nil {
		return "", err
	}
	dt := time.Unix(msg.Timestamp, 0).In(loc)
	dtStr := dt.Format(dtfmt)
	key, err := gen()
	if err != nil {
		return "", err
	}
	msgId := fmt.Sprintf("<%d.%s@localhost>", msg.Timestamp, key)

	mformat := fmt.Sprintf("From: %s\n", msg.From)
	toStr := strings.Join(msg.To, ",")
	mformat += fmt.Sprintf("To: %s\n", toStr)
	mformat += fmt.Sprintf("Subject: %s\n", msg.Subject)
	mformat += fmt.Sprintf("Date: %s\n", dtStr)
	mformat += fmt.Sprintf("Message-ID: %s\n", msgId)
	mformat += fmt.Sprintf("MIME-Version: 1.0\n")
	mformat += fmt.Sprintf("Content-Type: text/plain; charset=\"UTF-8\"\n")
	mformat += "\n"
	mformat += msg.Body
	return mformat, nil
}

func initMaildir(d maildir.Dir) error {
	mu.Lock()
	defer mu.Unlock()
	err := d.Init()
	return err
}
