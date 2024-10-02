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

package conf

import (
	"strings"
	"time"

	"github.com/spf13/cast"

	"github.com/lvan100/go-conf/internal/cfgr"
	"github.com/lvan100/go-conf/internal/conf"
	"github.com/lvan100/go-conf/internal/expr"
	"github.com/lvan100/go-conf/reader/json"
	"github.com/lvan100/go-conf/reader/prop"
	"github.com/lvan100/go-conf/reader/toml"
	"github.com/lvan100/go-conf/reader/yaml"
)

func init() {
	SetValidator(&expr.Validator{})

	RegisterReader(json.Read, ".json")
	RegisterReader(prop.Read, ".properties")
	RegisterReader(yaml.Read, ".yaml", ".yml")
	RegisterReader(toml.Read, ".toml", ".tml")

	RegisterConverter(func(s string) (time.Time, error) {
		return cast.ToTimeE(strings.TrimSpace(s))
	})

	RegisterConverter(func(s string) (time.Duration, error) {
		return time.ParseDuration(strings.TrimSpace(s))
	})
}

/****************************** conf.Properties ******************************/

var (
	ErrNotExist      = conf.ErrNotExist
	ErrInvalidSyntax = conf.ErrInvalidSyntax
)

type (
	Reader    = conf.Reader
	Splitter  = conf.Splitter
	Converter = conf.Converter
)

// RegisterReader registers its Reader for some kind of file extension.
func RegisterReader(r Reader, ext ...string) {
	conf.RegisterReader(r, ext...)
}

// RegisterSplitter registers a Splitter and named it.
func RegisterSplitter(name string, fn Splitter) {
	conf.RegisterSplitter(name, fn)
}

// RegisterConverter registers its converter for non-primitive type such as
// time.Time, time.Duration, or other user-defined value type.
func RegisterConverter(fn Converter) {
	conf.RegisterConverter(fn)
}

type (
	ValidatorInterface = conf.ValidatorInterface
)

// SetValidator sets the validator.
func SetValidator(i ValidatorInterface) {
	conf.Validator = i
}

// Def used to set default value for conf.Get().
func Def(v string) conf.GetOption {
	return conf.Def(v)
}

// Key binds properties using a key for conf.Bind().
func Key(key string) conf.BindArg {
	return conf.Key(key)
}

// Tag binds properties using a tag for conf.Bind().
func Tag(tag string) conf.BindArg {
	return conf.Tag(tag)
}

type BindParam = conf.BindParam

// Param binds properties using BindParam for conf.Bind().
func Param(param BindParam) conf.BindArg {
	return conf.Param(param)
}

type ParsedTag = conf.ParsedTag

func ParseTag(tag string) (ret ParsedTag, err error) {
	return conf.ParseTag(tag)
}

/**************************** cfgr.Configuration *****************************/

// Configuration is a layered configuration manager.
type Configuration = cfgr.Configuration

func NewConfiguration() *Configuration {
	return cfgr.New()
}

// ReadOnlyProperties is the interface for read-only properties.
type ReadOnlyProperties = cfgr.ReadOnlyProperties
