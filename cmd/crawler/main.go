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

package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/brurbanko/mercury/internal/scrapper"

	"github.com/brurbanko/mercury/internal/publisher"

	"github.com/brurbanko/mercury/config"
	"github.com/brurbanko/mercury/database"
	"github.com/brurbanko/mercury/server"

	"github.com/brurbanko/mercury/service/hearings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	l := log.Logger.With().
		Str("service", "crawler").
		Logger()

	// Read and load configuration
	cfg, err := config.Load()
	if err != nil {
		l.Fatal().Err(err).Msg("failed load service configuration")
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if cfg.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		l.Debug().Msg("debug mode")
	}

	l.Debug().Msgf("config: %+v", cfg)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)

	err = run(ctx, cancel, cfg, &l)
	if err != nil {
		l.Fatal().Err(err).Msg("service failed start")
	}
	l.Info().Msg("stopped")
}

func run(ctx context.Context, cancel context.CancelFunc, cfg *config.Config, logger *zerolog.Logger) error {
	defer cancel()

	db, err := database.New(cfg.Database.DSN, logger)
	if err != nil {
		return fmt.Errorf("failed connect to database: %w", err)
	}
	defer func() {
		if cerr := db.Close(); cerr != nil {
			logger.Error().Err(cerr).Msg("failed close database")
		}
	}()

	s := scrapper.New(&scrapper.Options{
		Logger:      logger,
		UserAgent:   "urbanist-public-hearings (https://t.me/public_bryansk_bot)",
		MaxBodySize: 1 << 20, // 1MB
		CacheDir:    "./cache",
	})
	p, err := publisher.New(&publisher.Options{
		Logger:       logger,
		URL:          cfg.Publish.URL,
		Method:       cfg.Publish.Method,
		BodyTemplate: cfg.Publish.Template,
		Headers:      cfg.Publish.Headers,
	})
	if err != nil {
		return fmt.Errorf("failed create publisher: %w", err)
	}

	srv := hearings.New(&hearings.Config{
		Database:  db,
		Scrapper:  s,
		Publisher: p,
		Logger:    logger,
	})

	http := server.New(server.Config{
		Host:     cfg.Server.Host,
		Port:     cfg.Server.Port,
		Token:    cfg.Server.Token,
		Logger:   logger,
		Hearings: srv,
	})

	go http.Start(ctx, cancel)

	<-ctx.Done()

	return nil
}
