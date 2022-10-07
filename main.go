package main

import (
	"context"
	"log"
	"net/http"

	"github.com/Moranilt/billing/logger"
	"github.com/Moranilt/billing/services"
	"github.com/Moranilt/billing/services/auth"
	"github.com/Moranilt/billing/services/cards"
	"github.com/Moranilt/billing/services/user"
	"github.com/Moranilt/billing/utils"
	"github.com/go-redis/redis/v8"

	"github.com/Moranilt/rou"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type ChannelsList struct {
	info chan string
	err  chan error
	warn chan string
}

type Repository struct {
	Authorization auth.Authentication
	User          user.UserMethods
	Cards         cards.CardsMethods
	logger        logger.LoggerWriter
	channels      ChannelsList
}

func (r *Repository) Login(ctx *rou.Context) {
	err := r.Authorization.CreateTokens(ctx.ResponseWriter(), 1, "user")
	if err != nil {
		r.channels.err <- err
		ctx.ErrorJSONResponse(http.StatusBadRequest, err.Error())
		return
	}

	userData, err := r.User.Get(2)
	if err != nil {
		r.channels.err <- err
		ctx.ErrorJSONResponse(http.StatusBadRequest, err.Error())
		return
	}

	ctx.SuccessJSONResponse(userData)
}

func (r *Repository) CardsList(ctx *rou.Context) {
	accessToken, err := r.Authorization.GetTokenFromCookie(ctx.Request(), auth.KeyAccessToken)
	if err != nil {
		r.channels.err <- err
		ctx.ErrorJSONResponse(http.StatusBadRequest, err.Error())
		return
	}

	accessDetails, err := r.Authorization.ExtractAccessMetaData(accessToken)
	if err != nil {
		r.channels.err <- err
		ctx.ErrorJSONResponse(http.StatusBadRequest, err.Error())
		return
	}

	card, err := r.Cards.GetCards(accessDetails.GetUserId())
	if err != nil {
		r.channels.err <- err
		ctx.ErrorJSONResponse(http.StatusBadRequest, err.Error())
		return
	}
	ctx.SuccessJSONResponse(card)
}

func (r *Repository) UserInfo(ctx *rou.Context) {
	accessToken, err := r.Authorization.GetTokenFromCookie(ctx.Request(), auth.KeyAccessToken)
	if err != nil {
		r.channels.err <- err
		ctx.ErrorJSONResponse(http.StatusBadRequest, err.Error())
		return
	}

	accessDetails, err := r.Authorization.ExtractAccessMetaData(accessToken)
	if err != nil {
		r.channels.err <- err
		ctx.ErrorJSONResponse(http.StatusBadRequest, err.Error())
		return
	}
	user, err := r.User.Get(accessDetails.GetUserId())

	if err != nil {
		r.channels.err <- err
		ctx.ErrorJSONResponse(http.StatusBadRequest, err.Error())
		return
	}

	ctx.SuccessJSONResponse(user)
}

func catchLogs(channels ChannelsList, logger logger.LoggerWriter) {
	for {
		select {
		case msg := <-channels.info:
			logger.Info(msg)
		case msg := <-channels.err:
			logger.Error(msg.Error())
		case msg := <-channels.warn:
			logger.Warning(msg)
		}
	}
}

var redisContext = context.Background()

func main() {
	router := rou.NewRouter()
	conn, err := sqlx.Connect("postgres", "user=root password=123456 dbname=billing sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	localLogger := logger.NewLogger()

	channels := ChannelsList{
		info: make(chan string),
		err:  make(chan error),
		warn: make(chan string),
	}

	authorization := auth.NewService(auth.AuthSettings{
		Db:           conn,
		Secret:       "secret phrase",
		RedisClient:  redisClient,
		RedisContext: redisContext,
	})
	queryTest := utils.NewQuery(conn)
	middleware := services.NewMiddlewareService(conn, authorization)
	cardsService := cards.NewService(conn)
	userService := user.NewService(conn, queryTest)

	repository := Repository{
		Authorization: authorization,
		Cards:         cardsService,
		User:          userService,
		logger:        localLogger,
		channels:      channels,
	}

	go catchLogs(channels, localLogger)

	router.Get("/login", repository.Login)
	router.Get("/user", repository.UserInfo).Middleware(middleware.AuthorizedUser)
	router.Get("/cards", repository.CardsList).Middleware(middleware.AuthorizedUser)

	log.Fatal(router.RunServer(":8080"))

}
