package lib

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/google/uuid"
)

func Insert(objToInsert Nobject) (string, error) {
	var attributeVals, err = dynamodbattribute.MarshalMap(objToInsert)
	if err != nil {
		return "", err
	}

	var newId string
	if custom, ok := objToInsert.(CustomId); ok {
		if newId = custom.GetId(); newId == "" {
			return "", errors.New("id field empty. It must be set when using non-default id field")
		}
	} else {
		newId = uuid.New().String()
		attr := attributeVals["Id"].SetS(newId)
		// without this, dynamodb throws error because more than
		// one of the supported datatypes is set to not nil
		attr.NULL = nil
	}

	input := &dynamodb.PutItemInput{
		Item:      attributeVals,
		TableName: aws.String(objToInsert.GetTypeName()),
	}

	_, err = DBClient.PutItem(input)
	if err != nil {
		return "", err
	}

	return newId, nil
}

func Upsert(objToInsert Nobject, id string) error {
	var attributeVals, err = dynamodbattribute.MarshalMap(objToInsert)
	if err != nil {
		return err
	}

	attr := attributeVals["Id"].SetS(id)
	// without this, dynamodb throws error because more than
	// one of the supported datatypes is set to not nil
	attr.NULL = nil

	input := &dynamodb.PutItemInput{
		Item:      attributeVals,
		TableName: aws.String(objToInsert.GetTypeName()),
	}

	_, err = DBClient.PutItem(input)
	if err != nil {
		return err
	}

	return nil
}

func Delete[T Nobject](id string) error {
	if id == "" {
		return fmt.Errorf("missing id of object to delete")
	}

	input := &dynamodb.DeleteItemInput{
		TableName: aws.String((*new(T)).GetTypeName()),
		Key: map[string]*dynamodb.AttributeValue{
			"Id": {
				S: aws.String(id),
			},
		},
	}

	_, err := DBClient.DeleteItem(input)
	return err
}

func GetObjectState[T Nobject](id string) (*T, error) {
	if id == "" {
		return nil, fmt.Errorf("missing id of object to get")
	}

	input := &dynamodb.GetItemInput{
		TableName: aws.String((*new(T)).GetTypeName()),
		Key: map[string]*dynamodb.AttributeValue{
			"Id": {
				S: aws.String(id),
			},
		},
	}

	item, err := DBClient.GetItem(input)
	if err != nil {
		return nil, err
	}

	var parsedItem = new(T)
	if item.Item != nil {
		err = dynamodbattribute.UnmarshalMap(item.Item, parsedItem)
		return parsedItem, err
	}

	return nil, err
}

func GetByIndex[T Nobject](attributeName string, attributeValue string) ([]string, error) {
	if attributeName == "" {
		return nil, fmt.Errorf("missing attributeName")
	}

	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String((*new(T)).GetTypeName()),
		IndexName:              aws.String((*new(T)).GetTypeName() + attributeName),
		KeyConditionExpression: aws.String("gsi1pk = :gsi1pk and gsi1sk > :gsi1sk"),
	}

	items, err := DBClient.Query(queryInput)
	_ = items
	return nil, err
}

func GetBatch[T Nobject](ids []string) (*[]T, error) {
	if ids == nil {
		return nil, fmt.Errorf("missing id of object to get")
	}

	keysToRetrieve := make([]map[string]*dynamodb.AttributeValue, len(ids))
	for i, id := range ids {
		keysToRetrieve[i] = map[string]*dynamodb.AttributeValue{"Id": {
			S: aws.String(id),
		}}
	}

	tableName := (*new(T)).GetTypeName()
	input := &dynamodb.BatchGetItemInput{
		RequestItems: map[string]*dynamodb.KeysAndAttributes{
			tableName: {
				Keys: keysToRetrieve,
			},
		},
	}

	items, err := DBClient.BatchGetItem(input)
	if err != nil {
		return nil, err
	}

	var parsedItem = new([]T)
	if items.Responses[tableName] != nil {

		err = dynamodbattribute.UnmarshalListOfMaps(items.Responses[tableName], parsedItem)
		return parsedItem, err
	}

	return nil, err
}

func Update[T Nobject](values aws.JSONValue) error {
	if len(values) == 0 {
		return fmt.Errorf("no values specified for update")
	}
	if values["id"] == "" {
		return fmt.Errorf("missing id of object to update")
	}

	update := expression.UpdateBuilder{}
	for k, v := range values {
		update = update.Set(expression.Name(k), expression.Value(v))
	}
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		return fmt.Errorf("error occurred when building dynamodb update expression %w", err)
	}

	_, err = DBClient.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String((*new(T)).GetTypeName()),
		Key: map[string]*dynamodb.AttributeValue{
			"Id": {
				S: aws.String(values["Id"].(string)),
			},
		},
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})

	return err
}

type GetFieldParam struct {
	FieldName string
	TypeName  string
}

func GetField(id string, param GetFieldParam) (interface{}, error) {
	if id == "" {
		return *new(interface{}), fmt.Errorf("missing id of object's field  to get")
	}
	if param.FieldName == "" {
		return *new(interface{}), fmt.Errorf("missing field name of object's field to get")
	}
	if param.TypeName == "" {
		return *new(interface{}), fmt.Errorf("missing type name of object's field to get")
	}

	input := &dynamodb.GetItemInput{
		TableName: aws.String(param.TypeName),
		Key: map[string]*dynamodb.AttributeValue{
			"Id": {
				S: aws.String(id),
			},
		},
		ProjectionExpression: &param.FieldName,
	}

	item, err := DBClient.GetItem(input)
	if err != nil {
		return *new(interface{}), err
	}

	if item.Item != nil {
		var parsedItem interface{}
		err = dynamodbattribute.Unmarshal(item.Item[param.FieldName], &parsedItem)
		return parsedItem, err

	}

	return nil, err
}

type SetFieldParam struct {
	FieldName string
	TypeName  string
	Value     interface{}
}

func SetField(id string, param SetFieldParam) error {
	if id == "" {
		return fmt.Errorf("missing id of object's field  to get")
	}
	if param.FieldName == "" {
		return fmt.Errorf("missing field name of object's field to get")
	}
	if param.TypeName == "" {
		return fmt.Errorf("missing type name of object's field to get")
	}

	update := expression.UpdateBuilder{}
	update = update.Set(expression.Name(param.FieldName), expression.Value(param.Value))
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		return fmt.Errorf("error occurred when building dynamodb update expression %w", err)
	}

	_, err = DBClient.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String(param.TypeName),
		Key: map[string]*dynamodb.AttributeValue{
			"Id": {
				S: aws.String(id),
			},
		},
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})

	return err
}
