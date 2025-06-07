package main

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
)

type Client struct {
	ID   int64
	Name string
	UUID string
	Key  string
}

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

type ClientDomain struct {
	ID       int64
	ClientID int64
	Domain   string
}

func NewClientDomain(c Client, domain string) ClientDomain {
	d := ClientDomain{
		ClientID: c.ID,
		Domain:   domain,
	}

	return d
}

type Comment struct {
	ID       int64
	ClientID int64
	DomainID int64
	Name     string
	Content  string
	PageURL  string
	Hidden   bool
}

func NewComment(c Client, d ClientDomain, name string, content string,
	page_url string) Comment {
	comment := Comment{
		ClientID: c.ID,
		DomainID: d.ID,
		Name:     name,
		Content:  content,
		PageURL:  page_url,
	}
	return comment
}

func GenUUID4() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	b[6] = (b[6] | 0x40) & 0x4F
	b[8] = (b[8] | 0x80) & 0xBF
	uuid := fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return uuid, nil
}

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
