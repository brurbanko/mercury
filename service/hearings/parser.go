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
	"errors"
	"regexp"
	"strconv"
	"time"

	"github.com/brurbanko/mercury/domain"
)

var spaces = `[\t\n\v\f\r\p{Zs}]`
var topicStartParagraph = spaces + "состоятся" + spaces + "+(публичные" + spaces + "+слушания" + spaces + "+)?"
var topicEndParagraph = "(?:Экспозиция" + spaces + "+проект|Участник)"
var proposalParagraph = "^При[её]м" + spaces
var timeAndPlace = `(?P<day>\d+)` + spaces + `+(?P<month>\p{L}+)(?:` + spaces + `+(?P<year>\d+)` + spaces + "+года)?" + spaces + "+в" + spaces + `+(?P<hours>\d+)[\.:](?P<minutes>\d+)` + spaces + `+в` + spaces + `(?P<place>.*)`
var clearLine = `^\p{Zs}*?[-—]?\p{Zs}*?(?P<line>.*)\p{Zs}*?[\.;]+?\p{Zs}*?$`

var serviceTimeLocation = time.Now().Location()
var beginnigTime = time.Date(2021, time.May, 1, 0, 0, 0, 0, serviceTimeLocation)

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
	reTS *regexp.Regexp
	reTE *regexp.Regexp
	reTP *regexp.Regexp
	reCL *regexp.Regexp
	rePP *regexp.Regexp
}

// NewParser return instance of public hearings parser
func NewParser() *Parser {
	return &Parser{
		reTS: regexp.MustCompile(topicStartParagraph),
		reTE: regexp.MustCompile(topicEndParagraph),
		reTP: regexp.MustCompile(timeAndPlace),
		reCL: regexp.MustCompile(clearLine),
		rePP: regexp.MustCompile(proposalParagraph),
	}
}

// Content return full data of public hearing
func (p *Parser) Content(content []string) (*domain.Hearing, error) {
	return p.prepare(content)
}

func (p *Parser) prepare(content []string) (*domain.Hearing, error) {
	ph := &domain.Hearing{}

	if len(content) == 0 {
		return ph, errors.New("empty content")
	}

	/* DEFINE TOPIC */
	start, next := p.defineTopicsParagraphs(content)

	// All data in one paragraph
	if start+1 == next {
		// Split content to parts: 1-st -- time and place, 2-nd -- topic
		parts := p.reTS.Split(content[start], -1)
		if len(parts) != 2 {
			return ph, errors.New("failed parse content. cannot split to time/place and topic")
		}
		top := p.clearString(parts[1])
		if top != "" {
			ph.Topic = append(ph.Topic, top)
		}
		ph.Place = parts[0]
	} else {
		// Some topics in different paragraphs
		for i := start; i < next; i++ {
			// First paragraph with date and place
			if i == start {
				parts := p.reTS.Split(content[i], -1)
				if len(parts) == 0 {
					return ph, errors.New("failed parse content. cannot split to time and place")
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
		return ph, errors.New("failed parse content. cannot get topics")
	}

	/* DEFINE PLACE AND TIME */
	if ph.Place != "" {
		match := p.reTP.FindStringSubmatch(ph.Place)

		paramsMap := make(map[string]string)
		for i, name := range p.reTP.SubexpNames() {
			if i > 0 && i <= len(match) {
				paramsMap[name] = match[i]
			}
		}

		year, err := strconv.Atoi(paramsMap["year"])
		if err != nil {
			year = time.Now().Year()
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
		// Replace place
		ph.Place = paramsMap["place"]
	} else {
		return ph, errors.New("failed parse date and place. empty string")
	}

	if ph.Time.Before(beginnigTime) {
		return ph, errors.New("failed parse date. raw strings with date: '" + ph.Place + "'")
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
		if p.reTS.MatchString(paragraph) {
			start = i
		}
		if p.reTE.MatchString(paragraph) {
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
		if p.rePP.MatchString(paragraph) {
			res = append(res, i)
		}
	}
	return
}

func (p *Parser) clearString(str string) string {
	match := p.reCL.FindStringSubmatch(str)

	paramsMap := make(map[string]string)
	for i, name := range p.reCL.SubexpNames() {
		if i > 0 && i <= len(match) {
			paramsMap[name] = match[i]
		}
	}

	return paramsMap["line"]
}
