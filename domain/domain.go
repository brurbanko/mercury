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

import "time"

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
