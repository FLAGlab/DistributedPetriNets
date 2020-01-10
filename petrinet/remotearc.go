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
}

func (rt RemoteArc) String() string {
	return fmt.Sprintf("{arc to service: %v, weight: %v}", rt.ServiceName, rt.Weight)
}

//@TODO
func (rt *RemoteArc) canFire() bool {
	return true
}

//@TODO
func (rt *RemoteArc) fire() {
	config := &sleuth.Config{LogLevel: "debug"}
	client, err := sleuth.New(config)
	if err != nil {
		panic(err.Error())
	}
	defer client.Close()
	client.WaitFor(rt.ServiceName)
	input := "1"
	body := bytes.NewBuffer([]byte(input))
	request, _ := http.NewRequest("POST", "sleuth://"+rt.ServiceName+"/", body)
	response, err := client.Do(request)
	if err != nil {
		panic(err.Error())
	}
	output, _ := ioutil.ReadAll(response.Body)
	if string(output) == input {
		fmt.Println("It works.")
	} else {
		fmt.Println("It doesn't work.")
	}
}
