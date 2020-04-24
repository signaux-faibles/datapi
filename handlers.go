package main

import (
	"net/http"

	"github.com/Nerzal/gocloak/v3"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	dalib "github.com/signaux-faibles/datapi/lib"
	"github.com/spf13/viper"
)

var wsChannel chan dalib.Object

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     check,
}

func check(request *http.Request) bool {
	return true
}

func loginHandler(c *gin.Context) {
	keycloakClientID := viper.GetString("keycloakClientID")
	keycloakClientSecret := viper.GetString("keycloakClientSecret")
	keycloakRealm := viper.GetString("keycloakRealm")

	vals := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}

	c.Bind(&vals)
	token, err := keycloak.Login(
		keycloakClientID,
		keycloakClientSecret,
		keycloakRealm,
		vals.Email,
		vals.Password)

	if err != nil {
		message := map[string]string{
			"message": "not ok",
			"error":   err.Error(),
		}
		c.JSON(401, message)
		return
	}

	c.JSON(200, token)
}

func connectionEmailHandler(c *gin.Context) {
	var request struct {
		Email string `json:"email"`
	}

	err := c.Bind(&request)
	if err != nil {
		return
	}

	keycloakRealm := viper.GetString("keycloakRealm")
	keycloakClientID := viper.GetString("keycloakClientID")
	keycloakAdmin := viper.GetString("keycloakAdmin")
	keycloakAdminRealm := viper.GetString("keycloakAdminRealm")
	keycloakPassword := viper.GetString("keycloakPassword")
	jwt, err := keycloak.LoginAdmin(
		keycloakAdmin,
		keycloakPassword,
		keycloakAdminRealm)
	if err != nil {
		c.JSON(500, "Problème de communication avec le système d'authentification: "+err.Error())
		return
	}

	user, err := keycloak.GetUsers(jwt.AccessToken, keycloakAdminRealm, gocloak.GetUsersParams{
		Email: request.Email,
	})

	if err != nil {
		c.JSON(500, "problème de communication avec le système d'authentification: "+err.Error())
	}

	if len(user) != 1 {
		return
	}

	lifespan := 60 * 15
	keycloak.ExecuteActionsEmail(jwt.AccessToken, keycloakRealm, gocloak.ExecuteActionsEmail{
		UserID:   user[0].ID,
		ClientID: keycloakClientID,
		Lifespan: lifespan,
		Actions:  []string{"UPDATE_PASSWORD"},
	})
}

func refreshTokenHandler(c *gin.Context) {
	var request struct {
		RefreshToken string `json:"refresh_token"`
	}
	c.Bind(&request)
	keycloakClientID := viper.GetString("keycloakClientID")
	keycloakClientSecret := viper.GetString("keycloakClientSecret")
	keycloakRealm := viper.GetString("keycloakRealm")

	token, err := keycloak.RefreshToken(request.RefreshToken, keycloakClientID, keycloakClientSecret, keycloakRealm)

	if err != nil {
		message := map[string]string{
			"message": "not ok",
			"error":   err.Error(),
		}
		c.JSON(401, message)
		return
	}

	c.JSON(200, token)
}

// For use in next release supporting websockets
// func wsTokenMiddleWare(c *gin.Context) {
// 	token := c.Request.URL.RequestURI()[4:]
// 	c.Request.Header["Authorization"] = []string{"Bearer " + token}
// 	c.Next()
// }

// func wsHandler(c *gin.Context) {
// 	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
// 	if err != nil {
// 		fmt.Printf("Failed to set websocket upgrade: %+v", err)
// 		return
// 	}

// 	// go func() {
// 	// 	for {
// 	// 		err := conn.WriteJSON([]byte("coucou"))
// 	// 		if err != nil {
// 	// 			return
// 	// 		}
// 	// 		time.Sleep(time.Second)
// 	// 	}
// 	// }()

// 	for {
// 		t, msg, err := conn.ReadMessage()
// 		fmt.Println(string(msg))
// 		if err != nil {
// 			break
// 		}
// 		myMsg := map[string]string{
// 			"test": "test",
// 		}
// 		jsonMsg, _ := json.Marshal(myMsg)
// 		conn.WriteMessage(t, jsonMsg)
// 	}

// 	fmt.Println("ended")
// }

func prepare(c *gin.Context) {
	err := dalib.CachePrepare()

	if err != nil {
		c.JSON(500, "bad request: "+err.Error())
	}
}

func put(c *gin.Context) {
	bucket := c.Params.ByName("bucket")

	var objects []dalib.Object
	err := c.Bind(&objects)

	if err != nil {
		c.JSON(400, "bad request: "+err.Error())
		return
	}

	u, ok := c.Get("user")
	user := u.(*dalib.User)
	if !ok {
		c.JSON(401, "malformed token")
	}

	vars := map[string]string{
		"$user.email": user.Email,
	}

	err = dalib.Insert(objects, bucket, user.Scope, user.Email, vars)
	if err != nil {
		if daerror, ok := err.(dalib.DAError); ok {
			c.JSON(daerror.Code, daerror.Message)
			return
		}
		c.JSON(500, "Server error: "+err.Error())
		return
	}

	c.JSON(200, "ok")

}

func get(c *gin.Context) {
	var params dalib.QueryParams

	err := c.ShouldBind(&params)
	if err != nil {
		c.JSON(400, err.Error())
	}
	bucket := c.Params.ByName("bucket")

	u, ok := c.Get("user")
	if !ok {
		c.JSON(500, "No user information in gin context, interrupting")
		return
	}
	user, ok := u.(*dalib.User)
	if !ok {
		c.JSON(500, "Erroneous user information in gin context, interrupting")
		return
	}
	rawToken, ok := c.Get("token")
	if !ok {
		c.JSON(500, "No JWT object in gin context, interrupting")
		return
	}
	token, ok := rawToken.(*jwt.Token)
	if !ok {
		c.JSON(500, "JWT object is not a string, interrupting")
		return
	}

	scope := user.Scope

	data, err := dalib.Query(bucket, params, scope, true, nil, token.Raw)

	if err != nil {
		if daerror, ok := err.(dalib.DAError); ok {
			c.JSON(daerror.Code, daerror.Message)
			return
		}
		c.JSON(500, err.Error())
		return
	}

	if len(data) == 0 {
		c.JSON(404, "no data")
		return
	}

	c.JSON(200, data)
}

func cache(c *gin.Context) {
	var params dalib.QueryParams

	err := c.ShouldBind(&params)
	if err != nil {
		c.JSON(400, err.Error())
	}
	bucket := c.Params.ByName("bucket")

	u, ok := c.Get("user")
	if !ok {
		c.JSON(500, "No user information in gin context, interrupting")
		return
	}
	user, ok := u.(*dalib.User)
	if !ok {
		c.JSON(500, "Erroneous user information in gin context, interrupting")
		return
	}
	rawToken, ok := c.Get("token")
	if !ok {
		c.JSON(500, "No JWT object in gin context, interrupting")
		return
	}
	token, ok := rawToken.(*jwt.Token)
	if !ok {
		c.JSON(500, "JWT object is not a string, interrupting")
		return
	}

	scope := user.Scope

	data, err := dalib.Cache(bucket, params, scope, true, nil, token.Raw, scope.ToHash())

	if err != nil {
		if daerror, ok := err.(dalib.DAError); ok {
			c.JSON(daerror.Code, daerror.Message)
			return
		}
		c.JSON(500, err.Error())
		return
	}

	if len(data) == 0 {
		c.JSON(404, "no data")
		return
	}

	c.JSON(200, data)
}
