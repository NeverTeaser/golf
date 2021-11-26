# GOLF (WIP)

GOLF(Go Light Filter), golf 工作依赖于 [Gorm](https://github.com/go-gorm/gorm)
。 golf 可以让你model 支持各种过滤。像Django Restful framework 的model filter 一样方便

## usage

### 定义你的model

```go
type testModel struct {
    ID       int    
    UserID   int    
    Username string 
}
```

## 实现 `GolfQuery` 接口

使用golf的话，你的model 必须实现这个接口。提示: map 里的key 值是struct 成员名,value golf 里支持的操作，下面定义的具体含义是，ID 支持等于操作，UserID 支持 等于、大于等于操作
```go

func (m *testModel) Field() map[string][]golf.Filter {
    return map[string][]golf.Filter{
        "ID":       {golf.Equal},
        "UserID":   {golf.Equal, golf.Gte},
        "Username": {golf.Equal, golf.Like},
    }
}
```

## 使用示例

你也可以去[example](https://github.com/NeverTeaser/golf/tree/master/example) 查看更详细的用法

```go
// URL path /ping?eq_id=1&like_username=test
// sql log should be  SELECT * FROM "test_model" WHERE username LIKE 'test' AND id = 1 LIMIT 10
golfQ := golf.NewGolf(globalDB)
var tests testModel
if err := golfQ.Build(&testModel{}, Request.URL.Query()).Find(&tests).Error; err != nil {
    log.Println("find failed", err)
}
```

