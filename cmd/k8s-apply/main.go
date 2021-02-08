package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

type ModuleArgs struct {
	File string
}

type Response struct {
	Msg     string `json:"msg"`
	Changed bool   `json:"changed"`
	Failed  bool   `json:"failed"`
}

func ExitJson(responseBody Response) {
	returnResponse(responseBody)
}

func FailJson(responseBody Response) {
	responseBody.Failed = true
	returnResponse(responseBody)
}

func returnResponse(responseBody Response) {
	var response []byte
	var err error
	response, err = json.Marshal(responseBody)
	if err != nil {
		response, _ = json.Marshal(Response{Msg: "Invalid response object"})
	}
	fmt.Println(string(response))
	if responseBody.Failed {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

func command(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	err := cmd.Run()
	buffer, outputErr := cmd.CombinedOutput()

	if err == nil {
		return nil
	}
	if outputErr != nil {
		return errors.New(err.Error() + outputErr.Error())
	}

	return errors.New(string(buffer))
}

func main() {
	var response Response

	if len(os.Args) != 2 {
		response.Msg = "No argument file provided"
		FailJson(response)
	}

	argsFile := os.Args[1]

	text, err := ioutil.ReadFile(argsFile)
	if err != nil {
		response.Msg = "Could not read configuration file: " + argsFile
		FailJson(response)
	}

	var moduleArgs ModuleArgs
	err = json.Unmarshal(text, &moduleArgs)
	if err != nil {
		response.Msg = "Configuration file not valid JSON: " + argsFile
		FailJson(response)
	}

	err = command("kubectl", "diff", "-f", moduleArgs.File)

	if err == nil {
		response.Msg = "No diff found"
		ExitJson(response)
	}

	err = command("kubectl", "apply", "-f", moduleArgs.File)
	if err != nil {
		response.Msg = err.Error()
		FailJson(response)
	}

	response.Msg = "Applied successfully"
	ExitJson(response)
}
