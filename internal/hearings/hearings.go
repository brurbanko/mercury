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

package hearings

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path"
	"sync/atomic"
	"time"

	"github.com/brurbanko/mercury/pkg/crawler"

	"github.com/brurbanko/mercury/internal/database"
	"github.com/brurbanko/mercury/internal/domain"
	"github.com/rs/zerolog"
)

var fid uint64

// Service to manage public hearings
type Service struct {
	logger  *zerolog.Logger
	crawler *crawler.Crawler
	parser  *Parser
	db      *database.Client
	tempDir string
}

// Config for hearings service
type Config struct {
	Database *database.Client
	Crawler  *crawler.Crawler
	Logger   *zerolog.Logger
	TempDir  string
}

// New returns an instance of hearings service
func New(cfg *Config) *Service {
	l := cfg.Logger.With().Str("service", "hearings").Logger()
	return &Service{
		logger:  &l,
		crawler: cfg.Crawler,
		db:      cfg.Database,
		tempDir: cfg.TempDir,
		parser:  NewParser(),
	}
}

// List of hearings (Cached)
func (s *Service) List(ctx context.Context) ([]domain.Hearing, error) {
	links, err := s.db.List(ctx)

	// Reverse slice. Older links will be at begin
	for i, j := 0, len(links)-1; i < j; i, j = i+1, j-1 {
		links[i], links[j] = links[j], links[i]
	}
	return links, err
}

// FetchLinks of public hearings
func (s *Service) FetchLinks() ([]string, error) {
	links, err := s.crawler.ExtractLinks(s.parser.LinkAndSelectorForAll())

	// Reverse slice. Older links will be at begin
	for i, j := 0, len(links)-1; i < j; i, j = i+1, j-1 {
		links[i], links[j] = links[j], links[i]
	}
	return links, err
}

// Fetch information about public hearing from site
func (s *Service) Fetch(link string) (*domain.Hearing, error) {
	content, err := s.crawler.ExtractContent(link, s.parser.SelectorForContent())
	if err != nil {
		return &domain.Hearing{
			URL: link,
		}, err
	}
	err = s.storeTempContent(link, content)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed store temporary content")
	}

	hearing, err := s.parser.Content(content)
	if err != nil {
		return &domain.Hearing{
			URL: link,
		}, err
	}
	hearing.URL = link
	return hearing, nil
}

// Find public hearing by URL
func (s *Service) Find(ctx context.Context, link string) (*domain.Hearing, error) {
	hearing, err := s.db.Find(ctx, link)
	return &hearing, err
}

// Save information about public hearing in database
func (s *Service) Save(ctx context.Context, hearing *domain.Hearing) error {
	return s.db.Create(ctx, *hearing)
}

func (s *Service) storeTempContent(link string, content []string) error {
	if s.tempDir == "" {
		return nil
	}
	ts := time.Now().UnixNano() / int64(time.Millisecond)
	id := atomic.AddUint64(&fid, 1)
	fp := path.Join(s.tempDir, fmt.Sprintf("%d.%d.txt", ts, id))
	f, err := os.Create(path.Clean(fp))
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	// write link
	_, err = f.WriteString(link + "\n")
	if err != nil {
		return err
	}

	// write content
	for _, s := range content {
		_, err = f.WriteString(s + "\n")
		if err != nil {
			return err
		}
	}
	err = f.Sync()
	if err != nil {
		return err
	}

	return f.Close()
}

func (s *Service) loadTempContent(fileName string) (content []string, link string, err error) {
	if fileName == "" {
		return content, link, fmt.Errorf("empty file name")
	}
	if s.tempDir == "" {
		return content, link, fmt.Errorf("empty tmpPathorary storage")
	}

	file := path.Join(s.tempDir, fileName)

	f, err := os.OpenFile(path.Clean(file), os.O_RDONLY, os.ModePerm)
	if err != nil {
		return content, link, err
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		content = append(content, scanner.Text())
	}
	if len(content) > 0 {
		link = content[0]
		content = content[1:]
	}

	if err := scanner.Err(); err != nil {
		return content, link, err
	}

	return content, link, nil
}
