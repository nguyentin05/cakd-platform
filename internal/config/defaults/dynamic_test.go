package defaults

import (
	"reflect"
	"testing"
)

func TestApplyValue_Direct(t *testing.T) {
	type TestStruct struct {
		VersionControl string
	}

	ts := &TestStruct{}
	parentVal := reflect.ValueOf(ts).Elem()
	fieldVal := parentVal.FieldByName("VersionControl")

	applyValue(parentVal, fieldVal, "Providers.VersionControl")

	if ts.VersionControl != "github" {
		t.Errorf("Expected 'github', got %q", ts.VersionControl)
	}
}

func TestApplyValue_Map(t *testing.T) {
	type TestStruct struct {
		Language string
		Version  string
	}

	ts := &TestStruct{
		Language: "java-spring-boot",
	}

	parentVal := reflect.ValueOf(ts).Elem()
	fieldVal := parentVal.FieldByName("Version")

	applyValue(parentVal, fieldVal, "map:LanguageVersion,key:Language")

	if ts.Version != "21" {
		t.Errorf("Expected '21', got %q", ts.Version)
	}
}

func TestDynamic_PointerInitialization(t *testing.T) {
	type Child struct {
		Field string `default:"Providers.VersionControl"`
	}
	type Parent struct {
		ChildPtr *Child
	}

	p := &Parent{}
	dynamic(reflect.ValueOf(p).Elem())

	if p.ChildPtr == nil {
		t.Fatalf("Expected ChildPtr to be initialized")
	}

	if p.ChildPtr.Field != "github" {
		t.Errorf("Expected ChildPtr.Field to be 'github', got %q", p.ChildPtr.Field)
	}
}
