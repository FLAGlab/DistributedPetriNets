package petrinet

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"encoding/json"

	"github.com/ursiform/sleuth"
)

// RemoteArc for arcs crossing nodes
type RemoteArc struct {
	ServiceName string
	Weight  int
	Client *sleuth.Client
}

func (rt *RemoteArc) Init() {
	config := &sleuth.Config{}//LogLevel: "debug"}
	client, err := sleuth.New(config)
	rt.Client = client
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
	rt.Client.WaitFor(rt.ServiceName)
	return true
}

//@TODO update fire to hanlde time outs
func (rt *RemoteArc) fire(t []Token) bool {
	rt.Client.WaitFor(rt.ServiceName)
	t = t[0:rt.Weight]
	fmt.Printf("%v\n",t)
	vals, err := json.Marshal(t)
	if err != nil {
		return false
	}
	fmt.Printf("%v\n",vals)
	body := bytes.NewBuffer(vals)
	request, _ := http.NewRequest("POST", "sleuth://"+rt.ServiceName+"/", body)
	response, err := rt.Client.Do(request)
	if err != nil {
		//panic(err.Error())
		return false
	}
	//fmt.Println("Hey si pude")
	output, _ := ioutil.ReadAll(response.Body)
	fmt.Printf("%v\n", output)
	return true
}
