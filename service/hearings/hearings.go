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
	"context"

	"github.com/brurbanko/mercury/database"
	"github.com/brurbanko/mercury/domain"

	"github.com/brurbanko/mercury/pkg/crawler"

	"github.com/rs/zerolog"
)

var fid uint64

// Service to manage public hearings
type Service struct {
	logger  *zerolog.Logger
	crawler *crawler.Crawler
	parser  *Parser
	db      *database.Client
}

// Config for hearings service
type Config struct {
	Database *database.Client
	Crawler  *crawler.Crawler
	Logger   *zerolog.Logger
}

// New returns an instance of hearings service
func New(cfg *Config) *Service {
	l := cfg.Logger.With().Str("service", "hearings").Logger()
	return &Service{
		logger:  &l,
		crawler: cfg.Crawler,
		db:      cfg.Database,
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

func (s *Service) FindNew(ctx context.Context) ([]domain.Hearing, error) {
	links, err := s.FetchLinks()
	if err != nil {
		return nil, err
	}
	var hearings []domain.Hearing
	for _, link := range links {
		_, err := s.Find(ctx, link)
		if err != nil {
			hearing, ferr := s.Fetch(link)
			if ferr != nil {
				s.logger.Err(ferr).Msg("Failed to fetch hearing")
				continue
			}
			serr := s.Save(ctx, hearing)
			if serr != nil {
				s.logger.Err(serr).Msg("Failed to save hearing")
				continue
			}
			hearings = append(hearings, *hearing)
		}
	}
	return hearings, nil
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
