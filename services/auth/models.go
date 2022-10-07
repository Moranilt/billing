package auth

import "time"

type Tokens struct {
	AccessToken  string
	RefreshToken string
}

type AccessTokenDetails struct {
	ATUuid string
	Role   string
}

type RefreshTokenDetails struct {
	RTUuid string
	Role   string
}

type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	ATUuid       string
	RTUuid       string
	ATExpires    time.Time
	RTExpires    time.Time
}
