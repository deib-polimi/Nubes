package main

import (
	"fmt"

	lib "github.com/Astenna/Nubes/lib"
	"github.com/aws/aws-lambda-go/lambda"
)

func AddToManyToManyHandler(input lib.AddToManyToManyParam) error {

	if err := input.Verify(); err != nil {
		return err
	}

	exists, err := lib.IsInstanceAlreadyCreated(lib.IsInstanceAlreadyCreatedParam{Id: input.NewId, TypeName: input.TypeName})
	if err != nil {
		return fmt.Errorf("error occurred while checking if typename %s with id %s exists. Error %w", input.TypeName, input.NewId, err)
	}
	if !exists {
		return fmt.Errorf("only existing instances can be added to many to many relationships. Typename %s with id %s not found", input.TypeName, input.NewId)
	}

	if input.UsesIndex {
		return lib.InsertToManyToManyTable(input.TypeName, input.OwnerTypeName, input.NewId, input.OwnerId)
	}
	return lib.InsertToManyToManyTable(input.OwnerTypeName, input.TypeName, input.OwnerId, input.NewId)
}

func main() {
	lambda.Start(AddToManyToManyHandler)
}