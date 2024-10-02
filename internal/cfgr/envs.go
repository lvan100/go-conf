// Copyright 2024 github.com/lvan100
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cfgr

import (
	"os"
	"regexp"
	"strings"

	"github.com/lvan100/go-conf/internal/conf"
)

const (
	IncludeEnvPatterns = "INCLUDE_ENV_PATTERNS"
	ExcludeEnvPatterns = "EXCLUDE_ENV_PATTERNS"
)

// Environment environment variable
type Environment struct {
	prefix  string
	environ []string
}

func NewEnvironment() *Environment {
	return &Environment{
		prefix:  "GS_",
		environ: os.Environ(),
	}
}

func (c *Environment) Reset(environ []string) {
	c.environ = environ
}

func (c *Environment) SetPrefix(prefix string) {
	c.prefix = prefix
}

func (c *Environment) lookupEnv(key string) (value string, found bool) {
	key = strings.TrimSpace(key) + "="
	for _, s := range c.environ {
		if strings.HasPrefix(s, key) {
			v := strings.TrimPrefix(s, key)
			return strings.TrimSpace(v), true
		}
	}
	return "", false
}

// copyTo add environment variables that matches IncludeEnvPatterns and
// exclude environment variables that matches ExcludeEnvPatterns.
func (c *Environment) copyTo(p *conf.Properties) error {

	toRex := func(patterns []string) ([]*regexp.Regexp, error) {
		var rex []*regexp.Regexp
		for _, v := range patterns {
			exp, err := regexp.Compile(v)
			if err != nil {
				return nil, err
			}
			rex = append(rex, exp)
		}
		return rex, nil
	}

	includes := []string{".*"}
	if s, ok := c.lookupEnv(IncludeEnvPatterns); ok {
		includes = strings.Split(s, ",")
	}
	includeRex, err := toRex(includes)
	if err != nil {
		return err
	}

	var excludes []string
	if s, ok := c.lookupEnv(ExcludeEnvPatterns); ok {
		excludes = strings.Split(s, ",")
	}
	excludeRex, err := toRex(excludes)
	if err != nil {
		return err
	}

	matches := func(rex []*regexp.Regexp, s string) bool {
		for _, r := range rex {
			if r.MatchString(s) {
				return true
			}
		}
		return false
	}

	for _, env := range c.environ {
		ss := strings.SplitN(env, "=", 2)
		k, v := ss[0], ""
		if len(ss) > 1 {
			v = ss[1]
		}

		var propKey string
		if strings.HasPrefix(k, c.prefix) {
			propKey = strings.TrimPrefix(k, c.prefix)
		} else if matches(includeRex, k) && !matches(excludeRex, k) {
			propKey = k
		} else {
			continue
		}

		propKey = strings.ReplaceAll(propKey, "_", ".")
		propKey = strings.ToLower(propKey)
		if err = p.Set(propKey, v); err != nil {
			return err
		}
	}
	return nil
}
