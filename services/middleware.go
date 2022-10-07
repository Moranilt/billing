package services

import (
	"io"
	"net/http"

	"github.com/Moranilt/billing/services/auth"
	"github.com/jmoiron/sqlx"
)

type Middleware struct {
	db   *sqlx.DB
	auth auth.Authentication
}

func NewMiddlewareService(db *sqlx.DB, auth auth.Authentication) *Middleware {
	return &Middleware{db: db, auth: auth}
}

func (mw *Middleware) AuthorizedUser(w http.ResponseWriter, r *http.Request) bool {
	if r.Header.Get("Authorization") == "" {
		io.WriteString(w, "Authorization header was not provided")
		return false
	}
	accessToken, err := mw.auth.GetTokenFromCookie(r, auth.KeyAccessToken)

	if err != nil && err.Error() == auth.ErrorNotValidToken {
		io.WriteString(w, err.Error())
		return false
	}

	if err != nil && err.Error() == auth.ErrorEmptyToken {
		refreshToken, err := mw.auth.GetTokenFromCookie(r, auth.KeyRefreshToken)
		if err != nil {
			io.WriteString(w, err.Error())
			return false
		}

		_, err = mw.auth.RefreshToken(w, refreshToken)
		if err != nil {
			io.WriteString(w, err.Error())
			return false
		}

		return true
	}

	_, err = mw.auth.ExtractAccessMetaData(accessToken)
	if err != nil {
		io.WriteString(w, err.Error())
		return false
	}

	return true
}
