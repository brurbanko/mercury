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

package config

import "github.com/cristalhq/aconfig"

// Config of service
type Config struct {
	Debug   bool   `env:"DEBUG"`
	TempDir string `env:"TMP_DIR"`
	Crawler struct {
		Domain    string `env:"DOMAIN" default:"bga32.ru"`
		UserAgent string `env:"USERAGENT" default:"urbanist-public-hearings (https://t.me/public_bryansk_bot)"`
	}
	Server struct {
		Host string `env:"HOST"`
		Port string `env:"PORT" default:"8080"`
	}
	Database struct {
		DSN string `env:"DATABASE_DSN" default:"database"`
	}
}

// Load tries to load config from env
func Load() (*Config, error) {
	var cfg Config
	loader := aconfig.LoaderFor(&cfg, aconfig.Config{
		SkipFlags: true,
		SkipFiles: true,
	})

	if err := loader.Load(); err != nil {
		return nil, err
	}

	return &cfg, nil
}
