package parlante

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type ctxKey string

const ctxClientKey ctxKey = "client"
const ctxDomainKey ctxKey = "domain"

type bodyReader func(io.Reader) ([]byte, error)
type jsonMarshaler func(v any) ([]byte, error)

type StatusedResponseWriter struct {
	http.ResponseWriter
	Status int
}

func (w *StatusedResponseWriter) WriteHeader(code int) {
	w.Status = code
	w.ResponseWriter.WriteHeader(code)
}

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

type Config struct {
	Port         int
	Host         string
	CertFilePath string
	KeyFilePath  string
	DBPath       string
	LogLevel     string
}

func (c Config) UsesSSL() bool {
	return c.CertFilePath != "" && c.KeyFilePath != ""
}

type ParlanteServer struct {
	ClientStorage       ClientStorage
	ClientDomainStorage ClientDomainStorage
	CommentStorage      CommentStorage
	mux                 *http.ServeMux
	BodyReader          bodyReader
	JsonMarshaler       jsonMarshaler
	Config              Config
}

func (s ParlanteServer) CreateComment(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		http.Error(w, "Missing body", http.StatusBadRequest)
		return
	}
	rawbody, err := s.BodyReader(r.Body)
	if err != nil {
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
	j, err := s.JsonMarshaler(resp)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}

func (s ParlanteServer) Run() {
	// notest
	addr := fmt.Sprintf("%s:%d", s.Config.Host, s.Config.Port)
	var err error
	if s.Config.UsesSSL() {
		err = http.ListenAndServeTLS(addr, s.Config.CertFilePath,
			s.Config.KeyFilePath, s.mux)
	} else {
		err = http.ListenAndServe(addr, s.mux)

	}
	if err != nil {
		panic(err.Error())
	}
}

func (s ParlanteServer) checkClient(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uuid := r.PathValue("uuid")
		c, err := s.ClientStorage.GetClientByUUID(uuid)
		if err != nil {
			Errorf(err.Error())
			http.Error(w, "Forbidden", http.StatusForbidden)
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
	s.mux.Handle("POST /comment/{uuid}",
		logRequest(s.checkClient(http.HandlerFunc(s.CreateComment))))

	s.mux.Handle("GET /comment/{uuid}",
		logRequest(s.checkClient(http.HandlerFunc(s.ListComments))))

}

func logRequest(h http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, req *http.Request) {
		sw := &StatusedResponseWriter{w, http.StatusOK}
		h.ServeHTTP(sw, req)
		remote := getIp(req)
		path := req.URL.Path
		method := req.Method
		ua := req.Header.Get("User-Agent")
		Infof("%s %s %s %d %s\n", remote, method, path, sw.Status, ua)
	}
	return http.HandlerFunc(handler)
}

func getIp(req *http.Request) string {
	ip := req.Header.Get("X-Real-Ip")
	if ip == "" {
		ip = req.Header.Get("X-Forwarded-For")
	}
	if ip == "" {
		ip = req.RemoteAddr
	}
	return ip
}

func NewServer(c Config) ParlanteServer {
	s := ParlanteServer{}
	s.mux = http.NewServeMux()
	s.BodyReader = io.ReadAll
	s.JsonMarshaler = json.Marshal
	s.Config = c
	s.ClientStorage = ClientStorageSQLite{}
	s.ClientDomainStorage = ClientDomainStorageSQLite{}
	s.CommentStorage = CommentStorageSQLite{}
	SetLogLevelStr(c.LogLevel)
	s.setupUrls()
	return s
}
