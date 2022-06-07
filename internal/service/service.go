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

package service

import (
	"context"
	"errors"
	"net/http"

	"github.com/brurbanko/mercury/internal/crawler"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

// Service is a HTTP server
type Service struct {
	server *http.Server
	logger *zerolog.Logger
}

// Config for creating a new service
type Config struct {
	Host     string
	Port     string
	Database *sqlx.DB
	Crawler  *crawler.Crawler
	Logger   *zerolog.Logger
}

// New instance of service
func New(cfg Config) *Service {
	l := cfg.Logger.With().Str("service", "server").Logger()

	return &Service{
		logger: &l,
		server: &http.Server{
			Addr:    cfg.Host + ":" + cfg.Port,
			Handler: handlers(cfg.Database, cfg.Crawler, &l),
		},
	}
}

func handlers(db *sqlx.DB, crawler *crawler.Crawler, l *zerolog.Logger) http.Handler {
	l.Debug().Msg("creating handlers")
	mux := chi.NewRouter()
	mux.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world!"))
	})

	return mux
}

// Start the web service
func (s *Service) Start(ctx context.Context, cancel context.CancelFunc) {
	s.logger.Info().Msg("starting server")
	go func() {
		err := s.server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Err(err).Msg("error serve service server")
			cancel()
		}
	}()

	<-ctx.Done()

	s.logger.Info().Msg("shutdown web server")

	err := s.server.Shutdown(context.Background())
	if err != nil {
		s.logger.Err(err).Msg("error shutdown service server")
	}
}
