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

package server

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/brurbanko/mercury/service/hearings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	"github.com/rs/zerolog"
)

// Server is HTTP server
type Server struct {
	server   *http.Server
	logger   *zerolog.Logger
	hearings *hearings.Service
}

// Config for creating a new http server
type Config struct {
	Host     string
	Port     string
	Token    string
	Logger   *zerolog.Logger
	Hearings *hearings.Service
}

// New instance of service
func New(cfg Config) *Server {
	l := cfg.Logger.With().Str("server", "http").Logger()

	s := &Server{
		logger: &l,
		server: &http.Server{
			Addr: cfg.Host + ":" + cfg.Port,
		},
		hearings: cfg.Hearings,
	}

	s.initRouter(cfg.Token)

	return s
}

// Start the web service
func (s Server) Start(ctx context.Context, cancel context.CancelFunc) {
	s.logger.Info().Msg("starting http server")
	go func() {
		err := s.server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Err(err).Msg("error serve service server")
			cancel()
		}
	}()

	s.logger.Info().Msgf("http server listening at %s", s.server.Addr)

	<-ctx.Done()

	s.logger.Info().Msg("shutdown web server")

	err := s.server.Shutdown(context.Background())
	if err != nil {
		s.logger.Err(err).Msg("error shutdown service server")
	}
}

func (s Server) initRouter(token string) {
	mux := chi.NewRouter()
	mux.Use(middleware.Logger)
	mux.Use(middleware.Recoverer)

	if token != "" {
		mux.Use(s.authTokenMiddleware(token))
	}

	mux.Route("/hearings", func(r chi.Router) {
		r.Get("/", s.listHearings)
		r.Get("/new", s.newHearings)
		r.Get("/links", s.hearingLinks)
	})

	s.server.Handler = mux
}

func (s Server) authTokenMiddleware(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")

			if header == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// remove bearer prefix
			if strings.HasPrefix(strings.ToLower(header), "bearer ") {
				header = header[7:]
			}

			if header != token {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

type errorResponse struct {
	Error string `json:"error"`
}

type dataResponse struct {
	Data interface{} `json:"data"`
}

func (s Server) listHearings(w http.ResponseWriter, r *http.Request) {
	s.logger.Debug().Msg("list hearings")
	h, err := s.hearings.List(r.Context())
	if err != nil {
		s.logger.Err(err).Msg("failed show hearings list")
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, errorResponse{err.Error()})
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, dataResponse{h})
}

func (s Server) newHearings(w http.ResponseWriter, r *http.Request) {
	s.logger.Debug().Msg("new hearings")
	h, err := s.hearings.NewHearings(r.Context())
	if err != nil {
		s.logger.Err(err).Msg("failed find new hearings")
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, errorResponse{err.Error()})
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, dataResponse{h})
}

func (s Server) hearingLinks(w http.ResponseWriter, r *http.Request) {
	s.logger.Debug().Msg("hearing links")
	links, err := s.hearings.FetchLinks(r.Context())
	if err != nil {
		s.logger.Err(err).Msg("failed to fetch hearing links")
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, errorResponse{err.Error()})
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, dataResponse{links})
}
