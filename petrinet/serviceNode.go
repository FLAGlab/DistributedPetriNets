package petrinet

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ursiform/sleuth"
)

//@TODO write a suitble handler
type echoHandler struct {
	place *Place
}

func (h *echoHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	body, _ := ioutil.ReadAll(req.Body)
	fmt.Println("Message")
	fmt.Printf("%v, %v\n", body, h.place.GetMarks())
	fmt.Fprintf(res, "%v", h.place.GetMarks())

}

//ServiceNode structure of a node associated to its place
type ServiceNode struct {
	PetriPlace  *Place
	ServiceName string
}

//RunService executes the node's server and client
func (sn *ServiceNode) RunService() {
	server(sn)
	client(sn)
}

func client(sn *ServiceNode) {
	if sn.ServiceName == "ping" {

	}
}

func server(sn *ServiceNode) {
	handler := &echoHandler{place: sn.PetriPlace}

	config := &sleuth.Config{
		Handler: handler,
		// this interface is for test purposes only
		//Interface: "wlp1s0",
		LogLevel: "debug",
		Port:     6000,
		Service:  sn.ServiceName,
	}
	server, err := sleuth.New(config)

	if config.Service == "ping" {

	} else if config.Service == "pong" {

	}
	if err != nil {
		panic(err.Error())
	}
	defer server.Close()
	http.ListenAndServe(":9000", handler)
}
