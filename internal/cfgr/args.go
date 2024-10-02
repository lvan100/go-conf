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
	"errors"
	"os"
	"strings"

	"github.com/lvan100/go-conf/internal/conf"
)

// CommandArgs command-line parameters
type CommandArgs struct {
	option  string
	cmdArgs []string
}

func NewCommandArgs() *CommandArgs {
	return &CommandArgs{
		option:  "-D",
		cmdArgs: os.Args[1:],
	}
}

func (c *CommandArgs) Reset(args []string) {
	c.cmdArgs = args
}

func (c *CommandArgs) SetOption(option string) {
	c.option = option
}

// copyTo loads parameters passed in the form of -D key[=value/true].
func (c *CommandArgs) copyTo(out *conf.Properties) error {
	n := len(c.cmdArgs)
	for i := 0; i < n; i++ {
		s := c.cmdArgs[i]
		if s == c.option {
			if i >= n-1 {
				return errors.New("cmd option -D needs arg")
			}
			next := c.cmdArgs[i+1]
			ss := strings.SplitN(next, "=", 2)
			if len(ss) == 1 {
				ss = append(ss, "true")
			}
			if err := out.Set(ss[0], ss[1]); err != nil {
				return err
			}
		}
	}
	return nil
}
