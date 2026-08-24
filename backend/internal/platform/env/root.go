package env

import (
	"errors"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	seconds = "s"
	minute  = "m"
	hour    = "h"
	day     = "d"
)

type FieldParams struct {
	Key             string
	KeyName         string
	Value           string
	DefaultValue    string
	HasDefaultValue bool
	Required        bool
}

func getKeyTags(v string) (string, []string) {
	opts := strings.Split(v, ",")

	return opts[0], opts[1:]
}

func getFieldParams(f reflect.StructField) (FieldParams, error) {
	value, ok := f.Tag.Lookup("env")
	if !ok {
		return FieldParams{}, errors.New("tag env no exist")
	}

	defaultValue, hasDefaultValue := f.Tag.Lookup("default")
	name, tags := getKeyTags(value)

	field := FieldParams{
		Key:             f.Name,
		KeyName:         name,
		Value:           os.Getenv(name),
		DefaultValue:    defaultValue,
		HasDefaultValue: hasDefaultValue,
	}

	for _, tag := range tags {
		if tag == "required" {
			field.Required = true
		}
	}

	return field, nil
}

func processString(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}

	return value
}

func convertDuration(value string) (time.Duration, error) {
	switch {
	case strings.Contains(value, seconds):
		strNS, ok := strings.CutSuffix(value, seconds)
		if !ok {
			return time.Duration(0), errors.New("")
		}

		ns, err := strconv.ParseInt(strNS, 10, 64)
		if err != nil {
			return time.Duration(0), err
		}

		return time.Second * time.Duration(ns), nil
	case strings.Contains(value, minute):
		strNM, ok := strings.CutSuffix(value, minute)
		if !ok {
			return time.Duration(0), errors.New("")
		}

		nm, err := strconv.ParseInt(strNM, 10, 64)
		if err != nil {
			return time.Duration(0), err
		}

		return time.Minute * time.Duration(nm), nil
	case strings.Contains(value, hour):
		strNH, ok := strings.CutSuffix(value, hour)
		if !ok {
			return time.Duration(0), errors.New("")
		}

		nh, err := strconv.ParseInt(strNH, 10, 64)
		if err != nil {
			return time.Duration(0), err
		}

		return time.Hour * time.Duration(nh), nil
	case strings.Contains(value, day):
		strND, ok := strings.CutSuffix(value, day)
		if !ok {
			return time.Duration(0), errors.New("")
		}

		nd, err := strconv.ParseInt(strND, 10, 64)
		if err != nil {
			return time.Duration(0), err
		}

		return time.Hour * 24 * time.Duration(nd), nil
	}

	return time.Duration(0), nil
}

func processSliceString(value, defaultValue string) []string {
	if value == "" {
		return strings.Split(defaultValue, ",")
	}

	return strings.Split(value, ",")
}

func processDuration(value, defaultValue string) (time.Duration, error) {
	if value == "" {
		return convertDuration(defaultValue)
	}

	return convertDuration(value)
}

func processInt(value, defaultValue string, bitSize int) (int64, error) {
	if value == "" {
		dv, err := strconv.ParseInt(defaultValue, 10, bitSize)
		if err != nil {
			return 0, err
		}

		return dv, nil
	}

	int64Value, err := strconv.ParseInt(value, 10, bitSize)
	if err != nil {
		dv, err := strconv.ParseInt(defaultValue, 10, bitSize)
		if err != nil {
			return 0, err
		}

		return dv, nil
	}

	return int64Value, nil
}

func processBool(value, defaultValue string) (bool, error) {
	if value == "" {
		dv, err := strconv.ParseBool(defaultValue)
		if err != nil {
			return false, err
		}

		return dv, nil
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		dv, err := strconv.ParseBool(defaultValue)
		if err != nil {
			return false, err
		}

		return dv, nil
	}

	return boolValue, nil
}

func setField(f reflect.Value, raw, defaultValue string) error {
	switch f.Type() {
	case reflect.TypeFor[string]():
		f.SetString(processString(raw, defaultValue))
	case reflect.TypeFor[int8](), reflect.TypeFor[int16](), reflect.TypeFor[int32](), reflect.TypeFor[int64](), reflect.TypeFor[int]():
		n, err := processInt(raw, defaultValue, f.Type().Bits())
		if err != nil {
			return err
		}

		f.SetInt(n)
	case reflect.TypeFor[bool]():
		b, err := processBool(raw, defaultValue)
		if err != nil {
			return err
		}

		f.SetBool(b)
	case reflect.TypeFor[time.Duration]():
		d, err := processDuration(raw, defaultValue)
		if err != nil {
			return err
		}

		f.SetInt(int64(d))
	case reflect.TypeFor[[]string]():
		f.Set(reflect.ValueOf(processSliceString(raw, defaultValue)))
	}

	return nil
}

func processField(fieldValue reflect.Value, fieldType reflect.Type) error {
	for i := range fieldValue.NumField() {
		fieldParams, err := getFieldParams(fieldType.Field(i))
		if err != nil {
			return err
		}

		if fieldParams.Required && fieldParams.HasDefaultValue {
			return errors.New("field " + fieldParams.KeyName + " is required and mustn't have a default value")
		}

		if fieldParams.Required && fieldParams.Value == "" {
			return errors.New("field " + fieldParams.KeyName + " is required must have a value")
		}

		if err := setField(fieldValue.Field(i), fieldParams.Value, fieldParams.DefaultValue); err != nil {
			return err
		}
	}

	return nil
}

func Parse(v any) error {
	ptrRef := reflect.ValueOf(v)
	if ptrRef.Kind() != reflect.Pointer {
		return errors.New("no pointer")
	}

	s := ptrRef.Elem()
	if s.Kind() != reflect.Struct {
		return errors.New("no struct")
	}

	if err := processField(s, s.Type()); err != nil {
		return err
	}

	return nil
}
