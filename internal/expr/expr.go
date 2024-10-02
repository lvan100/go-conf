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

package expr

import (
	"fmt"

	"github.com/expr-lang/expr"
)

type Validator struct{}

// Name returns the name of the validator.
func (d *Validator) Name() string {
	return "expr"
}

// Field validates the field with the given tag and value.
func (d *Validator) Field(tag string, i interface{}) error {
	r, err := expr.Eval(tag, map[string]interface{}{"$": i})
	if err != nil {
		return fmt.Errorf("eval %q returns error, %w", tag, err)
	}
	ret, ok := r.(bool)
	if !ok {
		return fmt.Errorf("eval %q doesn't return bool value", tag)
	}
	if !ret {
		return fmt.Errorf("validate failed on %q for value %v", tag, i)
	}
	return nil
}
