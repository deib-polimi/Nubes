package main

import (
	lib "github.com/Astenna/Nubes/lib"
	"github.com/aws/aws-lambda-go/lambda"
)

func GetStateHandler(input lib.GetStateParam) (interface{}, error) {
	var output interface{}
	var err error

	if input.GetStub {
		output, err = lib.GetObjectStateWithTypeNameAsArg(input.Id, input.TypeName)
	} else {
		output, err = lib.GetField(input)
	}

	if err != nil {
		return *new(interface{}), err
	}
	return output, nil
}

func main() {
	lambda.Start(GetStateHandler)
}
