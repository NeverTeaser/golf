# GOLF (WIP)

GOLF(Go Light Filter), golf dependents [Gorm](https://github.com/go-gorm/gorm)
and [Gin](https://github.com/gin-gonic/gin). golf can help you build model query as fast as，build model query like
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

## implement `Golf Query`

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

## use with Gin

TODO you can read the example

```go
// URL path /ping?eq_id=1&like_username=test
// sql log should be  SELECT * FROM "test_model" WHERE username LIKE 'test' AND id = 1 LIMIT 10
```

