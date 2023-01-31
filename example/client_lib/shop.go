package client_lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/Astenna/Nubes/lib"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
)

type shop struct {
	id string

	Products referenceNavigationList[product, ProductStub]

	Owners referenceNavigationList[user, UserStub]
}

// ALL THE CODE BELOW IS GENERATED ONLY FOR NOBJECTS TYPES
func (shop) GetTypeName() string {
	return "Shop"
}

// LOAD AND EXPORT

func LoadShop(id string) (*shop, error) {
	newInstance := new(shop)

	params := lib.HandlerParameters{
		Id:       id,
		TypeName: newInstance.GetTypeName(),
	}
	jsonParam, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	out, _err := LambdaClient.Invoke(&lambda.InvokeInput{FunctionName: aws.String("Load"), Payload: jsonParam})
	if _err != nil {
		return nil, _err
	}
	if out.FunctionError != nil {
		return nil, fmt.Errorf("lambda function designed to verify if instance exists failed. Error: %s", string(out.Payload))
	}

	newInstance.id = id
	newInstance.init()
	return newInstance, nil
}

func loadShopWithoutCheckIfExists(id string) *shop {
	newInstance := new(shop)
	newInstance.id = id
	return newInstance
}

// setId interface for initilization in ReferenceNavigationList
func (u *shop) setId(id string) {
	u.id = id
}

func (r *shop) init() {

	r.Products = *newReferenceNavigationList[product, ProductStub](r.id, r.GetTypeName(), "SoldBy", false)

	r.Owners = *newReferenceNavigationList[user, UserStub](r.id, r.GetTypeName(), "", true)

}

func ExportShop(input ShopStub) (*shop, error) {
	newInstance := new(shop)

	params := lib.HandlerParameters{
		TypeName:  newInstance.GetTypeName(),
		Parameter: input,
	}
	jsonParam, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	out, _err := LambdaClient.Invoke(&lambda.InvokeInput{FunctionName: aws.String("Export"), Payload: jsonParam})
	if _err != nil {
		return nil, _err
	}
	if out.FunctionError != nil {
		return nil, fmt.Errorf("lambda function designed to verify if instance exists failed. Error: %s", string(out.Payload[:]))
	}

	newInstance.id, err = strconv.Unquote(string(out.Payload[:]))
	newInstance.init()
	return newInstance, err
}

// DELETE

func DeleteShop(id string) error {
	newInstance := new(shop)

	params := lib.HandlerParameters{
		Id:       id,
		TypeName: newInstance.GetTypeName(),
	}
	jsonParam, err := json.Marshal(params)
	if err != nil {
		return err
	}

	out, _err := LambdaClient.Invoke(&lambda.InvokeInput{FunctionName: aws.String("Delete"), Payload: jsonParam})
	if _err != nil {
		return _err
	}
	if out.FunctionError != nil {
		return fmt.Errorf("lambda function failed. Error: %s", string(out.Payload))
	}

	return nil
}

// GETID

func (s shop) GetId() string {
	return s.id
}

// GETTERS AND SETTERS

func (s shop) GetName() (string, error) {
	if s.id == "" {
		return *new(string), errors.New("id of the type not set, use  LoadShop or ExportShop to create new instance of the type")
	}

	params := lib.GetStateParam{
		Id:        s.GetId(),
		TypeName:  s.GetTypeName(),
		FieldName: "Name",
	}
	jsonParam, err := json.Marshal(params)
	if err != nil {
		return *new(string), err
	}

	out, _err := LambdaClient.Invoke(&lambda.InvokeInput{FunctionName: aws.String("GetField"), Payload: jsonParam})
	if _err != nil {
		return *new(string), _err
	}
	if out.FunctionError != nil {
		return *new(string), fmt.Errorf(string(out.Payload[:]))
	}

	result := new(string)
	err = json.Unmarshal(out.Payload, result)
	if err != nil {
		return *new(string), err
	}
	return *result, err

}

func (s shop) SetName(newValue string) error {
	if s.id == "" {
		return errors.New("id of the type not set, use LoadShop or ExportShop to create new instance of the type")
	}

	params := lib.SetFieldParam{
		Id:        s.GetId(),
		TypeName:  s.GetTypeName(),
		FieldName: "Name",
		Value:     newValue,
	}
	jsonParam, err := json.Marshal(params)
	if err != nil {
		return err
	}

	out, _err := LambdaClient.Invoke(&lambda.InvokeInput{FunctionName: aws.String("SetField"), Payload: jsonParam})
	if _err != nil {
		return _err
	}
	if out.FunctionError != nil {
		return fmt.Errorf(string(out.Payload[:]))
	}
	return nil
}

// (STATE-CHANGING) METHODS

func (r shop) GetStub() (ShopStub, error) {
	if r.id == "" {
		return *new(ShopStub), errors.New("id of the type not set, use  LoadShop or ExportShop to create new instance of the type")
	}

	params := lib.GetStateParam{
		Id:       r.GetId(),
		TypeName: r.GetTypeName(),
		GetStub:  true,
	}
	jsonParam, err := json.Marshal(params)
	if err != nil {
		return *new(ShopStub), err
	}

	out, _err := LambdaClient.Invoke(&lambda.InvokeInput{FunctionName: aws.String("GetState"), Payload: jsonParam})
	if _err != nil {
		return *new(ShopStub), _err
	}
	if out.FunctionError != nil {
		return *new(ShopStub), fmt.Errorf(string(out.Payload[:]))
	}

	result := new(ShopStub)
	err = json.Unmarshal(out.Payload, result)
	if err != nil {
		return *new(ShopStub), err
	}
	return *result, err
}