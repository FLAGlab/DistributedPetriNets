package main

import (
	"io/ioutil"
	"net/http"

	"github.com/FLAGlab/sleuth"
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
		Interface: "enp2s0",
		LogLevel: "debug",
		Service:  "echo-service",
		Port: 5670,
	}
	server, err := sleuth.New(config)
	if err != nil {
		panic(err.Error())
	}
	defer server.Close()
	http.ListenAndServe(":5000", handler)
}
