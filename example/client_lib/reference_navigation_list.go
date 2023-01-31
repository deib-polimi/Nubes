package client_lib

import (
	"encoding/json"
	"fmt"

	"github.com/Astenna/Nubes/lib"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
)

type setId interface {
	setId(id string)
	init()
}

type referenceNavigationList[T lib.Nobject, Stub any] struct {
	ownerId       string
	ownerTypeName string

	isManyToMany             bool
	usesIndex                bool
	queryByIndexParam        lib.QueryByIndexParam
	queryByPartitionKeyParam lib.QueryByPartitionKeyParam
}

func newReferenceNavigationList[T lib.Nobject, Stub any](ownerId, ownerTypeName, referringFieldName string, isManyToMany bool) *referenceNavigationList[T, Stub] {
	r := new(referenceNavigationList[T, Stub])
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

func (r referenceNavigationList[T, Stub]) GetIds() ([]string, error) {

	if r.usesIndex {
		out, err := r.getByIndex()
		return out, err
	}

	if r.isManyToMany && !r.usesIndex {
		out, err := r.getSortKeysByPartitionKey()
		return out, err
	}

	return nil, fmt.Errorf("invalid initialization of ReferenceNavigationList")
}

func (r referenceNavigationList[T, Stub]) Get() ([]T, error) {
	var ids []string
	var err error
	if r.usesIndex {
		ids, err = r.getByIndex()
		if err != nil {
			return nil, err
		}
	} else if r.isManyToMany && !r.usesIndex {
		ids, err = r.getSortKeysByPartitionKey()
		if err != nil {
			return nil, err
		}
	} else {
		fmt.Println("")
		return nil, fmt.Errorf("invalid initialization of ReferenceNavigationList")
	}

	result := make([]T, len(ids))
	typeName := (*new(T)).GetTypeName()
	// TODO: do it in one call
	for i, id := range ids {
		params := lib.HandlerParameters{
			Id:       id,
			TypeName: typeName,
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

		newInstance := new(T)
		casted := any(newInstance)
		setIdInterf, _ := casted.(setId)
		setIdInterf.setId(id)
		setIdInterf.init()
		result[i] = *newInstance
	}

	return result, err
}

func (r referenceNavigationList[T, Stub]) GetStubs() ([]Stub, error) {
	var ids []string
	var err error
	if r.usesIndex {
		ids, err = r.getByIndex()
		if err != nil {
			return nil, err
		}
	} else if r.isManyToMany && !r.usesIndex {
		ids, err = r.getSortKeysByPartitionKey()
		if err != nil {
			return nil, err
		}
	} else {
		fmt.Println("")
		return nil, fmt.Errorf("invalid initialization of ReferenceNavigationList")
	}

	if len(ids) < 1 {
		return *(new([]Stub)), nil
	}

	params := lib.GetBatchParam{
		Ids:      ids,
		TypeName: (*new(T)).GetTypeName(),
	}
	jsonParam, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	out, _err := LambdaClient.Invoke(&lambda.InvokeInput{FunctionName: aws.String("GetBatch"), Payload: jsonParam})
	if _err != nil {
		return nil, _err
	}
	if out.FunctionError != nil {
		return nil, fmt.Errorf("lambda function designed to verify if instance exists failed. Error: %s", string(out.Payload))
	}

	stubs := make([]Stub, len(ids))
	err = json.Unmarshal(out.Payload, &stubs)
	if err != nil {
		return nil, err
	}

	return stubs, err
}

func (r referenceNavigationList[T, Stub]) AddToManyToMany(newId string) error {
	if newId == "" {
		return fmt.Errorf("missing id")
	}

	if r.isManyToMany {
		typeName := (*new(T)).GetTypeName()
		params := lib.AddToManyToManyParam{
			TypeName:      typeName,
			NewId:         newId,
			OwnerTypeName: r.ownerTypeName,
			OwnerId:       r.ownerId,
			UsesIndex:     r.usesIndex,
		}

		jsonParam, err := json.Marshal(params)
		if err != nil {
			return err
		}

		out, _err := LambdaClient.Invoke(&lambda.InvokeInput{FunctionName: aws.String("ReferenceAddToManyToMany"), Payload: jsonParam})
		if _err != nil {
			return _err
		}
		if out.FunctionError != nil {
			return fmt.Errorf(string(out.Payload[:]))
		}

		return nil
	}

	return fmt.Errorf("can not add elements to ReferenceNavigationList of OneToMany relationship")
}

func (r referenceNavigationList[T, Stub]) getByIndex() ([]string, error) {

	jsonParam, err := json.Marshal(r.queryByIndexParam)
	if err != nil {
		return nil, err
	}

	out, _err := LambdaClient.Invoke(&lambda.InvokeInput{FunctionName: aws.String("ReferenceGetByIndex"), Payload: jsonParam})
	if _err != nil {
		return nil, _err
	}
	if out.FunctionError != nil {
		return nil, fmt.Errorf(string(out.Payload[:]))
	}

	var result []string
	err = json.Unmarshal(out.Payload, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r referenceNavigationList[T, Stub]) getSortKeysByPartitionKey() ([]string, error) {

	jsonParam, err := json.Marshal(r.queryByPartitionKeyParam)
	if err != nil {
		return nil, err
	}

	out, _err := LambdaClient.Invoke(&lambda.InvokeInput{FunctionName: aws.String("ReferenceGetSortKeysByPartitionKey"), Payload: jsonParam})
	if _err != nil {
		return nil, _err
	}
	if out.FunctionError != nil {
		return nil, fmt.Errorf(string(out.Payload[:]))
	}

	var result []string
	err = json.Unmarshal(out.Payload, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *referenceNavigationList[T, Stub]) setupOneToManyRelationship(referringFieldName string) {
	otherTypeName := (*(new(T))).GetTypeName()
	r.queryByIndexParam.KeyAttributeName = referringFieldName
	r.queryByIndexParam.OutputAttributeName = "Id"
	r.usesIndex = true
	r.queryByIndexParam.TableName = otherTypeName
	r.queryByIndexParam.IndexName = otherTypeName + referringFieldName
}

func (r *referenceNavigationList[T, Stub]) setupManyToManyRelationship() {
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