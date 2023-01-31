package types

import (
	"fmt"

	"github.com/Astenna/Nubes/lib"
)

type User struct {
	FirstName     string
	LastName      string
	Email         string `nubes:"id,readonly" dynamodbav:"Id"`
	Password      string `nubes:"readonly"`
	Address       string
	Shops         lib.ReferenceNavigationList[Shop] `nubes:"hasMany-Owners" dynamodbav:"-"`
	Orders        lib.ReferenceList[Order]
	isInitialized bool
}

func DeleteUser(id string) error {
	_libError := lib.Delete[User](id)
	if _libError != nil {
		return _libError
	}
	return nil
}

func (User) GetTypeName() string {
	return "User"
}

func (u User) GetId() string {
	return u.Email
}

func (u *User) SetLastName(lastName string) error {
	u.LastName = lastName
	if u.isInitialized {
		_libError := lib.SetField(lib.SetFieldParam{Id: u.Email, TypeName: "User", FieldName: "LastName", Value: u.LastName})
		if _libError != nil {
			return _libError
		}
	}
	return nil
}

func (u *User) GetLastName() (string, error) {
	if u.isInitialized {
		fieldValue, _libError := lib.GetFieldOfType[string](lib.GetFieldParam{Id: u.Email, TypeName: "User", FieldName: "LastName"})
		if _libError != nil {
			return *new(string), _libError
		}
		u.LastName = fieldValue
	}
	return u.LastName, nil
}

func (u User) GetShops() ([]string, error) {
	if !u.isInitialized {
		return nil, fmt.Errorf(`fields of type ReferenceNavigationList can be used only after instance initialization. 
			Use lib.Load or lib.Export from the Nubes library to create initialized instances`)
	}
	return u.Shops.GetIds()
}

func (u User) VerifyPassword(password string) (bool, error) {
	if u.isInitialized {
		tempReceiverName, _libError := lib.GetObjectState[User](u.Email)
		if _libError != nil {
			return *new(bool), _libError
		}
		u = *tempReceiverName
		u.Init()
	}

	if u.Password == password {
		return true, nil
	}
	if u.isInitialized {
		_libError := lib.Upsert(u, u.Email)
		if _libError != nil {
			return *new(bool), _libError
		}
	}
	return false, nil
}

func (receiver *User) Init() {
	receiver.isInitialized = true
	receiver.Shops = *lib.NewReferenceNavigationList[Shop](receiver.Email, receiver.GetTypeName(), "", true)
}