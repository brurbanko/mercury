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
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/brurbanko/mercury/domain"
)

var spaces = `[\x{00A0}\s\t\n\v\f\r\p{Zs}]`
var topicStartParagraph = spaces + "состоятся" + spaces + "+(публичные" + spaces + "+слушания" + spaces + "+)?"
var topicFromMisprintParagraph = "по" + spaces + "+(?:проект|объект).*$"
var topicEndParagraph = "^(?:Экспозици." + spaces + "+проект|Участник|В" + spaces + "+проект|Публичные" + spaces + "+слушания)"
var proposalParagraph = "^При[её]м" + spaces
var timeAndPlace = `(?P<day>\d+)` + spaces + `+(?P<month>\p{L}+)(?:` + spaces + `+(?P<year>\d+)` + spaces + "+года)?" + spaces + `+в` + spaces + `+(?P<hours>\d+)[\.:](?P<minutes>\d+)` + spaces + "+(?:в|по" + spaces + "+адресу:?)" + spaces + `(?P<place>.*)`
var clearLine = `^[\s\p{Zs}]*[-—]?[\s\p{Zs}]*(?P<line>.*)[\s\p{Zs}]*[\.;]+?[\s\p{Zs}]*$`
var year = `(?P<year>\d{4})(?:-goda/)?`

var serviceTimeLocation = time.Now().Location()
var beginnigTime = time.Date(2021, time.January, 1, 0, 0, 0, 0, serviceTimeLocation)

var months = map[string]time.Month{
	"январь":   time.January,
	"января":   time.January,
	"февраль":  time.February,
	"февраля":  time.February,
	"март":     time.March,
	"марта":    time.March,
	"апрель":   time.April,
	"апреля":   time.April,
	"май":      time.May,
	"мая":      time.May,
	"июнь":     time.June,
	"июня":     time.June,
	"июль":     time.July,
	"июля":     time.July,
	"август":   time.August,
	"августа":  time.August,
	"сентябрь": time.September,
	"сентября": time.September,
	"октябрь":  time.October,
	"октября":  time.October,
	"ноябрь":   time.November,
	"ноября":   time.November,
	"декабрь":  time.December,
	"декабря":  time.December,
}

// Parser is preparing data about public hearings
type Parser struct {
	reTopicStart        *regexp.Regexp
	reTopicEnd          *regexp.Regexp
	reTimePlace         *regexp.Regexp
	reClearLine         *regexp.Regexp
	reProposalParagraph *regexp.Regexp
	reYear              *regexp.Regexp
	reMissprintTopic    *regexp.Regexp
}

// NewParser return instance of public hearings parser
func NewParser() *Parser {
	return &Parser{
		reTopicStart:        regexp.MustCompile(topicStartParagraph),
		reTopicEnd:          regexp.MustCompile(topicEndParagraph),
		reTimePlace:         regexp.MustCompile(timeAndPlace),
		reClearLine:         regexp.MustCompile(clearLine),
		reProposalParagraph: regexp.MustCompile(proposalParagraph),
		reYear:              regexp.MustCompile(year),
		reMissprintTopic:    regexp.MustCompile(topicFromMisprintParagraph),
	}
}

// Content return full data of public hearing
func (p *Parser) Content(hearing domain.Hearing) (domain.Hearing, error) {
	return p.prepare(hearing)
}

func (p *Parser) prepare(hearing domain.Hearing) (domain.Hearing, error) {
	ph := domain.Hearing{
		URL: hearing.URL,
		Raw: hearing.Raw,
	}
	content := hearing.Raw

	if len(content) == 0 {
		return ph, fmt.Errorf("empty content")
	}

	/* DEFINE TOPIC */
	start, next := p.defineTopicsParagraphs(content)

	// All data in one paragraph
	if start+1 == next {
		// Split content to parts: 1-st -- time and place, 2-nd -- topic
		parts := p.reTopicStart.Split(content[start], -1)
		if len(parts) != 2 {
			topics := p.reMissprintTopic.FindAllString(content[start], -1)
			if len(topics) == 0 {
				return ph, fmt.Errorf("failed parse content. cannot split to time/place and topic")
			}
			// Add topic to it position
			parts = append(parts, topics[0])
			// extract topic from place
			parts[0] = strings.TrimSpace(strings.TrimSuffix(content[start], topics[0]))
		}
		top := p.clearString(parts[1])
		if top != "" {
			ph.Topic = append(ph.Topic, top)
		}
		ph.Place = parts[0]
	} else {
		// Multiple topics in different paragraphs
		for i := start; i < next; i++ {
			// First paragraph with date and place
			if i == start {
				parts := p.reTopicStart.Split(content[i], -1)
				if len(parts) == 0 {
					return ph, fmt.Errorf("failed parse content. cannot split to time and place")
				}
				ph.Place = parts[0]
			} else {
				//	Next paragraphs with topics
				top := p.clearString(content[i])
				if top != "" {
					ph.Topic = append(ph.Topic, top)
				}
			}
		}
	}

	if len(ph.Topic) == 0 {
		return ph, fmt.Errorf("failed parse content. cannot get topics")
	}

	/* DEFINE PLACE AND TIME */
	if ph.Place != "" {
		match := p.reTimePlace.FindStringSubmatch(ph.Place)

		paramsMap := make(map[string]string)
		for i, name := range p.reTimePlace.SubexpNames() {
			if i > 0 && i <= len(match) {
				paramsMap[name] = match[i]
			}
		}

		year, err := strconv.Atoi(paramsMap["year"])
		if err != nil {
			// If year not defined try extract it from url
			year, err = p.extractYear(hearing.URL)
			if err != nil {
				// If year is still not defined, use current year
				year = time.Now().Year()

			}
		}

		day, err := strconv.Atoi(paramsMap["day"])
		if err != nil {
			day = 1
		}

		hours, err := strconv.Atoi(paramsMap["hours"])
		if err != nil {
			hours = 0
		}

		minutes, err := strconv.Atoi(paramsMap["minutes"])
		if err != nil {
			minutes = 0
		}

		ph.Time = time.Date(year, months[paramsMap["month"]], day, hours, minutes, 0, 0, serviceTimeLocation)
		if ph.Time.Before(beginnigTime) {
			return ph, fmt.Errorf("failed parse date. the extracted date (%s) is earlier than the beginning time (%s): %s", ph.Time, beginnigTime, ph.Place)
		}
		// Replace place
		ph.Place = paramsMap["place"]
	} else {
		return ph, fmt.Errorf("failed parse date and place. empty string")
	}

	/* DEFINE PROPOSALS */
	prop := p.defineProposalParagraphs(content)
	for _, p := range prop {
		ph.Proposals = append(ph.Proposals, content[p])
	}

	return ph, nil
}

func (p *Parser) defineTopicsParagraphs(content []string) (start, next int) {
	for i, paragraph := range content {
		if p.reTopicStart.MatchString(paragraph) {
			start = i
		}
		if p.reTopicEnd.MatchString(paragraph) {
			next = i
			break
		}
	}
	if next == start {
		next++
	}
	return start, next
}

func (p *Parser) defineProposalParagraphs(content []string) (res []int) {
	for i, paragraph := range content {
		if p.reProposalParagraph.MatchString(paragraph) {
			res = append(res, i)
		}
	}
	return
}

func (p *Parser) clearString(str string) string {
	match := p.reClearLine.FindStringSubmatch(str)
	if len(match) == 0 {
		return str
	}

	paramsMap := make(map[string]string)
	for i, name := range p.reClearLine.SubexpNames() {
		if i > 0 && i <= len(match) {
			paramsMap[name] = match[i]
		}
	}

	return paramsMap["line"]
}

func (p *Parser) extractYear(link string) (int, error) {
	match := p.reYear.FindStringSubmatch(link)
	if len(match) == 0 {
		return 0, fmt.Errorf("failed parse year. cannot extract year from url")
	}
	year, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, fmt.Errorf("failed parse year. cannot convert year to int")
	}
	return year, nil
}
