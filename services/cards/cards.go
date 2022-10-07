package cards

import (
	"encoding/json"

	"github.com/jmoiron/sqlx"
)

type Cards struct {
	db *sqlx.DB
}

type CardsMethods interface {
	GetCards(userId int) ([]Card, error)
	Create(userId int, cardState CardCreate) (*Card, error)
}

func NewService(db *sqlx.DB) CardsMethods {
	return &Cards{db: db}
}

func (c *Cards) Create(userId int, cardState CardCreate) (*Card, error) {
	var card Card

	query := `INSERT INTO cards 
	(user_id, number, mask, cvc, state_id) 
	VALUES($1,$2,$3,$4,$5) 
	RETURNING id, number, mask, cvc, state_id`

	var newCard Card
	err := c.db.QueryRowx(
		query,
		userId,
		cardState.Number,
		cardState.Mask,
		cardState.CVC,
		DEFAULT_CARD_STATE,
	).StructScan(&newCard)
	if err != nil {
		return nil, err
	}

	return &card, nil
}

func (c *Cards) GetCards(userId int) ([]Card, error) {
	var cards []Card
	rows, err := c.db.Queryx(`SELECT 
	cards.number as number,
	cards.mask as mask,
	cards.cvc as cvc,
	cards.balance as balance,
	json_build_object(
		'id', cs.id,
		'name', cs.description
	) as state
	FROM cards
	INNER JOIN card_states as cs ON cs.id = cards.state_id
	WHERE cards.user_id=$1`, userId)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var dirtyCard CardSelect

		err := rows.StructScan(&dirtyCard)

		if err != nil {
			return nil, err
		}

		var state CardState
		err = json.Unmarshal(dirtyCard.State, &state)
		if err != nil {
			return nil, err
		}

		card := Card{
			Number:  dirtyCard.Number,
			Mask:    dirtyCard.Mask,
			CVC:     dirtyCard.CVC,
			Balance: dirtyCard.Balance,
			State:   state,
		}

		cards = append(cards, card)
	}

	return cards, nil
}
