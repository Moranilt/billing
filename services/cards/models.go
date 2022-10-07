package cards

type Card struct {
	Number  string    `json:"number" db:"number"`
	Mask    string    `json:"mask" db:"mask"`
	CVC     int       `json:"cvc" db:"cvc"`
	Balance float64   `json:"balance" db:"balance"`
	State   CardState `json:"state" db:"state"`
}

type CardSelect struct {
	Number  string  `json:"number" db:"number"`
	Mask    string  `json:"mask" db:"mask"`
	CVC     int     `json:"cvc" db:"cvc"`
	Balance float64 `json:"balance" db:"balance"`
	State   []byte  `json:"state" db:"state"`
}

type CardState struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type CardCreate struct {
	Number string `json:"number" db:"number"`
	Mask   string `json:"mask" db:"mask"`
	CVC    int    `json:"cvc" db:"cvc"`
}
