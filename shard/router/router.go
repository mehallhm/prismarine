package router

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func Configure() {
	router := gin.New()

	router.GET("/", getHome)
	err := router.Run("localhost:8000")
	if err != nil {
		return
	}
}

func getHome(c *gin.Context) {
	c.String(http.StatusOK, "Hello World")
}
