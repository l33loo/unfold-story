package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"log"
	"net/http"
	"strings"
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

// Payload length < 126
func TestRecvShortMessage(t *testing.T) {
	want := strings.Repeat("a", 125)
	msg, err := mockWebsocketMessage(want)
	if err != nil {
		log.Println("error creating mock websocket message: ", err.Error())
	}
	r := bytes.NewReader(msg)
	rr := bufio.NewReader(r)
	rw := bufio.NewReadWriter(rr, nil)
	req, _ := http.NewRequest("GET", "localhost:8080", rw.Reader)
	ws := Ws{
		conn:    nil,
		bufrw:   rw,
		request: req,
	}
	unmaskedMsg, _, err := ws.Recv()

	if want != unmaskedMsg {
		t.Fatalf(`TestRecvShortMessage wrong msg: got "%s", want "%s"`, unmaskedMsg, want)
	}

	if err != nil {
		t.Fatalf(`TestRecvShortMessage error: "%s", should not have error`, err.Error())
	}
}

// Payload length == 126
func TestRecvMediumMessageLowerLimit(t *testing.T) {
	want := strings.Repeat("a", 126)
	msg, err := mockWebsocketMessage(want)
	if err != nil {
		log.Println("error creating mock websocket message: ", err.Error())
	}
	r := bytes.NewReader(msg)
	rr := bufio.NewReader(r)
	rw := bufio.NewReadWriter(rr, nil)
	req, _ := http.NewRequest("GET", "localhost:8080", rw.Reader)
	ws := Ws{
		conn:    nil,
		bufrw:   rw,
		request: req,
	}
	unmaskedMsg, _, err := ws.Recv()

	if want != unmaskedMsg {
		t.Fatalf(`TestRecvShortMessage wrong msg: got "%s", want "%s"`, unmaskedMsg, want)
	}

	if err != nil {
		t.Fatalf(`TestRecvShortMessage error: "%s", should not have error`, err.Error())
	}
}

// Payload length == 126
func TestRecvMediumMessageUpperLimit(t *testing.T) {
	want := strings.Repeat("a", 65535)
	msg, err := mockWebsocketMessage(want)
	if err != nil {
		log.Println("error creating mock websocket message: ", err.Error())
	}
	r := bytes.NewReader(msg)
	rr := bufio.NewReader(r)
	rw := bufio.NewReadWriter(rr, nil)
	req, _ := http.NewRequest("GET", "localhost:8080", rw.Reader)
	ws := Ws{
		conn:    nil,
		bufrw:   rw,
		request: req,
	}
	unmaskedMsg, _, err := ws.Recv()

	if want != unmaskedMsg {
		t.Fatalf(`TestRecvShortMessage wrong msg: got "%s", want "%s"`, unmaskedMsg, want)
	}

	if err != nil {
		t.Fatalf(`TestRecvShortMessage error: "%s", should not have error`, err.Error())
	}
}

// Payload length == 127 bytes
func TestRecvLongMessage(t *testing.T) {
	want := strings.Repeat("a", 65536)
	msg, err := mockWebsocketMessage(want)
	if err != nil {
		log.Println("error creating mock websocket message: ", err.Error())
	}
	r := bytes.NewReader(msg)
	rr := bufio.NewReader(r)
	rw := bufio.NewReadWriter(rr, nil)
	req, _ := http.NewRequest("GET", "localhost:8080", rw.Reader)
	ws := Ws{
		conn:    nil,
		bufrw:   rw,
		request: req,
	}
	unmaskedMsg, _, err := ws.Recv()

	if want != unmaskedMsg {
		t.Fatalf(`TestRecvShortMessage wrong msg: got "%s", want "%s"`, unmaskedMsg, want)
	}

	if err != nil {
		t.Fatalf(`TestRecvShortMessage error: "%s", should not have error`, err.Error())
	}
}

var maskKey = []byte{0xbb, 0x76, 0x44, 0xbf}

func mockWebsocketMessage(msg string) ([]byte, error) {
	maskedPay := maskPayload(msg)
	payLen := len(msg)

	var fin uint8 = 1
	var rsv1 uint8 = 0
	var rsv2 uint8 = 0
	var rsv3 uint8 = 0
	var opcode uint8 = 1
	var masked uint8 = 1

	frame := new(bytes.Buffer)
	byte1 := (fin << 7) | (rsv1 << 6) | (rsv2 << 5) | (rsv3 << 4) | opcode
	err := frame.WriteByte(byte1)
	if err != nil {
		return []byte{}, err
	}

	switch {
	case payLen < 126:
		// 7 bits to denote payload length (in bytes)
		byte2 := (masked << 7) | uint8(payLen)
		err = frame.WriteByte(byte2)
		if err != nil {
			return []byte{}, err
		}
	case payLen < (1 << 16):
		// 7+16 bits
		byte2 := (masked << 7) | (uint8(126))
		err = frame.WriteByte(byte2)
		if err != nil {
			return []byte{}, err
		}
		bytes34 := uint16(payLen)
		err = binary.Write(frame, binary.BigEndian, bytes34)
		if err != nil {
			return []byte{}, err
		}
	default:
		// 7+64 bits
		byte2 := (masked << 7) | (uint8(127))
		err = frame.WriteByte(byte2)
		if err != nil {
			return []byte{}, err
		}
		nextBytes := uint64(payLen)
		err = binary.Write(frame, binary.BigEndian, nextBytes)
		if err != nil {
			return []byte{}, err
		}
	}

	for _, b := range maskKey {
		err = binary.Write(frame, binary.BigEndian, b)
		if err != nil {
			return []byte{}, err
		}
	}
	_, err = frame.Write(maskedPay)
	if err != nil {
		return []byte{}, err
	}

	return frame.Bytes(), nil
}

func maskPayload(payload string) []byte {
	payLen := len(payload)

	masked := make([]byte, payLen)
	keyIdx := 0
	for i, e := range []byte(payload) {
		masked[i] = e ^ maskKey[keyIdx]
		if keyIdx == len(maskKey)-1 {
			keyIdx = 0
			continue
		}
		keyIdx++
	}

	return masked
}
