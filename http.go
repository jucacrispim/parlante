package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

type ctxKey string

const ctxClientKey ctxKey = "client"
const ctxDomainKey ctxKey = "domain"

type bodyReader func(io.Reader) ([]byte, error)

type CreateCommentRequest struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type CreateCommentResponse struct {
	Msg string `json:"msg"`
}

type ListCommentsResponse struct {
	Total    int       `json:"total"`
	Comments []Comment `json:"comments"`
}

type ParlanteServer struct {
	ClientStorage       ClientStorage
	ClientDomainStorage ClientDomainStorage
	CommentStorage      CommentStorage
	Server              *http.ServeMux
	BodyReader          bodyReader
}

func (s ParlanteServer) CreateComment(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		http.Error(w, "Missing body", http.StatusBadRequest)
		return
	}
	rawbody, err := s.BodyReader(r.Body)
	if err != nil {
		// notest
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var body CreateCommentRequest
	err = json.Unmarshal(rawbody, &body)
	if err != nil {
		http.Error(w, "Malformed json", http.StatusBadRequest)
		return
	}

	c := r.Context().Value(ctxClientKey).(Client)
	cd := r.Context().Value(ctxDomainKey).(ClientDomain)

	page_url := r.Header.Get("Referer")

	_, err = s.CommentStorage.CreateComment(
		c, cd, body.Name, body.Content, page_url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp := CreateCommentResponse{Msg: "Ok"}
	j, _ := json.Marshal(resp)
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}

func (s ParlanteServer) ListComments(w http.ResponseWriter, r *http.Request) {
	c := r.Context().Value(ctxClientKey).(Client)
	cd := r.Context().Value(ctxDomainKey).(ClientDomain)
	page_url := r.Header.Get("referer")
	filter := CommentsFilter{
		ClientID: &c.ID,
		DomainID: &cd.ID,
		PageURL:  &page_url,
	}

	comments, err := s.CommentStorage.ListComments(filter)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	total := len(comments)
	resp := ListCommentsResponse{
		Total:    total,
		Comments: comments,
	}
	j, err := json.Marshal(resp)
	if err != nil {
		// notest
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)

}

func (s ParlanteServer) checkClient(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uuid := r.PathValue("uuid")
		c, err := s.ClientStorage.GetClientByUUID(uuid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		zeroClient := Client{}
		if c == zeroClient {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		origin := r.Header.Get("Origin")
		parts := strings.Split(origin, "://")
		if len(parts) != 2 {
			http.Error(w, "Forbibben", http.StatusForbidden)
			return
		}
		noscheme := parts[1]
		domain_port := strings.Split(noscheme, "/")[0]
		domain := strings.Split(domain_port, ":")[0]

		cd, err := s.ClientDomainStorage.GetClientDomain(c, domain)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		zeroDomain := ClientDomain{}
		if cd == zeroDomain {
			http.Error(w, "Forbidden", http.StatusForbidden)
		}
		ctx := context.WithValue(r.Context(), ctxClientKey, c)
		ctx = context.WithValue(ctx, ctxDomainKey, cd)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s ParlanteServer) setupUrls() {
	s.Server.Handle("POST /comment/{uuid}",
		s.checkClient(http.HandlerFunc(s.CreateComment)))

	s.Server.Handle("GET /comment/{uuid}",
		s.checkClient(http.HandlerFunc(s.ListComments)))

}

func NewServer() ParlanteServer {
	s := ParlanteServer{}
	s.Server = http.NewServeMux()
	s.BodyReader = io.ReadAll
	s.setupUrls()
	return s
}
