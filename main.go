package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	jwt "github.com/appleboy/gin-jwt"
	_ "github.com/signaux-faibles/datapi/docs"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	dalib "github.com/signaux-faibles/datapi/lib"

	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

func main() {
	api := flag.String("api", "", "Runs API on TCP adresse (e.g. -api :3000)")
	createdb := flag.String("createdb", "", "Initialize postgres Database with provided pgsql credentials.")
	connstr := flag.String("pg", "", "Postgres connection string")
	jwtsecret := flag.String("jwt", "", "Secret key for JWT signature")
	createuser := flag.String("createuser", "", "Create the first datapi user with user:password syntax")

	flag.Parse()

	if *connstr == "" {
		panic("abort: empty connection string")
	}

	if *createuser != "" {
		credentials := strings.Split(*createuser, ":")
		user := credentials[0]
		password := credentials[1]
		if user == "" || password == "" {
			panic("CreateUser failed: user or pasword empty")
		}

		err := dalib.CreateUser(*connstr, user, password)
		if err != nil {
			log.Fatal("CreateDB failed: " + err.Error())
		}
		return
	}

	if *createdb != "" {
		dalib.CreateDB(*createdb, *connstr)
		return
	}

	if *jwtsecret == "" {
		panic("empty JWT secret not allowed")
	}

	if *api != "" {
		runAPI(*api, *jwtsecret, *connstr)
		return
	}

	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()

}

func runAPI(bind, jwtsecret, connstr string) {
	dalib.Warmup(connstr)

	var authMiddleware, err = jwt.New(&jwt.GinJWTMiddleware{
		Realm:           "Signaux-Faibles",
		Key:             []byte(jwtsecret),
		SendCookie:      false,
		Timeout:         time.Hour,
		MaxRefresh:      time.Hour,
		IdentityKey:     "id",
		PayloadFunc:     payloadHandler,
		IdentityHandler: identityHandler,
		Authenticator:   authenticatorHandler,
		Authorizator:    authorizatorHandler,
		Unauthorized:    unauthorizedHandler,
		TokenLookup:     "header: Authorization, query: token",
		TokenHeadName:   "Bearer",
		TimeFunc:        time.Now,
	})

	if err != nil {
		panic("authMiddleWare can't be instanciated: " + err.Error())
	}
	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	config.AddAllowHeaders("Authorization")
	config.AddAllowMethods("GET", "POST")
	router.Use(cors.New(config))

	router.POST("/login", authMiddleware.LoginHandler)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/refresh", authMiddleware.MiddlewareFunc(), authMiddleware.RefreshHandler)
	router.GET("/ws/:token", wsTokenMiddleWare, authMiddleware.MiddlewareFunc(), wsHandler)

	data := router.Group("data")

	data.Use(authMiddleware.MiddlewareFunc())
	data.POST("/get/:bucket", get)
	data.POST("/put/:bucket", put)

	router.Run(bind)
}
