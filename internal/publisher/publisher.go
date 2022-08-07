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

package publisher

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/rs/zerolog"
)

// Publisher is library to publish messages with HTTP
type Publisher struct {
	logger *zerolog.Logger

	url          string
	method       string
	bodyTemplate *template.Template
	headers      map[string][]string
}

// Options for creating a new publisher
type Options struct {
	Logger *zerolog.Logger

	URL          string
	Method       string
	BodyTemplate string
	Headers      map[string][]string
}

// New instance of publisher
func New(opt *Options) (*Publisher, error) {
	var l zerolog.Logger
	if opt.Logger == nil {
		l = zerolog.Nop()
	} else {
		l = opt.Logger.With().Str("package", "publisher").Logger()
	}

	if opt.Method == "" {
		opt.Method = "POST"
	}

	tmpl, err := template.New("bodyTemplate").Parse(opt.BodyTemplate)
	if err != nil {
		l.Error().Err(err).Msg("error parsing bodyTemplate")
		return nil, err
	}

	return &Publisher{
		logger: &l,

		url:          opt.URL,
		method:       opt.Method,
		bodyTemplate: tmpl,
		headers:      opt.Headers,
	}, nil
}

// Publish message
func (p *Publisher) Publish(ctx context.Context, message string) error {
	p.logger.Debug().Msg("Publishing")
	if p.url == "" {
		p.logger.Error().Msg("URL is not set")
		return fmt.Errorf("URL is not set")
	}

	var body bytes.Buffer
	err := p.bodyTemplate.Execute(&body, struct{ Message string }{Message: message})
	if err != nil {
		p.logger.Error().Err(err).Msg("error creating body")
		return err
	}
	// prepare request
	req, err := http.NewRequestWithContext(ctx, p.method, p.url, &body)
	if err != nil {
		p.logger.Error().Err(err).Msg("error creating request")
		return err
	}
	// add headers
	if p.headers != nil {
		for k, v := range p.headers {
			req.Header.Set(k, strings.Join(v, ","))
		}
	}
	// send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		p.logger.Error().Err(err).Msg("error sending request")
		return err
	}
	defer func() {
		cerr := resp.Body.Close()
		if cerr != nil {
			p.logger.Error().Err(cerr).Msg("error closing response body")
		}
	}()
	p.logger.Debug().Msg("response received")
	if resp.StatusCode != http.StatusOK {
		p.logger.Error().Msgf("response status code is not OK: %d", resp.StatusCode)
		p.logger.Error().Msgf("response bodyTemplate: %s", resp.Body)
		return fmt.Errorf("response status code is not OK: %d", resp.StatusCode)
	}

	return nil
}
