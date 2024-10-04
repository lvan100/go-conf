# go-conf

[英文](README.md)

一个分层的配置管理器。它支持多个配置源(环境变量、命令行参数和配置文件等)，并且允许向前覆盖。

- Layer 1: 通过函数 `SetProperty` 设置的属性
- Layer 2: 通过本地静态文件设置的属性
- Layer 3: 通过环境变量设置的属性
- Layer 4: 通过命令行参数设置的属性
- Layer 5: 通过远程动态文件设置的属性

较低层级的属性会被较高层级的属性覆盖。

### 使用

1. 初始化配置管理器。

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

2. 将属性绑定到值上。

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
