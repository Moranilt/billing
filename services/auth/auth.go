package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type authService struct {
	db           *sqlx.DB
	redisClient  *redis.Client
	redisContext context.Context
	secret       string
}

type Authentication interface {
	// Create access- and refresh-token and store it to Set-Cookie header
	// using http.ResponseWriter. Stores keys of tokens to Redis
	CreateTokens(w http.ResponseWriter, userId int, role string) error
	// Get token from Cookie request by token name
	GetTokenFromCookie(r *http.Request, key string) (*jwt.Token, error)
	// Extract fields from access token
	ExtractAccessMetaData(token *jwt.Token) (AccessDetails, error)
	// Extract fields from refresh token
	ExtractRefreshMetaData(token *jwt.Token) (AccessDetails, error)
	// Check for existing refresh-token, delete refresh-token from Redis
	// and create new tokens using CreateTokens method
	RefreshToken(w http.ResponseWriter, refreshToken *jwt.Token) (AccessDetails, error)
}

type AccessDetails interface {
	GetUserId() int
	GetRole() string
}

type AuthSettings struct {
	Db           *sqlx.DB
	RedisClient  *redis.Client
	RedisContext context.Context
	Secret       string
}

func NewService(settings AuthSettings) Authentication {
	return &authService{
		db:           settings.Db,
		secret:       settings.Secret,
		redisClient:  settings.RedisClient,
		redisContext: settings.RedisContext,
	}
}

type accessDetails struct {
	Role   string
	UserId int
}

func (ad *accessDetails) GetUserId() int {
	return ad.UserId
}

func (ad *accessDetails) GetRole() string {
	return ad.Role
}

func (auth *authService) CreateTokens(w http.ResponseWriter, userId int, role string) error {
	now := time.Now()
	td, err := auth.generateTokens(userId, role)
	if err != nil {
		return err
	}

	err = auth.redisClient.Set(auth.redisContext, td.ATUuid, fmt.Sprint(userId), td.ATExpires.Sub(now)).Err()
	if err != nil {
		return errors.New(ErrorRedisCannotSetAccessKey)
	}

	err = auth.redisClient.Set(auth.redisContext, td.RTUuid, fmt.Sprint(userId), td.RTExpires.Sub(now)).Err()
	if err != nil {
		return errors.New(ErrorRedisCannotSetRefreshKey)
	}

	accessCookie := &http.Cookie{
		Name:     KeyAccessToken,
		Value:    td.AccessToken,
		Expires:  td.ATExpires,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   true,
	}
	refreshCookie := &http.Cookie{
		Name:     KeyRefreshToken,
		Value:    td.RefreshToken,
		Expires:  td.RTExpires,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   true,
	}
	http.SetCookie(w, accessCookie)
	http.SetCookie(w, refreshCookie)
	return nil
}

func (auth *authService) GetTokenFromCookie(r *http.Request, key string) (*jwt.Token, error) {
	cookieToken, err := r.Cookie(key)
	if err != nil {
		return nil, errors.New(ErrorEmptyToken)
	}

	if cookieToken.Value == "" {
		return nil, errors.New(ErrorEmptyToken)
	}

	token, err := auth.verifyToken(cookieToken.Value)
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New(ErrorNotValidToken)
	}

	return token, nil
}

func (auth *authService) ExtractAccessMetaData(token *jwt.Token) (AccessDetails, error) {
	accessTokenDetails, err := auth.parseAccessToken(token)
	if err != nil {
		return nil, err
	}

	userIdString, err := auth.redisClient.Get(auth.redisContext, accessTokenDetails.ATUuid).Result()
	if err != nil {
		return nil, err
	}

	userId, err := strconv.Atoi(userIdString)
	if err != nil {
		return nil, err
	}

	return &accessDetails{
		UserId: userId,
		Role:   accessTokenDetails.Role,
	}, nil
}

func (auth *authService) ExtractRefreshMetaData(token *jwt.Token) (AccessDetails, error) {
	refreshDetails, err := auth.parseRefreshToken(token)
	if err != nil {
		return nil, err
	}

	userIdString, err := auth.redisClient.Get(auth.redisContext, refreshDetails.RTUuid).Result()
	if err != nil {
		return nil, errors.New("cannot get value by key")
	}

	userId, err := strconv.Atoi(userIdString)
	if err != nil {
		return nil, err
	}

	return &accessDetails{
		UserId: userId,
		Role:   refreshDetails.Role,
	}, nil
}

func (auth *authService) RefreshToken(w http.ResponseWriter, refreshToken *jwt.Token) (AccessDetails, error) {

	refreshDetails, err := auth.parseRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	userIdString, err := auth.redisClient.Get(auth.redisContext, refreshDetails.RTUuid).Result()

	if err != nil {
		return nil, errors.New("cannot get user_id from redis")
	}

	userId, err := strconv.Atoi(userIdString)
	if err != nil {
		return nil, err
	}

	err = auth.redisClient.Del(auth.redisContext, refreshDetails.RTUuid).Err()

	if err != nil {
		return nil, errors.New("cannot delete key from redis")
	}

	err = auth.CreateTokens(w, userId, refreshDetails.Role)
	if err != nil {
		return nil, err
	}

	return &accessDetails{
		Role:   refreshDetails.Role,
		UserId: userId,
	}, nil
}

func (auth *authService) parseAccessToken(token *jwt.Token) (*AccessTokenDetails, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && claims.Valid() == nil {
		accessUuid, ok := claims["access_uuid"].(string)
		if !ok {
			return nil, errors.New(ErrorNotIncludedUUIDClaim)
		}

		role, ok := claims["role"].(string)
		if !ok {
			return nil, errors.New(ErrorNotIncludedRoleClaim)
		}

		return &AccessTokenDetails{
			ATUuid: accessUuid,
			Role:   role,
		}, nil
	}

	return nil, errors.New(ErrorNotValidClaims)
}

func (auth *authService) parseRefreshToken(token *jwt.Token) (*RefreshTokenDetails, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && claims.Valid() == nil {
		refreshUuid, ok := claims["refresh_uuid"].(string)
		if !ok {
			return nil, errors.New(ErrorNotIncludedUUIDClaim)
		}

		role, ok := claims["role"].(string)
		if !ok {
			return nil, errors.New(ErrorNotIncludedRoleClaim)
		}

		return &RefreshTokenDetails{
			RTUuid: refreshUuid,
			Role:   role,
		}, nil
	}

	return nil, errors.New(ErrorNotValidClaims)
}

func (auth *authService) generateTokens(userId int, role string) (*TokenDetails, error) {
	now := time.Now()
	td := &TokenDetails{
		ATUuid:    uuid.NewString(),
		RTUuid:    uuid.NewString(),
		ATExpires: now.Add(TTLAccessToken),
		RTExpires: now.Add(TTLRefreshToken),
	}

	var err error

	atClaims := jwt.MapClaims{}
	atClaims["access_uuid"] = td.ATUuid
	atClaims["role"] = role
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString([]byte(auth.secret))
	if err != nil {
		return nil, err
	}

	rtClaims := jwt.MapClaims{}
	rtClaims["refresh_uuid"] = td.RTUuid
	rtClaims["role"] = role
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = rt.SignedString([]byte(auth.secret))

	if err != nil {
		return nil, err
	}

	return td, nil
}

func (auth *authService) verifyToken(requestToken string) (*jwt.Token, error) {
	token, err := jwt.Parse(requestToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(auth.secret), nil
	})
	if err != nil {
		return nil, err
	}

	return token, nil
}
