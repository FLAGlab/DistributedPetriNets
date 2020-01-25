package petrinet

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"fmt"

	"github.com/ursiform/sleuth"
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
	//fmt.Println("Addign Token")
	var tokens []Token
	err := json.Unmarshal(body,&tokens)
	if err != nil {
		fmt.Printf("Fallo :v \n")
	}
	//fmt.Printf("====Old Marks %v \n", h.place.GetMarks())
	h.place.AddMarks(tokens)
	//fmt.Printf("====New marks %v \n", h.place.GetMarks())
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
		Service:  sn.ServiceName,
	}
	server, err := sleuth.New(config)

	if err != nil {
		panic(err.Error())
	}
	defer server.Close()
	http.ListenAndServe(":0", handler)
}
