package validator

import (
	"fmt"
	"reflect"
)

var isValidType = reflect.TypeOf((*IsValid)(nil)).Elem()

type IsValid interface {
	IsValid() error
}

var Default Validator = Validator{}

type Validator struct{}

func (Validator) IsValid(iv interface{}) (err error) {
	if err = isInterfaceValid(iv); err != nil {
		return
	}
	traverseFields(iv, func(tmp interface{}) {
		if isValidImplementor(tmp) {
			returned := reflect.ValueOf(tmp).MethodByName("IsValid").Call([]reflect.Value{})
			if returned[0].Interface() != nil {
				err = returned[0].Interface().(error)
			}
			return
		}
	})
	return
}

func isValidImplementor(iv interface{}) bool {
	return reflect.ValueOf(iv).Type().Implements(isValidType)
}

func isInterfaceValid(iv interface{}) (err error) {
	if iv == nil {
		err = fmt.Errorf("invalid input: input is nil")
		return
	}
	inVal := reflect.ValueOf(iv)
	if inVal.Type().Kind() != reflect.Ptr {
		err = fmt.Errorf("invalid input: input type %s, expecting ptr", inVal.Type().Kind())
		return
	}
	return
}

func traverseFields(iv interface{}, f func(i interface{})) {
	f(iv)
	intype := reflect.ValueOf(iv).Type().Elem()
	for i := 0; i < intype.NumField(); i++ {
		switch intype.Field(i).Type.Kind() {
		case reflect.Struct:
			traverseFields(reflect.ValueOf(iv).Elem().Field(i).Addr().Interface(), f)
		default:
		}
	}
}
