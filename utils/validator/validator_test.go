package validator_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/darren-west/app/utils/validator"
)

type Foo struct {
	Bar       Bar
	Item      string
	Something int
}

type Bar struct {
	Item    string
	Another int
}

func (f Foo) IsValid() (err error) {
	if f.Bar == (Bar{}) {
		err = fmt.Errorf("missing field: bar")
		return
	}
	return
}

func (Bar) IsValid() (err error) {
	return
}

func TestValidateMissingField(t *testing.T) {
	err := validator.Validator{}.IsValid(&Foo{Item: "Something"})

	assert.EqualError(t, err, "missing field: bar")
}

func TestValidatorIncorrectType(t *testing.T) {
	err := validator.Validator{}.IsValid(struct{}{})
	assert.EqualError(t, err, "invalid input: input type struct, expecting ptr")
}

func TestValidatorIncorrectTypeNil(t *testing.T) {
	err := validator.Validator{}.IsValid(nil)
	assert.EqualError(t, err, "invalid input: input is nil")
}
