package petrinet

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ursiform/sleuth"
)

// RemoteArc for arcs crossing nodes
type RemoteArc struct {
	ServiceName string
	Weight  int
	Client sleuth.Client
}

func (rt *RemoteArc) Init() {
	config := &sleuth.Config{LogLevel: "debug"}
	rt.Client, err := sleuth.New(config)
	if err != nil {
		panic(err.Error())
	}
	//defer client.Close()
}


func (rt RemoteArc) String() string {
	return fmt.Sprintf("{arc to service: %v, weight: %v}", rt.ServiceName, rt.Weight)
}

//@TODO
func (rt *RemoteArc) canFire() bool {
	return true
}

//@TODO
func (rt *RemoteArc) fire() bool {
	rt.Client.WaitFor(rt.ServiceName)
	input := "This is the value I am inputting."
	body := bytes.NewBuffer([]byte(input))
	fmt.Println("Hey Hey llegue aca")
	request, _ := http.NewRequest("POST", "sleuth://"+rt.ServiceName+"/", body)
	response, err := rt.Client.Do(request)
	fmt.Println("Hey si pude")
	if err != nil {
		//panic(err.Error())
		return false
	}
	output, _ := ioutil.ReadAll(response.Body)
	if string(output) == input {
		fmt.Println("It works.")
	} else {
		fmt.Println("It doesn't work.")
	}
	return true
}
