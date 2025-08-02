create table if not exists clients (
       id INTEGER PRIMARY KEY AUTOINCREMENT,
       name STRING unique not null,
       uuid string unique not null,
       key string not null
);

CREATE INDEX IF NOT EXISTS client_uuid_idx ON clients(uuid);
CREATE INDEX IF NOT EXISTS client_key_idx ON clients(key);

create table if not exists client_domains (
       id INTEGER PRIMARY KEY AUTOINCREMENT,
       client_id integer not null,
       domain string not null,
       FOREIGN KEY(client_id) REFERENCES clients(id),
       Unique(client_id, domain) on conflict fail
);

CREATE INDEX IF NOT EXISTS client_domain_client_idx ON client_domains(client_id);
CREATE INDEX IF NOT EXISTS client_domain_domain_idx ON client_domains(domain);

create table if not exists comments (
       id INTEGER PRIMARY KEY AUTOINCREMENT,
       client_id integer not null,
       domain_id integer not null,
       name STRING not null,
       content text not null,
       page_url string not null,
       hidden boolean not null default false,
       timestamp timestamp not null,
       FOREIGN KEY(domain_id) REFERENCES client_domains(id),
       FOREIGN KEY(client_id) REFERENCES clients(id)
);


CREATE INDEX IF NOT EXISTS comment_client_idx ON comments(client_id);
CREATE INDEX IF NOT EXISTS comment_domain_idx ON comments(domain_id);
CREATE INDEX IF NOT EXISTS comment_page_url_idx ON comments(page_url);
