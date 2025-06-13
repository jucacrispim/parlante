package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type ClientStorageInMemory struct {
	data          map[string]Client
	BadClientUUID string
}

func (s ClientStorageInMemory) CreateClient(name string) (Client, string, error) {
	c, key, _ := NewClient(name)
	s.data[c.UUID] = c
	return c, key, nil
}

func (s ClientStorageInMemory) GetClientByUUID(uuid string) (Client, error) {
	if uuid == s.BadClientUUID {
		return Client{}, errors.New("Bad!")
	}
	c, ok := s.data[uuid]
	if !ok {
		return Client{}, nil
	}
	return c, nil
}

func NewClientStorageInMemory() ClientStorageInMemory {
	c := ClientStorageInMemory{}
	c.data = make(map[string]Client)
	u, _ := GenUUID4()
	c.BadClientUUID = u
	return c
}

type ClientDomainStorageInMemory struct {
	data      map[string]ClientDomain
	BadDomain string
}

func (s ClientDomainStorageInMemory) AddClientDomain(c Client, domain string) (
	ClientDomain, error) {
	d := NewClientDomain(c, domain)
	key := c.UUID + "-" + d.Domain
	s.data[key] = d
	return d, nil
}

func (s ClientDomainStorageInMemory) RemoveClientDomain(c Client, domain string) error {
	key := c.UUID + "-" + domain
	delete(s.data, key)
	return nil
}

func (s ClientDomainStorageInMemory) GetClientDomain(c Client, domain string) (
	ClientDomain, error) {
	if domain == s.BadDomain {
		return ClientDomain{}, errors.New("bad")
	}
	key := c.UUID + "-" + domain
	d, ok := s.data[key]
	if !ok {
		return ClientDomain{}, nil
	}
	return d, nil
}

func NewClientDomainStorageInMemory() ClientDomainStorageInMemory {
	d := ClientDomainStorageInMemory{}
	d.data = make(map[string]ClientDomain)
	d.BadDomain = "bad.net"
	return d
}

type CommentStorageInMemory struct {
	clientComments map[int64][]Comment
	domainComments map[int64][]Comment
	pageComments   map[string][]Comment
	BadCommenter   string
	BadPage        string
}

func (s CommentStorageInMemory) CreateComment(c Client, d ClientDomain,
	name string, content string, page_url string) (Comment, error) {
	if name == s.BadCommenter {
		return Comment{}, errors.New("bad")
	}

	comment := NewComment(c, d, name, content, page_url)
	s.clientComments[c.ID] = append(s.clientComments[c.ID], comment)
	s.domainComments[d.ID] = append(s.domainComments[d.ID], comment)
	s.pageComments[comment.PageURL] = append(
		s.pageComments[comment.PageURL], comment)

	return comment, nil
}

func (s CommentStorageInMemory) ListComments(filter CommentsFilter) (
	[]Comment, error) {

	if *filter.PageURL == s.BadPage {
		return []Comment{}, errors.New("bad")
	}

	if filter.ClientID != nil {
		return s.clientComments[*filter.ClientID], nil
	}
	if filter.DomainID != nil {
		return s.domainComments[*filter.DomainID], nil
	}

	if filter.PageURL != nil {
		return s.pageComments[*filter.PageURL], nil
	}
	return nil, nil
}

func NewCommentStorageInMemory() CommentStorageInMemory {
	c := CommentStorageInMemory{}
	c.clientComments = make(map[int64][]Comment)
	c.domainComments = make(map[int64][]Comment)
	c.pageComments = make(map[string][]Comment)
	c.BadCommenter = "bad"
	c.BadPage = "http://bla.net/bad"
	return c
}

func TestCreateComment(t *testing.T) {

	s := NewServer()
	s.ClientStorage = NewClientStorageInMemory()
	s.ClientDomainStorage = NewClientDomainStorageInMemory()
	s.CommentStorage = NewCommentStorageInMemory()
	s.Server = http.NewServeMux()
	s.BodyReader = io.ReadAll
	s.setupUrls()

	c, _, _ := s.ClientStorage.CreateClient("test client")
	s.ClientDomainStorage.AddClientDomain(c, "bla.net")

	var test_data = []struct {
		testName string
		req      *http.Request
		status   int
	}{
		{
			"comment with bad client",
			func() *http.Request {
				uuid, _ := GenUUID4()
				payload := CreateCommentRequest{
					Name:    "Zé",
					Content: "A comment",
				}
				j, _ := json.Marshal(payload)
				body := bytes.NewBuffer(j)
				req, _ := http.NewRequest("POST", "/comment/"+uuid, body)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("Referer", "https://bla.net/post")
				return req

			}(),
			403},
		{
			"comment with wrong origin",
			func() *http.Request {
				payload := CreateCommentRequest{
					Name:    "Zé",
					Content: "A comment",
				}
				j, _ := json.Marshal(payload)
				body := bytes.NewBuffer(j)
				req, _ := http.NewRequest("POST", "/comment/"+c.UUID, body)
				req.Header.Set("Origin", "https://bleble.net")
				req.Header.Set("Referer", "https://bleble.net/post")
				return req

			}(),
			403},
		{
			"comment missing body",
			func() *http.Request {
				req, _ := http.NewRequest("POST", "/comment/"+c.UUID, nil)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("Referer", "https://bla.net/post")
				return req

			}(),
			400},
		{
			"comment bad body",
			func() *http.Request {
				body := bytes.NewBuffer([]byte(""))
				req, _ := http.NewRequest("POST", "/comment/"+c.UUID, body)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("Referer", "https://bla.net/post")
				return req

			}(),
			400},
		{
			"comment ok",
			func() *http.Request {
				payload := CreateCommentRequest{
					Name:    "Zé",
					Content: "A comment",
				}
				j, _ := json.Marshal(payload)
				body := bytes.NewBuffer(j)
				req, _ := http.NewRequest("POST", "/comment/"+c.UUID, body)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("Referer", "https://bla.net/post")
				return req

			}(),
			201},
	}

	for _, test := range test_data {
		t.Run(test.testName, func(t *testing.T) {
			w := httptest.NewRecorder()
			s.Server.ServeHTTP(w, test.req)

			if w.Code != test.status {
				t.Fatalf("bad status for %d", w.Code)
			}

		})
	}

}

func TestListComments(t *testing.T) {

	s := NewServer()
	s.ClientStorage = NewClientStorageInMemory()
	s.ClientDomainStorage = NewClientDomainStorageInMemory()
	s.CommentStorage = NewCommentStorageInMemory()
	s.Server = http.NewServeMux()
	s.BodyReader = io.ReadAll
	s.setupUrls()

	c, _, _ := s.ClientStorage.CreateClient("test client")
	d, _ := s.ClientDomainStorage.AddClientDomain(c, "bla.net")

	comments := []Comment{
		{
			Name:    "Zé",
			Content: "The comment",
			PageURL: "http://bla.net/post1",
		},
		{
			Name:    "Tião",
			Content: "The other comment",
			PageURL: "http://bla.net/post1",
		},
		{
			Name:    "Jão",
			Content: "The new comment",
			PageURL: "http://bla.net/post1",
		},
		{
			Name:    "Zé",
			Content: "The new new comment",
			PageURL: "http://bla.net/post2",
		},
		{
			Name:    "Jão",
			Content: "Another comment",
			PageURL: "http://bla.net/post2",
		},
	}

	for _, co := range comments {
		s.CommentStorage.CreateComment(c, d, co.Name, co.Content, co.PageURL)
	}

	var test_data = []struct {
		testName string
		req      *http.Request
		status   int
	}{
		{
			"list comments with bad client",
			func() *http.Request {
				uuid, _ := GenUUID4()
				req, _ := http.NewRequest("GET", "/comment/"+uuid, nil)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("Referer", "https://bla.net/pos1")
				return req
			}(),
			403,
		},
		{
			"list comments with wrong origin",
			func() *http.Request {
				req, _ := http.NewRequest("GET", "/comment/"+c.UUID, nil)
				req.Header.Set("Origin", "https://bleble.net")
				req.Header.Set("Referer", "https://bla.net/pos1")
				return req
			}(),
			403,
		},
		{
			"list comments without origin",
			func() *http.Request {
				req, _ := http.NewRequest("GET", "/comment/"+c.UUID, nil)
				req.Header.Set("Referer", "https://bla.net/pos1")
				return req
			}(),
			403,
		},
		{
			"list comments ok",
			func() *http.Request {
				req, _ := http.NewRequest("GET", "/comment/"+c.UUID, nil)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("Referer", "https://bla.net/pos1")
				return req
			}(),
			200,
		},
	}

	for _, test := range test_data {
		t.Run(test.testName, func(t *testing.T) {
			w := httptest.NewRecorder()
			s.Server.ServeHTTP(w, test.req)

			if w.Code != test.status {
				t.Fatalf("bad status for %d", w.Code)
			}

		})
	}
}

func TestComments_WithDBErrors(t *testing.T) {

	s := NewServer()
	client_storage := NewClientStorageInMemory()
	s.ClientStorage = client_storage
	s.ClientDomainStorage = NewClientDomainStorageInMemory()
	comment_storage := NewCommentStorageInMemory()
	s.CommentStorage = comment_storage
	s.Server = http.NewServeMux()
	s.BodyReader = io.ReadAll
	s.setupUrls()

	c, _, _ := s.ClientStorage.CreateClient("test client")
	s.ClientDomainStorage.AddClientDomain(c, "bla.net")

	var test_data = []struct {
		testName string
		req      *http.Request
		status   int
	}{
		{
			"comment with db error on get client",
			func() *http.Request {
				payload := CreateCommentRequest{
					Name:    "Zé",
					Content: "A comment",
				}
				j, _ := json.Marshal(payload)
				body := bytes.NewBuffer(j)
				uuid := client_storage.BadClientUUID
				req, _ := http.NewRequest("POST", "/comment/"+uuid, body)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("Referer", "https://bla.net/post")
				return req

			}(),
			500,
		},
		{
			"comment with db error on get domain",
			func() *http.Request {
				payload := CreateCommentRequest{
					Name:    "Zé",
					Content: "A comment",
				}
				j, _ := json.Marshal(payload)
				body := bytes.NewBuffer(j)
				req, _ := http.NewRequest("POST", "/comment/"+c.UUID, body)
				req.Header.Set("Origin", "https://bad.net")
				req.Header.Set("Referer", "https://bla.net/post")
				return req

			}(),
			500,
		},
		{
			"comment with db error creating comment",
			func() *http.Request {
				payload := CreateCommentRequest{
					Name:    comment_storage.BadCommenter,
					Content: "A comment",
				}
				j, _ := json.Marshal(payload)
				body := bytes.NewBuffer(j)
				req, _ := http.NewRequest("POST", "/comment/"+c.UUID, body)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("Referer", "https://bla.net/post")
				return req

			}(),
			400,
		},
		{
			"list comments with db error getting comments",
			func() *http.Request {
				req, _ := http.NewRequest("GET", "/comment/"+c.UUID, nil)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("Referer", comment_storage.BadPage)
				return req

			}(),
			500,
		},
	}

	for _, test := range test_data {
		t.Run(test.testName, func(t *testing.T) {
			w := httptest.NewRecorder()
			s.Server.ServeHTTP(w, test.req)

			if w.Code != test.status {
				t.Fatalf("bad status for %d", w.Code)
			}

		})
	}

}
