package petrinet

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"fmt"

	"github.com/FLAGlab/sleuth"
)

//@TODO Update the handler to manage the requests
type petriHandler struct {
	place *Place
}

func (ph *petriHandler) Init(p *Place) {
	ph.place = p
}

func (h *petriHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	body, _ := ioutil.ReadAll(req.Body)
	//fmt.Printf("%v",body)
	var tokens []Token
	err := json.Unmarshal(body,&tokens)
	if err != nil {
		fmt.Printf("Fallo :v \n")
	}
	h.place.AddMarks(tokens)
	res.Write(body)
}

//ServiceNode structure of a node associated to its place
type ServiceNode struct {
	Interface string
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
		Interface: sn.Interface,
		LogLevel: "info",
		Service:  sn.ServiceName,
	}
	server, err := sleuth.New(config)

	if err != nil {
		fmt.Println("failed here")
		panic(err.Error())
	}
	defer server.Close()
	http.ListenAndServe(":9000", handler)
}
