package telegram

import "fmt"

type User struct {
	ID    int64 `json:"id"`
	IsBot bool  `json:"is_bot,omitempty"`
}

func (user *User) StringID() string {
	return fmt.Sprintf("%d", user.ID)
}
