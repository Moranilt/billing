package user

import (
	"encoding/json"

	"github.com/Moranilt/billing/services/cards"
	"github.com/Moranilt/billing/utils"
	"github.com/jmoiron/sqlx"
)

type UserService struct {
	db          *sqlx.DB
	customQuery utils.Query
}

type UserMethods interface {
	Get(userId int) (*User, error)
	Delete()
	Update(UpdateUser)
}

func NewService(db *sqlx.DB, query utils.Query) UserMethods {
	return &UserService{
		db:          db,
		customQuery: query,
	}
}

func (u *UserService) Get(userId int) (*User, error) {
	var user User

	var unpUser GetUser

	query := `SELECT
	users.id,
	users.email,
	users.firstname,
	users.lastname,
	users.patronymic,
	users.created_at,
	users.phone,
	json_agg(
		json_build_object(
			'id', cards.id,
			'number', cards.number,
			'mask', cards.mask,
			'cvc', cards.cvc,
			'balance', cards.balance,
			'until_date', cards.until_date,
			'state', json_build_object(
				'id', cs.id,
				'name', cs.description
			)
		)
	) as cards
	FROM users
	INNER JOIN cards ON cards.user_id=users.id
	INNER JOIN card_states as cs ON cs.id=cards.state_id
	WHERE users.id=$1
	GROUP BY users.id`

	err := u.db.Get(&unpUser, query, userId)
	if err != nil {
		return nil, err
	}

	var cards []cards.Card
	err = json.Unmarshal(unpUser.Cards, &cards)
	if err != nil {
		return nil, err
	}

	user = User{
		Id:         unpUser.Id,
		Email:      unpUser.Email,
		Firstname:  unpUser.Firstname,
		Lastname:   unpUser.Lastname,
		Patronymic: unpUser.Patronymic,
		Phone:      unpUser.Phone,
		CreatedAt:  unpUser.CreatedAt,
		Cards:      cards,
	}

	return &user, nil
}

func (u *UserService) Delete() {

}

func (u *UserService) Update(updateUser UpdateUser) {

}
