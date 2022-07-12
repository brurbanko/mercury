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

package crawler

import (
	"fmt"

	"github.com/gocolly/colly/v2"
)

// Crawler is main package struct for parsing passed URL
type Crawler struct {
	collector *colly.Collector
}

// New returns Crawler instance
func New(domain, userAgent string) *Crawler {
	col := colly.NewCollector(
		colly.AllowedDomains(domain),
		colly.MaxDepth(1),
		colly.AllowURLRevisit(),
		colly.UserAgent(userAgent),
	)

	return &Crawler{
		collector: col,
	}
}

// ExtractLinks return all links from passed selector (GetAllLinks)
func (c *Crawler) ExtractLinks(url, querySelector string) ([]string, error) {
	links := make([]string, 0)
	c.collector.OnHTML(querySelector, func(e *colly.HTMLElement) {
		links = append(links, e.Attr("href"))
	})

	err := c.collector.Visit(url)
	c.collector.Wait()

	return links, err
}

// ExtractContent returns content of passed link (GetSelectorContent)
func (c *Crawler) ExtractContent(url, querySelector string) ([]string, error) {
	content := make([]string, 0)
	// On every an element which has href attribute call callback
	c.collector.OnHTML(querySelector, func(e *colly.HTMLElement) {
		// Paragraph content length is greater 2 bytes
		if len(e.Text) > 2 {
			content = append(content, e.Text)
		}
	})

	// Set error handler
	c.collector.OnError(func(r *colly.Response, err error) {
		fmt.Printf("request to url %s failed with status code %d. \nerror: %v\n", r.Request.URL, r.StatusCode, err)
	})

	err := c.collector.Visit(url)
	c.collector.Wait()

	return content, err
}
