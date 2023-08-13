package domain

import "time"

// User 领域对象，是 DDD 中的 entity
// BO(business object)
type User struct {
	Id       int64     `json:"id"`
	Email    string    `json:"email"`
	Password string    `json:"-"`
	Name     string    `json:"name"`
	Birthday string    `json:"birthday"`
	Resume   string    `json:"resume"`
	Ctime    time.Time `json:"-"`
}
