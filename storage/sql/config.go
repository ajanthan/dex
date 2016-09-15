package sql

import (
	"database/sql"
	"fmt"
	"time"
)

// SQLite options for creating an SQL db.
type SQLite struct {
	// File to
	File string `yaml:"file"`
}

func (s *SQLite) open() (*conn, error) {
	db, err := sql.Open("sqlite3", s.File)
	if err != nil {
		return nil, err
	}
	if s.File == ":memory:" {
		// sqlite3 uses file locks to coordinate concurrent access. In memory
		// doesn't support this, so limit the number of connections to 1.
		db.SetMaxOpenConns(1)
	}
	c := &conn{db, flavorSQLite3}
	if _, err := c.migrate(); err != nil {
		return nil, fmt.Errorf("failed to perform migrations: %v", err)
	}
	return c, nil
}

// Postgres options for creating an SQL db.
type Postgres struct {
	Database string
	User     string
	Password string
	Host     string

	SSLCAFile string

	SSLKeyFile  string
	SSLCertFile string

	ConnectionTimeout time.Duration
}
