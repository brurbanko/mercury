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

	"github.com/brurbanko/mercury/internal/publisher"

	"github.com/brurbanko/mercury/internal/scrapper"

	"github.com/brurbanko/mercury/database"
	"github.com/brurbanko/mercury/domain"

	"github.com/rs/zerolog"
)

// Service to manage public hearings
type Service struct {
	logger *zerolog.Logger
	parser *Parser
	db     *database.Client

	scrapper  *scrapper.Scrapper
	publisher *publisher.Publisher
}

// Config for hearings service
type Config struct {
	Database  *database.Client
	Logger    *zerolog.Logger
	Scrapper  *scrapper.Scrapper
	Publisher *publisher.Publisher
}

// New returns an instance of hearing service
func New(cfg *Config) *Service {
	l := cfg.Logger.With().Str("service", "hearings").Logger()
	return &Service{
		logger: &l,
		db:     cfg.Database,
		parser: NewParser(),

		scrapper:  cfg.Scrapper,
		publisher: cfg.Publisher,
	}
}

// List of hearings
func (s Service) List(ctx context.Context) ([]domain.Hearing, error) {
	links, err := s.db.List(ctx)

	// Reverse slice. Older links will be at begin
	for i, j := 0, len(links)-1; i < j; i, j = i+1, j-1 {
		links[i], links[j] = links[j], links[i]
	}
	return links, err
}

// FetchLinks of public hearings
func (s Service) FetchLinks(ctx context.Context) ([]string, error) {
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

// ProcessLink and get information about public hearing
func (s Service) ProcessLink(ctx context.Context, link string) (domain.Hearing, error) {
	l := s.logger.With().
		Str("method", "ProcessLink").
		Str("link", link).
		Logger()
	l.Info().Msg("processing hearing")
	hearing := domain.Hearing{URL: link}
	content, err := s.scrapper.ExtractContent(ctx, link, ".thecontent p", false)
	if err != nil {
		l.Error().Err(err).Msg("failed to extract content")
		return hearing, err
	}
	hearing.Raw = content

	hp, err := s.parser.Content(hearing)
	if err != nil {
		l.Error().Err(err).Msg("failed to parse hearing content")
		return hp, err
	}
	return hp, nil
}

// Find public hearing by URL
func (s Service) Find(ctx context.Context, link string) (domain.Hearing, error) {
	return s.db.Find(ctx, link)
}

// NewHearings returns list of new hearings from site
func (s Service) NewHearings(ctx context.Context) ([]domain.Hearing, error) {
	l := s.logger.With().Str("method", "NewHearings").Logger()
	l.Info().Msg("fetching new hearings")
	links, err := s.FetchLinks(ctx)
	if err != nil {
		l.Error().Err(err).Msg("failed to get list of links")
		return nil, err
	}

	l.Debug().Msg("retrieving processed hearings")
	list, err := s.db.List(ctx)
	if err != nil {
		l.Error().Err(err).Msg("failed to get list of hearings")
		return nil, err
	}

	l.Debug().Msg("filtering new hearings")
	processedLinks := make(map[string]struct{})
	for _, hearing := range list {
		processedLinks[hearing.URL] = struct{}{}
	}

	newLinks := make([]string, 0)
	for _, link := range links {
		if _, ok := processedLinks[link]; !ok {
			newLinks = append(newLinks, link)
		}
	}

	l.Info().Msgf("found %d new hearings", len(newLinks))
	hearings := make([]domain.Hearing, 0)
	for _, link := range newLinks {
		hearing, err := s.ProcessLink(ctx, link)
		if err != nil {
			l.Error().Err(err).Msg("failed to process hearing")
			continue
		}

		err = s.db.Create(ctx, hearing)
		if err != nil {
			l.Error().Err(err).Str("link", link).Msg("failed to save hearing")
			continue
		}

		hearings = append(hearings, hearing)
	}

	return hearings, nil
}

// ListUnpublished returns list of unpublished hearings
func (s Service) ListUnpublished(ctx context.Context, mark bool) ([]domain.IHearing, error) {
	l := s.logger.With().Str("method", "ListUnpublished").Logger()
	l.Info().Msg("listing unpublished hearings")
	unpublished, err := s.db.Unpublished(ctx, mark)
	if err != nil {
		l.Error().Err(err).Msg("failed to get unpublished hearings")
		return nil, err
	}
	// Cast slice of domain.Hearing to slice of domain.IHearing
	hearings := make([]domain.IHearing, 0)
	for _, h := range unpublished {
		hearings = append(hearings, h)
	}
	return hearings, nil
}

// Publish all unpublished hearings.
// Get it from DB and publish them one by one to URL
func (s Service) Publish(ctx context.Context, format string) (int, error) {
	l := s.logger.With().Str("method", "Publish").Logger()
	l.Info().Msg("publishing new hearings")
	unpublished, err := s.db.Unpublished(ctx, false)
	if err != nil {
		l.Error().Err(err).Msg("failed to get unpublished hearings")
		return 0, err
	}

	for i, h := range unpublished {
		var message string
		if format == "markdown" {
			message = h.Markdown()
		} else {
			message = h.String()
		}
		err = s.publisher.Publish(ctx, message)
		if err != nil {
			l.Error().Err(err).Str("link", h.URL).Msg("failed to publish hearing")
			return i, err
		}
		l.Info().Str("link", h.URL).Msg("hearing published")
		err = s.db.MarkPublished(ctx, h.URL)
		if err != nil {
			l.Error().Err(err).Str("link", h.URL).Msg("failed to mark hearing as published")
			return i, err
		}

	}
	return len(unpublished), nil
}
