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
	"fmt"

	"github.com/brurbanko/mercury/internal/scrapper"

	"github.com/brurbanko/mercury/database"
	"github.com/brurbanko/mercury/domain"

	"github.com/rs/zerolog"
)

var fid uint64

// Service to manage public hearings
type Service struct {
	logger *zerolog.Logger
	parser *Parser
	db     *database.Client

	scrapper *scrapper.Scrapper
}

// Config for hearings service
type Config struct {
	Database *database.Client
	Logger   *zerolog.Logger
}

// New returns an instance of hearing service
func New(cfg *Config) *Service {
	l := cfg.Logger.With().Str("service", "hearings").Logger()
	return &Service{
		logger: &l,
		db:     cfg.Database,
		parser: NewParser(),

		scrapper: scrapper.New(&scrapper.Options{
			Logger:      &l,
			UserAgent:   "urbanist-public-hearings (https://t.me/public_bryansk_bot)",
			MaxBodySize: 1024 * 1024,
			CacheDir:    "./cache",
		}),
	}
}

// List of hearings
func (s *Service) List(ctx context.Context) ([]domain.Hearing, error) {
	links, err := s.db.List(ctx)

	// Reverse slice. Older links will be at begin
	for i, j := 0, len(links)-1; i < j; i, j = i+1, j-1 {
		links[i], links[j] = links[j], links[i]
	}
	return links, err
}

// FetchLinks of public hearings
func (s *Service) FetchLinks(ctx context.Context) ([]string, error) {
	links, err := s.scrapper.ExtractLinks(ctx,
		"https://bga32.ru/arxitektura-i-gradostroitelstvo/publichnye-slushaniya/",
		".thecontent ol li a",
		true)

	// Reverse slice. Older links will be at begin
	for i, j := 0, len(links)-1; i < j; i, j = i+1, j-1 {
		links[i], links[j] = links[j], links[i]
	}
	return links, err
}

// Fetch information about public hearing from site
func (s *Service) Fetch(ctx context.Context, link string) (domain.Hearing, error) {
	l := s.logger.With().
		Str("method", "Fetch").
		Str("link", link).
		Logger()
	hearing := domain.Hearing{URL: link}
	content, err := s.scrapper.ExtractContent(ctx, link, ".thecontent p", false)
	if err != nil {
		l.Error().Err(err).Msg("failed to extract content")
		return hearing, err
	}
	hearing.Raw = content

	err = s.db.Create(ctx, hearing)
	if err != nil {
		l.Error().Err(err).Msg("failed to save hearing")
		return hearing, err
	}

	hearing, err = s.parser.Content(content)
	hearing.URL = link
	if err != nil {
		l.Error().Err(err).Msg("failed to parse hearing content")
		return hearing, err
	}
	return hearing, nil
}

// Find public hearing by URL
func (s *Service) Find(ctx context.Context, link string) (domain.Hearing, error) {
	return s.db.Find(ctx, link)
}

func (s *Service) FindUnpublished(_ context.Context) ([]domain.Hearing, error) {
	return nil, fmt.Errorf("not implemented")
}
