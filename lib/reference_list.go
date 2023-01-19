package lib

import "fmt"

type ReferenceList[T Nobject] []string

func NewReferenceList[T Nobject](ids []string) ReferenceList[T] {
	result := ReferenceList[T](ids)
	return result
}

func (r ReferenceList[T]) GetIds() []string {
	return []string(r)
}

func (r ReferenceList[T]) Get() ([]T, error) {
	result := make([]T, len(r))

	instance := *new(T)
	casted := any(instance)
	if customId, ok := casted.(CustomId); ok {
		_ = customId.GetId()
	}

	for index, id := range r {
		instance, err := Get[T](id)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve object with id: %d. Error: %w", index, err)
		}
		result[index] = *instance
	}

	return result, nil
}

func (r ReferenceList[T]) GetAt(index int) (*T, error) {
	if len(r)-1 < index || index < 0 {
		return nil, fmt.Errorf("provided index: %d is out of bounds of the list", index)
	}
	instance, err := Get[T](r[index])
	if err != nil {
		return nil, fmt.Errorf("could not retrieve object with id: %d. Error: %w", index, err)
	}

	return instance, nil
}