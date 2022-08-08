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
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog"
)

// Publisher is library to publish messages with HTTP
type Publisher struct {
	logger *zerolog.Logger

	skip bool
	url  string
	chat string
}

// Options for creating a new publisher
type Options struct {
	Logger *zerolog.Logger

	Token  string
	ChatID string
}

type tgMessage struct {
	ChatID         string `json:"chat_id"`
	ParseMode      string `json:"parse_mode"`
	Text           string `json:"text"`
	DisablePreview bool   `json:"disable_web_page_preview"`
}

// New instance of publisher
func New(opt *Options) (*Publisher, error) {
	var l zerolog.Logger
	if opt.Logger == nil {
		l = zerolog.Nop()
	} else {
		l = opt.Logger.With().Str("package", "publisher").Logger()
	}

	return &Publisher{
		logger: &l,

		skip: opt.Token == "",
		url:  fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", opt.Token),
		chat: opt.ChatID,
	}, nil
}

// Publish message
func (p Publisher) Publish(ctx context.Context, message string) error {
	if p.skip {
		p.logger.Debug().Msg("Token is empty. Publish skipped")
		return nil
	}
	p.logger.Debug().Msg("Publishing")

	msg := tgMessage{
		ParseMode:      "MarkdownV2",
		DisablePreview: true,
		ChatID:         p.chat,
		Text:           message,
	}
	var body bytes.Buffer
	err := json.NewEncoder(&body).Encode(msg)
	if err != nil {
		p.logger.Error().Err(err).Msg("error creating body")
		return err
	}

	// prepare request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.url, &body)
	if err != nil {
		p.logger.Error().Err(err).Msg("error creating request")
		return err
	}

	req.Header.Set("Content-Type", "application/json")

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
		response, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		p.logger.Error().Msgf("response: %s", response)
		return fmt.Errorf("response status code is not OK: %d", resp.StatusCode)
	}

	return nil
}
