package main

import (
	"encoding/json"
	"log"
	"testing"
)

// Test different scenarios for passing
// an arbitrary JSON from the browser
// to the server

func TestUnmarshalJsonStruct(t *testing.T) {
	type User struct {
		User string
	}
	type Obj struct {
		Broadcast User
	}
	str := `{"Broadcast": {"User": "name"}}`
	b := Obj{}
	err := json.Unmarshal([]byte(str), &b)
	if err != nil {
		t.Fatalf(`JSON struct unmarshal error: "%s", should not have error`, err.Error())
	}
}

func TestUnmarshalJsonStructString(t *testing.T) {
	type User struct {
		User string
	}
	type Obj struct {
		Broadcast User
	}
	str := `{"Broadcast": "{\"User\": \"name\"}"}`
	b := Obj{}
	err := json.Unmarshal([]byte(str), &b)
	log.Println(err.Error())
	if err == nil {
		t.Fatalf(`JSON struct unmarshal gives no error, there should be an error`)
	}
}

func TestUnmarshalJsonString(t *testing.T) {
	type Obj struct {
		Broadcast string
	}
	str := `{"Broadcast": "{\"User\": \"name\"}"}`
	b := Obj{}
	err := json.Unmarshal([]byte(str), &b)
	if err != nil {
		t.Fatalf(`JSON string unmarshal error: "%s", should not have error`, err.Error())
	}

	type User struct {
		User string
	}
	s := User{}
	err = json.Unmarshal([]byte(b.Broadcast), &s)
	if err != nil {
		t.Fatalf(`JSON sub-struct unmarshal error: "%s", should not have error`, err.Error())
	}
}

func TestUnmarshalArbitraryJson(t *testing.T) {
	str := `{"Broadcast": {"User": "name"}}`
	b := ClientMessage{}
	err := json.Unmarshal([]byte(str), &b)
	if err != nil {
		t.Fatalf(`JSON arbitrary struct unmarshal error: "%s", should not have error`, err.Error())
	}

	if b.Broadcast["User"] != "name" {
		t.Fatalf(`JSON arbitrary struct unmarshal: got "%s", want "name"`, b.Broadcast["User"])
	}
}

func TestMarshalArbitraryJson(t *testing.T) {
	s := ClientMessage{Broadcast: map[string]interface{}{
		"User": "name",
	}}
	s.Broadcast["User"] = "name"
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatalf(`JSON arbitrary json marshal error: "%s", should not have error`, err.Error())
	}

	want := `{"Broadcast":{"User":"name"}}`
	if string(b) != want {
		t.Fatalf(`JSON arbitrary struct unmarshal: got "%s", want "%s"`, string(b), want)
	}
}
