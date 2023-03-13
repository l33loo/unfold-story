package main

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

var testConn net.Conn

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-type", "text/html")
		f, err := os.Open("./public/index.html")
		if err != nil {
			fmt.Println(err.Error())
		}

		io.Copy(w, f)
	})

	http.HandleFunc("/public/styles.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-type", "text/css")
		f, err := os.Open("./public/styles.css")

		if err != nil {
			fmt.Println(err.Error())
		}

		io.Copy(w, f)
	})

	http.HandleFunc("/public/scripts.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-type", "text/javascript")
		f, err := os.Open("./public/scripts.js")

		if err != nil {
			fmt.Println(err.Error())
		}

		io.Copy(w, f)
	})

	http.HandleFunc("/public/game/styles.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-type", "text/css")
		f, err := os.Open("./public/game/styles.css")

		if err != nil {
			fmt.Println(err.Error())
		}

		io.Copy(w, f)
	})

	http.HandleFunc("/public/game/scripts.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-type", "text/javascript")
		f, err := os.Open("./public/game/scripts.js")

		if err != nil {
			fmt.Println(err.Error())
		}

		io.Copy(w, f)
	})

	http.HandleFunc("/game", func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("./public/game/index.html")
		if err != nil {
			fmt.Println(err.Error())
		}

		io.Copy(w, f)
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		err := WsHandler(w, r)
		// TODO: Change error handling because may no longer use HTTP
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func WsHandler(w http.ResponseWriter, req *http.Request) error {
	ws, err := handshake(w, req)
	if err != nil {
		return err
	}

	// Frame
	ws.Send("hello, this is Lila's ws server <3")

	// not frame
	ws.write([]byte("hello websocket <3"))

	return nil
}

func handshake(w http.ResponseWriter, r *http.Request) (*Ws, error) {
	httpStatus, err := validateWsRequest(r)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), httpStatus)
	}
	// As per RFC6455:
	// For this header field [Sec-WebSocket-Key], the server has to take the value (as present
	// in the header field, e.g., the base64-encoded [RFC4648] version minus
	// any leading and trailing whitespace) and concatenate this with the
	// Globally Unique Identifier (GUID, [RFC4122]) "258EAFA5-E914-47DA-
	// 95CA-C5AB0DC85B11" in string form, which is unlikely to be used by
	// network endpoints that do not understand the WebSocket Protocol.  A
	// SHA-1 hash (160 bits) [FIPS.180-3], base64-encoded (see Section 4 of
	// [RFC4648]), of this concatenation is then returned in the server's
	// handshake.
	wsKeyConcat := strings.TrimSpace(r.Header.Get("Sec-WebSocket-Key")) + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	wsBytes := []byte(wsKeyConcat)
	hasher := sha1.New()
	hasher.Write(wsBytes)
	sha := base64.StdEncoding.EncodeToString(hasher.Sum(nil))

	w.Header().Add("Upgrade", "websocket")
	w.Header().Add("Connection", "Upgrade")
	w.Header().Add("Sec-WebSocket-Accept", sha)
	w.Header().Add("Sec-WebSocket-Protocol", "chat")
	w.WriteHeader(101)

	hj, ok := w.(http.Hijacker)
	if !ok {
		return nil, errors.New("webserver doesn't support http hijacking")
	}
	conn, bufwr, err := hj.Hijack()
	if err != nil {
		return nil, err
	}

	log.Println("HERE! <3")

	ws := &Ws{conn, bufwr, r}
	testConn = ws.conn
	// defer ws.conn.Close()

	return ws, nil
}

func validateWsRequest(r *http.Request) (int, error) {
	// An HTTP/1.1 or higher GET request, including a "Request-URI"
	// [RFC2616] that should be interpreted as a /resource name/
	// defined in Section 3 (or an absolute HTTP/HTTPS URI containing
	// the /resource name/).
	method := r.Method
	if method != "GET" {
		return http.StatusForbidden, fmt.Errorf("forbidden method %s\n, must be 'GET'", method)
	}

	if !r.ProtoAtLeast(1, 1) {
		return http.StatusForbidden, fmt.Errorf("forbidden HTTP protocol version %s\n, must be 1.1 or higher", r.Proto)
	}

	// As per RFC6455:
	// The client includes the hostname in the |Host| header field of its
	// handshake as per [RFC2616], so that both the client and the server
	// can verify that they agree on which host is in use.
	host := r.Host
	fmt.Printf("host: %s\n", host)
	if host != "localhost:8080" {
		return http.StatusForbidden, fmt.Errorf("forbidden host: %s\n", host)
	}

	// An |Upgrade| header field containing the value "websocket",
	// treated as an ASCII case-insensitive value.
	upgrade := r.Header.Get("Upgrade")
	if upgrade != "websocket" {
		return http.StatusBadRequest, fmt.Errorf("invalid Upgrade header %s, must be 'websocket'", upgrade)
	}

	wsKey := r.Header.Get("Sec-WebSocket-Key")

	// As per RFC6455:
	// The request MUST include a header field with the name
	// |Sec-WebSocket-Key|.  The value of this header field MUST be a
	// nonce consisting of a randomly selected 16-byte value that has
	// been base64-encoded (see Section 4 of [RFC4648]).  The nonce
	// MUST be selected randomly for each connection.
	wsKeyBytes, err := base64.StdEncoding.DecodeString(wsKey)
	if err != nil {
		return http.StatusInternalServerError, errors.New("error decoding Sec-WebSocket-Key header")
	}
	if len(wsKeyBytes) != 16 {
		return http.StatusBadRequest, fmt.Errorf("invalid Sec-WebSocket-Key header length of %d, must be 16-bytes long", len(wsKeyBytes))
	}

	return 0, nil
}

func (ws *Ws) write(data []byte) error {
	_, err := ws.bufrw.Write(data)
	if err != nil {
		return err
	}
	return ws.bufrw.Flush()
}

func (ws *Ws) Send(msg string) {
	pay := []byte(msg)
	payLen := len(pay)

	var fin uint8 = 1
	var rsv1 uint8 = 0
	var rsv2 uint8 = 0
	var rsv3 uint8 = 0
	var upcode uint8 = 1
	var masked uint8 = 0

	frame := new(bytes.Buffer)
	byte1 := (fin << 7) | (rsv1 << 6) | (rsv2 << 5) | (rsv3 << 4) | upcode
	err := frame.WriteByte(byte1)
	if err != nil {
		log.Fatal(err)
	}

	switch {
	case payLen < 126:
		// 7 bits to denote payload length (in bytes)
		byte2 := (masked << 7) | uint8(payLen)
		err = frame.WriteByte(byte2)
		if err != nil {
			log.Fatal(err)
		}
	case payLen < (1 << 16):
		// 7+16 bits
		byte2 := (masked << 7) | (uint8(126))
		err = frame.WriteByte(byte2)
		if err != nil {
			log.Fatal(err)
		}
		bytes34 := uint16(payLen)
		err = binary.Write(frame, binary.BigEndian, bytes34)
		if err != nil {
			log.Fatal(err)
		}
	default:
		// 7+64 bits
		byte2 := (masked << 7) | (uint8(127))
		err = frame.WriteByte(byte2)
		if err != nil {
			log.Fatal(err)
		}
		nextBytes := uint64(payLen)
		err = binary.Write(frame, binary.BigEndian, nextBytes)
		if err != nil {
			log.Fatal(err)
		}
	}
	_, err = frame.Write(pay)
	if err != nil {
		log.Fatal(err)
	}
	ws.write(frame.Bytes())
}

type Ws struct {
	conn    net.Conn
	bufrw   *bufio.ReadWriter
	request *http.Request
}
