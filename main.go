package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/Nerzal/gocloak"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/signaux-faibles/datapi/docs"
	dalib "github.com/signaux-faibles/datapi/lib"
	"github.com/spf13/viper"

	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

var keycloak gocloak.GoCloak

func main() {
	api := flag.Bool("api", false, "Runs API")
	createdb := flag.String("createdb", "", "Initialize postgres Database with provided pgsql credentials.")
	forge := flag.String("forge", "", "Forge Long Term JWT for auth")

	flag.Parse()

	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	viper.ReadInConfig()

	postgres := viper.GetString("postgres")

	bind := viper.GetString("bind")

	if *createdb != "" {
		dalib.CreateDB(*createdb, postgres)
		return
	}

	if *forge != "" {
		return
	}

	if *api {
		log.Println("starting api, listening " + bind)
		keycloak = gocloak.NewClient(viper.GetString("keycloakHostname"))

		runAPI(bind, postgres, &keycloak)
		return
	}

}

func runAPI(bind, postgres string, keycloak *gocloak.GoCloak) {

	dalib.Warmup(postgres)

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	config.AddAllowHeaders("Authorization")
	config.AddAllowMethods("GET", "POST")
	router.Use(cors.New(config))

	router.POST("/login", loginHandler)
	router.POST("/connectionEmail", connectionEmailHandler)
	router.POST("/refreshToken", refreshTokenHandler)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/prepare", prepare)
	router.GET("/refresh")
	// router.GET("/ws/:token", wsTokenMiddleWare, authMiddleware.MiddlewareFunc(), wsHandler)

	data := router.Group("data")

	data.Use(keycloakMiddleware)
	data.POST("/get/:bucket", get)
	data.POST("/put/:bucket", put)

	data.POST("/cache/:bucket", cache)

	err := router.Run(bind)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func keycloakMiddleware(c *gin.Context) {
	header := c.Request.Header["Authorization"][0]
	rawToken := strings.Split(header, " ")[1]

	token, claims, err := keycloak.DecodeAccessToken(rawToken, viper.GetString("keycloakRealm"))
	if errValid := claims.Valid(); err != nil && errValid != nil {
		c.AbortWithStatus(401)
	}
	var emailString, nameString, firstNameString string

	email, ok := (*claims)["email"]
	if ok {
		emailString = email.(string)
	}
	name, ok := (*claims)["family_name"]
	if ok {
		nameString = name.(string)
	}
	firstName, ok := (*claims)["given_name"]
	if ok {
		firstNameString = firstName.(string)
	}

	user := dalib.User{
		Email:     emailString,
		Name:      nameString,
		FirstName: firstNameString,
		Scope:     scopeFromClaims(claims),
	}
	c.Set("token", token)
	c.Set("claims", claims)
	c.Set("user", &user)

	c.Next()
}

func scopeFromClaims(claims *jwt.MapClaims) dalib.Tags {
	var resourceAccess = make(map[string]interface{})
	var client = make(map[string]interface{})
	var scope []interface{}

	resourceAccessInterface, ok := (*claims)["resource_access"]
	if ok {
		resourceAccess, _ = resourceAccessInterface.(map[string]interface{})
	}

	clientInterface, ok := (resourceAccess)["signauxfaibles"]
	if ok {
		client, _ = clientInterface.(map[string]interface{})
	}

	scopeInterface, ok := (client)["roles"]
	if ok {
		scope, _ = scopeInterface.([]interface{})
	}

	var tags dalib.Tags
	for _, tag := range scope {
		tags = append(tags, tag.(string))
	}

	return tags
}
