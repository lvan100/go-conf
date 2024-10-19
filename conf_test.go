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

package conf_test

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/lvan100/go-conf"
	"github.com/lvan100/go-conf/internal/flat"
)

func TestConverter(t *testing.T) {
	defer func() {
		r := recover()
		expect := "converter is func(string)(type,error)"
		if fmt.Sprint(r) != expect {
			t.Fatal(r)
		}
	}()
	conf.RegisterConverter(func() {})
}

func TestParseTag(t *testing.T) {

	{
		_, gotErr := conf.ParseTag("")
		expectErr := errors.New("parse tag '' error: invalid syntax")
		if fmt.Sprint(gotErr) != expectErr.Error() {
			t.Fatalf("got %v, expect %v", gotErr, expectErr)
		}
	}

	{
		_, gotErr := conf.ParseTag("{}")
		expectErr := errors.New("parse tag '{}' error: invalid syntax")
		if fmt.Sprint(gotErr) != expectErr.Error() {
			t.Fatalf("got %v, expect %v", gotErr, expectErr)
		}
	}

	{
		_, gotErr := conf.ParseTag(">>point")
		expectErr := errors.New("parse tag '>>point' error: invalid syntax")
		if fmt.Sprint(gotErr) != expectErr.Error() {
			t.Fatalf("got %v, expect %v", gotErr, expectErr)
		}
	}

	{
		gotTag, err := conf.ParseTag(`${points:=(1,2)|(3,4)|(5,6)}>>point`)
		if err != nil {
			t.Fatal(err)
		}
		expectTag := conf.ParsedTag{
			Def:      "(1,2)|(3,4)|(5,6)",
			HasDef:   true,
			Key:      "points",
			Splitter: "point",
		}
		if gotTag.String() != expectTag.String() {
			t.Fatalf("got %v, expect %v", gotTag, expectTag)
		}
	}
}

func TestStorage(t *testing.T) {

	{
		testcase := func(t *testing.T, fn func(t *testing.T, c *conf.Configuration), expect map[string]interface{}) {
			c := conf.NewConfiguration()
			c.Env().Reset([]string{})
			fn(t, c)
			p, err := c.Refresh()
			if err != nil {
				t.Fatal(err)
			}
			gotData := p.Data()
			expectData := flat.FlattenMap(expect)
			if !reflect.DeepEqual(gotData, expectData) {
				t.Fatalf("got %v, expect %v", gotData, expectData)
			}
		}

		testcase(t, func(t *testing.T, c *conf.Configuration) {
			err := c.SetProperty("key", nil)
			if err != nil {
				t.Fatal(err)
			}
			err = c.SetProperty("key", "abc")
			if err != nil {
				t.Fatal(err)
			}
		}, map[string]interface{}{
			"key": "abc",
		})

		testcase(t, func(t *testing.T, c *conf.Configuration) {
			err := c.SetProperty("key", "123")
			if err != nil {
				t.Fatal(err)
			}
			err = c.SetProperty("key", "abc")
			if err != nil {
				t.Fatal(err)
			}
		}, map[string]interface{}{
			"key": "abc",
		})

		testcase(t, func(t *testing.T, c *conf.Configuration) {
			err := c.SetProperty("key", nil)
			if err != nil {
				t.Fatal(err)
			}
			err = c.SetProperty("key", []string{"abc"})
			if err != nil {
				t.Fatal(err)
			}
		}, map[string]interface{}{
			"key": []string{"abc"},
		})

		testcase(t, func(t *testing.T, c *conf.Configuration) {
			err := c.SetProperty("key", []interface{}{})
			if err != nil {
				t.Fatal(err)
			}
			err = c.SetProperty("key", []string{"abc"})
			if err != nil {
				t.Fatal(err)
			}
		}, map[string]interface{}{
			"key": []string{"abc"},
		})

		testcase(t, func(t *testing.T, c *conf.Configuration) {
			err := c.SetProperty("key", []interface{}{"123"})
			if err != nil {
				t.Fatal(err)
			}
			err = c.SetProperty("key[0]", "abc")
			if err != nil {
				t.Fatal(err)
			}
			err = c.SetProperty("key[1]", "xyz")
			if err != nil {
				t.Fatal(err)
			}
		}, map[string]interface{}{
			"key": []string{"abc", "xyz"},
		})

		testcase(t, func(t *testing.T, c *conf.Configuration) {
			err := c.SetProperty("key", nil)
			if err != nil {
				t.Fatal(err)
			}
			err = c.SetProperty("key", map[string]string{
				"val": "abc",
			})
			if err != nil {
				t.Fatal(err)
			}
		}, map[string]interface{}{
			"key": map[string]interface{}{
				"val": "abc",
			},
		})

		testcase(t, func(t *testing.T, c *conf.Configuration) {
			err := c.SetProperty("key", map[string]interface{}{})
			if err != nil {
				t.Fatal(err)
			}
			err = c.SetProperty("key", map[string]string{
				"val": "abc",
			})
			if err != nil {
				t.Fatal(err)
			}
		}, map[string]interface{}{
			"key": map[string]interface{}{
				"val": "abc",
			},
		})

		testcase(t, func(t *testing.T, c *conf.Configuration) {
			err := c.SetProperty("key", map[string]interface{}{
				"val": "abc",
			})
			if err != nil {
				t.Fatal(err)
			}
			err = c.SetProperty("key.val", "123")
			if err != nil {
				t.Fatal(err)
			}
		}, map[string]interface{}{
			"key": map[string]interface{}{
				"val": "123",
			},
		})
	}

	{
		c := conf.NewConfiguration()
		c.Env().Reset([]string{})
		err := c.SetProperty("key", "abc")
		if err != nil {
			t.Fatal(err)
		}
		{
			gotErr := c.SetProperty("key", []string{"abc"})
			expectErr := errors.New("property 'key' is a value but 'key[0]' wants other type")
			if fmt.Sprint(gotErr) != expectErr.Error() {
				t.Fatalf("got %v, expect %v", gotErr, expectErr)
			}
		}
		{
			gotErr := c.SetProperty("key", map[string]interface{}{
				"val": "abc",
			})
			expectErr := errors.New("property 'key' is a value but 'key.val' wants other type")
			if fmt.Sprint(gotErr) != expectErr.Error() {
				t.Fatalf("got %v, expect %v", gotErr, expectErr)
			}
		}
	}

	{
		c := conf.NewConfiguration()
		c.Env().Reset([]string{})
		err := c.SetProperty("key", []string{"abc"})
		if err != nil {
			t.Fatal(err)
		}
		{
			gotErr := c.SetProperty("key", "abc")
			expectErr := errors.New("property 'key' is an array but 'key' wants other type")
			if fmt.Sprint(gotErr) != expectErr.Error() {
				t.Fatalf("got %v, expect %v", gotErr, expectErr)
			}
		}
		{
			gotErr := c.SetProperty("key", map[string]interface{}{
				"val": "abc",
			})
			expectErr := errors.New("property 'key' is an array but 'key.val' wants other type")
			if fmt.Sprint(gotErr) != expectErr.Error() {
				t.Fatalf("got %v, expect %v", gotErr, expectErr)
			}
		}
	}

	{
		c := conf.NewConfiguration()
		c.Env().Reset([]string{})
		err := c.SetProperty("key", map[string]interface{}{
			"val": "abc",
		})
		if err != nil {
			t.Fatal(err)
		}
		{
			gotErr := c.SetProperty("key", "abc")
			expectErr := errors.New("property 'key' is a map but 'key' wants other type")
			if fmt.Sprint(gotErr) != expectErr.Error() {
				t.Fatalf("got %v, expect %v", gotErr, expectErr)
			}
		}
		{
			gotErr := c.SetProperty("key", []string{"abc"})
			expectErr := errors.New("property 'key' is a map but 'key[0]' wants other type")
			if fmt.Sprint(gotErr) != expectErr.Error() {
				t.Fatalf("got %v, expect %v", gotErr, expectErr)
			}
		}
	}
}

func TestProperties(t *testing.T) {
	c := conf.New()
	err := c.Load("testdata/conf.json")
	if err != nil {
		t.Fatal(err)
	}
	err = c.Load("testdata/conf.toml")
	if err != nil {
		t.Fatal(err)
	}
	got := c.Data()
	expect := flat.FlattenMap(map[string]interface{}{
		"toml": map[string]interface{}{
			"int": 1,
			"str": "abc",
			"arr": []string{"a", "b", "c"},
			"map": map[string]interface{}{
				"a": "1",
				"b": "2",
			},
			"empty_arr": []interface{}{},
			"empty_map": map[string]interface{}{},
		},
		"json": map[string]interface{}{
			"int": 1,
			"str": "abc",
			"arr": []string{"a", "b", "c"},
			"map": map[string]interface{}{
				"a": "1",
				"b": "2",
			},
			"empty_arr": []interface{}{},
			"empty_map": map[string]interface{}{},
		},
	})
	if !reflect.DeepEqual(got, expect) {
		t.Fatalf("got %v, expect %v", got, expect)
	}
}

func TestConfiguration(t *testing.T) {

	type Point struct {
		X int
		Y int
	}

	conf.RegisterSplitter("point", func(s string) ([]string, error) {
		if strings.HasPrefix(s, "error:") {
			return nil, errors.New(strings.TrimPrefix(s, "error:"))
		}
		return strings.Split(s, "|"), nil
	})

	conf.RegisterConverter(func(s string) (Point, error) {
		s = strings.TrimSpace(s)
		s = strings.TrimPrefix(s, "(")
		s = strings.TrimSuffix(s, ")")
		ss := strings.Split(s, ",")
		x, _ := strconv.Atoi(ss[0])
		y, _ := strconv.Atoi(ss[1])
		return Point{X: x, Y: y}, nil
	})

	c := conf.NewConfiguration()
	c.Env().Reset([]string{})

	{
		c.Args().Reset([]string{
			"-D",
			"args.int=1",
			"-D",
		})
		_, gotErr := c.Refresh()
		expectErr := errors.New("cmd option -D needs arg")
		if fmt.Sprint(gotErr) != expectErr.Error() {
			t.Fatalf("got %v, expect %v", gotErr, expectErr)
		}
		c.Args().Reset([]string{})
	}

	{
		c.Args().Reset([]string{
			"-D",
			"args.int=1",
			"-D",
			"args.int.sub=1",
		})
		_, gotErr := c.Refresh()
		expectErr := errors.New("property 'args.int' is a value but 'args.int.sub' wants other type")
		if fmt.Sprint(gotErr) != expectErr.Error() {
			t.Fatalf("got %v, expect %v", gotErr, expectErr)
		}
		c.Args().Reset([]string{})
	}

	{
		c.Env().Reset([]string{
			"INCLUDE_ENV_PATTERNS=(",
		})
		_, gotErr := c.Refresh()
		expectErr := errors.New("error parsing regexp: missing closing ): `(`")
		if fmt.Sprint(gotErr) != expectErr.Error() {
			t.Fatalf("got %v, expect %v", gotErr, expectErr)
		}
		c.Env().Reset([]string{})
	}

	{
		c.Env().Reset([]string{
			"EXCLUDE_ENV_PATTERNS=(",
		})
		_, gotErr := c.Refresh()
		expectErr := errors.New("error parsing regexp: missing closing ): `(`")
		if fmt.Sprint(gotErr) != expectErr.Error() {
			t.Fatalf("got %v, expect %v", gotErr, expectErr)
		}
		c.Env().Reset([]string{})
	}

	{
		c.Env().Reset([]string{
			"GS_ENVS_INT=1",
			"GS_ENVS_INT_SUB=1",
		})
		_, gotErr := c.Refresh()
		expectErr := errors.New("property 'envs.int' is a value but 'envs.int.sub' wants other type")
		if fmt.Sprint(gotErr) != expectErr.Error() {
			t.Fatalf("got %v, expect %v", gotErr, expectErr)
		}
		c.Env().Reset([]string{})
	}

	{
		c.Env().Reset([]string{
			"INCLUDE_ENV_PATTERNS=ENVS_.*",
			"ENVS_INT=1",
			"ENVS_INT_SUB=1",
		})
		_, gotErr := c.Refresh()
		expectErr := errors.New("property 'envs.int' is a value but 'envs.int.sub' wants other type")
		if fmt.Sprint(gotErr) != expectErr.Error() {
			t.Fatalf("got %v, expect %v", gotErr, expectErr)
		}
		c.Env().Reset([]string{})
	}

	{
		c.File().Add("testdata/conf-${dev}.yaml")
		_, gotErr := c.Refresh()
		if !errors.Is(gotErr, conf.ErrNotExist) {
			t.Fatalf("got %v, expect %v", gotErr, conf.ErrNotExist)
		}
		c.File().Clear()
	}

	{
		c.File().Add("testdata/invalid.json")
		_, gotErr := c.Refresh()
		expectErr := errors.New(`invalid character 't' looking for beginning of object key string`)
		if fmt.Sprint(gotErr) != expectErr.Error() {
			t.Fatalf("got %v, expect %v", gotErr, expectErr)
		}
		c.File().Clear()
	}

	{
		c.File().Add("testdata/invalid.toml")
		_, gotErr := c.Refresh()
		expectErr := errors.New(`(1, 6): was expecting token =, but got "is" instead`)
		if fmt.Sprint(gotErr) != expectErr.Error() {
			t.Fatalf("got %v, expect %v", gotErr, expectErr)
		}
		c.File().Clear()
	}

	{
		c.File().Add("testdata/invalid.yaml")
		_, gotErr := c.Refresh()
		expectErr := errors.New("yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `this is...` into map[string]interface {}")
		if fmt.Sprint(gotErr) != expectErr.Error() {
			t.Fatalf("got %v, expect %v", gotErr, expectErr)
		}
		c.File().Clear()
	}

	{
		c.File().Add("testdata/conf.toml", "testdata/invalid.properties")
		_, gotErr := c.Refresh()
		expectErr := errors.New("property 'toml.int' is a value but 'toml.int.sub' wants other type")
		if fmt.Sprint(gotErr) != expectErr.Error() {
			t.Fatalf("got %v, expect %v", gotErr, expectErr)
		}
		c.File().Clear()
	}

	{
		c.File().Add("testdata/conf.unknown_ext")
		_, gotErr := c.Refresh()
		expectErr := errors.New(`unsupported file type ".unknown_ext"`)
		if fmt.Sprint(gotErr) != expectErr.Error() {
			t.Fatalf("got %v, expect %v", gotErr, expectErr)
		}
		c.File().Clear()
	}

	{
		gotErr := c.SetProperty("", "")
		expectErr := errors.New("key is empty")
		if fmt.Sprint(gotErr) != expectErr.Error() {
			t.Fatalf("got %v, expect %v", gotErr, expectErr)
		}
	}

	Property := func(key string, val interface{}) {
		err := c.SetProperty(key, val)
		if err != nil {
			t.Fatal(err)
		}
	}

	Property("spring.active.profile", "online")
	Property("converter.time", "2024-10-01 00:00:00 +0800")

	Property("maps.a", "1")
	Property("maps.b", "a")

	{
		dir, _ := os.Getwd()
		c.SetWorkDir(dir)
	}

	c.File().Add("testdata/xxx.yaml")
	c.File().Clear()
	c.File().Add("testdata/conf.toml", "testdata/conf-${spring.active.profile}.yaml")
	c.File().Add("testdata/xxx.yaml", "testdata/conf.json")

	c.Env().SetPrefix("SK_")
	c.Env().Reset([]string{
		"INCLUDE_ENV_PATTERNS=^ENVS_.*",
		"EXCLUDE_ENV_PATTERNS=^ENVS_INT_SUB.*",
		"SK_ENVS_INT=1",
		"SK_ENVS_STR=abc",
		"SK_ENVS_ARR=a,b,c",
		"ENVS_MAP_A=1",
		"ENVS_MAP_B=2",
		"ENVS_INT_SUB=1",
	})

	c.Args().SetOption("-X")
	c.Args().Reset([]string{
		"-X",
		"args.int=1",
		"-X",
		"args.str=abc",
		"-X",
		"args.arr=a,b,c",
		"-X",
		"args.map.a=1",
		"-X",
		"args.map.b=2",
		"-X",
		"args.bool",
	})

	c.Dync().Add("testdata/xxx.yaml")
	c.Dync().Clear()
	c.Dync().Add("testdata-${spring.active.profile}/dync.properties")

	{
		err := os.RemoveAll("testdata-online")
		if err != nil {
			t.Fatal(err)
		}
		err = os.MkdirAll("testdata-online", os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}
		var data []byte
		data, err = os.ReadFile("testdata/dync-temp/dync.properties")
		if err != nil {
			t.Fatal(err)
		}
		err = os.WriteFile("testdata-online/dync.properties", data, os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}
	}

	var p conf.ReadOnlyProperties
	{
		var err error
		p, err = c.Refresh()
		if err != nil {
			t.Fatal(err)
		}
	}

	gotData := p.Data()
	expectData := flat.FlattenMap(map[string]interface{}{
		"spring.active.profile": "online",
		"converter": map[string]interface{}{
			"time": "2024-10-01 00:00:00 +0800",
		},
		"maps": map[string]interface{}{
			"a": "1",
			"b": "a",
		},
		"toml": map[string]interface{}{
			"int": 1,
			"str": "abc",
			"arr": []string{"a", "b", "c"},
			"map": map[string]interface{}{
				"a": "1",
				"b": "2",
			},
			"empty_arr": []interface{}{},
			"empty_map": map[string]interface{}{},
		},
		"yaml": map[string]interface{}{
			"int": 1,
			"str": "abc",
			"arr": []string{"a", "b", "c"},
			"map": map[string]interface{}{
				"a": "1",
				"b": "2",
			},
			"empty_arr": []interface{}{},
			"empty_map": map[string]interface{}{},
		},
		"prop": map[string]interface{}{
			"int": 1,
			"str": "abc",
			"arr": "a, b, c",
			"map": map[string]interface{}{
				"a": "1",
				"b": "2",
			},
			"empty_arr": []interface{}{},
			"empty_map": map[string]interface{}{},
		},
		"json": map[string]interface{}{
			"int": 1,
			"str": "abc",
			"arr": []string{"a", "b", "c"},
			"map": map[string]interface{}{
				"a": "1",
				"b": "2",
			},
			"empty_arr": []interface{}{},
			"empty_map": map[string]interface{}{},
		},
		"args": map[string]interface{}{
			"int": 1,
			"str": "abc",
			"arr": "a,b,c",
			"map": map[string]interface{}{
				"a": "1",
				"b": "2",
			},
			"bool": true,
		},
		"envs": map[string]interface{}{
			"int": 1,
			"str": "abc",
			"arr": "a,b,c",
			"map": map[string]interface{}{
				"a": "1",
				"b": "2",
			},
		},
	})

	if !maps.Equal(gotData, expectData) {
		t.Fatalf("got %v, expect %v", gotData, expectData)
	}

	{
		var gotTime time.Time
		err := p.Bind(&gotTime, conf.Key("converter.time"))
		if err != nil {
			t.Fatal(err)
		}

		location := time.FixedZone("CST+0800", 8*60*60)
		expectTime := time.Date(2024, 10, 1, 0, 0, 0, 0, location)
		if !gotTime.Equal(expectTime) {
			t.Fatalf("got %v, expect %v", gotTime, expectTime)
		}
	}

	{
		var gotDuration time.Duration
		err := p.Bind(&gotDuration, conf.Tag("${converter.duration:=12h}"))
		if err != nil {
			t.Fatal(err)
		}

		expectDuration := 12 * time.Hour
		if gotDuration != expectDuration {
			t.Fatalf("got %v, expect %v", gotDuration, expectDuration)
		}
	}

	{
		var gotTime time.Time
		gotErr := p.Bind(&gotTime, conf.Tag("${time:=123456789}"))
		expectErr := errors.New("bind Time error, unable to parse date: 123456789")
		if !strings.Contains(gotErr.Error(), expectErr.Error()) {
			t.Fatalf("got %v, expect %v", gotErr, expectErr)
		}
	}

	{
		var i int
		gotErr := p.Bind(&i, conf.Tag("${int:=abc}"))
		expectErr := errors.New(`parsing "abc": invalid syntax`)
		if !strings.Contains(gotErr.Error(), expectErr.Error()) {
			t.Fatalf("got %v, expect %v", gotErr, expectErr)
		}
	}

	{
		var u uint
		gotErr := p.Bind(&u, conf.Tag("${uint:=abc}"))
		expectErr := errors.New(`parsing "abc": invalid syntax`)
		if !strings.Contains(gotErr.Error(), expectErr.Error()) {
			t.Fatalf("got %v, expect %v", gotErr, expectErr)
		}
	}

	{
		var f float32
		gotErr := p.Bind(&f, conf.Tag("${float:=abc}"))
		expectErr := errors.New(`parsing "abc": invalid syntax`)
		if !strings.Contains(gotErr.Error(), expectErr.Error()) {
			t.Fatalf("got %v, expect %v", gotErr, expectErr)
		}
	}

	{
		var b bool
		gotErr := p.Bind(&b, conf.Tag("${bool:=abc}"))
		expectErr := errors.New(`parsing "abc": invalid syntax`)
		if !strings.Contains(gotErr.Error(), expectErr.Error()) {
			t.Fatalf("got %v, expect %v", gotErr, expectErr)
		}
	}

	{
		var x complex64
		gotErr := p.Bind(&x, conf.Tag("${number:=2+i}"))
		expectErr := errors.New("target should be value type")
		if !strings.Contains(gotErr.Error(), expectErr.Error()) {
			t.Fatalf("got %v, expect %v", gotErr, expectErr)
		}
	}

	bindArray := func(key string) {
		var gotArray []string
		err := p.Bind(&gotArray, conf.Key(key))
		if err != nil {
			t.Fatal(err)
		}
		expectArray := []string{"a", "b", "c"}
		if slices.Compare(gotArray, expectArray) != 0 {
			t.Fatalf("got %v, expect %v", gotArray, expectArray)
		}
	}

	bindArray("toml.arr")
	bindArray("yaml.arr")
	bindArray("prop.arr")
	bindArray("args.arr")
	bindArray("envs.arr")

	bindMap := func(tag string) {
		var gotMap map[string]string
		err := p.Bind(&gotMap, conf.Tag(tag))
		if err != nil {
			t.Fatal(err)
		}
		expectMap := map[string]string{
			"a": "1",
			"b": "2",
		}
		if !maps.Equal(gotMap, expectMap) {
			t.Fatalf("got %v, expect %v", gotMap, expectMap)
		}
	}

	bindMap("${toml.map}")
	bindMap("${yaml.map}")
	bindMap("${prop.map}")
	bindMap("${args.map}")
	bindMap("${envs.map}")

	bindObject := func(key string) {
		type Object struct {
			Int  int               `value:"${int}" expr:"$>0&&$<10"`
			Str  string            `value:"${str}"`
			Arr  []string          `value:"${arr}"`
			Map  map[string]string `value:"${map}"`
			Time time.Time         `value:"${time:=2024-10-01}"`
		}
		var gotObj Object
		err := p.Bind(&gotObj, conf.Key(key))
		if err != nil {
			t.Fatal(err)
		}
		expectObj := Object{
			Int:  1,
			Str:  "abc",
			Arr:  []string{"a", "b", "c"},
			Map:  map[string]string{"a": "1", "b": "2"},
			Time: time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC),
		}
		if !reflect.DeepEqual(gotObj, expectObj) {
			t.Fatalf("got %v, expect %v", gotObj, expectObj)
		}
	}

	bindObject("toml")
	bindObject("yaml")
	bindObject("prop")
	bindObject("args")
	bindObject("envs")

	{
		var obj struct {
			Int int `value:"${toml.int}" expr:"len($)"`
		}
		gotErr := p.Bind(&obj)
		expectStr := "invalid argument for len (type int)"
		if !strings.Contains(fmt.Sprint(gotErr), expectStr) {
			t.Fatalf("unexpect error %s", gotErr)
		}
	}

	{
		var obj struct {
			Str string `value:"${yaml.str}" expr:"len($)"`
		}
		gotErr := p.Bind(&obj)
		expectStr := `eval "len($)" doesn't return bool value`
		if !strings.Contains(fmt.Sprint(gotErr), expectStr) {
			t.Fatalf("unexpect error %s", gotErr)
		}
	}

	{
		var obj struct {
			Map map[string]int `value:"${prop.map}" expr:"len($)>3"`
		}
		gotErr := p.Bind(&obj)
		expectStr := `validate failed on "len($)>3" for value map[a:1 b:2]`
		if !strings.Contains(fmt.Sprint(gotErr), expectStr) {
			t.Fatalf("unexpect error %s", gotErr)
		}
	}

	bindFromInt := func(key string) {
		type Object struct {
			Int     int     `value:"${int}"`
			Int8    int8    `value:"${int}"`
			Int16   int16   `value:"${int}"`
			Int32   int32   `value:"${int}"`
			Int64   int64   `value:"${int}"`
			Uint    uint    `value:"${int}"`
			Uint8   uint8   `value:"${int}"`
			Uint16  uint16  `value:"${int}"`
			Uint32  uint32  `value:"${int}"`
			Uint64  uint64  `value:"${int}"`
			Float32 float32 `value:"${int}"`
			Float64 float64 `value:"${int}"`
			String  string  `value:"${int}"`
			Bool    bool    `value:"${int}"`
		}
		var gotObj Object
		err := p.Bind(&gotObj, conf.Param(conf.BindParam{
			Key: key,
		}))
		if err != nil {
			t.Fatal(err)
		}
		expectObj := Object{
			Int:     1,
			Int8:    1,
			Int16:   1,
			Int32:   1,
			Int64:   1,
			Uint:    1,
			Uint8:   1,
			Uint16:  1,
			Uint32:  1,
			Uint64:  1,
			Float32: 1,
			Float64: 1,
			String:  "1",
			Bool:    true,
		}
		if !reflect.DeepEqual(gotObj, expectObj) {
			t.Fatalf("got %v, expect %v", gotObj, expectObj)
		}
	}

	bindFromInt("toml")
	bindFromInt("yaml")
	bindFromInt("prop")
	bindFromInt("args")
	bindFromInt("envs")

	bindFromStr := func(key string) {
		type Object struct {
			Int     int     `value:"${map.a}"`
			Int8    int8    `value:"${map.a}"`
			Int16   int16   `value:"${map.a}"`
			Int32   int32   `value:"${map.a}"`
			Int64   int64   `value:"${map.a}"`
			Uint    uint    `value:"${map.a}"`
			Uint8   uint8   `value:"${map.a}"`
			Uint16  uint16  `value:"${map.a}"`
			Uint32  uint32  `value:"${map.a}"`
			Uint64  uint64  `value:"${map.a}"`
			Float32 float32 `value:"${map.a}"`
			Float64 float64 `value:"${map.a}"`
			String  string  `value:"${map.a}"`
			Bool    bool    `value:"${map.a}"`
		}
		var gotObj Object
		addr := reflect.ValueOf(&gotObj).Elem()
		err := p.Bind(addr, conf.Key(key))
		if err != nil {
			t.Fatal(err)
		}
		expectObj := Object{
			Int:     1,
			Int8:    1,
			Int16:   1,
			Int32:   1,
			Int64:   1,
			Uint:    1,
			Uint8:   1,
			Uint16:  1,
			Uint32:  1,
			Uint64:  1,
			Float32: 1,
			Float64: 1,
			String:  "1",
			Bool:    true,
		}
		if !reflect.DeepEqual(gotObj, expectObj) {
			t.Fatalf("got %v, expect %v", gotObj, expectObj)
		}
	}

	bindFromStr("toml")
	bindFromStr("yaml")
	bindFromStr("prop")
	bindFromStr("args")
	bindFromStr("envs")

	{
		v := p.Get("key.undef")
		if v != "" {
			t.Fatalf("unexpect value %s", v)
		}
		v = p.Get("key.undef", conf.Def("abc"))
		if v != "abc" {
			t.Fatalf("unexpect value %s", v)
		}
	}

	{
		gotErr := p.Bind(nil)
		expectErr := errors.New("should be a ptr")
		if fmt.Sprint(gotErr) != expectErr.Error() {
			t.Fatalf("got %v, expect %v", gotErr, expectErr)
		}
	}

	{
		var i int
		gotErr := p.Bind(&i, conf.Tag("$"))
		expectErr := conf.ErrInvalidSyntax
		if !errors.Is(gotErr, expectErr) {
			t.Fatalf("got %v, expect %v", gotErr, expectErr)
		}
	}

	{
		var gotPoints []Point
		err := p.Bind(&gotPoints, conf.Tag(`${points:=(1,2)|(3,4)|(5,6)}>>point`))
		if err != nil {
			t.Fatal(err)
		}
		expectPoints := []Point{
			{1, 2}, {3, 4}, {5, 6},
		}
		if !reflect.DeepEqual(gotPoints, expectPoints) {
			t.Fatalf("got %v, expect %v", gotPoints, expectPoints)
		}
	}

	{
		var gotPoints [3]Point
		gotErr := p.Bind(&gotPoints, conf.Tag(`${points:=(1,2)|(3,4)|(5,6)}>>point`))
		expectErr := errors.New("use slice instead of array")
		if !strings.Contains(fmt.Sprint(gotErr), expectErr.Error()) {
			t.Fatalf("got %v, expect %v", gotErr, expectErr)
		}
	}

	{
		var gotPoints []Point
		gotErr := p.Bind(&gotPoints, conf.Tag(`${points:=error:injected error}>>point`))
		expectErr := errors.New(`split error: injected error, value: `)
		if !strings.Contains(fmt.Sprint(gotErr), expectErr.Error()) {
			t.Fatalf("got %v, expect %v", gotErr, expectErr)
		}
	}

	{
		var gotPoints []Point
		gotErr := p.Bind(&gotPoints, conf.Tag(`${points:=error:injected error}>>splitPoints`))
		expectErr := errors.New(`unknown splitter "splitPoints"`)
		if !strings.Contains(fmt.Sprint(gotErr), expectErr.Error()) {
			t.Fatalf("got %v, expect %v", gotErr, expectErr)
		}
	}

	{
		var i int
		v := reflect.ValueOf(&i)
		gotErr := p.Bind(v)
		expectErr := errors.New("bind *int error, target should be value type")
		if !strings.Contains(fmt.Sprint(gotErr), expectErr.Error()) {
			t.Fatalf("got %v, expect %v", gotErr, expectErr)
		}
	}

	{
		type Object struct {
			Ints []int `value:"${ints}"`
		}
		var gotObj Object
		gotErr := p.Bind(&gotObj)
		expectErr := errors.New(`property "ints" not exist`)
		if !strings.Contains(fmt.Sprint(gotErr), expectErr.Error()) {
			t.Fatalf("got %v, expect %v", gotErr, expectErr)
		}
	}

	{
		type Object struct {
			Ints []int `value:"${ints:=}"`
		}
		var gotObj Object
		err := p.Bind(&gotObj)
		if err != nil {
			t.Fatal(err)
		}
		expectObj := Object{
			Ints: make([]int, 0),
		}
		if !reflect.DeepEqual(gotObj, expectObj) {
			t.Fatalf("got %v, expect %v", gotObj, expectObj)
		}
	}

	{
		type Object struct {
			Ints []int `value:"${ints:=1,2,3}"`
		}
		var gotObj Object
		err := p.Bind(&gotObj)
		if err != nil {
			t.Fatal(err)
		}
		expectObj := Object{
			Ints: []int{1, 2, 3},
		}
		if !reflect.DeepEqual(gotObj, expectObj) {
			t.Fatalf("got %v, expect %v", gotObj, expectObj)
		}
	}

	{
		type Handler interface {
			Func()
		}
		type Inner struct {
			Ints []int `value:"${ints:=1,2,3}"`
		}
		type Object struct {
			Inner
			Handler
			Arr []int `value:"${toml.empty_arr}"`
		}
		var gotObj Object
		err := p.Bind(&gotObj)
		if err != nil {
			t.Fatal(err)
		}
		expectObj := Object{
			Inner: Inner{
				Ints: []int{1, 2, 3},
			},
			Arr: make([]int, 0),
		}
		if !reflect.DeepEqual(gotObj, expectObj) {
			t.Fatalf("got %v, expect %v", gotObj, expectObj)
		}
	}

	{
		var obj struct {
			Map map[string]string `value:"${maps:=a=1,b=1}"`
		}
		gotErr := p.Bind(&obj)
		expectErr := errors.New("map can't have a non-empty default value")
		if !strings.Contains(fmt.Sprint(gotErr), expectErr.Error()) {
			t.Fatalf("got %v, expect %v", gotErr, expectErr)
		}
	}

	{
		var obj struct {
			Struct struct{} `value:"${struct:=a=1,b=1}"`
		}
		gotErr := p.Bind(&obj)
		expectErr := errors.New("struct can't have a non-empty default value")
		if !strings.Contains(fmt.Sprint(gotErr), expectErr.Error()) {
			t.Fatalf("got %v, expect %v", gotErr, expectErr)
		}
	}

	{
		var arr []int
		gotErr := p.Bind(&arr, conf.Tag("${arr:=1,a,3}"))
		expectErr := errors.New(`parsing "a": invalid syntax`)
		if !strings.Contains(fmt.Sprint(gotErr), expectErr.Error()) {
			t.Fatalf("got %v, expect %v", gotErr, expectErr)
		}
	}

	{
		type Point3D struct {
			X, Y, Z int
		}
		var obj struct {
			Points []Point3D `value:"${points:=(1,2,3)}"`
		}
		gotErr := p.Bind(&obj)
		expectErr := errors.New(`can't find converter for conf_test.Point3D`)
		if !strings.Contains(fmt.Sprint(gotErr), expectErr.Error()) {
			t.Fatalf("got %v, expect %v", gotErr, expectErr)
		}
	}

	{
		type Point3D struct {
			X, Y, Z int
		}
		type Inner struct {
			Points []Point3D `value:"${points:=(1,2,3)}"`
		}
		var obj struct {
			Inner
		}
		gotErr := p.Bind(&obj)
		expectErr := errors.New(`can't find converter for conf_test.Point3D`)
		if !strings.Contains(fmt.Sprint(gotErr), expectErr.Error()) {
			t.Fatalf("got %v, expect %v", gotErr, expectErr)
		}
	}

	{
		var obj struct {
			Map map[string]string `value:"${toml.map.a}"`
		}
		gotErr := p.Bind(&obj)
		expectErr := errors.New(`property 'toml.map.a' is value`)
		if !strings.Contains(fmt.Sprint(gotErr), expectErr.Error()) {
			t.Fatalf("got %v, expect %v", gotErr, expectErr)
		}
	}

	{
		var m map[string]int
		gotErr := p.Bind(&m, conf.Key("maps"))
		expectErr := errors.New(`parsing "a": invalid syntax`)
		if !strings.Contains(fmt.Sprint(gotErr), expectErr.Error()) {
			t.Fatalf("got %v, expect %v", gotErr, expectErr)
		}
	}

	{
		var obj struct {
			s int `value:"${private-int:=3}"`
		}
		err := p.Bind(&obj)
		if err != nil {
			t.Fatal(err)
		}
		if obj.s != 0 {
			t.Fatalf("got %v, expect %v", obj.s, 0)
		}
	}

	{
		var obj struct {
			Int int `value:"{}"`
		}
		gotErr := p.Bind(&obj)
		expectErr := errors.New(`parse tag '{}' error: invalid syntax`)
		if !strings.Contains(fmt.Sprint(gotErr), expectErr.Error()) {
			t.Fatalf("got %v, expect %v", gotErr, expectErr)
		}
	}

	{
		type Object struct {
			A int
		}
		var gotObj Object
		err := p.Bind(&gotObj, conf.Tag(`${toml.map}`))
		if err != nil {
			t.Fatal(err)
		}
		expectObj := Object{
			A: 1,
		}
		if !reflect.DeepEqual(gotObj, expectObj) {
			t.Fatalf("got %v, expect %v", gotObj, expectObj)
		}
	}

	{
		var obj struct {
			B int
		}
		gotErr := p.Bind(&obj, conf.Tag(`${maps}`))
		expectErr := errors.New(`parsing "a": invalid syntax`)
		if !strings.Contains(fmt.Sprint(gotErr), expectErr.Error()) {
			t.Fatalf("got %v, expect %v", gotErr, expectErr)
		}
	}

	{
		type Object struct {
			Args_Int int
		}
		var gotObj Object
		err := p.Bind(&gotObj)
		if err != nil {
			t.Fatal(err)
		}
		expectObj := Object{Args_Int: 1}
		if !reflect.DeepEqual(gotObj, expectObj) {
			t.Fatalf("got %v, expect %v", gotObj, expectObj)
		}
	}

	{
		got, err := p.Resolve("current profile is ${spring.active.profile}")
		if err != nil {
			t.Fatal(err)
		}
		expect := "current profile is online"
		if got != expect {
			t.Fatalf("got %v, expect %v", got, expect)
		}
	}

	{
		got, err := p.Resolve("current profile is ${spring.profiles.active:=${spring.active.profile}}")
		if err != nil {
			t.Fatal(err)
		}
		expect := "current profile is online"
		if got != expect {
			t.Fatalf("got %v, expect %v", got, expect)
		}
	}

	{
		_, gotErr := p.Resolve("current profile is ${spring.active.profile")
		expectErr := errors.New(`invalid syntax`)
		if !strings.Contains(fmt.Sprint(gotErr), expectErr.Error()) {
			t.Fatalf("got %v, expect %v", gotErr, expectErr)
		}
	}

	{
		_, gotErr := p.Resolve("current profile is ${${spring.active.profile}")
		expectErr := errors.New(`invalid syntax`)
		if !strings.Contains(fmt.Sprint(gotErr), expectErr.Error()) {
			t.Fatalf("got %v, expect %v", gotErr, expectErr)
		}
	}

	{
		_, gotErr := p.Resolve("current profile is ${maps}")
		expectErr := errors.New(`property "maps" isn't simple value`)
		if !strings.Contains(fmt.Sprint(gotErr), expectErr.Error()) {
			t.Fatalf("got %v, expect %v", gotErr, expectErr)
		}
	}

	{
		got, err := p.Resolve("current profile is ${spring.active.profile} and ${another.active.profile:=cluster_a}")
		if err != nil {
			t.Fatal(err)
		}
		expect := "current profile is online and cluster_a"
		if got != expect {
			t.Fatalf("got %v, expect %v", got, expect)
		}
	}

	{
		_, gotErr := p.Resolve("current profile is ${spring.active.profile} and ${another.active.profile")
		expectErr := errors.New(`invalid syntax`)
		if !strings.Contains(fmt.Sprint(gotErr), expectErr.Error()) {
			t.Fatalf("got %v, expect %v", gotErr, expectErr)
		}
	}
}
