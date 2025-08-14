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
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

// A client that is allowed to use parlante.
type Client struct {
	ID   int64
	Name string
	UUID string
	// The key is used to authenticate the client. It is always stored
	// as a hashed value.
	Key string
}

// UpdateKey creates a new key to the client. Returns the plain text
// version of the key
func (c *Client) UpdateKey() (string, error) {
	key, err := GenKey()
	if err != nil {
		return "", err
	}
	hashed, err := HashStr(key)
	if err != nil {
		return "", err
	}
	c.Key = hashed
	return key, nil
}

// NewClient instantiate a new client, generating
// an uuid and a key for the client. Returns the
// plain text version of the key
func NewClient(name string) (Client, string, error) {
	uuid, err := GenUUID4()
	if err != nil {
		return Client{}, "", err
	}

	key, err := GenKey()
	if err != nil {
		return Client{}, "", err
	}
	encr, err := HashStr(key)
	if err != nil {
		return Client{}, "", err
	}
	c := Client{
		Name: name,
		UUID: uuid,
		Key:  encr,
	}

	return c, key, nil
}

// ClientStorage is an interface to save/retrieve information
// about clients.
type ClientStorage interface {
	CreateClient(name string) (Client, string, error)
	GetClientByUUID(uuid string) (Client, error)
	ListClients() ([]Client, error)
	RemoveClient(uuid string) error
}

// ClientDomain is a domain allowed by a client to have comments
type ClientDomain struct {
	ID       int64
	ClientID int64
	Domain   string
	Client   *Client
}

// NewClientDomain instantiate a new  ClientDomain
func NewClientDomain(c Client, domain string) ClientDomain {
	d := ClientDomain{
		ClientID: c.ID,
		Domain:   domain,
		Client:   &c,
	}
	return d
}

// ClientDomainStorage is an interface to save/retrieve information
// about client domains
type ClientDomainStorage interface {
	AddClientDomain(c Client, domain string) (ClientDomain, error)
	RemoveClientDomain(c Client, domain string) error
	GetClientDomain(c Client, domain string) (ClientDomain, error)
	ListDomains() ([]ClientDomain, error)
}

// CommentsFilter contains the fields used to filter a query for
// comments
type CommentsFilter struct {
	ClientID *int64
	DomainID *int64
	PageURL  *string
	Hidden   *bool
}

// Comment is a comment made by an user in a web page.
type Comment struct {
	ID       int64
	ClientID int64
	DomainID int64
	Author   string
	Content  string
	PageURL  string
	Hidden   bool
	Client   *Client
	Domain   *ClientDomain
	// unix timestamp for comment creating. It must be in UTC timezone
	Timestamp int64
}

// CommentCount has the count of comments made in a web page.
type CommentCount struct {
	PageURL string `json:"page_url"`
	Count   int64  `json:"count"`
}

// NewComment returns a new instance of Comment. Checks for blank strings
// for author, content e page_url. If missing returns an error.
func NewComment(c Client, d ClientDomain, author string, content string,
	page_url string) (Comment, error) {
	if author == "" || content == "" || page_url == "" {
		return Comment{}, errors.New("Missing required field")
	}
	comment := Comment{
		ClientID:  c.ID,
		DomainID:  d.ID,
		Author:    author,
		Content:   content,
		PageURL:   page_url,
		Client:    &c,
		Domain:    &d,
		Timestamp: time.Now().Unix(),
	}
	return comment, nil
}

// CommentStorage is an interface to save/retrive information
// from comments
type CommentStorage interface {
	CreateComment(
		c Client,
		d ClientDomain,
		name string,
		content string,
		page_url string) (Comment, error)

	ListComments(filter CommentsFilter) ([]Comment, error)
	RemoveComment(comment Comment) error
	CountComments(urls ...string) ([]CommentCount, error)
}

// EmailMessage represents an email to be sent. Note that as this have
// no content type and the body is a string, only text/plain bodies are
// supported.
type EmailMessage struct {
	From  string
	To    []string
	Title string
	Body  string
}

// GenUUID4 returns a new uuid v4 converted to string
func GenUUID4() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	b[6] = (b[6] | 0x40) & 0x4F
	b[8] = (b[8] | 0x80) & 0xBF
	uuid := fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	uuid = strings.ToLower(uuid)
	return uuid, nil
}

// GenKey returns a random 32 chars string.
func GenKey() (string, error) {
	letters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	klen := 32
	b := make([]byte, klen)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	for i := range b {
		b[i] = letters[b[i]%62]
	}
	return string(b), nil
}

// HashStr creates a new sha512 hash from a string
func HashStr(s string) (string, error) {
	hash := sha512.New()
	_, err := hash.Write([]byte(s))
	if err != nil {
		return "", err
	}
	summed := hash.Sum(nil)
	encoded := hex.EncodeToString(summed)
	return encoded, nil
}
