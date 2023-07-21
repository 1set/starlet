package starlet

import (
	"reflect"
	"testing"

	"go.starlark.net/starlark"
)

func TestCastStringDictToAnyMap(t *testing.T) {
	// Create a starlark.StringDict
	m := starlark.StringDict{
		"key1": starlark.String("value1"),
		"key2": starlark.String("value2"),
		"key3": starlark.NewList([]starlark.Value{starlark.String("value3")}),
	}

	// Convert to StringAnyMap
	anyMap := castStringDictToAnyMap(m)

	// Check that the conversion was successful
	if anyMap["key1"].(starlark.String) != "value1" {
		t.Errorf("Expected 'value1', got '%s'", anyMap["key1"].(starlark.String))
	}
	if anyMap["key2"].(starlark.String) != "value2" {
		t.Errorf("Expected 'value2', got '%s'", anyMap["key2"].(starlark.String))
	}
	if !reflect.DeepEqual(anyMap["key3"], m["key3"]) {
		t.Errorf("Expected '%v', got '%v'", m["key3"], anyMap["key3"])
	}
}

func TestCastStringAnyMapToStringDict(t *testing.T) {
	// Create a StringAnyMap
	m := StringAnyMap{
		"key1": starlark.String("value1"),
		"key2": starlark.String("value2"),
		"key3": starlark.NewList([]starlark.Value{starlark.String("value3")}),
	}

	// Convert to starlark.StringDict
	stringDict, err := castStringAnyMapToStringDict(m)

	// Check that the conversion was successful
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if stringDict["key1"].(starlark.String) != "value1" {
		t.Errorf("Expected 'value1', got '%s'", stringDict["key1"].(starlark.String))
	}
	if stringDict["key2"].(starlark.String) != "value2" {
		t.Errorf("Expected 'value2', got '%s'", stringDict["key2"].(starlark.String))
	}
	if !reflect.DeepEqual(stringDict["key3"], m["key3"]) {
		t.Errorf("Expected '%v', got '%v'", m["key3"], stringDict["key3"])
	}

	// Test with a non-starlark.Value in the map
	m["key4"] = "value4"
	_, err = castStringAnyMapToStringDict(m)

	// Check that an error was returned
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestSetInputConversionEnabled(t *testing.T) {
	m := NewDefault()
	if m.enableInConv != true {
		t.Errorf("Expected input conversion to be enabled by default, but it wasn't")
	}

	m.SetInputConversionEnabled(false)
	if m.enableInConv != false {
		t.Errorf("Expected input conversion to be disabled, but it wasn't")
	}

	m.SetInputConversionEnabled(true)
	if m.enableInConv != true {
		t.Errorf("Expected input conversion to be enabled, but it wasn't")
	}
}

func TestSetOutputConversionEnabled(t *testing.T) {
	m := NewDefault()
	if m.enableOutConv != true {
		t.Errorf("Expected output conversion to be enabled by default, but it wasn't")
	}

	m.SetOutputConversionEnabled(false)
	if m.enableOutConv != false {
		t.Errorf("Expected output conversion to be disabled, but it wasn't")
	}

	m.SetOutputConversionEnabled(true)
	if m.enableOutConv != true {
		t.Errorf("Expected output conversion to be enabled, but it wasn't")
	}
}
