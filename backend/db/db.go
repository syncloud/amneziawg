// Package db owns the SQLite schema and queries for peer state and
// persisted settings. Uses modernc.org/sqlite (pure Go) so the whole
// backend compiles with CGO_ENABLED=0.
package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
}

func Open(path string) (*DB, error) {
	raw, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if err := raw.Ping(); err != nil {
		return nil, err
	}
	db := &DB{raw}
	if err := db.migrate(); err != nil {
		return nil, err
	}
	return db, nil
}

func (db *DB) migrate() error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS peers (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			name        TEXT NOT NULL UNIQUE,
			public_key  TEXT NOT NULL UNIQUE,
			private_key TEXT,
			address_v4  TEXT NOT NULL UNIQUE,
			created_at  TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS settings (
			key   TEXT PRIMARY KEY,
			value TEXT NOT NULL
		);
	`)
	return err
}

type Peer struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"-"`
	AddressV4  string `json:"address_v4"`
	CreatedAt  string `json:"created_at"`
}

func (db *DB) ListPeers() ([]Peer, error) {
	rows, err := db.Query(`SELECT id, name, public_key, private_key, address_v4, created_at FROM peers ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Peer
	for rows.Next() {
		var p Peer
		var priv sql.NullString
		if err := rows.Scan(&p.ID, &p.Name, &p.PublicKey, &priv, &p.AddressV4, &p.CreatedAt); err != nil {
			return nil, err
		}
		if priv.Valid {
			p.PrivateKey = priv.String
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (db *DB) InsertPeer(p Peer) (int64, error) {
	res, err := db.Exec(
		`INSERT INTO peers (name, public_key, private_key, address_v4) VALUES (?, ?, ?, ?)`,
		p.Name, p.PublicKey, nullable(p.PrivateKey), p.AddressV4,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (db *DB) GetPeer(id int64) (Peer, error) {
	var p Peer
	var priv sql.NullString
	err := db.QueryRow(
		`SELECT id, name, public_key, private_key, address_v4, created_at FROM peers WHERE id = ?`, id,
	).Scan(&p.ID, &p.Name, &p.PublicKey, &priv, &p.AddressV4, &p.CreatedAt)
	if err != nil {
		return p, err
	}
	if priv.Valid {
		p.PrivateKey = priv.String
	}
	return p, nil
}

func (db *DB) DeletePeer(id int64) error {
	res, err := db.Exec(`DELETE FROM peers WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("peer %d not found", id)
	}
	return nil
}

// UsedAddresses returns the set of /32 addresses already assigned to a
// peer, so the service layer can pick the next free one.
func (db *DB) UsedAddresses() (map[string]bool, error) {
	rows, err := db.Query(`SELECT address_v4 FROM peers`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]bool)
	for rows.Next() {
		var addr string
		if err := rows.Scan(&addr); err != nil {
			return nil, err
		}
		out[addr] = true
	}
	return out, rows.Err()
}

func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}
