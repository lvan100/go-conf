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
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/lvan100/go-conf/internal/util"
)

var (
	Validator ValidatorInterface
)

// ValidatorInterface is the interface for validating a field.
type ValidatorInterface interface {
	Name() string
	Field(tag string, i interface{}) error
}

var (
	ErrNotExist      = errors.New("not exist")
	ErrInvalidSyntax = errors.New("invalid syntax")
)

// ParsedTag a value tag includes at most three parts: required key, optional
// default value, and optional splitter, the syntax is ${key:=value}>>splitter.
type ParsedTag struct {
	Key      string // short property key
	Def      string // default value
	HasDef   bool   // has default value
	Splitter string // splitter's name
}

func (tag ParsedTag) String() string {
	var sb strings.Builder
	sb.WriteString("${")
	sb.WriteString(tag.Key)
	if tag.HasDef {
		sb.WriteString(":=")
		sb.WriteString(tag.Def)
	}
	sb.WriteString("}")
	if tag.Splitter != "" {
		sb.WriteString(">>")
		sb.WriteString(tag.Splitter)
	}
	return sb.String()
}

// ParseTag parses a value tag, returns its key, and default value, and splitter.
func ParseTag(tag string) (ret ParsedTag, err error) {
	i := strings.LastIndex(tag, ">>")
	if i == 0 {
		err = fmt.Errorf("parse tag '%s' error: %w", tag, ErrInvalidSyntax)
		return
	}
	j := strings.LastIndex(tag, "}")
	if j <= 0 {
		err = fmt.Errorf("parse tag '%s' error: %w", tag, ErrInvalidSyntax)
		return
	}
	k := strings.Index(tag, "${")
	if k < 0 {
		err = fmt.Errorf("parse tag '%s' error: %w", tag, ErrInvalidSyntax)
		return
	}
	if i > j {
		ret.Splitter = strings.TrimSpace(tag[i+2:])
	}
	ss := strings.SplitN(tag[k+2:j], ":=", 2)
	ret.Key = ss[0]
	if len(ss) > 1 {
		ret.HasDef = true
		ret.Def = ss[1]
	}
	return
}

type BindParam struct {
	Key      string            // full key
	Path     string            // full path
	Tag      ParsedTag         // parsed tag
	Validate reflect.StructTag // full field tag
}

func (param *BindParam) BindTag(tag string, validate reflect.StructTag) error {
	parsedTag, err := ParseTag(tag)
	if err != nil {
		return err
	}
	if parsedTag.Key == "ROOT" {
		parsedTag.Key = ""
	} else if parsedTag.Key == "" {
		parsedTag.Key = "ANONYMOUS"
	}
	param.Tag = parsedTag
	if param.Key == "" {
		param.Key = parsedTag.Key
	} else if parsedTag.Key != "" {
		param.Key = param.Key + "." + parsedTag.Key
	}
	param.Validate = validate
	return nil
}

type Filter func(i interface{}, param BindParam) (bool, error)

// BindValue binds properties to a value.
func BindValue(p *Properties, v reflect.Value, t reflect.Type, param BindParam, filter Filter) (RetErr error) {

	if !IsValueType(t) {
		err := errors.New("target should be value type")
		return fmt.Errorf("%s: bind %s error, %w", util.FileLine(), param.Path, err)
	}

	defer func() {
		if RetErr == nil {
			tag, ok := param.Validate.Lookup(Validator.Name())
			if ok && len(tag) > 0 {
				err := Validator.Field(tag, v.Interface())
				if err != nil {
					RetErr = err
				}
			}
		}
	}()

	switch v.Kind() {
	case reflect.Map:
		return bindMap(p, v, t, param, filter)
	case reflect.Slice:
		return bindSlice(p, v, t, param, filter)
	case reflect.Array:
		err := errors.New("use slice instead of array")
		return fmt.Errorf("%s: bind %s error, %w", util.FileLine(), param.Path, err)
	default: // for linter
	}

	fn := converters[t]
	if fn == nil && v.Kind() == reflect.Struct {
		if err := bindStruct(p, v, t, param, filter); err != nil {
			return fmt.Errorf("%s: bind %s error, %w", util.FileLine(), param.Path, err)
		}
		return nil
	}

	val, err := resolve(p, param)
	if err != nil {
		return fmt.Errorf("%s: bind %s error, %w", util.FileLine(), param.Path, err)
	}

	if fn != nil {
		fnValue := reflect.ValueOf(fn)
		out := fnValue.Call([]reflect.Value{reflect.ValueOf(val)})
		if !out[1].IsNil() {
			err = out[1].Interface().(error)
			return fmt.Errorf("%s: bind %s error, %w", util.FileLine(), param.Path, err)
		}
		v.Set(out[0])
		return nil
	}

	switch v.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var u uint64
		if u, err = strconv.ParseUint(val, 0, 0); err == nil {
			v.SetUint(u)
			return nil
		}
		return fmt.Errorf("%s: bind %s error, %w", util.FileLine(), param.Path, err)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var i int64
		if i, err = strconv.ParseInt(val, 0, 0); err == nil {
			v.SetInt(i)
			return nil
		}
		return fmt.Errorf("%s: bind %s error, %w", util.FileLine(), param.Path, err)
	case reflect.Float32, reflect.Float64:
		var f float64
		if f, err = strconv.ParseFloat(val, 64); err == nil {
			v.SetFloat(f)
			return nil
		}
		return fmt.Errorf("%s: bind %s error, %w", util.FileLine(), param.Path, err)
	case reflect.Bool:
		var b bool
		if b, err = strconv.ParseBool(val); err == nil {
			v.SetBool(b)
			return nil
		}
		return fmt.Errorf("%s: bind %s error, %w", util.FileLine(), param.Path, err)
	case reflect.String:
		v.SetString(val)
		return nil
	default: // for linter
	}

	err = fmt.Errorf("unsupported bind type %q", t.String())
	return fmt.Errorf("%s: bind %s error, %w", util.FileLine(), param.Path, err)
}

// bindSlice binds properties to a slice value.
func bindSlice(p *Properties, v reflect.Value, t reflect.Type, param BindParam, filter Filter) error {

	et := t.Elem()
	p, err := getSlice(p, et, param)
	if err != nil {
		return fmt.Errorf("%s: bind %s error, %w", util.FileLine(), param.Path, err)
	}

	slice := reflect.MakeSlice(t, 0, 0)
	defer func() { v.Set(slice) }()

	if p == nil {
		return nil
	}

	for i := 0; ; i++ {
		e := reflect.New(et).Elem()
		subParam := BindParam{
			Key:  fmt.Sprintf("%s[%d]", param.Key, i),
			Path: fmt.Sprintf("%s[%d]", param.Path, i),
		}
		err = BindValue(p, e, et, subParam, filter)
		if errors.Is(err, ErrNotExist) {
			break
		}
		if err != nil {
			return fmt.Errorf("%s: bind %s error, %w", util.FileLine(), param.Path, err)
		}
		slice = reflect.Append(slice, e)
	}
	return nil
}

func getSlice(p *Properties, et reflect.Type, param BindParam) (*Properties, error) {

	// properties that defined as list.
	if p.Has(param.Key + "[0]") {
		return p, nil
	}

	// properties that defined as string and needs to split into []string.
	var strVal string
	{
		if p.Has(param.Key) {
			strVal = p.Get(param.Key)
		} else {
			if !param.Tag.HasDef {
				return nil, fmt.Errorf("%s: property %q %w", util.FileLine(), param.Key, ErrNotExist)
			}
			if param.Tag.Def == "" {
				return nil, nil
			}
			if !IsPrimitiveValueType(et) && converters[et] == nil {
				return nil, fmt.Errorf("%s: can't find converter for %s", util.FileLine(), et.String())
			}
			strVal = param.Tag.Def
		}
	}
	if strVal == "" {
		return nil, nil
	}

	var (
		err    error
		arrVal []string
	)

	if s := param.Tag.Splitter; s == "" {
		arrVal = strings.Split(strVal, ",")
		for i := range arrVal {
			arrVal[i] = strings.TrimSpace(arrVal[i])
		}
	} else if fn, ok := splitters[s]; ok && fn != nil {
		if arrVal, err = fn(strVal); err != nil {
			return nil, fmt.Errorf("%s: split error: %w, value: %q", util.FileLine(), err, strVal)
		}
	} else {
		return nil, fmt.Errorf("%s: unknown splitter %q", util.FileLine(), s)
	}

	p = New()
	for i, s := range arrVal {
		k := fmt.Sprintf("%s[%d]", param.Key, i)
		_ = p.store(k, s)
	}
	return p, nil
}

// bindMap binds properties to a map value.
func bindMap(p *Properties, v reflect.Value, t reflect.Type, param BindParam, filter Filter) error {

	if param.Tag.HasDef && param.Tag.Def != "" {
		err := errors.New("map can't have a non-empty default value")
		return fmt.Errorf("%s: bind %s error, %w", util.FileLine(), param.Path, err)
	}

	et := t.Elem()
	ret := reflect.MakeMap(t)
	defer func() { v.Set(ret) }()

	keys, err := p.storage.SubKeys(param.Key)
	if err != nil {
		return fmt.Errorf("%s: bind %s error, %w", util.FileLine(), param.Path, err)
	}

	for _, key := range keys {
		e := reflect.New(et).Elem()
		subKey := key
		if param.Key != "" {
			subKey = param.Key + "." + key
		}
		subParam := BindParam{
			Key:  subKey,
			Path: param.Path,
		}
		err = BindValue(p, e, et, subParam, filter)
		if err != nil {
			return fmt.Errorf("%s: bind %s error, %w", util.FileLine(), param.Path, err)
		}
		ret.SetMapIndex(reflect.ValueOf(key), e)
	}
	return nil
}

// bindStruct binds properties to a struct value.
func bindStruct(p *Properties, v reflect.Value, t reflect.Type, param BindParam, filter Filter) error {

	if param.Tag.HasDef && param.Tag.Def != "" {
		err := errors.New("struct can't have a non-empty default value")
		return fmt.Errorf("%s: bind %s error, %w", util.FileLine(), param.Path, err)
	}

	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i)
		fv := v.Field(i)

		if !fv.CanInterface() {
			continue
		}

		subParam := BindParam{
			Key:  param.Key,
			Path: param.Path + "." + ft.Name,
		}

		if tag, ok := ft.Tag.Lookup("value"); ok {
			if err := subParam.BindTag(tag, ft.Tag); err != nil {
				return fmt.Errorf("%s: bind %s error, %w", util.FileLine(), param.Path, err)
			}
			if filter != nil {
				ret, err := filter(fv.Addr().Interface(), subParam)
				if err != nil {
					return err
				}
				if ret {
					continue
				}
			}
			if err := BindValue(p, fv, ft.Type, subParam, filter); err != nil {
				return fmt.Errorf("%s: bind %s error, %w", util.FileLine(), param.Path, err)
			}
			continue
		}

		if ft.Anonymous {
			// embed pointer type may lead to infinite recursion.
			if ft.Type.Kind() != reflect.Struct {
				continue
			}
			if err := bindStruct(p, fv, ft.Type, subParam, filter); err != nil {
				return fmt.Errorf("%s: bind %s error, %w", util.FileLine(), param.Path, err)
			}
			continue
		}

		if IsValueType(ft.Type) {
			if subParam.Key == "" {
				subParam.Key = ft.Name
			} else {
				subParam.Key = subParam.Key + "." + ft.Name
			}
			subParam.Key = strings.ToLower(subParam.Key)
			subParam.Key = strings.ReplaceAll(subParam.Key, "_", ".")
			if err := BindValue(p, fv, ft.Type, subParam, filter); err != nil {
				return fmt.Errorf("%s: bind %s error, %w", util.FileLine(), param.Path, err)
			}
		}
	}
	return nil
}

// resolve returns property references processed property value.
func resolve(p *Properties, param BindParam) (string, error) {
	if val, ok := p.storage.Get(param.Key); ok {
		return resolveString(p, val)
	}
	if p.storage.Has(param.Key) {
		err := fmt.Errorf("property %q isn't simple value", param.Key)
		return "", fmt.Errorf("%s: resolve property %q error, %w", util.FileLine(), param.Key, err)
	}
	if param.Tag.HasDef {
		return resolveString(p, param.Tag.Def)
	}
	err := fmt.Errorf("property %q %w", param.Key, ErrNotExist)
	return "", fmt.Errorf("%s: resolve property %q error, %w", util.FileLine(), param.Key, err)
}

// resolveString returns property references processed string.
func resolveString(p *Properties, s string) (string, error) {

	var (
		length = len(s)
		count  = 0
		start  = -1
		end    = -1
	)

	for i := 0; i < length; i++ {
		if s[i] == '$' {
			if i < length-1 && s[i+1] == '{' {
				if count == 0 {
					start = i
				}
				count++
			}
		} else if s[i] == '}' {
			if count > 0 {
				count--
				if count == 0 {
					end = i
					break
				}
			}
		}
	}

	if start < 0 {
		return s, nil
	}

	if end < 0 || count > 0 {
		err := ErrInvalidSyntax
		return "", fmt.Errorf("%s: resolve string %q error, %w", util.FileLine(), s, err)
	}

	var param BindParam
	_ = param.BindTag(s[start:end+1], "")

	s1, err := resolve(p, param)
	if err != nil {
		return "", fmt.Errorf("%s: resolve string %q error, %w", util.FileLine(), s, err)
	}

	s2, err := resolveString(p, s[end+1:])
	if err != nil {
		return "", fmt.Errorf("%s: resolve string %q error, %w", util.FileLine(), s, err)
	}

	return s[:start] + s1 + s2, nil
}
