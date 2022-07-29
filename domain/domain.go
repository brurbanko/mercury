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

package domain

import (
	"strings"
	"time"
)

// IHearing is interface for hearing
// Have two methods: String and Markdown
type IHearing interface {
	String() string
	Markdown() string
}

// Hearing of BGA32
type Hearing struct {
	ID        string    `json:"id"`
	Topic     []string  `json:"topic"`
	Proposals []string  `json:"proposals"`
	Place     string    `json:"place"`
	URL       string    `json:"url"`
	Time      time.Time `json:"time"`
	Published bool      `json:"published"`
	Raw       []string  `json:"raw"`
}

// String returns text representation of hearing
func (h Hearing) String() string {
	if h.Place == "" {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(h.Time.Format("02.01.2006"))
	sb.WriteString(" в ")
	sb.WriteString(h.Time.Format("15:04"))
	sb.WriteString(" в ")
	sb.WriteString(h.Place)

	if len(h.Topic) == 1 {
		sb.WriteString(" состоятся публичные слушания ")
		sb.WriteString(h.Topic[0])
	} else {
		sb.WriteString(" состоятся публичные слушания:")
		for _, t := range h.Topic {
			if len(t) > 0 {
				sb.WriteString("\n - ")
				sb.WriteString(t)
			}
		}
	}
	sb.WriteString("\n")

	for _, p := range h.Proposals {
		sb.WriteString(p)
		sb.WriteString("\n")
	}

	sb.WriteString("Ссылка на публикацию: ")
	sb.WriteString(h.URL)
	sb.WriteString("\n")

	return sb.String()
}

// Markdown returns formatted representation of hearing
func (h Hearing) Markdown() string {
	if h.Place == "" {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("*")
	sb.WriteString(h.Time.Format("02.01.2006"))
	sb.WriteString(" в ")
	sb.WriteString(h.Time.Format("15:04"))
	sb.WriteString(" в ")
	sb.WriteString(h.Place)
	sb.WriteString("*")

	if len(h.Topic) == 1 {
		sb.WriteString(" состоятся публичные слушания ")
		sb.WriteString(h.Topic[0])
	} else {
		sb.WriteString(" состоятся публичные слушания:")
		for _, t := range h.Topic {
			sb.WriteString("\n\n - ")
			sb.WriteString(t)
		}
	}
	sb.WriteString("\n\n")

	for _, p := range h.Proposals {
		sb.WriteString(p)
		sb.WriteString("\n\n")
	}

	sb.WriteString("[Ссылка на публикацию](")
	sb.WriteString(h.URL)
	sb.WriteString(")\n")

	return sb.String()
}
