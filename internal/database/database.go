//   Copyright 2022 Alexander <sattellite> Groshev
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package database

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/rs/zerolog"

	_ "github.com/glebarez/go-sqlite" // pure go SQLite driver
	"github.com/jmoiron/sqlx"
)

type Client struct {
	db     *sqlx.DB
	logger *zerolog.Logger
}

// New connection to database
func New(dsn string, logger *zerolog.Logger) (*Client, error) {
	if dsn == "" {
		return nil, fmt.Errorf("dsn is empty")
	}
	if !strings.HasSuffix(dsn, ".sqlite3") {
		dsn += ".sqlite3"
	}

	l := logger.With().Str("component", "database").Logger()
	client := &Client{
		logger: &l,
	}

	err := client.prepareDBFile(dsn)
	if err != nil {
		return nil, err
	}

	db, err := sqlx.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	client.db = db
	return client, nil
}

func (c *Client) prepareDBFile(filename string) error {
	// Проверка существования файла БД
	fi, err := os.Stat(filename)
	if err != nil || fi.Size() == 0 {
		// Создание файла БД
		f, err := os.Create(filename)
		if err != nil {
			return fmt.Errorf("could not create database file: %w", err)
		}
		defer func() {
			cerr := f.Close()
			if cerr != nil {
				log.Printf("could not close database file: %v", cerr)
			}
		}()
	}
	return nil
}

func (c *Client) prepareSchema() error {
	version, err := c.getSchemaVersion()
	if err != nil {
		return fmt.Errorf("could not get schema version: %w", err)
	}

	c.logger.Info().Msgf("database current schema version: %d", version)

	queries := []string{
		"CREATE TABLE IF NOT EXISTS hearings(id INTEGER PRIMARY KEY, link TEXT DEFAULT '' NOT NULL UNIQUE, topics TEXT DEFAULT '', proposals TEXT DEFAULT '', place TEXT DEFAULT '', date TEXT DEFAULT '1970-01-01 00:00:00', published BOOLEAN DEFAULT false",
	}

	if version == len(queries) {
		c.logger.Info().Msg("database schema is up to date")
		return nil
	}

	c.logger.Info().Msg("upgrading database schema...")

	for i := version; i < len(queries); i++ {
		c.logger.Debug().Msgf("database executing query:", queries[i])
		_, err = c.db.Exec(queries[i])
		if err != nil {
			return fmt.Errorf("could not execute query: %w", err)
		}
		err = c.setSchemaVersion(i + 1)
		if err != nil {
			return fmt.Errorf("could not set schema version: %w", err)
		}
		c.logger.Debug().Msgf("database schema version:", i+1)
	}
	c.logger.Info().Msg("database schema is up to date")
	return nil
}

func (c *Client) getSchemaVersion() (int, error) {
	row := c.db.QueryRow("PRAGMA user_version")
	if row == nil {
		return 0, fmt.Errorf("PRAGMA user_version not found")
	}
	var version int
	if err := row.Scan(&version); err != nil {
		return 0, err
	}
	return version, nil
}

func (c *Client) setSchemaVersion(version int) error {
	_, err := c.db.Exec(fmt.Sprintf("PRAGMA user_version = %d", version))
	return err
}
