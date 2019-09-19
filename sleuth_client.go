package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ursiform/sleuth"
)

func main() {
	service := "echo-service"
	// In the real world, the Interface field of the sleuth.Config object
	// should be set so that all services are on the same subnet.
	config := &sleuth.Config{LogLevel: "debug"}
	client, err := sleuth.New(config)
	if err != nil {
		panic(err.Error())
	}
	defer client.Close()
	client.WaitFor(service)
	input := "This is the value I am inputting."
	body := bytes.NewBuffer([]byte(input))
	request, _ := http.NewRequest("POST", "sleuth://"+service+"/", body)
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
