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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func readFn(reader io.Reader) ([]byte, error) {
	r, err := io.ReadAll(reader)
	if string(r) == "bad body" {
		return nil, errors.New("bad")
	}
	return r, err
}

func errorMarshal(v any) ([]byte, error) {
	return nil, errors.New("bad")
}

func errorHtmlRender(
	s string,
	lang string,
	tz string,
	d map[string]any) ([]byte, error) {
	return nil, errors.New("bad")
}

func TestCreateComment(t *testing.T) {

	co := Config{}
	s := NewServer(co)
	s.ClientStorage = NewClientStorageInMemory()
	s.ClientDomainStorage = NewClientDomainStorageInMemory()
	s.CommentStorage = NewCommentStorageInMemory()
	s.mux = http.NewServeMux()
	s.BodyReader = io.ReadAll
	s.EmailSender = TestMailSender{}
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
				req, _ := http.NewRequest("POST", "/comment/", body)
				req.Header.Set("X-ClientUUID", uuid)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-PageURL", "https://bla.net/post")
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
				req, _ := http.NewRequest("POST", "/comment/", body)
				req.Header.Set("X-ClientUUID", c.UUID)
				req.Header.Set("Origin", "https://bleble.net")
				req.Header.Set("X-PageURL", "https://bleble.net/post")
				return req

			}(),
			403},
		{
			"comment missing body",
			func() *http.Request {
				req, _ := http.NewRequest("POST", "/comment/", nil)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-PageURL", "https://bla.net/post")
				req.Header.Set("X-ClientUUID", c.UUID)
				return req

			}(),
			400},
		{
			"comment bad body",
			func() *http.Request {
				body := bytes.NewBuffer([]byte(""))
				req, _ := http.NewRequest("POST", "/comment/", body)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-PageURL", "https://bla.net/post")
				req.Header.Set("X-ClientUUID", c.UUID)
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
				req, _ := http.NewRequest("POST", "/comment/", body)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-PageURL", "https://bla.net/post")
				req.Header.Set("X-ClientUUID", c.UUID)
				return req

			}(),
			201},
	}

	for _, test := range test_data {
		t.Run(test.testName, func(t *testing.T) {
			w := httptest.NewRecorder()
			s.mux.ServeHTTP(w, test.req)

			if w.Code != test.status {
				t.Fatalf("bad status for %d", w.Code)
			}

		})
	}

}

func TestCreateComment_Auth(t *testing.T) {

	co := Config{}
	s := NewServer(co)
	s.ClientStorage = NewClientStorageInMemory()
	s.ClientDomainStorage = NewClientDomainStorageInMemory()
	s.CommentStorage = NewCommentStorageInMemory()
	s.mux = http.NewServeMux()
	s.BodyReader = io.ReadAll
	s.EmailSender = TestMailSender{}
	s.Config.Auth = true
	s.setupUrls()

	c, key, _ := s.ClientStorage.CreateClient("test client")
	s.ClientDomainStorage.AddClientDomain(c, "bla.net")

	var test_data = []struct {
		testName string
		req      *http.Request
		status   int
	}{
		{
			"comment with bad key",
			func() *http.Request {
				payload := CreateCommentRequest{
					Name:    "Zé",
					Content: "A comment",
				}
				j, _ := json.Marshal(payload)
				body := bytes.NewBuffer(j)
				req, _ := http.NewRequest("POST", "/comment/", body)
				req.Header.Set("X-APIKey", "bad")
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-PageURL", "https://bla.net/post")
				req.Header.Set("X-ClientUUID", c.UUID)
				return req

			}(),
			403},
		{
			"comment ok",
			func() *http.Request {
				payload := CreateCommentRequest{
					Name:    "Zé",
					Content: "A comment",
				}
				j, _ := json.Marshal(payload)
				body := bytes.NewBuffer(j)
				req, _ := http.NewRequest("POST", "/comment/", body)
				req.Header.Set("X-APIKey", key)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-ClientUUID", c.UUID)
				req.Header.Set("X-PageURL", "https://bla.net/post")
				return req

			}(),
			201},
	}

	for _, test := range test_data {
		t.Run(test.testName, func(t *testing.T) {
			w := httptest.NewRecorder()
			s.mux.ServeHTTP(w, test.req)

			if w.Code != test.status {
				t.Fatalf("bad status for %d", w.Code)
			}

		})
	}

}

func TestListComments(t *testing.T) {
	co := Config{}
	s := NewServer(co)
	s.ClientStorage = NewClientStorageInMemory()
	s.ClientDomainStorage = NewClientDomainStorageInMemory()
	s.CommentStorage = NewCommentStorageInMemory()
	s.mux = http.NewServeMux()
	s.BodyReader = io.ReadAll
	s.setupUrls()

	c, _, _ := s.ClientStorage.CreateClient("test client")
	d, _ := s.ClientDomainStorage.AddClientDomain(c, "bla.net")

	comments := []Comment{
		{
			Author:  "Zé",
			Content: "The comment",
			PageURL: "http://bla.net/post1",
		},
		{
			Author:  "Tião",
			Content: "The other comment",
			PageURL: "http://bla.net/post1",
		},
		{
			Author:  "Jão",
			Content: "The new comment",
			PageURL: "http://bla.net/post1",
		},
		{
			Author:  "Zé",
			Content: "The new new comment",
			PageURL: "http://bla.net/post2",
		},
		{
			Author:  "Jão",
			Content: "Another comment",
			PageURL: "http://bla.net/post2",
		},
	}

	for _, co := range comments {
		s.CommentStorage.CreateComment(c, d, co.Author, co.Content, co.PageURL)
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
				req, _ := http.NewRequest("GET", "/comment/", nil)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-PageURL", "https://bla.net/post1")
				req.Header.Set("X-ClientUUID", uuid)
				return req
			}(),
			403,
		},
		{
			"list comments with wrong origin",
			func() *http.Request {
				req, _ := http.NewRequest("GET", "/comment/", nil)
				req.Header.Set("Origin", "https://bleble.net")
				req.Header.Set("X-PageURL", "https://bla.net/post1")
				req.Header.Set("X-ClientUUID", c.UUID)
				return req
			}(),
			403,
		},
		{
			"list comments without origin",
			func() *http.Request {
				req, _ := http.NewRequest("GET", "/comment/", nil)
				req.Header.Set("X-PageURL", "https://bla.net/post1")
				req.Header.Set("X-ClientUUID", c.UUID)
				return req
			}(),
			403,
		},
		{
			"list comments options",
			func() *http.Request {
				req, _ := http.NewRequest("OPTIONS", "/comment/", nil)
				req.Header.Set("X-PageURL", "https://bla.net/post1")
				return req
			}(),
			204,
		},
		{
			"list comments ok",
			func() *http.Request {
				req, _ := http.NewRequest("GET", "/comment/", nil)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-PageURL", "https://bla.net/post1")
				req.Header.Set("X-ClientUUID", c.UUID)
				return req
			}(),
			200,
		},
	}

	for _, test := range test_data {
		t.Run(test.testName, func(t *testing.T) {
			w := httptest.NewRecorder()
			s.mux.ServeHTTP(w, test.req)

			if w.Code != test.status {
				t.Fatalf("bad status for %d", w.Code)
			}

		})
	}
}

func TestListComments_auth(t *testing.T) {
	co := Config{}
	s := NewServer(co)
	s.ClientStorage = NewClientStorageInMemory()
	s.ClientDomainStorage = NewClientDomainStorageInMemory()
	s.CommentStorage = NewCommentStorageInMemory()
	s.mux = http.NewServeMux()
	s.BodyReader = io.ReadAll
	s.Config.Auth = true
	s.setupUrls()

	c, key, _ := s.ClientStorage.CreateClient("test client")
	d, _ := s.ClientDomainStorage.AddClientDomain(c, "bla.net")

	comments := []Comment{
		{
			Author:  "Zé",
			Content: "The comment",
			PageURL: "http://bla.net/post1",
		},
		{
			Author:  "Tião",
			Content: "The other comment",
			PageURL: "http://bla.net/post1",
		},
		{
			Author:  "Jão",
			Content: "The new comment",
			PageURL: "http://bla.net/post1",
		},
		{
			Author:  "Zé",
			Content: "The new new comment",
			PageURL: "http://bla.net/post2",
		},
		{
			Author:  "Jão",
			Content: "Another comment",
			PageURL: "http://bla.net/post2",
		},
	}

	for _, co := range comments {
		s.CommentStorage.CreateComment(c, d, co.Author, co.Content, co.PageURL)
	}

	var test_data = []struct {
		testName string
		req      *http.Request
		status   int
	}{
		{
			"list comments with bad key",
			func() *http.Request {
				req, _ := http.NewRequest("GET", "/comment/", nil)
				req.Header.Set("X-APIKey", "bad")
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-PageURL", "https://bla.net/post1")
				req.Header.Set("X-ClientUUID", c.UUID)
				return req
			}(),
			403,
		},
		{
			"list comments ok",
			func() *http.Request {
				req, _ := http.NewRequest("GET", "/comment/", nil)
				req.Header.Set("X-APIKey", key)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-PageURL", "https://bla.net/post1")
				req.Header.Set("X-ClientUUID", c.UUID)
				return req
			}(),
			200,
		},
	}

	for _, test := range test_data {
		t.Run(test.testName, func(t *testing.T) {
			w := httptest.NewRecorder()
			s.mux.ServeHTTP(w, test.req)

			if w.Code != test.status {
				t.Fatalf("bad status for %d", w.Code)
			}

		})
	}
}

func TestListCommentsHTML(t *testing.T) {
	co := Config{}
	s := NewServer(co)
	s.ClientStorage = NewClientStorageInMemory()
	s.ClientDomainStorage = NewClientDomainStorageInMemory()
	s.CommentStorage = NewCommentStorageInMemory()
	s.mux = http.NewServeMux()
	s.BodyReader = io.ReadAll
	s.setupUrls()

	c, _, _ := s.ClientStorage.CreateClient("test client")
	d, _ := s.ClientDomainStorage.AddClientDomain(c, "bla.net")

	comments := []Comment{
		{
			Author:  "Zé",
			Content: "The comment",
			PageURL: "http://bla.net/post1",
		},
		{
			Author:  "Tião",
			Content: "The other comment",
			PageURL: "http://bla.net/post1",
		},
		{
			Author:  "Jão",
			Content: "The new comment",
			PageURL: "http://bla.net/post1",
		},
		{
			Author:  "Zé",
			Content: "The new new comment",
			PageURL: "http://bla.net/post2",
		},
		{
			Author:  "Jão",
			Content: "Another comment",
			PageURL: "http://bla.net/post2",
		},
	}

	for _, co := range comments {
		s.CommentStorage.CreateComment(c, d, co.Author, co.Content, co.PageURL)
	}

	var test_data = []struct {
		testName string
		req      *http.Request
		status   int
	}{
		{
			"list comments html with bad client",
			func() *http.Request {
				uuid, _ := GenUUID4()
				req, _ := http.NewRequest("GET", "/comment/html", nil)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-PageURL", "https://bla.net/post1")
				req.Header.Set("X-ClientUUID", uuid)
				return req
			}(),
			403,
		},
		{
			"list comments html with wrong origin",
			func() *http.Request {
				req, _ := http.NewRequest(
					"GET", "/comment/html", nil)
				req.Header.Set("Origin", "https://bleble.net")
				req.Header.Set("X-PageURL", "https://bla.net/post1")
				req.Header.Set("X-ClientUUID", c.UUID)
				return req
			}(),
			403,
		},
		{
			"list comments html without origin",
			func() *http.Request {
				req, _ := http.NewRequest(
					"GET", "/comment/html", nil)
				req.Header.Set("X-PageURL", "https://bla.net/post1")
				req.Header.Set("X-ClientUUID", c.UUID)
				return req
			}(),
			403,
		},
		{
			"list comments options",
			func() *http.Request {
				req, _ := http.NewRequest("OPTIONS", "/comment/html", nil)
				req.Header.Set("X-PageURL", "https://bla.net/post1")
				return req
			}(),
			204,
		},
		{
			"list comments html ok",
			func() *http.Request {
				req, _ := http.NewRequest(
					"GET", "/comment/html", nil)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-PageURL", "https://bla.net/post1")
				req.Header.Set("X-ClientUUID", c.UUID)
				return req
			}(),
			200,
		},
		{
			"list comments html ok with language",
			func() *http.Request {
				req, _ := http.NewRequest(
					"GET", "/comment/html", nil)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-PageURL", "https://bla.net/post1")
				req.Header.Set("X-ClientUUID", c.UUID)
				req.Header.Set("Accepted-Language", "pt-BR")
				return req
			}(),
			200,
		},
		{
			"list comments html ok with missing language",
			func() *http.Request {
				req, _ := http.NewRequest(
					"GET", "/comment/html", nil)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-PageURL", "https://bla.net/post1")
				req.Header.Set("X-ClientUUID", c.UUID)
				req.Header.Set("Accepted-Language", "es-AR")
				return req
			}(),
			200,
		},
	}

	for _, test := range test_data {
		t.Run(test.testName, func(t *testing.T) {
			w := httptest.NewRecorder()
			s.mux.ServeHTTP(w, test.req)

			if w.Code != test.status {
				t.Fatalf("bad status for %d", w.Code)
			}

		})
	}
}

func TestListCommentsHTML_Auth(t *testing.T) {
	co := Config{}
	s := NewServer(co)
	s.ClientStorage = NewClientStorageInMemory()
	s.ClientDomainStorage = NewClientDomainStorageInMemory()
	s.CommentStorage = NewCommentStorageInMemory()
	s.mux = http.NewServeMux()
	s.BodyReader = io.ReadAll
	s.Config.Auth = true
	s.setupUrls()

	c, key, _ := s.ClientStorage.CreateClient("test client")
	d, _ := s.ClientDomainStorage.AddClientDomain(c, "bla.net")

	comments := []Comment{
		{
			Author:  "Zé",
			Content: "The comment",
			PageURL: "http://bla.net/post1",
		},
		{
			Author:  "Tião",
			Content: "The other comment",
			PageURL: "http://bla.net/post1",
		},
		{
			Author:  "Jão",
			Content: "The new comment",
			PageURL: "http://bla.net/post1",
		},
		{
			Author:  "Zé",
			Content: "The new new comment",
			PageURL: "http://bla.net/post2",
		},
		{
			Author:  "Jão",
			Content: "Another comment",
			PageURL: "http://bla.net/post2",
		},
	}

	for _, co := range comments {
		s.CommentStorage.CreateComment(c, d, co.Author, co.Content, co.PageURL)
	}

	var test_data = []struct {
		testName string
		req      *http.Request
		status   int
	}{
		{
			"list comments html with bad key",
			func() *http.Request {
				req, _ := http.NewRequest("GET", "/comment/html", nil)
				req.Header.Set("X-APIKey", "bad")
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-PageURL", "https://bla.net/post1")
				req.Header.Set("X-ClientUUID", c.UUID)
				return req
			}(),
			403,
		},
		{
			"list comments html ok",
			func() *http.Request {
				req, _ := http.NewRequest(
					"GET", "/comment/html", nil)
				req.Header.Set("X-APIKey", key)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-PageURL", "https://bla.net/post1")
				req.Header.Set("X-ClientUUID", c.UUID)
				return req
			}(),
			200,
		},
	}

	for _, test := range test_data {
		t.Run(test.testName, func(t *testing.T) {
			w := httptest.NewRecorder()
			s.mux.ServeHTTP(w, test.req)

			if w.Code != test.status {
				t.Fatalf("bad status for %d", w.Code)
			}

		})
	}
}

func TestCountComments(t *testing.T) {
	co := Config{}
	s := NewServer(co)
	s.ClientStorage = NewClientStorageInMemory()
	s.ClientDomainStorage = NewClientDomainStorageInMemory()
	s.CommentStorage = NewCommentStorageInMemory()
	s.mux = http.NewServeMux()
	s.BodyReader = io.ReadAll
	s.setupUrls()

	c, _, _ := s.ClientStorage.CreateClient("test client")
	d, _ := s.ClientDomainStorage.AddClientDomain(c, "bla.net")

	comments := []Comment{
		{
			Author:  "Zé",
			Content: "The comment",
			PageURL: "http://bla.net/post1",
		},
		{
			Author:  "Tião",
			Content: "The other comment",
			PageURL: "http://bla.net/post1",
		},
		{
			Author:  "Jão",
			Content: "The new comment",
			PageURL: "http://bla.net/post1",
		},
		{
			Author:  "Zé",
			Content: "The new new comment",
			PageURL: "http://bla.net/post2",
		},
		{
			Author:  "Jão",
			Content: "Another comment",
			PageURL: "http://bla.net/post2",
		},
	}

	for _, co := range comments {
		s.CommentStorage.CreateComment(c, d, co.Author, co.Content, co.PageURL)
	}

	var test_data = []struct {
		testName string
		req      *http.Request
		status   int
	}{
		{
			"count comments with bad client",
			func() *http.Request {
				uuid, _ := GenUUID4()
				payload := CountCommentsRequest{
					PageURLs: []string{"http://bla.net/post1", "http://bla.net/post2"},
				}
				j, _ := json.Marshal(payload)
				body := bytes.NewBuffer(j)
				req, _ := http.NewRequest("POST", "/comment/count", body)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-ClientUUID", uuid)
				return req
			}(),
			403,
		},
		{
			"count comments with wrong origin",
			func() *http.Request {
				payload := CountCommentsRequest{
					PageURLs: []string{"http://bla.net/post1", "http://bla.net/post2"},
				}
				j, _ := json.Marshal(payload)
				body := bytes.NewBuffer(j)
				req, _ := http.NewRequest("POST", "/comment/count", body)
				req.Header.Set("Origin", "https://bleble.net")
				req.Header.Set("X-ClientUUID", c.UUID)
				return req
			}(),
			403,
		},
		{
			"count comments without origin",
			func() *http.Request {
				payload := CountCommentsRequest{
					PageURLs: []string{"http://bla.net/post1", "http://bla.net/post2"},
				}
				j, _ := json.Marshal(payload)
				body := bytes.NewBuffer(j)
				req, _ := http.NewRequest("POST", "/comment/count", body)
				req.Header.Set("X-ClientUUID", c.UUID)
				return req
			}(),
			403,
		},
		{
			"count comments malformed payload",
			func() *http.Request {
				body := bytes.NewBuffer([]byte("bad"))
				req, _ := http.NewRequest("POST", "/comment/count", body)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-ClientUUID", c.UUID)
				return req
			}(),
			400,
		},
		{
			"count comments options",
			func() *http.Request {
				req, _ := http.NewRequest("OPTIONS", "/comment/count", nil)
				req.Header.Set("Origin", "https://bla.net")
				return req
			}(),
			204,
		},

		{
			"count comments ok",
			func() *http.Request {
				payload := CountCommentsRequest{
					PageURLs: []string{"http://bla.net/post1", "http://bla.net/post2"},
				}
				j, _ := json.Marshal(payload)
				body := bytes.NewBuffer(j)
				req, _ := http.NewRequest("POST", "/comment/count", body)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-ClientUUID", c.UUID)
				return req
			}(),
			200,
		},
	}

	for _, test := range test_data {
		t.Run(test.testName, func(t *testing.T) {
			w := httptest.NewRecorder()
			s.mux.ServeHTTP(w, test.req)

			if w.Code != test.status {
				t.Fatalf("bad status for %d", w.Code)
			}

		})
	}
}

func TestCountComments_Auth(t *testing.T) {
	co := Config{}
	s := NewServer(co)
	s.ClientStorage = NewClientStorageInMemory()
	s.ClientDomainStorage = NewClientDomainStorageInMemory()
	s.CommentStorage = NewCommentStorageInMemory()
	s.mux = http.NewServeMux()
	s.BodyReader = io.ReadAll
	s.Config.Auth = true
	s.setupUrls()

	c, key, _ := s.ClientStorage.CreateClient("test client")
	d, _ := s.ClientDomainStorage.AddClientDomain(c, "bla.net")

	comments := []Comment{
		{
			Author:  "Zé",
			Content: "The comment",
			PageURL: "http://bla.net/post1",
		},
		{
			Author:  "Tião",
			Content: "The other comment",
			PageURL: "http://bla.net/post1",
		},
		{
			Author:  "Jão",
			Content: "The new comment",
			PageURL: "http://bla.net/post1",
		},
		{
			Author:  "Zé",
			Content: "The new new comment",
			PageURL: "http://bla.net/post2",
		},
		{
			Author:  "Jão",
			Content: "Another comment",
			PageURL: "http://bla.net/post2",
		},
	}

	for _, co := range comments {
		s.CommentStorage.CreateComment(c, d, co.Author, co.Content, co.PageURL)
	}

	var test_data = []struct {
		testName string
		req      *http.Request
		status   int
	}{
		{
			"count comments with bad key",
			func() *http.Request {
				payload := CountCommentsRequest{
					PageURLs: []string{"http://bla.net/post1", "http://bla.net/post2"},
				}
				j, _ := json.Marshal(payload)
				body := bytes.NewBuffer(j)
				req, _ := http.NewRequest("POST", "/comment/count", body)
				req.Header.Set("X-APIKey", "bad")
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-ClientUUID", c.UUID)
				return req
			}(),
			403,
		},
		{
			"count comments ok",
			func() *http.Request {
				payload := CountCommentsRequest{
					PageURLs: []string{"http://bla.net/post1", "http://bla.net/post2"},
				}
				j, _ := json.Marshal(payload)
				body := bytes.NewBuffer(j)
				req, _ := http.NewRequest("POST", "/comment/count", body)
				req.Header.Set("X-APIKey", key)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-ClientUUID", c.UUID)
				return req
			}(),
			200,
		},
	}

	for _, test := range test_data {
		t.Run(test.testName, func(t *testing.T) {
			w := httptest.NewRecorder()
			s.mux.ServeHTTP(w, test.req)

			if w.Code != test.status {
				t.Fatalf("bad status for %d", w.Code)
			}

		})
	}
}

func TestCountCommentsHTML(t *testing.T) {
	co := Config{}
	s := NewServer(co)
	s.ClientStorage = NewClientStorageInMemory()
	s.ClientDomainStorage = NewClientDomainStorageInMemory()
	s.CommentStorage = NewCommentStorageInMemory()
	s.mux = http.NewServeMux()
	s.BodyReader = io.ReadAll
	s.setupUrls()

	c, _, _ := s.ClientStorage.CreateClient("test client")
	d, _ := s.ClientDomainStorage.AddClientDomain(c, "bla.net")

	comments := []Comment{
		{
			Author:  "Zé",
			Content: "The comment",
			PageURL: "http://bla.net/post1",
		},
		{
			Author:  "Tião",
			Content: "The other comment",
			PageURL: "http://bla.net/post1",
		},
		{
			Author:  "Jão",
			Content: "The new comment",
			PageURL: "http://bla.net/post1",
		},
		{
			Author:  "Zé",
			Content: "The new new comment",
			PageURL: "http://bla.net/post2",
		},
		{
			Author:  "Jão",
			Content: "Another comment",
			PageURL: "http://bla.net/post2",
		},
	}

	for _, co := range comments {
		s.CommentStorage.CreateComment(c, d, co.Author, co.Content, co.PageURL)
	}

	var test_data = []struct {
		testName string
		req      *http.Request
		status   int
	}{
		{
			"count comments html with bad client",
			func() *http.Request {
				uuid, _ := GenUUID4()
				payload := CountCommentsRequest{
					PageURLs: []string{"http://bla.net/post1", "http://bla.net/post2"},
				}
				j, _ := json.Marshal(payload)
				body := bytes.NewBuffer(j)
				req, _ := http.NewRequest("POST", "/comment/count/html", body)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-ClientUUID", uuid)
				return req
			}(),
			403,
		},
		{
			"count comments html with wrong origin",
			func() *http.Request {
				payload := CountCommentsRequest{
					PageURLs: []string{"http://bla.net/post1", "http://bla.net/post2"},
				}
				j, _ := json.Marshal(payload)
				body := bytes.NewBuffer(j)
				req, _ := http.NewRequest("POST", "/comment/count/html", body)
				req.Header.Set("Origin", "https://bleble.net")
				req.Header.Set("X-ClientUUID", c.UUID)
				return req
			}(),
			403,
		},
		{
			"count comments html without origin",
			func() *http.Request {
				payload := CountCommentsRequest{
					PageURLs: []string{"http://bla.net/post1", "http://bla.net/post2"},
				}
				j, _ := json.Marshal(payload)
				body := bytes.NewBuffer(j)
				req, _ := http.NewRequest("POST", "/comment/count/html", body)
				req.Header.Set("X-PageURL", "https://bla.net/post1")
				req.Header.Set("X-ClientUUID", c.UUID)
				return req
			}(),
			403,
		},
		{
			"count comments html malformed payload",
			func() *http.Request {
				body := bytes.NewBuffer([]byte("bad"))
				req, _ := http.NewRequest("POST", "/comment/count/html", body)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-PageURL", "https://bla.net/post1")
				req.Header.Set("X-ClientUUID", c.UUID)
				return req
			}(),
			400,
		},

		{
			"count comments html options",
			func() *http.Request {
				req, _ := http.NewRequest("OPTIONS", "/comment/count/html", nil)
				req.Header.Set("Origin", "https://bla.net")
				return req
			}(),
			204,
		},

		{
			"count comments html ok",
			func() *http.Request {
				payload := CountCommentsRequest{
					PageURLs: []string{"http://bla.net/post1", "http://bla.net/post2"},
				}
				j, _ := json.Marshal(payload)
				body := bytes.NewBuffer(j)
				req, _ := http.NewRequest("POST", "/comment/count/html", body)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-ClientUUID", c.UUID)
				return req
			}(),
			200,
		},
	}

	for _, test := range test_data {
		t.Run(test.testName, func(t *testing.T) {
			w := httptest.NewRecorder()
			s.mux.ServeHTTP(w, test.req)

			if w.Code != test.status {
				t.Fatalf("bad status for %d", w.Code)
			}

		})
	}
}

func TestCountCommentsHTML_Auth(t *testing.T) {
	co := Config{}
	s := NewServer(co)
	s.ClientStorage = NewClientStorageInMemory()
	s.ClientDomainStorage = NewClientDomainStorageInMemory()
	s.CommentStorage = NewCommentStorageInMemory()
	s.mux = http.NewServeMux()
	s.BodyReader = io.ReadAll
	s.Config.Auth = true
	s.setupUrls()

	c, key, _ := s.ClientStorage.CreateClient("test client")
	d, _ := s.ClientDomainStorage.AddClientDomain(c, "bla.net")

	comments := []Comment{
		{
			Author:  "Zé",
			Content: "The comment",
			PageURL: "http://bla.net/post1",
		},
		{
			Author:  "Tião",
			Content: "The other comment",
			PageURL: "http://bla.net/post1",
		},
		{
			Author:  "Jão",
			Content: "The new comment",
			PageURL: "http://bla.net/post1",
		},
		{
			Author:  "Zé",
			Content: "The new new comment",
			PageURL: "http://bla.net/post2",
		},
		{
			Author:  "Jão",
			Content: "Another comment",
			PageURL: "http://bla.net/post2",
		},
	}

	for _, co := range comments {
		s.CommentStorage.CreateComment(c, d, co.Author, co.Content, co.PageURL)
	}

	var test_data = []struct {
		testName string
		req      *http.Request
		status   int
	}{
		{
			"count comments html with bad key",
			func() *http.Request {
				uuid, _ := GenUUID4()
				payload := CountCommentsRequest{
					PageURLs: []string{"http://bla.net/post1", "http://bla.net/post2"},
				}
				j, _ := json.Marshal(payload)
				body := bytes.NewBuffer(j)
				req, _ := http.NewRequest("POST", "/comment/count/html", body)
				req.Header.Set("X-APIKey", "Bad")
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-ClientUUID", uuid)
				return req
			}(),
			403,
		},
		{
			"count comments html ok",
			func() *http.Request {
				payload := CountCommentsRequest{
					PageURLs: []string{"http://bla.net/post1", "http://bla.net/post2"},
				}
				j, _ := json.Marshal(payload)
				body := bytes.NewBuffer(j)
				req, _ := http.NewRequest("POST", "/comment/count/html", body)
				req.Header.Set("X-APIKey", key)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-ClientUUID", c.UUID)
				return req
			}(),
			200,
		},
	}

	for _, test := range test_data {
		t.Run(test.testName, func(t *testing.T) {
			w := httptest.NewRecorder()
			s.mux.ServeHTTP(w, test.req)

			if w.Code != test.status {
				t.Fatalf("bad status for %d", w.Code)
			}

		})
	}
}

func TestParlanteJS(t *testing.T) {
	co := Config{}
	s := NewServer(co)

	req, _ := http.NewRequest("GET", "/parlante.js", nil)
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("bad status for parlante.js %d", w.Code)
	}
}

func TestGetPingMeForm(t *testing.T) {
	co := Config{}
	s := NewServer(co)
	s.ClientStorage = NewClientStorageInMemory()
	s.ClientDomainStorage = NewClientDomainStorageInMemory()
	s.CommentStorage = NewCommentStorageInMemory()
	s.mux = http.NewServeMux()
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
			"pingme form bad client",
			func() *http.Request {
				req, _ := http.NewRequest("GET", "/pingme/", nil)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-ClientUUID", "xxx")
				return req

			}(),
			403,
		},
		{
			"pingme form wrong origin",
			func() *http.Request {
				req, _ := http.NewRequest("GET", "/pingme/", nil)
				req.Header.Set("Origin", "https://blable.net")
				req.Header.Set("X-ClientUUID", c.UUID)
				return req

			}(),
			403,
		},
		{
			"pingme form GET",
			func() *http.Request {
				req, _ := http.NewRequest("GET", "/pingme/", nil)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-ClientUUID", c.UUID)
				return req

			}(),
			200,
		},
		{
			"pingme form OPTIONS",
			func() *http.Request {
				req, _ := http.NewRequest("OPTIONS", "/pingme/", nil)
				req.Header.Set("Origin", "https://bla.net")
				return req

			}(),
			204,
		},
	}

	for _, test := range test_data {
		t.Run(test.testName, func(t *testing.T) {
			w := httptest.NewRecorder()
			s.mux.ServeHTTP(w, test.req)

			if w.Code != test.status {
				t.Fatalf("bad status for pingme form %d", w.Code)
			}
		})
	}
}

func TestGetPingMeForm_Auth(t *testing.T) {
	co := Config{}
	s := NewServer(co)
	s.ClientStorage = NewClientStorageInMemory()
	s.ClientDomainStorage = NewClientDomainStorageInMemory()
	s.CommentStorage = NewCommentStorageInMemory()
	s.mux = http.NewServeMux()
	s.BodyReader = io.ReadAll
	s.Config.Auth = true
	s.setupUrls()

	c, key, _ := s.ClientStorage.CreateClient("test client")
	s.ClientDomainStorage.AddClientDomain(c, "bla.net")

	var test_data = []struct {
		testName string
		req      *http.Request
		status   int
	}{
		{
			"pingme form bad client",
			func() *http.Request {
				req, _ := http.NewRequest("GET", "/pingme/", nil)
				req.Header.Set("X-APIKey", "bad")
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-ClientUUID", c.UUID)
				return req

			}(),
			403,
		},
		{
			"pingme form GET",
			func() *http.Request {
				req, _ := http.NewRequest("GET", "/pingme/", nil)
				req.Header.Set("X-APIKey", key)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-ClientUUID", c.UUID)
				return req

			}(),
			200,
		},
	}

	for _, test := range test_data {
		t.Run(test.testName, func(t *testing.T) {
			w := httptest.NewRecorder()
			s.mux.ServeHTTP(w, test.req)

			if w.Code != test.status {
				t.Fatalf("bad status for pingme form %d", w.Code)
			}
		})
	}
}

func TestPingMe(t *testing.T) {

	co := Config{}
	s := NewServer(co)
	s.ClientStorage = NewClientStorageInMemory()
	s.ClientDomainStorage = NewClientDomainStorageInMemory()
	s.CommentStorage = NewCommentStorageInMemory()
	sender := TestMailSender{}
	s.EmailSender = &sender
	s.mux = http.NewServeMux()
	s.BodyReader = io.ReadAll
	s.setupUrls()

	c, _, _ := s.ClientStorage.CreateClient("test client")
	s.ClientDomainStorage.AddClientDomain(c, "bla.net")

	var test_data = []struct {
		testName string
		req      *http.Request
		status   int
		setup    func()
		teardown func()
	}{
		{
			"pingme with bad client",
			func() *http.Request {
				uuid, _ := GenUUID4()
				payload := PingMeRequest{
					Name:    "Zé",
					Message: "A message",
					Email:   "a@a.com",
				}
				j, _ := json.Marshal(payload)
				body := bytes.NewBuffer(j)
				req, _ := http.NewRequest("POST", "/pingme/", body)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-ClientUUID", uuid)
				return req

			}(),
			403,
			nil,
			nil,
		},
		{
			"pingme with wrong origin",
			func() *http.Request {
				payload := PingMeRequest{
					Name:    "Zé",
					Message: "A message",
					Email:   "a@a.com",
				}
				j, _ := json.Marshal(payload)
				body := bytes.NewBuffer(j)
				req, _ := http.NewRequest("POST", "/pingme/", body)
				req.Header.Set("Origin", "https://bleble.net")
				req.Header.Set("X-ClientUUID", c.UUID)
				return req

			}(),
			403,
			nil,
			nil,
		},
		{
			"pingme missing body",
			func() *http.Request {
				req, _ := http.NewRequest("POST", "/pingme/", nil)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-PageURL", "https://bla.net/post")
				req.Header.Set("X-ClientUUID", c.UUID)
				return req

			}(),
			400,
			nil,
			nil,
		},
		{
			"pingme bad body",
			func() *http.Request {
				body := bytes.NewBuffer([]byte(""))
				req, _ := http.NewRequest("POST", "/pingme/", body)
				req.Header.Set("X-ClientUUID", c.UUID)
				req.Header.Set("Origin", "https://bla.net")
				return req

			}(),
			400,
			nil,
			nil,
		},
		{
			"pingme without name",
			func() *http.Request {
				payload := PingMeRequest{
					Message: "A message",
					Email:   "a@a.com",
				}
				j, _ := json.Marshal(payload)
				body := bytes.NewBuffer(j)
				req, _ := http.NewRequest("POST", "/pingme/", body)
				req.Header.Set("X-ClientUUID", c.UUID)
				req.Header.Set("Origin", "https://bla.net")
				return req

			}(),
			400,
			nil,
			nil,
		},
		{
			"pingme without message",
			func() *http.Request {
				payload := PingMeRequest{
					Name:  "Zé",
					Email: "a@a.com",
				}
				j, _ := json.Marshal(payload)
				body := bytes.NewBuffer(j)
				req, _ := http.NewRequest("POST", "/pingme/", body)
				req.Header.Set("X-ClientUUID", c.UUID)
				req.Header.Set("Origin", "https://bla.net")
				return req

			}(),
			400,
			nil,
			nil,
		},
		{
			"pingme without email",
			func() *http.Request {
				payload := PingMeRequest{
					Name:    "Zé",
					Message: "A message",
				}
				j, _ := json.Marshal(payload)
				body := bytes.NewBuffer(j)
				req, _ := http.NewRequest("POST", "/pingme/", body)
				req.Header.Set("X-ClientUUID", c.UUID)
				req.Header.Set("Origin", "https://bla.net")
				return req

			}(),
			400,
			nil,
			nil,
		},
		{
			"pingme error sending email",
			func() *http.Request {
				payload := PingMeRequest{
					Name:    "Zé",
					Message: "A message",
					Email:   "a@a.com",
				}
				j, _ := json.Marshal(payload)
				body := bytes.NewBuffer(j)
				req, _ := http.NewRequest("POST", "/pingme/", body)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-ClientUUID", c.UUID)
				return req

			}(),
			500,
			func() { sender.ForceError(true) },
			func() { sender.ForceError(false) },
		},
		{
			"pingme ok",
			func() *http.Request {
				payload := PingMeRequest{
					Name:    "Zé",
					Message: "A message",
					Email:   "a@a.com",
				}
				j, _ := json.Marshal(payload)
				body := bytes.NewBuffer(j)
				req, _ := http.NewRequest("POST", "/pingme/", body)
				req.Header.Set("X-ClientUUID", c.UUID)
				req.Header.Set("Origin", "https://bla.net")
				return req

			}(),
			201,
			nil,
			nil,
		},
	}

	for _, test := range test_data {
		t.Run(test.testName, func(t *testing.T) {
			w := httptest.NewRecorder()
			if test.setup != nil {
				test.setup()
			}
			s.mux.ServeHTTP(w, test.req)

			if test.teardown != nil {
				test.teardown()
			}
			if w.Code != test.status {
				t.Fatalf("bad status for %d", w.Code)
			}

		})
	}

}

func TestPingMe_Auth(t *testing.T) {

	co := Config{}
	s := NewServer(co)
	s.ClientStorage = NewClientStorageInMemory()
	s.ClientDomainStorage = NewClientDomainStorageInMemory()
	s.CommentStorage = NewCommentStorageInMemory()
	sender := TestMailSender{}
	s.EmailSender = &sender
	s.mux = http.NewServeMux()
	s.BodyReader = io.ReadAll
	s.Config.Auth = true
	s.setupUrls()

	c, key, _ := s.ClientStorage.CreateClient("test client")
	s.ClientDomainStorage.AddClientDomain(c, "bla.net")

	var test_data = []struct {
		testName string
		req      *http.Request
		status   int
		setup    func()
		teardown func()
	}{
		{
			"pingme with bad key",
			func() *http.Request {
				payload := PingMeRequest{
					Name:    "Zé",
					Message: "A message",
					Email:   "a@a.com",
				}
				j, _ := json.Marshal(payload)
				body := bytes.NewBuffer(j)
				req, _ := http.NewRequest("POST", "/pingme/", body)
				req.Header.Set("X-APIKey", "bad")
				req.Header.Set("X-ClientUUID", c.UUID)
				req.Header.Set("Origin", "https://bla.net")
				return req

			}(),
			403,
			nil,
			nil,
		},
		{
			"pingme ok",
			func() *http.Request {
				payload := PingMeRequest{
					Name:    "Zé",
					Message: "A message",
					Email:   "a@a.com",
				}
				j, _ := json.Marshal(payload)
				body := bytes.NewBuffer(j)
				req, _ := http.NewRequest("POST", "/pingme/", body)
				req.Header.Set("X-APIKey", key)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-ClientUUID", c.UUID)
				return req

			}(),
			201,
			nil,
			nil,
		},
	}

	for _, test := range test_data {
		t.Run(test.testName, func(t *testing.T) {
			w := httptest.NewRecorder()
			if test.setup != nil {
				test.setup()
			}
			s.mux.ServeHTTP(w, test.req)

			if test.teardown != nil {
				test.teardown()
			}
			if w.Code != test.status {
				t.Fatalf("bad status for %d", w.Code)
			}

		})
	}

}

func TestComments_WithErrors(t *testing.T) {

	co := Config{}
	s := NewServer(co)
	client_storage := NewClientStorageInMemory()
	s.ClientStorage = client_storage
	s.ClientDomainStorage = NewClientDomainStorageInMemory()
	comment_storage := NewCommentStorageInMemory()
	s.CommentStorage = comment_storage
	s.mux = http.NewServeMux()
	s.BodyReader = readFn
	s.JsonMarshaler = errorMarshal
	s.HtmlRenderer = errorHtmlRender
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
				req, _ := http.NewRequest("POST", "/comment/", body)
				req.Header.Set("X-ClientUUID", uuid)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-PageURL", "https://bla.net/post")
				return req

			}(),
			403,
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
				req, _ := http.NewRequest("POST", "/comment/", body)
				req.Header.Set("X-ClientUUID", c.UUID)
				req.Header.Set("Origin", "https://bad.net")
				req.Header.Set("X-PageURL", "https://bla.net/post")
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
				req, _ := http.NewRequest("POST", "/comment/", body)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-ClientUUID", c.UUID)
				req.Header.Set("X-PageURL", "https://bla.net/post")
				return req

			}(),
			400,
		},
		{
			"create comment error reading body",
			func() *http.Request {
				body := bytes.NewBuffer([]byte("bad body"))

				req, _ := http.NewRequest("POST", "/comment/", body)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-ClientUUID", c.UUID)
				req.Header.Set("X-PageURL", "https://bla.net/post1")
				return req

			}(),
			400,
		},
		{
			"list comments with db error getting comments",
			func() *http.Request {
				req, _ := http.NewRequest("GET", "/comment/", nil)
				req.Header.Set("X-ClientUUID", c.UUID)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-PageURL", comment_storage.BadPage)
				return req

			}(),
			500,
		},
		{
			"list comments error marshal json",
			func() *http.Request {
				req, _ := http.NewRequest("GET", "/comment/", nil)
				req.Header.Set("X-ClientUUID", c.UUID)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-PageURL", "https://bla.net/post1")
				return req

			}(),
			500,
		},
		{
			"list comments html with db error getting comments",
			func() *http.Request {
				req, _ := http.NewRequest("GET", "/comment/html", nil)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-ClientUUID", c.UUID)
				req.Header.Set("X-PageURL", comment_storage.BadPage)
				return req

			}(),
			500,
		},
		{
			"list comments html error render html",
			func() *http.Request {
				req, _ := http.NewRequest("GET", "/comment/html", nil)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-PageURL", "https://bla.net/post1")
				req.Header.Set("X-ClientUUID", c.UUID)
				return req

			}(),
			500,
		},
		{
			"count comments missing body",
			func() *http.Request {
				req, _ := http.NewRequest("POST", "/comment/count", nil)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-ClientUUID", c.UUID)
				req.Header.Set("X-PageURL", "https://bla.net/post1")
				return req

			}(),
			400,
		},
		{
			"count comments error reading body",
			func() *http.Request {
				body := bytes.NewBuffer([]byte("bad body"))

				req, _ := http.NewRequest("POST", "/comment/count", body)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-ClientUUID", c.UUID)
				req.Header.Set("X-PageURL", "https://bla.net/post1")
				return req

			}(),
			400,
		},
		{
			"count comments error getting comments from db",
			func() *http.Request {
				payload := CountCommentsRequest{
					PageURLs: []string{"http://bla.net/post1", "http://bla.net/post2",
						comment_storage.BadPage},
				}
				j, _ := json.Marshal(payload)
				body := bytes.NewBuffer(j)
				req, _ := http.NewRequest("POST", "/comment/count", body)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-ClientUUID", c.UUID)
				req.Header.Set("X-PageURL", "https://bla.net/post1")
				return req

			}(),
			500,
		},
		{
			"count comments html missing body",
			func() *http.Request {
				req, _ := http.NewRequest("POST", "/comment/count/html", nil)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-PageURL", "https://bla.net/post1")
				req.Header.Set("X-ClientUUID", c.UUID)
				return req

			}(),
			400,
		},
		{
			"count comments hmtl error reading body",
			func() *http.Request {
				body := bytes.NewBuffer([]byte("bad body"))

				req, _ := http.NewRequest("POST", "/comment/count/html", body)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-ClientUUID", c.UUID)
				req.Header.Set("X-PageURL", "https://bla.net/post1")
				return req

			}(),
			400,
		},
		{
			"count comments html error getting comments from db",
			func() *http.Request {
				payload := CountCommentsRequest{
					PageURLs: []string{"http://bla.net/post1", "http://bla.net/post2",
						comment_storage.BadPage},
				}
				j, _ := json.Marshal(payload)
				body := bytes.NewBuffer(j)
				req, _ := http.NewRequest("POST", "/comment/count/html", body)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-ClientUUID", c.UUID)
				req.Header.Set("X-PageURL", "https://bla.net/post1")
				return req

			}(),
			500,
		},

		{
			"count comments html error rendering template",
			func() *http.Request {
				payload := CountCommentsRequest{
					PageURLs: []string{"http://bla.net/post1", "http://bla.net/post2"},
				}
				j, _ := json.Marshal(payload)
				body := bytes.NewBuffer(j)
				req, _ := http.NewRequest("POST", "/comment/count/html", body)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-ClientUUID", c.UUID)
				req.Header.Set("X-PageURL", "https://bla.net/post1")
				return req

			}(),
			500,
		},
		{
			"get pingme form error rendering html",
			func() *http.Request {
				req, _ := http.NewRequest("GET", "/pingme/", nil)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-ClientUUID", c.UUID)
				return req

			}(),
			500,
		},
		{
			"pingme error reading body",
			func() *http.Request {
				body := bytes.NewBuffer([]byte("bad body"))
				req, _ := http.NewRequest("POST", "/pingme/", body)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("X-ClientUUID", c.UUID)
				return req

			}(),
			400,
		},
	}

	for _, test := range test_data {
		t.Run(test.testName, func(t *testing.T) {
			w := httptest.NewRecorder()
			s.mux.ServeHTTP(w, test.req)

			if w.Code != test.status {
				t.Fatalf("bad status for %d", w.Code)
			}

		})
	}

}

func TestConfig(t *testing.T) {
	type checkConf func(c Config)
	var test_data = []struct {
		testName string
		config   Config
		checkFn  checkConf
	}{
		{
			"test config without ssl",
			Config{Port: 9000, Host: "0.0.0.0"},
			func(c Config) {
				if c.UsesSSL() != false {
					t.Fatalf("Bad config whithout ssl")
				}
			},
		},
		{
			"test config wit ssl",
			Config{Port: 9000, Host: "0.0.0.0",
				CertFilePath: "/bla",
				KeyFilePath:  "/ble"},
			func(c Config) {
				if c.UsesSSL() != true {
					t.Fatalf("Bad config whith ssl")
				}
			},
		},
	}

	for _, test := range test_data {
		t.Run(test.testName, func(t *testing.T) {
			test.checkFn(test.config)
		})
	}
}

func TestRequestLogger(t *testing.T) {
	var s string
	fn := func(format string, v ...any) {
		s = fmt.Sprintf(format, v)
	}
	req, _ := http.NewRequest("GET", "/parlante.js", nil)
	w := httptest.NewRecorder()
	logger := RequestLogger{loggerFn: fn}
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	loggedHandler := logger.Log(http.HandlerFunc(handler))
	loggedHandler.ServeHTTP(w, req)

	if !strings.Contains(s, "GET /parlante.js") {
		t.Fatalf("bad request log %s", s)
	}
}
