package main

import (
	"io/ioutil"
	"net/http"

	"github.com/ursiform/sleuth"
)

type echoHandler struct{}

func (h *echoHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	body, _ := ioutil.ReadAll(req.Body)
	res.Write(body)
}

func main() {
	handler := new(echoHandler)
	// In the real world, the Interface field of the sleuth.Config object
	// should be set so that all services are on the same subnet.
	config := &sleuth.Config{
		Handler:  handler,
		//Interface: "wlp1s0",
		LogLevel: "debug",
		Service:  "echo-service",
	}
	server, err := sleuth.New(config)
	if err != nil {
		panic(err.Error())
	}
	defer server.Close()
	http.ListenAndServe(":8080", handler)
}
