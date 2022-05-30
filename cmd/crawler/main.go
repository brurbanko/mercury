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

	"github.com/brurbanko/mercury/internal/config"
	"github.com/brurbanko/mercury/internal/crawler"
	"github.com/brurbanko/mercury/internal/database"
	"github.com/brurbanko/mercury/internal/service"

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

	db, err := database.New(cfg.Database.DSN)
	if err != nil {
		return fmt.Errorf("failed connect to database: %w", err)
	}

	bga := crawler.New(cfg.Crawler.Domain, cfg.Crawler.UserAgent)

	// like https://github.com/queuedb/queuedb/blob/master/cmd/queuedb/main.go
	srv := service.New(service.Config{
		Host:     cfg.HTTP.Host,
		Port:     cfg.HTTP.Port,
		Database: db,
		Crawler:  bga,
		Logger:   logger,
	})

	go srv.Start(ctx, cancel)

	<-ctx.Done()

	return nil
}
