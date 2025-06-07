create table if not exists clients (
       id INTEGER PRIMARY KEY AUTOINCREMENT,
       name STRING not null,
       uuid string unique not null,
       key string not null
);

create table if not exists client_domains (
       id INTEGER PRIMARY KEY AUTOINCREMENT,
       client_id integer not null,
       domain string not null,
       FOREIGN KEY(client_id) REFERENCES clients(id)
);

create table if not exists comments (
       id INTEGER PRIMARY KEY AUTOINCREMENT,
       client_id integer not null,
       domain_id integer not null,
       name STRING not null,
       content text not null,
       page_url string not null,
       hidden boolean not null default false,
       FOREIGN KEY(domain_id) REFERENCES client_domains(id),
       FOREIGN KEY(client_id) REFERENCES clients(id)
);
