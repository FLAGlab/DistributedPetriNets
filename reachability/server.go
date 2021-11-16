package reachability

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var graph = Rgraph{
	ID:        1,
	Timestamp: time.Now().Unix(),
}

func getGraph(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, graph)
}

func updateGraph() {
	for range time.Tick(time.Second * 30) {
		graph.Timestamp = time.Now().Unix()
	}
}

func Run() {
	router := gin.Default()
	router.GET("/graph", getGraph)
	go updateGraph()
	router.Run("localhost:9000")
}
