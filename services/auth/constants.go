package auth

import "time"

const (
	KeyAccessToken  = "access_token"
	KeyRefreshToken = "refresh_token"

	TTLRefreshToken = time.Hour * 24 * 7
	TTLAccessToken  = time.Minute * 15

	ErrorEmptyToken               = "token was not provided"
	ErrorNotValidToken            = "token is not valid"
	ErrorRedisCannotSetAccessKey  = "cannot set new access_token key"
	ErrorRedisCannotSetRefreshKey = "cannot set new refresh_token key"
	ErrorNotIncludedUUIDClaim     = "not included uuid claim"
	ErrorNotIncludedRoleClaim     = "not included role claim access"
	ErrorNotValidClaims           = "not valid claims"

	ROLE_USER    = "user"
	ROLE_VISITOR = "visitor"
)
