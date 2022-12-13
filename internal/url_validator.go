package internal

import (
	"errors"
	"net/url"
	"reflect"
)

func UrlValidator(val interface{}) error {
	// the reflect value of the result
	value := reflect.ValueOf(val)

	// if the value passed in is the zero value of the appropriate type
	if isZero(value) && value.Kind() != reflect.Bool {
		return errors.New("Value is required")
	}

	s := val.(string)

	u, err := url.Parse(s)
	if err != nil {
		return err
	}

	if u.Host == "" || u.Scheme == "" {
		return errors.New("Full URL is required, e.g. `https://prod-uk-a.online.tableau.com`")
	}

	return nil
}

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Slice, reflect.Map:
		return v.Len() == 0
	}

	// compare the types directly with more general coverage
	return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}
