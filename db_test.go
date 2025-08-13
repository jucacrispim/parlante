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
	"os"
	"testing"

	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const DBFILE = "/var/tmp/parlante-test.sqlite"
const MIGRATIONS_DIR = "./migrations/"

func TestClient(t *testing.T) {

	err := setupTestDB()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(DBFILE)
	s := ClientStorageSQLite{}
	c, _, err := s.CreateClient("A test client")

	if err != nil {
		t.Fatalf(err.Error())
	}

	if c.ID == 0 {
		t.Fatalf("Bad id for new client")
	}

	c2, err := s.GetClientByUUID(c.UUID)
	if err != nil {
		t.Fatal(err)
	}

	if c2.ID != c.ID {
		t.Fatalf("bad id for get client by uuid")
	}

	clients, _ := s.ListClients()

	if len(clients) != 1 {
		t.Fatalf("bad clients list len")
	}

	err = s.RemoveClient(c.UUID)

	if err != nil {
		t.Fatalf("Error removing client %s", err.Error())
	}

	clients, _ = s.ListClients()

	if len(clients) != 0 {
		t.Fatalf("client not removed")
	}

}

func TestClientDomain(t *testing.T) {
	err := setupTestDB()
	defer os.Remove(DBFILE)
	if err != nil {
		t.Fatal(err)
	}

	cs := ClientStorageSQLite{}
	cds := ClientDomainStorageSQLite{}
	c, _, _ := cs.CreateClient("the test client")
	d, err := cds.AddClientDomain(c, "mydomain.net")
	if err != nil {
		t.Fatal(err)
	}

	if d.ID == 0 {
		t.Fatalf("bad id for add domain")
	}

	d2, err := cds.GetClientDomain(c, d.Domain)
	if err != nil {
		t.Fatal(err)
	}

	if d2.Domain != d.Domain {
		t.Fatalf("Bad domain for get domain %s", d2.Domain)
	}

	domains, err := cds.ListDomains()
	if err != nil {
		t.Fatalf(err.Error())
	}
	if len(domains) != 1 {
		t.Fatalf("bad domains list %d", len(domains))
	}
	if domains[0].Client.UUID != c.UUID {
		t.Fatalf("bad client for domain %s", domains[0].Client.UUID)
	}
	err = cds.RemoveClientDomain(c, d2.Domain)
	if err != nil {
		t.Fatal(err)
	}

	d3, err := cds.GetClientDomain(c, d.Domain)
	if err != nil {
		t.Fatal(err)
	}

	empty := ClientDomain{}
	if d3 != empty {
		t.Fatalf("did not remove client domain")
	}

}

func TestComments(t *testing.T) {

	err := setupTestDB()
	defer os.Remove(DBFILE)
	if err != nil {
		t.Fatal(err)
	}
	cs := ClientStorageSQLite{}
	cds := ClientDomainStorageSQLite{}
	comms := CommentStorageSQLite{}
	c, _, _ := cs.CreateClient("the test client")
	d, _ := cds.AddClientDomain(c, "bla.net")

	var tests = []struct {
		name     string
		content  string
		page_url string
	}{
		{"zé", "some comment", "http://bla.net/post"},
		{"jão", "other comment", "http://bla.net/post"},
		{"ble", "other other comment", "http://bla.net/post2"},
		{"bli", "new comment", "http://bla.net/post2"},
	}

	for _, test := range tests {
		c, err := comms.CreateComment(
			c, d, test.name, test.content, test.page_url)

		if err != nil {
			t.Fatal(err)
		}

		if c.ID == 0 {
			t.Fatalf("bad id for create comment")
		}
	}

	allcomments, err := comms.ListComments(CommentsFilter{})
	if err != nil {
		t.Fatal(err)
	}

	if len(allcomments) != 4 {
		t.Fatalf("Bad len for allcomments %d", len(allcomments))
	}
	url := "http://bla.net/post"
	p1comments, err := comms.ListComments(CommentsFilter{PageURL: &url})
	if err != nil {
		t.Fatal(err)
	}

	if len(p1comments) != 2 {
		t.Fatalf("Bad len for p1comments %d", len(p1comments))
	}

	comms.RemoveComment(allcomments[0])

	allcomments, err = comms.ListComments(CommentsFilter{})
	if len(allcomments) != 3 {
		t.Fatalf("Bad len for allcomments after remove %d", len(allcomments))
	}

}

func TestCommentCount_NoURLs(t *testing.T) {
	comms := CommentStorageSQLite{}
	_, err := comms.CountComments()
	if err == nil {
		t.Fatalf("No error for no urls on comment count")
	}
}

func TestCommentCount(t *testing.T) {
	err := setupTestDB()
	defer os.Remove(DBFILE)
	if err != nil {
		t.Fatal(err)
	}

	cs := ClientStorageSQLite{}
	cds := ClientDomainStorageSQLite{}
	comms := CommentStorageSQLite{}
	c, _, _ := cs.CreateClient("the test client")
	d, _ := cds.AddClientDomain(c, "bla.net")

	urls := []string{"http://bla.net/count-1", "http://bla.net/count-2", "http://bla.net/count-3"}
	for _, url := range urls[:2] {
		_, err := comms.CreateComment(c, d, "zé", "blabla", url)
		if err != nil {
			t.Fatalf("error creating comment %s", err.Error())
		}
	}
	count, err := comms.CountComments(urls...)
	if err != nil {
		t.Fatalf("error comment count! %s", err.Error())
	}

	if len(count) != 3 {
		t.Fatalf("bad len for comment count %d", len(count))
	}
}

func setupTestDB() error {
	SetupDB(DBFILE)
	err := MigrateDB(DBFILE)

	return err
}
