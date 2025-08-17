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
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type ctxKey string

const ctxClientKey ctxKey = "client"
const ctxDomainKey ctxKey = "domain"

const emailAddr = "blog@pdj01.poraodojuca.dev"

//go:embed js/parlante.js
var parlanteJS []byte

type bodyReader func(io.Reader) ([]byte, error)
type jsonMarshaler func(v any) ([]byte, error)
type htmlRenderer func(s string, lang string, tz string, d map[string]any) ([]byte, error)
type loggerFn func(format string, v ...any)

// RequestLogger logs a request made to the parlante server
type RequestLogger struct {
	loggerFn loggerFn
}

// Log logs the ip, method, path, status and user agent
func (l RequestLogger) Log(h http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, req *http.Request) {
		sw := &StatusedResponseWriter{w, http.StatusOK}
		h.ServeHTTP(sw, req)
		remote := l.getIp(req)
		path := req.URL.Path
		method := req.Method
		ua := req.Header.Get("User-Agent")
		l.loggerFn("%s %s %s %d %s\n", remote, method, path, sw.Status, ua)
	}
	return http.HandlerFunc(handler)
}

func (l RequestLogger) getIp(req *http.Request) string {
	ip := req.Header.Get("X-Real-Ip")
	if ip == "" {
		ip = req.Header.Get("X-Forwarded-For")
	}
	if ip == "" {
		ip = req.RemoteAddr
	}
	return ip
}

// StatusedResponseWriter is a reponse writer that knows the
// return status of the request
type StatusedResponseWriter struct {
	http.ResponseWriter
	Status int
}

// WriteHeader writes the Status header in the response.
func (w *StatusedResponseWriter) WriteHeader(code int) {
	w.Status = code
	w.ResponseWriter.WriteHeader(code)
}

// CreateCommentRequest is the structure of a json sent in the
// body of a request to create a new comment
type CreateCommentRequest struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// MsgResponse is a response with a single msg string field
type MsgResponse struct {
	Msg string `json:"msg"`
}

// CommentResponse is the information about a comment returned in the
// comments list json
type CommentResponse struct {
	Author    string `json:"author"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
}

type ListCommentsResponse struct {
	Total    int               `json:"total"`
	Comments []CommentResponse `json:"comments"`
}

type CountCommentsRequest struct {
	PageURLs []string `json:"page_urls"`
}

type CountCommentsHTMLRequest struct {
	PageURLs       []string `json:"page_urls"`
	CommentsAnchor string   `json:"comments_anchor"`
}

type CountCommentsResponse struct {
	Total        int            `json:"total"`
	CommentCount []CommentCount `json:"comment_count"`
}

type CountCommentsHTMLItem struct {
	PageURL string `json:"page_url"`
	Content string `json:"content"`
}

type CountCommentsHTMLResponse struct {
	Total int                     `json:"total"`
	Items []CountCommentsHTMLItem `json:"items"`
}

// PingMeRequest represents the json structure sent in the body of the request
// to send a new pingme message
type PingMeRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

// Config holds the configuration values used by the parlante server
type Config struct {
	Port         int
	Host         string
	CertFilePath string
	KeyFilePath  string
	DBPath       string
	MaildirPath  string
	LogLevel     string
}

func (c Config) UsesSSL() bool {
	return c.CertFilePath != "" && c.KeyFilePath != ""
}

// ParlanteServer is the server for the parlante http api
type ParlanteServer struct {
	ClientStorage       ClientStorage
	ClientDomainStorage ClientDomainStorage
	CommentStorage      CommentStorage
	EmailSender         EmailSender
	mux                 *http.ServeMux
	BodyReader          bodyReader
	JsonMarshaler       jsonMarshaler
	HtmlRenderer        htmlRenderer
	Config              Config
}

// CreateComment creates a new comment for a given web page. The page
// is the url in the `X-PageURL` header.
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

	page_url := r.Header.Get("X-PageURL")

	_, err = s.CommentStorage.CreateComment(
		c, cd, body.Name, body.Content, page_url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	loc := GetDefaultLocale()
	go func() {
		data := make(map[string]any)
		data["name"] = body.Name
		data["domain"] = cd.Domain
		subject := Tprintf(loc.Get("New comment from {{.name}} at {{.domain}}"), data)
		mailBody := fmt.Sprintf("url: %s\n\n%s", page_url, body.Content)
		err := s.sendEmail(subject, mailBody)
		if err != nil {
			Errorf("error sending email %s", err.Error())
		}

	}()
	resp := MsgResponse{Msg: "Ok"}
	j, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.WriteHeader(http.StatusCreated)
	w.Write(j)
}

// ListComments returns a json with the commments for a given page. The page
// is the URL in the `X-PageURL` header
func (s ParlanteServer) ListComments(w http.ResponseWriter, r *http.Request) {
	c := r.Context().Value(ctxClientKey).(Client)
	cd := r.Context().Value(ctxDomainKey).(ClientDomain)
	page_url := r.Header.Get("X-PageURL")
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
	cresp := make([]CommentResponse, 0)
	for _, c := range comments {
		resp := CommentResponse{
			Author:    c.Author,
			Content:   c.Content,
			Timestamp: c.Timestamp,
		}
		cresp = append(cresp, resp)
	}
	resp := ListCommentsResponse{
		Total:    total,
		Comments: cresp,
	}
	j, err := s.JsonMarshaler(resp)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.WriteHeader(http.StatusOK)
	w.Write(j)
}

// ListCommentsHTML returns a html snipet to be included in a web page.
// User the `Accepted-Language` header for translations and the `X-Timezone`
// header to set the comments display timezone.
func (s ParlanteServer) ListCommentsHTML(w http.ResponseWriter, r *http.Request) {

	c := r.Context().Value(ctxClientKey).(Client)
	cd := r.Context().Value(ctxDomainKey).(ClientDomain)

	page_url := r.Header.Get("X-PageURL")
	lang := getRequestLanguage(r)
	tz := r.Header.Get("X-Timezone")

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

	loc := GetLocale(lang)
	tmplCtx := make(map[string]any)
	total := len(comments)

	header := loc.Get("Comments (%d)", total)
	tmplCtx["header"] = header
	tmplCtx["addCommentHeader"] = loc.Get("Leave your comment!")
	tmplCtx["noComments"] = loc.Get("No comments.")
	tmplCtx["comments"] = comments
	tmplCtx["nameLabel"] = loc.Get("Name")
	tmplCtx["commentLabel"] = loc.Get("Comment")
	tmplCtx["submitComment"] = loc.Get("Send comment")
	tmplCtx["commentAddOkMsg"] = loc.Get("Comment sent. Thank you!")
	tmplCtx["commentAddErrorMsg"] = loc.Get("Error sending comment.")

	b, err := s.HtmlRenderer("comments.html", lang, tz, tmplCtx)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

// CountComments returns a json with the comments count for each url passed
// in the request
func (s ParlanteServer) CountComments(w http.ResponseWriter, r *http.Request) {
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

	var body CountCommentsRequest
	err = json.Unmarshal(rawbody, &body)
	if err != nil {
		http.Error(w, "Malformed json", http.StatusBadRequest)
		return
	}

	cd := r.Context().Value(ctxDomainKey).(ClientDomain)
	valid := getValidURLsForDomain(cd, body.PageURLs)

	count, err := s.CommentStorage.CountComments(valid...)
	resp := CountCommentsResponse{
		Total:        len(count),
		CommentCount: count,
	}
	j, err := s.JsonMarshaler(resp)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.WriteHeader(http.StatusOK)
	w.Write(j)

}

// CountCommentsHTML returns a json with a html snipet as the value for
// the url count.
func (s ParlanteServer) CountCommentsHTML(w http.ResponseWriter, r *http.Request) {
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

	var body CountCommentsHTMLRequest
	err = json.Unmarshal(rawbody, &body)
	if err != nil {
		http.Error(w, "Malformed json", http.StatusBadRequest)
		return
	}

	cd := r.Context().Value(ctxDomainKey).(ClientDomain)
	valid := getValidURLsForDomain(cd, body.PageURLs)
	count, err := s.CommentStorage.CountComments(valid...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	lang := getRequestLanguage(r)
	tz := ""
	loc := GetLocale(lang)

	count_items := make([]CountCommentsHTMLItem, 0)
	resp := CountCommentsHTMLResponse{}
	for _, item := range count {
		tmplCtx := make(map[string]any)
		header := loc.Get("Comments (%d)", item.Count)
		tmplCtx["header"] = header
		tmplCtx["commentsURL"] = item.PageURL + "#" + body.CommentsAnchor
		tmplCtx["addCommentHeader"] = loc.Get("Leave your comment!")
		content, err := s.HtmlRenderer("comments_count.html", lang, tz, tmplCtx)
		if err != nil {
			Errorf(err.Error())
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		count_items = append(count_items, CountCommentsHTMLItem{
			PageURL: item.PageURL, Content: string(content)})
	}
	resp.Items = count_items
	resp.Total = len(count_items)

	j, err := s.JsonMarshaler(resp)
	if err != nil {
		// notest
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.WriteHeader(http.StatusOK)
	w.Write(j)
}

// GetPingMeForm returns a html snippet to be used as a contact form
// in allowed domains
func (s ParlanteServer) GetPingMeForm(w http.ResponseWriter, r *http.Request) {
	lang := getRequestLanguage(r)
	tz := ""
	loc := GetLocale(lang)
	tmplCtx := make(map[string]any)
	tmplCtx["nameLabel"] = loc.Get("Name")
	tmplCtx["messageLabel"] = loc.Get("Your message")
	tmplCtx["submitMessage"] = loc.Get("Send message")
	tmplCtx["pingMeAddOkMsg"] = loc.Get("Message sent. Thank you!")
	tmplCtx["pingMeAddErrorMsg"] = loc.Get("Error sending message.")
	b, err := s.HtmlRenderer("pingme.html", lang, tz, tmplCtx)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

// PingMe send a message to a maildir
func (s ParlanteServer) PingMe(w http.ResponseWriter, r *http.Request) {
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

	var body PingMeRequest
	err = json.Unmarshal(rawbody, &body)
	if err != nil {
		http.Error(w, "Malformed json", http.StatusBadRequest)
		return
	}

	if body.Name == "" || body.Email == "" || body.Message == "" {
		http.Error(w, "Missing required params", http.StatusBadRequest)
		return
	}

	cd := r.Context().Value(ctxDomainKey).(ClientDomain)
	loc := GetDefaultLocale()
	data := make(map[string]any)
	data["name"] = body.Name
	data["domain"] = cd.Domain
	subject := Tprintf(loc.Get("New message from {{.name}} at {{.domain}}"), data)
	mailBody := fmt.Sprintf("email: %s\n\n%s", body.Email, body.Message)

	err = s.sendEmail(subject, mailBody)
	if err != nil {
		Errorf(err.Error())
		http.Error(w, "Error sending message", http.StatusInternalServerError)
		return
	}
	resp := MsgResponse{Msg: "Ok"}
	j, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.WriteHeader(http.StatusCreated)
	w.Write(j)
}

// ServeParlanteJS returns the parlante.js file that is used to render the
// comments in a web page.
func (s ParlanteServer) ServeParlanteJS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	w.Write(parlanteJS)
}

// Run starts the parlente sever
func (s ParlanteServer) Run() {
	// notest
	addr := fmt.Sprintf("%s:%d", s.Config.Host, s.Config.Port)
	var err error
	logger := RequestLogger{loggerFn: Infof}
	loggedMux := logger.Log(s.mux)
	if s.Config.UsesSSL() {
		err = http.ListenAndServeTLS(addr, s.Config.CertFilePath,
			s.Config.KeyFilePath, loggedMux)
	} else {
		err = http.ListenAndServe(addr, loggedMux)

	}
	if err != nil {
		panic(err.Error())
	}
}

// NewServer returns a new instance of ParlanteServer. Only one per process
// must be used
func NewServer(c Config) ParlanteServer {
	s := ParlanteServer{}
	s.mux = http.NewServeMux()
	s.BodyReader = io.ReadAll
	s.JsonMarshaler = json.Marshal
	s.HtmlRenderer = RenderTemplate
	s.Config = c
	s.ClientStorage = ClientStorageSQLite{}
	s.ClientDomainStorage = ClientDomainStorageSQLite{}
	s.CommentStorage = CommentStorageSQLite{}
	sender := NewMaildirSender(s.Config.MaildirPath)
	s.EmailSender = sender
	SetLogLevelStr(c.LogLevel)
	s.setupUrls()
	return s
}

// checkClient checks if the client exists and the request origin
// is a registered domain
func (s ParlanteServer) checkClient(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uuid := r.PathValue("uuid")
		uuid = strings.ToLower(uuid)
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
		domain, err := getDomainFromURL(origin)
		if err != nil {
			http.Error(w, "Forbibben", http.StatusForbidden)
			return
		}

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

func (s ParlanteServer) sendEmail(subject string, body string) error {
	msg, err := NewEmailMessage(emailAddr, []string{emailAddr}, subject, body)
	if err != nil {
		return err
	}

	return s.EmailSender.SendEmail(msg)
}

func getDomainFromURL(url string) (string, error) {
	parts := strings.Split(url, "://")
	if len(parts) != 2 {
		return "", errors.New("bad url")
	}
	noscheme := parts[1]
	domain_port := strings.Split(noscheme, "/")[0]
	domain := strings.Split(domain_port, ":")[0]
	return domain, nil
}

func getValidURLsForDomain(d ClientDomain, urls []string) []string {
	valid := make([]string, 0)

	for _, url := range urls {
		domain, err := getDomainFromURL(url)
		if domain == d.Domain || err != nil {
			valid = append(valid, url)
		}
	}
	return valid
}

func (s ParlanteServer) setupUrls() {
	s.mux.Handle("POST /comment/{uuid}",
		s.checkClient(http.HandlerFunc(s.CreateComment)))

	s.mux.Handle("GET /comment/{uuid}",
		s.checkClient(http.HandlerFunc(s.ListComments)))
	s.mux.Handle("OPTIONS /comment/{uuid}", http.HandlerFunc(handleCORS))

	s.mux.Handle("GET /comment/{uuid}/html",
		s.checkClient(http.HandlerFunc(s.ListCommentsHTML)))
	s.mux.Handle("OPTIONS /comment/{uuid}/html", http.HandlerFunc(handleCORS))

	s.mux.Handle("GET /parlante.js", http.HandlerFunc(s.ServeParlanteJS))

	s.mux.Handle("POST /comment/{uuid}/count",
		s.checkClient(http.HandlerFunc(s.CountComments)))
	s.mux.Handle("OPTIONS /comment/{uuid}/count",
		s.checkClient(http.HandlerFunc(handleCORS)))

	s.mux.Handle("POST /comment/{uuid}/count/html",
		s.checkClient(http.HandlerFunc(s.CountCommentsHTML)))
	s.mux.Handle("OPTIONS /comment/{uuid}/count/html",
		s.checkClient(http.HandlerFunc(handleCORS)))

	s.mux.Handle("GET /pingme/{uuid}",
		s.checkClient(http.HandlerFunc(s.GetPingMeForm)))
	s.mux.Handle("POST /pingme/{uuid}",
		s.checkClient(http.HandlerFunc(s.PingMe)))
	s.mux.Handle("OPTIONS /pingme/{uuid}",
		s.checkClient(http.HandlerFunc(handleCORS)))

}

func handleCORS(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	w.Header().Set("Access-Control-Allow-Origin", origin)

	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	h := "Content-Type, Authorization, Accepted-Language, X-Timezone, X-PageURL"
	w.Header().Set("Access-Control-Allow-Headers", h)
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.WriteHeader(http.StatusNoContent)
}

func getRequestLanguage(r *http.Request) string {
	lang := r.Header.Get("Accepted-Language")
	if lang == "" {
		return ""
	}
	preferred := strings.Split(lang, ",")[0]
	return strings.ReplaceAll(preferred, "-", "_")
}
