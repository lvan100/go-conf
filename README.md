# go-conf

<div>
 <img src="https://img.shields.io/github/license/lvan100/go-conf" alt="license"/>
 <img src="https://img.shields.io/github/go-mod/go-version/lvan100/go-conf" alt="go-version"/>
 <img src="https://img.shields.io/github/v/release/lvan100/go-conf?include_prereleases" alt="release"/>
 <img src='https://coveralls.io/repos/github/lvan100/go-conf/badge.svg?branch=main' alt='Coverage Status' />
</div>

[中文](README_CN.md)

A layered configuration manager. It supports multiple configuration sources
(environment variables, command line arguments, and configuration files) and
it allows for overriding configuration values at runtime.

- Layer 1: Properties set by func `SetProperty`
- Layer 2: Properties set by local static files
- Layer 3: Properties set by environment variables
- Layer 4: Properties set by command line arguments
- Layer 5: Properties set by remote dynamic sources

The properties in the lower layers are overridden by the higher ones.

### Usage

1. Initialize a configuration manager.

```go
import "github.com/lvan100/go-conf"

c := conf.NewConfiguration()

// Set a property
c.SetProperty("spring.active.profile", "online")

// Add local static files, the properties in the later added files will override
// the properties in the earlier added files.
c.File().Add("testdata/conf.toml", "testdata/conf-${spring.active.profile}.yaml")
c.File().Add("testdata/xxx.yaml", "testdata/conf.json")

// Add remote dynamic files, the properties in the later added files will override
// the properties in the earlier added files.
c.Dync().Add("testdata-${spring.active.profile}/dync.properties")

var p conf.ReadOnlyProperties
p, err = c.Refresh() // You get a merged read-only properties here.
```

2. Bind the properties into values.

```go
type SomeOne struct {
    // Age configged by `age` and it has a default value of 16,
    // and it must be between 14 and 18.
    Age int32 `value:"${age:=16}" expr:"$>=14&&$<=18"`
    
    // Name configged by `name`. Because it hasn't a default value,
    // so it must be configged by files or environment variable, and so on.
    Addr string `value:"${addr}"`
    
    Phones []string `value:"${phones}"`
    Grades map[string]int `value:"${grades}"`
}

type Point struct {
	X int32
	Y int32
}

type Tomor struct {

    // SomeOne its all fields will be configged by the prefix `someone`. 
    SomeOne `value:"${someone}"`
    
    // Another its all fields will be configged without a prefix, because
    // there isn't a `value` tag on the field. 
    Another SomeOne
    
    // Points configged by `points` and it has a splitter `PointSplitter`
    // waht splits the string `(1,2)(3,4)` to slice `[]Point{{1,2},{3,4}}`.
    Points []Point `value:"${points}>>PointSplitter"`
}

var obj Tomor
p.Bind(&obj) // all fields has no prefix.

var obj Tomor
p.Bind(&obj, conf.Key("conf")) // all fields has a prefix `conf`.
```
