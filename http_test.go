package parlante

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
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
				req, _ := http.NewRequest("GET", "/comment/"+uuid, nil)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("Referer", "https://bla.net/post1")
				return req
			}(),
			403,
		},
		{
			"list comments with wrong origin",
			func() *http.Request {
				req, _ := http.NewRequest("GET", "/comment/"+c.UUID, nil)
				req.Header.Set("Origin", "https://bleble.net")
				req.Header.Set("Referer", "https://bla.net/post1")
				return req
			}(),
			403,
		},
		{
			"list comments without origin",
			func() *http.Request {
				req, _ := http.NewRequest("GET", "/comment/"+c.UUID, nil)
				req.Header.Set("Referer", "https://bla.net/post1")
				return req
			}(),
			403,
		},
		{
			"list comments ok",
			func() *http.Request {
				req, _ := http.NewRequest("GET", "/comment/"+c.UUID, nil)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("Referer", "https://bla.net/post1")
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
				req, _ := http.NewRequest("GET", "/comment/"+uuid+"/html", nil)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("Referer", "https://bla.net/post1")
				return req
			}(),
			403,
		},
		{
			"list comments html with wrong origin",
			func() *http.Request {
				req, _ := http.NewRequest(
					"GET", "/comment/"+c.UUID+"/html", nil)
				req.Header.Set("Origin", "https://bleble.net")
				req.Header.Set("Referer", "https://bla.net/post1")
				return req
			}(),
			403,
		},
		{
			"list comments html without origin",
			func() *http.Request {
				req, _ := http.NewRequest(
					"GET", "/comment/"+c.UUID+"/html", nil)
				req.Header.Set("Referer", "https://bla.net/post1")
				return req
			}(),
			403,
		},
		{
			"list comments html ok",
			func() *http.Request {
				req, _ := http.NewRequest(
					"GET", "/comment/"+c.UUID+"/html", nil)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("Referer", "https://bla.net/post1")
				return req
			}(),
			200,
		},
		{
			"list comments html ok with language",
			func() *http.Request {
				req, _ := http.NewRequest(
					"GET", "/comment/"+c.UUID+"/html", nil)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("Referer", "https://bla.net/post1")
				req.Header.Set("Accepted-Language", "pt-BR")
				return req
			}(),
			200,
		},
		{
			"list comments html ok with missing language",
			func() *http.Request {
				req, _ := http.NewRequest(
					"GET", "/comment/"+c.UUID+"/html", nil)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("Referer", "https://bla.net/post1")
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
				req, _ := http.NewRequest("POST", "/comment/"+uuid, body)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("Referer", "https://bla.net/post")
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
			"create comment error reading body",
			func() *http.Request {
				body := bytes.NewBuffer([]byte("bad body"))

				req, _ := http.NewRequest("POST", "/comment/"+c.UUID, body)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("Referer", "https://bla.net/post1")
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
		{
			"list comments error marshal json",
			func() *http.Request {
				req, _ := http.NewRequest("GET", "/comment/"+c.UUID, nil)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("Referer", "https://bla.net/post1")
				return req

			}(),
			500,
		},
		{
			"list comments html with db error getting comments",
			func() *http.Request {
				req, _ := http.NewRequest(
					"GET", "/comment/"+c.UUID+"/html", nil)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("Referer", comment_storage.BadPage)
				return req

			}(),
			500,
		},
		{
			"list comments html error render html",
			func() *http.Request {
				req, _ := http.NewRequest(
					"GET", "/comment/"+c.UUID+"/html", nil)
				req.Header.Set("Origin", "https://bla.net")
				req.Header.Set("Referer", "https://bla.net/post1")
				return req

			}(),
			500,
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
