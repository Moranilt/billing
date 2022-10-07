package user

import "github.com/Moranilt/billing/services/cards"

type User struct {
	Id         int          `json:"id"`
	Email      string       `json:"email" db:"email"`
	Firstname  string       `json:"firstname" db:"firstname"`
	Lastname   string       `json:"lastname" db:"lastname"`
	Patronymic string       `json:"patronymic" db:"patronymic"`
	Phone      string       `json:"phone" db:"phone"`
	CreatedAt  string       `json:"created_at" db:"created_at"`
	Cards      []cards.Card `json:"card"`
}

type GetUser struct {
	Id         int    `db:"id"`
	Email      string `db:"email"`
	Firstname  string `db:"firstname"`
	Lastname   string `db:"lastname"`
	Patronymic string `db:"patronymic"`
	Phone      string `db:"phone"`
	CreatedAt  string `db:"created_at"`
	Cards      []byte `db:"cards"`
}

type UpdateUser struct {
	Firstname  *string
	Lastname   *string
	Patronymic *string
}
