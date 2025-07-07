package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
)

type TestController struct{}

func (t *TestController) TestHandler(c *gin.Context) {
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Convert the body bytes to a string
	bodyString := string(bodyBytes)

	// Print the request body
	fmt.Println("Request Body as received:", bodyString)
	c.JSON(http.StatusOK, gin.H{"message": "Hello, this is a test"})
}
