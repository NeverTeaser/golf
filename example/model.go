package main

import "github.com/neverteaser/golf"

type testModel struct {
	ID       int    `json:"id"`
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
}

func (m *testModel) Field() map[string][]golf.Filter {
	return map[string][]golf.Filter{
		"ID":       {golf.Equal},
		"UserID":   {golf.Equal, golf.Gte},
		"Username": {golf.Equal, golf.Like},
	}
}
