package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestParlante(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	err := setupTestDB()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(DBFILE)

	startServer()
	defer stopServer()
	type checkFn func(r *http.Response)

	cli := ClientStorageSQLite{}
	clid := ClientDomainStorageSQLite{}
	client, _, _ := cli.CreateClient("test client")
	clid.AddClientDomain(client, "localhost")

	var tests = []struct {
		testName string
		req      *http.Request
		checkFn  checkFn
	}{
		{
			"test create comment",
			func() *http.Request {
				payload := CreateCommentRequest{
					Name:    "Someone",
					Content: "Bla bla bla",
				}
				j, _ := json.Marshal(payload)
				body := io.NopCloser(strings.NewReader(string(j)))
				r, _ := http.NewRequest(
					"POST", "http://localhost:8080/comment/"+client.UUID,
					body)
				r.Header.Set("Origin", "https://localhost:8080")
				r.Header.Set("Referer", "https://localhost:8080/post")
				return r
			}(),
			func(r *http.Response) {
				if r.StatusCode != 201 {
					t.Fatalf("bad status for create comment %d", r.StatusCode)
				}
			},
		},
		{
			"test list comments",
			func() *http.Request {
				r, _ := http.NewRequest(
					"GET", "http://localhost:8080/comment/"+client.UUID, nil)
				r.Header.Set("Origin", "https://localhost:8080")
				r.Header.Set("Referer", "https://localhost:8080/post")
				return r
			}(),
			func(r *http.Response) {
				if r.StatusCode != 200 {
					t.Fatalf("bad status for list comments %d", r.StatusCode)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			c := http.Client{}
			resp, err := c.Do(test.req)

			if err != nil {
				t.Fatal(err)
			}
			test.checkFn(resp)

		})
	}
}

func startServer() {
	cmd := exec.Command("./build/parlante", "-dbpath", DBFILE)
	if cmd.Err != nil {
		panic(cmd.Err.Error())
	}
	err := cmd.Start()
	if err != nil {
		panic(err.Error())
	}
	time.Sleep(time.Millisecond * 200)
}

func stopServer() {
	exec.Command("killall", "parlante")
}
