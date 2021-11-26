# GOLF (WIP)

### [中文文档](https://github.com/NeverTeaser/golf/blob/master/README_Zh-cn.md)
GOLF(Go Light Filter), golf dependents [Gorm](https://github.com/go-gorm/gorm) . golf can help you build model query as fast as，build model query like
Django Rest Framework.

## usage

### define model

```go
type testModel struct {
    ID       int    `json:"id"`
    UserID   int    `json:"user_id"`
    Username string `json:"username"`
}
```

## implement `GolfQuery` interface 

use golf must implement this interface , map key is go struct member name

```go

func (m *testModel) Field() map[string][]golf.Filter {
    return map[string][]golf.Filter{
        "ID":       {golf.Equal},
        "UserID":   {golf.Equal, golf.Gte},
        "Username": {golf.Equal, golf.Like},
    }
}
```

## use with request url 

TODO you can read the example

```go
// URL path /ping?eq_id=1&like_username=test
// sql log should be  SELECT * FROM "test_model" WHERE username LIKE 'test' AND id = 1 LIMIT 10
golfQ := golf.NewGolf(globalDB)
var tests testModel
if err := golfQ.Build(&testModel{}, Request.URL.Query()).Find(&tests).Error; err != nil {
    log.Println("find failed", err)
}
```

