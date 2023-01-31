package main

import (
	org "github.com/Astenna/Nubes/evaluation/movie_review/faas/types"
	"github.com/Astenna/Nubes/lib"
	"github.com/aws/aws-lambda-go/lambda"
)

func VerifyPasswordHandler(input lib.HandlerParameters) (bool, error) {
	instance := new(org.Account)
	instance.Email = input.Id
	instance.Init()

	result, _err := instance.VerifyPassword(input.Parameter.(string))
	if _err != nil {
		return result, _err
	}

	return result, _err
}

func main() {
	lambda.Start(VerifyPasswordHandler)
}