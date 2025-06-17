package parlante

import (
	"database/sql"
	"reflect"
	"strings"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

type ClientStorageSQLite struct {
}

func (s ClientStorageSQLite) CreateClient(name string) (Client, string, error) {
	c, plain_text, err := NewClient(name)
	if err != nil {
		return Client{}, "", err
	}
	err = insertClient(&c)
	if err != nil {
		return Client{}, "", err
	}
	return c, plain_text, nil
}

func (s ClientStorageSQLite) GetClientByUUID(uuid string) (Client, error) {
	raw_query := "select * from clients where uuid = ?"
	row := DB.QueryRow(raw_query, uuid)
	client := Client{}
	err := row.Scan(&client.ID, &client.Name, &client.UUID, &client.Key)
	if err != nil {
		return Client{}, err
	}
	return client, nil
}

type ClientDomainStorageSQLite struct {
}

func (s ClientDomainStorageSQLite) AddClientDomain(c Client, domain string) (
	ClientDomain, error) {
	raw_query := "insert into client_domains (client_id, domain) values (?, ?)"
	d := NewClientDomain(c, domain)
	row, err := DB.Exec(raw_query, c.ID, d.Domain)
	if err != nil {
		return ClientDomain{}, err
	}
	id, err := row.LastInsertId()
	if err != nil {
		return ClientDomain{}, err
	}
	d.ID = id
	return d, nil
}

func (s ClientDomainStorageSQLite) RemoveClientDomain(c Client,
	domain string) error {
	raw_query := "delete from client_domains where domain = ? "
	raw_query += "and client_id = ?"
	_, err := DB.Exec(raw_query, domain, c.ID)
	if err != nil {
		return err
	}
	return nil
}

func (s ClientDomainStorageSQLite) GetClientDomain(c Client, domain string) (
	ClientDomain, error) {
	raw_query := "select * from client_domains where client_id = ? "
	raw_query += "and domain = ?"
	row := DB.QueryRow(raw_query, c.ID, domain)
	d := ClientDomain{}
	err := row.Scan(&d.ID, &d.ClientID, &d.Domain)
	if err != nil {
		return ClientDomain{}, nil
	}
	return d, nil
}

type CommentStorageSQLite struct {
}

func (s CommentStorageSQLite) CreateComment(c Client, d ClientDomain,
	name string, content string, page_url string) (Comment, error) {
	raw_query := "insert into comments (client_id, domain_id, name, content, page_url) "
	raw_query += " values (?, ?, ?, ?, ?)"

	comment := NewComment(c, d, name, content, page_url)
	row, err := DB.Exec(raw_query, c.ID, d.ID, comment.Name,
		comment.Content, comment.PageURL)
	if err != nil {
		return Comment{}, err
	}
	id, err := row.LastInsertId()
	if err != nil {
		return Comment{}, err
	}
	comment.ID = id
	return comment, nil

}

func (s CommentStorageSQLite) ListComments(filter CommentsFilter) (
	[]Comment, error) {

	where, args := []string{"1 = 1"}, []any{}
	tb := make(map[string]any)

	tb["client_id = ?"] = filter.ClientID
	tb["domain_id = ?"] = filter.DomainID
	tb["page_url = ?"] = filter.PageURL
	tb["hidden = ?"] = filter.Hidden

	for k, v := range tb {
		if !reflect.ValueOf(v).IsNil() {
			where, args = append(where, k), append(args, v)
		}
	}

	raw_query := "select * from comments where " + strings.Join(where, " and ")
	rows, err := DB.Query(raw_query, args...)
	if err != nil {
		return nil, err
	}
	comments := make([]Comment, 0)

	for rows.Next() {
		comment := Comment{}
		err := rows.Scan(&comment.ID, &comment.ClientID, &comment.DomainID,
			&comment.Name, &comment.Content, &comment.PageURL, &comment.Hidden)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	return comments, nil
}

func SetupDB(connUri string) error {
	db, err := sql.Open("sqlite", connUri)
	if err != nil {
		return err
	}
	DB = db
	return nil
}

func insertClient(client *Client) error {
	raw_query := `insert into clients (name, uuid, key) values (?, ?, ?)`
	stmt, err := DB.Prepare(raw_query)
	if err != nil {
		return err
	}
	res, err := stmt.Exec(client.Name, client.UUID, client.Key)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()

	if err != nil {
		return err
	}

	client.ID = id
	return nil
}
