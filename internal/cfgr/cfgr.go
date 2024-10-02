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
	"path/filepath"

	"github.com/lvan100/go-conf/internal/conf"
)

// ReadOnlyProperties is the interface for read-only properties.
type ReadOnlyProperties interface {

	// Data returns key-value pairs of the properties.
	Data() map[string]string

	// Keys returns keys of the properties.
	Keys() []string

	// Has returns whether the key exists.
	Has(key string) bool

	// Get returns key's value, using Def to return a default value.
	Get(key string, opts ...conf.GetOption) string

	// Resolve resolves string that contains references.
	Resolve(s string) (string, error)

	// Bind binds properties into a value.
	Bind(i interface{}, args ...conf.BindArg) error
}

/******************************* Configuration *******************************/

// Configuration is a layered configuration manager.
// prop - used to store the properties set by code,
// file - used to load properties from local static files,
// env - used to load properties from environment variables,
// args - used to load properties from command line arguments,
// dync - used to load properties from remote dynamic sources.
type Configuration struct {
	prop *conf.Properties
	file *PropertySources
	env  *Environment
	args *CommandArgs
	dync *PropertySources
}

func New() *Configuration {
	return &Configuration{
		prop: conf.New(),
		file: NewPropertySources(),
		env:  NewEnvironment(),
		args: NewCommandArgs(),
		dync: NewPropertySources(),
	}
}

// SetWorkDir sets the working directory.
func (c *Configuration) SetWorkDir(dir string) {
	c.file.workDir = dir
	c.dync.workDir = dir
}

// SetProperty sets a property that will be stored in the prop layer.
func (c *Configuration) SetProperty(key string, val interface{}) error {
	return c.prop.Set(key, val)
}

func (c *Configuration) File() *PropertySources {
	return c.file
}

func (c *Configuration) Env() *Environment {
	return c.env
}

func (c *Configuration) Args() *CommandArgs {
	return c.args
}

func (c *Configuration) Dync() *PropertySources {
	return c.dync
}

func merge(p *conf.Properties, sources ...interface {
	copyTo(out *conf.Properties) error
}) (*conf.Properties, error) {
	for _, s := range sources {
		if err := s.copyTo(p); err != nil {
			return nil, err
		}
	}
	return p, nil
}

// Refresh merges all layers and returned as a read-only properties.
func (c *Configuration) Refresh() (ReadOnlyProperties, error) {
	return merge(c.prop.Copy(), c.file, c.env, c.args, c.dync)
}

/****************************** PropertySources ******************************/

// PropertySources is a collection of property locations.
type PropertySources struct {
	workDir   string
	locations [][]string
}

func NewPropertySources() *PropertySources {
	return &PropertySources{}
}

func (p *PropertySources) Add(location ...string) {
	p.locations = append(p.locations, location)
}

// Clear removes all locations.
func (p *PropertySources) Clear() {
	p.locations = nil
}

// copyTo copies properties from the current layer to the output.
func (p *PropertySources) copyTo(out *conf.Properties) error {
	workDir := p.workDir
	if workDir == "" {
		workDir, _ = os.Getwd()
	}
	for _, ss := range p.locations {
		for _, s := range ss {
			// resolve filename that maybe contains references
			filename, err := out.Resolve(s)
			if err != nil {
				return err
			}
			if !filepath.IsAbs(filename) {
				filename = filepath.Join(workDir, filename)
			}
			c, err := conf.Load(filename)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return err
			}
			keys := c.Keys()
			for _, key := range keys {
				err = out.Set(key, c.Get(key))
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
