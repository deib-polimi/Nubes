package lib

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type ReferenceNavigationList[T Nobject] struct {
	ownerId       string
	ownerTypeName string

	isManyToMany             bool
	usesIndex                bool
	queryByIndexParam        QueryByIndexParam
	queryByPartitionKeyParam QueryByPartitionKeyParam
}

func NewReferenceNavigationList[T Nobject](ownerId, ownerTypeName, referringFieldName string, isManyToMany bool) *ReferenceNavigationList[T] {
	r := new(ReferenceNavigationList[T])
	r.ownerId = ownerId
	r.ownerTypeName = ownerTypeName
	r.isManyToMany = isManyToMany

	if isManyToMany {
		r.setupManyToManyRelationship()
	} else {
		r.setupOneToManyRelationship(referringFieldName)
	}

	if r.usesIndex {
		r.queryByIndexParam.KeyAttributeValue = r.ownerId
	}

	return r
}

func (r ReferenceNavigationList[T]) GetIds() ([]string, error) {

	if r.usesIndex {
		out, err := GetByIndex(r.queryByIndexParam)
		return out, err
	}

	if r.isManyToMany && !r.usesIndex {
		out, err := GetSortKeysByPartitionKey(r.queryByPartitionKeyParam)
		return out, err
	}

	return nil, fmt.Errorf("invalid initialization of ReferenceNavigationList")
}

func (r ReferenceNavigationList[T]) Get() ([]T, error) {
	var ids []string
	var err error
	if r.usesIndex {
		ids, err = GetByIndex(r.queryByIndexParam)
		if err != nil {
			return nil, err
		}
	} else if r.isManyToMany && !r.usesIndex {
		ids, err = GetSortKeysByPartitionKey(r.queryByPartitionKeyParam)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("invalid initialization of ReferenceNavigationList")
	}

	result := make([]T, len(ids))
	for i, id := range ids {
		initialisedObj, err := Load[T](id)
		if err != nil {
			return nil, err
		}
		result[i] = *initialisedObj
	}

	return result, err
}

func (r ReferenceNavigationList[T]) Add(id string) error {

	if r.isManyToMany {

		typeName := (*new(T)).GetTypeName()

		exists, err := IsInstanceAlreadyCreated(IsInstanceAlreadyCreatedParam{Id: id, TypeName: typeName})
		if err != nil {
			return fmt.Errorf("error occured while checking if typename %s with id %s exists. Error %w", typeName, id, err)
		}
		if !exists {
			return fmt.Errorf("only existing instances can be added to many to many relationships. Typename %s with id %s not found", typeName, id)
		}

		if r.usesIndex {
			return insertToManyToManyTable(r.ownerTypeName, typeName, r.ownerId, id)
		}
		return insertToManyToManyTable(typeName, r.ownerTypeName, id, r.ownerId)
	}

	return fmt.Errorf("can not add elements to ReferenceNavigationList of OneToMany relationship")
}

func (r *ReferenceNavigationList[T]) setupOneToManyRelationship(referringFieldName string) {
	otherTypeName := (*(new(T))).GetTypeName()
	r.queryByIndexParam.KeyAttributeName = referringFieldName
	r.queryByIndexParam.OutputAttributeName = "Id"
	r.usesIndex = true
	r.queryByIndexParam.TableName = otherTypeName
	r.queryByIndexParam.IndexName = otherTypeName + referringFieldName
}

func (r *ReferenceNavigationList[T]) setupManyToManyRelationship() {
	otherTypeName := (*(new(T))).GetTypeName()

	for index := 0; ; index++ {

		if index >= len(r.ownerTypeName) {
			r.queryByPartitionKeyParam.TableName = r.ownerTypeName + otherTypeName
			r.usesIndex = false
			break
		}
		if index >= len(otherTypeName) {
			r.queryByIndexParam.TableName = otherTypeName + r.ownerTypeName
			r.queryByIndexParam.IndexName = r.queryByIndexParam.TableName + "Reversed"
			r.usesIndex = true
			break
		}

		if r.ownerTypeName[index] < otherTypeName[index] {
			r.queryByPartitionKeyParam.TableName = r.ownerTypeName + otherTypeName
			r.usesIndex = false
			break
		} else if r.ownerTypeName[index] > otherTypeName[index] {
			r.queryByIndexParam.TableName = otherTypeName + r.ownerTypeName
			r.queryByIndexParam.IndexName = r.queryByIndexParam.TableName + "Reversed"
			r.usesIndex = true
			break
		}
	}

	if r.usesIndex {
		r.queryByIndexParam.KeyAttributeName = r.ownerTypeName
		r.queryByIndexParam.OutputAttributeName = otherTypeName
	} else {
		r.queryByPartitionKeyParam.PartitionAttributeName = r.ownerTypeName
		r.queryByPartitionKeyParam.PatritionAttributeValue = r.ownerId
		r.queryByPartitionKeyParam.OutputAttributeName = otherTypeName
	}
}

func insertToManyToManyTable(partitiokKeyName, sortKeyName, partitonKey, sortKey string) error {
	input := &dynamodb.PutItemInput{
		TableName: aws.String(partitiokKeyName + sortKeyName),
		Item: map[string]*dynamodb.AttributeValue{
			partitiokKeyName: {
				S: aws.String(partitonKey),
			},
			sortKeyName: {
				S: aws.String(sortKey),
			},
		},
	}

	_, err := DBClient.PutItem(input)
	return err
}
