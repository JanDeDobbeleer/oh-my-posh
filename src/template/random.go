package template

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
)

func random(list interface{}) (string, error) {
	v := reflect.ValueOf(list)

	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return "", errors.New("input must be a slice or array")
	}

	if v.Len() == 0 {
		return "", errors.New("input slice or array is empty")
	}

	return fmt.Sprintf("%v", v.Index(rand.Intn(v.Len()))), nil
}
