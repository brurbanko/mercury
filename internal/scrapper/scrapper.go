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

package scrapper

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"

	"golang.org/x/net/html/charset"

	"github.com/PuerkitoBio/goquery"

	"github.com/rs/zerolog"
)

// ErrCacheNotSet is returned when cache directory is not set
var ErrCacheNotSet = fmt.Errorf("cache directory is not set")

// Scrapper is a scrapper for public hearings
type Scrapper struct {
	client *http.Client
	logger *zerolog.Logger

	reSpaces *regexp.Regexp

	cacheDir string

	ua          string
	maxBodySize int64
}

// Options for scrapper
type Options struct {
	Logger   *zerolog.Logger
	LogLevel zerolog.Level

	CacheDir string

	UserAgent   string
	MaxBodySize int64
}

// New scrapper instance
func New(opt *Options) *Scrapper {
	var l zerolog.Logger
	if opt.Logger == nil {
		l = zerolog.Nop()
	} else {
		l = opt.Logger.With().Str("package", "scrapper").Logger().Level(opt.LogLevel)
	}
	if opt.UserAgent == "" {
		opt.UserAgent = "urbanist-public-hearings (https://t.me/public_bryansk_bot)"
	}
	if opt.MaxBodySize == 0 {
		opt.MaxBodySize = 1024 * 1024
	}

	return &Scrapper{
		logger: &l,
		client: &http.Client{},

		reSpaces: regexp.MustCompile(`[\x{00A0}\s\t\n\v\f\r\p{Zs}]+`),

		cacheDir: opt.CacheDir,

		ua:          opt.UserAgent,
		maxBodySize: opt.MaxBodySize,
	}
}

// ExtractContent returns content of passed link.
// Option "force" forces to fetch the page from the network instead of from the cache.
func (s Scrapper) ExtractContent(ctx context.Context, link, selector string, force bool) ([]string, error) {
	l := s.logger.With().
		Str("link", link).
		Str("selector", selector).
		Bool("force", force).
		Logger()
	l.Debug().Msg("extracting content")
	body, err := s.fetch(ctx, link, false)
	if err != nil {
		l.Error().Err(err).Msg("error fetching")
		return nil, err
	}
	l.Debug().Msg("creating document")
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		l.Error().Err(err).Msg("error creating document from body")
		return nil, err
	}

	l.Debug().Msg("extracting content")
	var content []string
	doc.Find(selector).Each(func(i int, sel *goquery.Selection) {
		t := strings.TrimSpace(s.reSpaces.ReplaceAllString(sel.Text(), " "))
		if len(t) > 0 {
			content = append(content, t)
		}
	})

	l.Debug().Msgf("content length: %d", len(content))
	return content, nil
}

// ExtractLinks return all links from passed selector.
// Option "force" forces to fetch the page from the network instead of from the cache.
func (s Scrapper) ExtractLinks(ctx context.Context, link, selector string, force bool) ([]string, error) {
	l := s.logger.With().
		Str("method", "ExtractLinks").
		Str("link", link).
		Str("selector", selector).
		Bool("force", force).
		Logger()
	l.Debug().Msg("extracting links")
	body, err := s.fetch(ctx, link, force)
	if err != nil {
		l.Error().Err(err).Msg("error fetching")
		return nil, err
	}
	l.Debug().Msg("creating document")
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		l.Error().Err(err).Msg("error creating document from body")
		return nil, err
	}

	l.Debug().Msg("extracting links")
	var content []string
	doc.Find(selector).Each(func(i int, s *goquery.Selection) {
		l, ok := s.Attr("href")
		if ok {
			content = append(content, l)
		}
	})

	l.Debug().Msgf("links length: %d", len(content))
	return content, nil
}

// fetch HTML from link
func (s Scrapper) fetch(ctx context.Context, link string, force bool) ([]byte, error) {
	l := s.logger.With().
		Str("method", "fetch").
		Str("link", link).
		Logger()

	if !force {
		body, err := s.loadFromCache(link)
		if err == nil {
			l.Debug().Msg("loaded from cache")
			return body, nil
		}
		l.Debug().Err(err).Msg("error loading from cache")
	}

	l.Debug().Msg("fetching HTML")
	req, err := http.NewRequestWithContext(ctx, "GET", link, http.NoBody)
	if err != nil {
		l.Error().Err(err).Msg("error creating request")
		return nil, err
	}
	l.Debug().Msg("setting user agent")
	req.Header.Add("User-Agent", s.ua)

	l.Debug().Msg("sending request")
	resp, err := s.client.Do(req)
	if err != nil {
		l.Error().Err(err).Msg("error fetching")
		return nil, err
	}
	defer func() {
		cerr := resp.Body.Close()
		if cerr != nil {
			l.Error().Err(err).Msg("error closing response body")
		}
	}()

	l.Debug().Msgf("response status code: %d", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		l.Error().Err(err).Msg("error fetching")
		return nil, err
	}

	l.Debug().Msg("decoding response body")
	utf8, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		l.Error().Err(err).Msg("error reading response body")
		return nil, err
	}

	body, err := io.ReadAll(&io.LimitedReader{R: utf8, N: s.maxBodySize})
	if err != nil {
		l.Error().Err(err).Msg("error reading response body")
		return nil, err
	}

	err = s.saveToCache(link, body)

	return body, err
}

// create cache directory if it doesn't exist
func (s Scrapper) createCacheDirectory() error {
	return os.MkdirAll(path.Clean(s.cacheDir), 0o750)
}

// save html page to cache
func (s Scrapper) saveToCache(link string, body []byte) error {
	l := s.logger.With().
		Str("method", "saveToCache").
		Str("link", link).
		Logger()
	l.Debug().Msg("saving to cache")
	if s.cacheDir == "" {
		l.Debug().Msg("cache directory is not set")
		return nil
	}
	err := s.createCacheDirectory()
	if err != nil {
		l.Error().Err(err).Msg("error creating cache directory")
		return err
	}
	file, err := os.Create(path.Clean(path.Join(s.cacheDir, s.safeFileNameFromLink(link))))
	if err != nil {
		l.Error().Err(err).Msg("error creating cache file")
		return err
	}
	_, err = file.Write(body)
	if err != nil {
		l.Error().Err(err).Msg("error writing to cache file")
		return err
	}

	return nil
}

// load html page from cache
func (s Scrapper) loadFromCache(link string) ([]byte, error) {
	l := s.logger.With().
		Str("method", "loadFromCache").
		Str("link", link).
		Logger()
	l.Debug().Msg("loading from cache")
	if s.cacheDir == "" {
		l.Debug().Msg("cache directory is not set")
		return nil, ErrCacheNotSet
	}
	file, err := os.Open(path.Clean(path.Join(s.cacheDir, s.safeFileNameFromLink(link))))
	if err != nil {
		l.Error().Err(err).Msg("error opening cache file")
		return nil, err
	}
	body, err := io.ReadAll(file)
	if err != nil {
		l.Error().Err(err).Msg("error reading cache file")
		return nil, err
	}

	return body, nil
}

// create safe file name from link
func (s Scrapper) safeFileNameFromLink(link string) string {
	str := &strings.Builder{}
	u, err := url.Parse(link)
	if err == nil {
		str.WriteString(u.Hostname())
		str.WriteString("_")
	}
	h := md5.New() // nolint:gosec // hash used for create file name
	_, _ = io.WriteString(h, link)
	str.WriteString(fmt.Sprintf("%x", h.Sum(nil)))
	return str.String()
}
