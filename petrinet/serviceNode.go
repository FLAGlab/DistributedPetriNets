package petrinet

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ursiform/sleuth"
)

//@TODO write a suitble handler
type petriHandler struct {
	place *Place
}

func (ph *petriHandler) Init(p *Place) {
	ph.place = p
}

func (h *petriHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	body, _ := ioutil.ReadAll(req.Body)
	fmt.Println("Addign Token")
	fmt.Printf("====Old Marks %v \n", h.place.GetMarks())
	h.place.Marks++
	fmt.Printf("====New marks %v \n", h.place.GetMarks())
	res.Write(body)
}

//ServiceNode structure of a node associated to its place
type ServiceNode struct {
	PetriPlace  *Place
	ServiceName string
}

//RunService executes the node's server and client
func (sn *ServiceNode) RunService() {
	server(sn)
}

func server(sn *ServiceNode) {
	handler := new(petriHandler)
	handler.Init(sn.PetriPlace)

	config := &sleuth.Config{
		Handler: handler,
		// this interface is for test purposes only
		//Interface: "wlp1s0",
		//LogLevel: "debug",
		//Port:     6000,
		Service:  sn.ServiceName,
	}
	server, err := sleuth.New(config)

	if err != nil {
		panic(err.Error())
	}
	defer server.Close()
	if sn.ServiceName == "ping" {
		http.ListenAndServe(":8080", handler)
	} else {
		http.ListenAndServe(":8081", handler)
	}
}
