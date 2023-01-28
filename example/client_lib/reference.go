package client_lib

import (
	"github.com/Astenna/Nubes/lib"
)

// REFERENCE

type Reference[T lib.Nobject] string

func NewReference[T lib.Nobject](id string) *Reference[T] {
	result := Reference[T](id)
	return &result
}

func (r Reference[T]) Id() string {
	return string(r)
}

// REFERENCE LIST

type ReferenceList[T lib.Nobject] []string

func NewReferenceList[T lib.Nobject](ids []string) *ReferenceList[T] {
	result := ReferenceList[T](ids)
	return &result
}

func (r ReferenceList[T]) Ids() []string {
	return []string(r)
}
