// Copyright 2025 Juca Crispim <juca@poraodojuca.net>

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
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "modernc.org/sqlite"
)

var DB *sql.DB

const DEFAULT_DB_PATH = "/var/local/parlante.sqlite"

//go:embed migrations/*.sql
var embeddedMigrations embed.FS

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

func (s ClientStorageSQLite) ListClients() ([]Client, error) {
	raw_query := "select * from clients"
	rows, err := DB.Query(raw_query)
	if err != nil {
		return nil, err
	}
	clients := make([]Client, 0)

	for rows.Next() {
		client := Client{}
		err := rows.Scan(&client.ID, &client.Name, &client.UUID, &client.Key)

		if err != nil {
			return nil, err
		}
		clients = append(clients, client)
	}
	return clients, nil

}

func (s ClientStorageSQLite) RemoveClient(uuid string) error {
	raw_query := "delete from clients where uuid = ?"
	_, err := DB.Exec(raw_query, uuid)
	return err
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

func (s ClientDomainStorageSQLite) RemoveClientDomain(c Client, domain string) error {
	raw_query := "delete from client_domains where domain = ? "
	raw_query += "and client_id = ?"
	_, err := DB.Exec(raw_query, domain, c.ID)
	return err
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

func (s ClientDomainStorageSQLite) ListDomains() ([]ClientDomain, error) {
	raw_query := `
select
  cd.id, cd.client_id, cd.domain,
  c.id, c.name, c.uuid, c.key

from
  client_domains cd

join
  clients c on c.id = cd.client_id
`

	rows, err := DB.Query(raw_query)
	if err != nil {
		return nil, err
	}
	domains := make([]ClientDomain, 0)
	for rows.Next() {
		c := Client{}
		cd := ClientDomain{}

		err := rows.Scan(
			&cd.ID,
			&cd.ClientID,
			&cd.Domain,
			&c.ID,
			&c.Name,
			&c.UUID,
			&c.Key,
		)

		if err != nil {
			return nil, err
		}
		cd.Client = &c
		domains = append(domains, cd)
	}
	return domains, nil
}

type CommentStorageSQLite struct {
}

func (s CommentStorageSQLite) CreateComment(
	c Client, d ClientDomain,
	name string,
	content string,
	page_url string) (Comment, error) {

	raw_query := `
insert into comments (client_id, domain_id, name, content, page_url, timestamp)
values (?, ?, ?, ?, ?, ?)`

	comment, err := NewComment(c, d, name, content, page_url)
	if err != nil {
		return Comment{}, err
	}
	row, err := DB.Exec(raw_query, c.ID, d.ID, comment.Author,
		comment.Content, comment.PageURL, comment.Timestamp)
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
	raw_query += " order by timestamp asc"
	rows, err := DB.Query(raw_query, args...)
	if err != nil {
		return nil, err
	}
	comments := make([]Comment, 0)

	for rows.Next() {
		comment := Comment{}
		err := rows.Scan(&comment.ID, &comment.ClientID, &comment.DomainID,
			&comment.Author, &comment.Content, &comment.PageURL, &comment.Hidden,
			&comment.Timestamp)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	return comments, nil
}

func (s CommentStorageSQLite) CountComments(urls ...string) ([]CommentCount, error) {
	if len(urls) == 0 {
		return nil, errors.New("At least one url is required")
	}
	in := make([]string, 0)
	anyurls := make([]any, 0)
	for _, url := range urls {
		in = append(in, "(?)")
		anyurls = append(anyurls, url)
	}
	instr := strings.Join(in, ",")
	raw_query := fmt.Sprintf(`
with urls(url) as (
    values %s
)
select
    u.url,
    coalesce(count(c.id), 0) as total_comments
from urls u
left join comments c
       on c.page_url = u.url
group by u.url
order by u.url;
`,
		instr)
	rows, err := DB.Query(raw_query, anyurls...)
	if err != nil {
		return nil, err
	}
	count := make([]CommentCount, 0)
	for rows.Next() {
		comment_count := CommentCount{}
		err := rows.Scan(&comment_count.PageURL, &comment_count.Count)
		if err != nil {
			return nil, err
		}
		count = append(count, comment_count)
	}
	return count, nil
}

func (s CommentStorageSQLite) RemoveComment(comment Comment) error {
	raw_query := "delete from comments where id = ? "
	_, err := DB.Exec(raw_query, comment.ID)
	return err

}

func SetupDB(connURI string) error {
	db, err := sql.Open("sqlite", connURI)
	if err != nil {
		return err
	}
	DB = db
	return nil
}

func MigrateDB(dbfile string) error {

	connURI := "sqlite3://" + dbfile
	d, err := iofs.New(embeddedMigrations, "migrations")
	if err != nil {
		return err
	}

	migr, err := migrate.NewWithSourceInstance("iofs", d, connURI)
	if err != nil {
		return err
	}
	err = migr.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}
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
