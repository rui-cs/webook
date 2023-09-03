package domain

import "time"

// User 领域对象，是 DDD 中的 entity
// BO(business object)
type User struct {
	Id       int64     `json:"id"`
	Email    string    `json:"email"`
	Phone    string    `json:"-"`
	Name     string    `json:"name"`
	Password string    `json:"-"`
	Birthday string    `json:"birthday"`
	Resume   string    `json:"resume"`
	Ctime    time.Time `json:"-"`
}
