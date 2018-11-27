package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"os"
	"os/exec"
)

func HandleRequest() {
	stdout, err := exec.Command("./bin/aws-iam-authenticator").Output()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("%s", stdout)
}

func main() {
	lambda.Start(HandleRequest)
}
