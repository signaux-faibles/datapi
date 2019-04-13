package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	dalib "github.com/signaux-faibles/datapi/lib"
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

func wsTokenMiddleWare(c *gin.Context) {
	token := c.Request.URL.RequestURI()[4:]
	c.Request.Header["Authorization"] = []string{"Bearer " + token}
	c.Next()
}

func wsHandler(c *gin.Context) {
	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Printf("Failed to set websocket upgrade: %+v", err)
		return
	}

	// go func() {
	// 	for {
	// 		err := conn.WriteJSON([]byte("coucou"))
	// 		if err != nil {
	// 			return
	// 		}
	// 		time.Sleep(time.Second)
	// 	}
	// }()

	for {
		t, msg, err := conn.ReadMessage()
		fmt.Println(string(msg))
		if err != nil {
			break
		}
		myMsg := map[string]string{
			"test": "test",
		}
		jsonMsg, _ := json.Marshal(myMsg)
		conn.WriteMessage(t, jsonMsg)
	}

	fmt.Println("ended")
}

func put(c *gin.Context) {
	bucket := c.Params.ByName("bucket")

	var objects []dalib.Object
	err := c.Bind(&objects)
	if err != nil {
		c.JSON(400, "bad request: "+err.Error())
		return
	}

	u, ok := c.Get("id")
	if !ok {
		c.JSON(401, "malformed token")
	}

	user, ok := u.(*dalib.User)
	if !ok {
		c.JSON(401, "malformed token")
	}

	err = dalib.Insert(objects, bucket, user.Scope, user.Email)
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

	u, _ := c.Get("id")
	user := u.(*dalib.User)
	scope := user.Scope

	data, err := dalib.Query(bucket, params, scope, true, nil)

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
