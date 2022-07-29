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
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/brurbanko/mercury/domain"

	"github.com/rs/zerolog"

	_ "github.com/glebarez/go-sqlite" // pure go SQLite driver
	"github.com/jmoiron/sqlx"
)

const (
	sliceDelimeter = "||"
	timeFormat     = "2006-01-02 15:04:05"
)

// Client to database
type Client struct {
	db     *sqlx.DB
	logger *zerolog.Logger
}

type hearing struct {
	ID        int    `json:"id" db:"id"`
	Link      string `json:"link" db:"link"`
	Topics    string `json:"topics" db:"topics"`
	Place     string `json:"place" db:"place"`
	Date      string `json:"date" db:"date"`
	Proposals string `json:"proposals" db:"proposals"`
	Published bool   `json:"published" db:"published"`
	Raw       string `json:"raw" db:"raw"`
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
		return nil, fmt.Errorf("could not prepare database file: %w", err)
	}

	db, err := sqlx.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("could not open database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("could not ping database: %w", err)
	}
	client.db = db

	err = client.prepareSchema()
	if err != nil {
		return nil, fmt.Errorf("could not prepare schema: %w", err)
	}

	return client, nil
}

// Close connection to database
func (c Client) Close() error {
	c.logger.Info().Msg("closing database connection")
	return c.db.Close()
}

func (c Client) prepareDBFile(filename string) error {
	// Проверка существования файла БД
	fi, err := os.Stat(filename)
	if err != nil || fi.Size() == 0 {
		// Создание файла БД
		f, err := os.Create(path.Clean(filename))
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

// clean query from "\n" and "\t"
func (c Client) cleanQuery(q string) string {
	return strings.ReplaceAll(strings.ReplaceAll(q, "\n", " "), "\t", "")
}

func (c Client) prepareSchema() error {
	version, err := c.getSchemaVersion()
	if err != nil {
		return fmt.Errorf("could not get schema version: %w", err)
	}

	c.logger.Info().Msgf("database current schema version: %d", version)

	queries := []string{
		`CREATE TABLE IF NOT EXISTS hearings(
			id INTEGER PRIMARY KEY,
			link TEXT DEFAULT '' NOT NULL UNIQUE,
			topics TEXT DEFAULT '',
			proposals TEXT DEFAULT '',
			place TEXT DEFAULT '',
			date TEXT DEFAULT '1970-01-01 00:00:00',
			published BOOLEAN DEFAULT false,
			raw TEXT DEFAULT ''
		)`,
	}

	if version == len(queries) {
		c.logger.Info().Msg("database schema is up to date")
		return nil
	}

	c.logger.Info().Msg("upgrading database schema...")

	for i := version; i < len(queries); i++ {
		c.logger.Debug().Msgf("database executing query: %s", c.cleanQuery(queries[i]))
		_, err = c.db.Exec(queries[i])
		if err != nil {
			return fmt.Errorf("could not execute query: %w", err)
		}
		err = c.setSchemaVersion(i + 1)
		if err != nil {
			return fmt.Errorf("could not set schema version: %w", err)
		}
		c.logger.Debug().Msgf("database schema version: %d", i+1)
	}
	c.logger.Info().Msg("database schema is up to date")
	return nil
}

func (c Client) getSchemaVersion() (int, error) {
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

func (c Client) setSchemaVersion(version int) error {
	_, err := c.db.Exec(fmt.Sprintf("PRAGMA user_version = %d", version))
	return err
}

// Create new hearing in database
func (c Client) Create(ctx context.Context, publicHearing domain.Hearing) error {
	// TODO add creation date
	query := "INSERT INTO hearings(link,topics,proposals,place,date,raw) VALUES($1, $2, $3, $4, $5, $6)"
	_, err := c.db.ExecContext(
		ctx,
		query,
		publicHearing.URL,
		strings.Join(publicHearing.Topic, sliceDelimeter),
		strings.Join(publicHearing.Proposals, sliceDelimeter),
		publicHearing.Place,
		publicHearing.Time.Format(timeFormat),
		strings.Join(publicHearing.Raw, sliceDelimeter),
	)
	return err
}

// Update hearing in database
func (c Client) Update(ctx context.Context, publicHearing domain.Hearing) error {
	if publicHearing.URL == "" {
		return fmt.Errorf("cannot update hearing: empty link")
	}

	currentHearing, err := c.Find(ctx, publicHearing.URL)
	if err != nil {
		return err
	}

	topic := publicHearing.Topic
	place := publicHearing.Place
	date := publicHearing.Time
	published := publicHearing.Published
	proposals := publicHearing.Proposals
	raw := publicHearing.Raw

	if len(topic) == 0 {
		topic = currentHearing.Topic
	}

	if place == "" {
		place = currentHearing.Place
	}

	if date.IsZero() {
		date = currentHearing.Time
	}
	dateStr := date.Format(timeFormat)

	if !published {
		published = currentHearing.Published
	}

	if len(proposals) == 0 {
		proposals = currentHearing.Proposals
	}

	if len(raw) == 0 {
		raw = currentHearing.Raw
	}

	query := "UPDATE hearings SET topic = $2, place = $3, date = $4, published = $5, proposals = $6, raw = $7 WHERE link = $1"
	_, err = c.db.ExecContext(ctx, query, publicHearing.URL, topic, place, dateStr, published, proposals, raw)

	return err
}

// Find one hearing in database
func (c Client) Find(ctx context.Context, link string) (domain.Hearing, error) {
	tempHearing := &hearing{}
	hp := domain.Hearing{}
	query := "SELECT id, link, topics, proposals, place, date, published, raw FROM hearings WHERE link = $1"
	err := c.db.QueryRowxContext(ctx, query, link).StructScan(tempHearing)
	if err != nil {
		return hp, err
	}

	hp.URL = tempHearing.Link
	hp.Time, _ = time.Parse(timeFormat, tempHearing.Date)
	hp.Place = tempHearing.Place
	hp.Topic = strings.Split(tempHearing.Topics, sliceDelimeter)
	hp.Proposals = strings.Split(tempHearing.Proposals, sliceDelimeter)
	hp.Published = tempHearing.Published
	hp.Raw = strings.Split(tempHearing.Raw, sliceDelimeter)

	return hp, nil
}

// List all hearings in database
func (c Client) List(ctx context.Context) ([]domain.Hearing, error) {
	tempHearings := make([]hearing, 0)
	res := make([]domain.Hearing, 0)
	query := "SELECT id, link, topics, proposals, place, date as date, published, raw FROM hearings ORDER BY date"
	err := c.db.SelectContext(ctx, &tempHearings, query)
	if err != nil {
		return res, err
	}
	res = c.castToHearing(tempHearings)
	return res, err
}

// Unpublished hearings in database
func (c Client) Unpublished(ctx context.Context, mark bool) ([]domain.Hearing, error) {
	tempHearings := make([]hearing, 0)
	res := make([]domain.Hearing, 0)
	query := "SELECT id, link, topics, proposals, place, date AS date, published, raw FROM hearings WHERE published IS NOT TRUE ORDER BY date"
	if mark {
		query = "UPDATE hearings SET published = TRUE WHERE published IS NOT TRUE RETURNING id, link, topics, proposals, place, date AS date, published, raw"
	}
	err := c.db.SelectContext(ctx, &tempHearings, query)
	if err != nil {
		return res, err
	}
	res = c.castToHearing(tempHearings)
	return res, err
}

func (c Client) castToHearing(h []hearing) []domain.Hearing {
	res := make([]domain.Hearing, 0)
	var err error
	for _, th := range h {
		hp := domain.Hearing{}
		hp.URL = th.Link
		hp.Time, err = time.Parse(timeFormat, th.Date)
		if err != nil {
			break
		}
		hp.Place = th.Place
		hp.Topic = strings.Split(th.Topics, sliceDelimeter)
		hp.Proposals = strings.Split(th.Proposals, sliceDelimeter)
		hp.Published = th.Published
		hp.Raw = strings.Split(th.Raw, sliceDelimeter)
		res = append(res, hp)
	}
	return res
}
