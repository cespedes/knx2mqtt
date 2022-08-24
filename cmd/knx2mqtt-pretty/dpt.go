package main

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/vapourismo/knx-go/knx/dpt"
)

// GetDPT returns the text representation of the internal value stored in a dpt.DatapointValue
// It works only if its underlying type is a bool, integer or float.
func GetDPTAsString(v dpt.DatapointValue) string {
	Val := reflect.ValueOf(v)
	if Val.Kind() != reflect.Ptr {
		return "" // Error: input value is not a pointer
	}
	Val = Val.Elem()
	switch Val.Kind() {
	case reflect.Bool:
		return strconv.FormatBool(Val.Elem().Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprint(Val.Elem().Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprint(Val.Elem().Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%f", Val.Elem().Float())
	default:
		return fmt.Sprint(v.Pack())
	}
}

// SetDPT sets the internal value of d to value.  Its kind must be integer, float or bool.
func SetDPTFromString(d dpt.DatapointValue, value string) error {
	Val := reflect.ValueOf(d)
	if Val.Kind() != reflect.Ptr {
		return fmt.Errorf("SetDPT: input variable is not a pointer")
	}
	if !Val.Elem().CanSet() {
		return fmt.Errorf("SetDPT: cannot set element value")
	}
	switch Val.Elem().Kind() {
	case reflect.Bool:
		var b bool
		switch strings.ToLower(value) {
		case "1", "t", "true", "on", "enable", "down", "close", "start", "heat":
			b = true
		case "0", "f", "false", "off", "disable", "up", "open", "stop", "cool":
			b = false
		default:
			return fmt.Errorf("SetDPT: %s is not a bool value", value)
		}
		Val.Elem().SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		Val.Elem().SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			switch strings.ToLower(value) { // for DPT 20.105 (HVACContrMode):
			case "heat":
				u = 1
			case "cool":
				u = 3
			case "fan", "fan only", "fan-only":
				u = 9
			default:
				return err
			}
		}
		Val.Elem().SetUint(u)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		Val.Elem().SetFloat(f)
	default:
		return fmt.Errorf("SetDPT: cannot set element (underlying type %v)", Val.Elem().Kind())
	}

	// Normalize:
	d.Unpack(d.Pack())
	return nil
}
