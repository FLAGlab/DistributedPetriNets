package communication

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ursiform/sleuth"
)

//@TODO write a suitble handler
type echoHandler struct{}

func (h *echoHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	body, _ := ioutil.ReadAll(req.Body)
	fmt.Println("Message")
	fmt.Printf("%v\n", body)
	res.Write(body)
}

type ServiceNode struct {
	ServiceName string
}

func (sn *ServiceNode) RunService() {

	handler := new(echoHandler)

	config := &sleuth.Config{
		Handler: handler,
		// this interface is for test purposes only
		Interface: "wlp1s0",
		LogLevel:  "debug",
		Service:   sn.ServiceName,
	}
	server, err := sleuth.New(config)
	if err != nil {
		panic(err.Error())
	}
	defer server.Close()
	http.ListenAndServe(":9000", handler)
}
