package main

import (
	"flag"
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
	// connstr := flag.String("pg", "", "Postgres connection string")
	// jwtsecret := flag.String("jwt", "", "Secret key for JWT signature")
	// createuser := flag.String("createuser", "", "Create the first datapi user with user:password syntax")

	flag.Parse()

	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	viper.ReadInConfig()

	postgres := viper.GetString("postgres")
	jwtSecret := viper.GetString("jwtSecret")
	bind := viper.GetString("bind")

	if *createdb != "" {
		dalib.CreateDB(*createdb, postgres)
		return
	}

	if *forge != "" {
		return
	}

	if *api {
		keycloak = gocloak.NewClient("http://localhost:8080/")

		runAPI(bind, jwtSecret, postgres, &keycloak)
		return
	}

}

func runAPI(bind, jwtsecret, postgres string, keycloak *gocloak.GoCloak) {

	dalib.Warmup(postgres)

	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	config.AddAllowHeaders("Authorization")
	config.AddAllowMethods("GET", "POST")
	router.Use(cors.New(config))

	router.POST("/login", loginHandler)
	router.POST("/refreshToken", refreshTokenHandler)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/refresh")
	// router.GET("/ws/:token", wsTokenMiddleWare, authMiddleware.MiddlewareFunc(), wsHandler)

	data := router.Group("data")

	data.Use(keycloakMiddleware)
	data.POST("/get/:bucket", get)
	data.POST("/put/:bucket", put)

	router.Run(bind)
}

func keycloakMiddleware(c *gin.Context) {
	header := c.Request.Header["Authorization"][0]

	rawToken := strings.Split(header, " ")[1]

	token, claims, err := keycloak.DecodeAccessToken(rawToken, "master")
	if errValid := claims.Valid(); err != nil && errValid != nil {
		c.AbortWithStatus(401)
	}
	user := dalib.User{
		Email:     (*claims)["email"].(string),
		Name:      (*claims)["family_name"].(string),
		FirstName: (*claims)["given_name"].(string),
		Scope:     scopeFromClaims(claims),
	}
	c.Set("token", token)
	c.Set("claims", claims)
	c.Set("user", &user)

	c.Next()
}

func scopeFromClaims(claims *jwt.MapClaims) dalib.Tags {
	resourceAccess := (*claims)["resource_access"].(map[string]interface{})
	client := (resourceAccess)["signauxfaibles"].(map[string]interface{})
	scope := (client)["roles"].([]interface{})

	var tags dalib.Tags
	for _, tag := range scope {
		tags = append(tags, tag.(string))
	}
	return tags
}
