package main

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

type Item struct {
	ID   string `json:"id"`
	Amount float32 `json:"amount"`
}

var items = []Item{
	{ID: "BTC/CHF", Amount: 49000.12},
	{ID: "BTC/EUR", Amount: 50000.12},
	{ID: "BTC/USD", Amount: 52000.12},
}

func main() {
	router := gin.Default()

	router.GET("/", getItems)

	router.Run(":8080") // Run on port 8080
}

func getItems(c *gin.Context) {
	c.JSON(http.StatusOK, items)
}
