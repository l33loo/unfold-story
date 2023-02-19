package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

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
		// An HTTP/1.1 or higher GET request, including a "Request-URI"
		// [RFC2616] that should be interpreted as a /resource name/
		// defined in Section 3 (or an absolute HTTP/HTTPS URI containing
		// the /resource name/).
		method := r.Method
		if method != "GET" {
			w.WriteHeader(403)
			fmt.Printf("forbidden method %s\n, must be 'GET'", method)
			return
		}

		if !r.ProtoAtLeast(1, 1) {
			w.WriteHeader(403)
			fmt.Printf("forbidden HTTP protocol version %s\n, must be 1.1 or higher", r.Proto)
			return
		}

		// As per RFC6455:
		// The client includes the hostname in the |Host| header field of its
		// handshake as per [RFC2616], so that both the client and the server
		// can verify that they agree on which host is in use.
		host := r.Host
		fmt.Printf("host: %s\n", host)
		if host != "localhost:8080" {
			w.WriteHeader(403)
			fmt.Printf("forbidden host: %s\n", host)
			return
		}

		// An |Upgrade| header field containing the value "websocket",
		// treated as an ASCII case-insensitive value.
		fmt.Println("HEADERS <3:")
		for name, values := range r.Header {
			for _, value := range values {
				fmt.Println(name, value)
			}
		}
		upgrade := r.Header.Get("Upgrade")
		if upgrade != "websocket" {
			w.WriteHeader(400)
			fmt.Printf("invalid Upgrade header %s, must be 'websocket'", upgrade)
			return
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
			http.Error(w, "error decoding Sec-WebSocket-Key header", http.StatusInternalServerError)
			return
		}
		if len(wsKeyBytes) != 16 {
			http.Error(w, "invalid Sec-WebSocket-Key header, must be 16-bytes long", http.StatusBadRequest)
			fmt.Printf("invalid Sec-WebSocket-Key header length of %d, must be 16-bytes long", len(wsKeyBytes))
			return
		}

		err = WsHandler(w, r)
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
	hj, ok := w.(http.Hijacker)
	if !ok {
		return errors.New("webserver doesn't support http hijacking")
	}
	conn, bufwr, err := hj.Hijack()
	if err != nil {
		return err
	}

	ws := &Ws{conn, bufwr, req}
	defer ws.conn.Close()

	err = ws.Handshake()
	if err != nil {
		fmt.Printf("unsuccessful ws handshake: %s", err.Error())
		return err
	}

	return nil
}

func (ws *Ws) Handshake() error {
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
	wsKeyConcat := strings.TrimSpace(ws.request.Header.Get("Sec-WebSocket-Key")) + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	wsBytes := []byte(wsKeyConcat)
	hasher := sha1.New()
	hasher.Write(wsBytes)
	sha := base64.StdEncoding.EncodeToString(hasher.Sum(nil))
	fmt.Printf("SHA <3: %s\n", sha)

	lines := []string{
		fmt.Sprintf("HTTP/%d.%d 101 Switching Protocols", ws.request.ProtoMajor, ws.request.ProtoMinor),
		"Upgrade: websocket",
		"Connection: Upgrade",
		fmt.Sprintf("Sec-WebSocket-Accept: %s", sha),
		"Sec-WebSocket-Protocol: chat",
		"\n\r",
	}

	return ws.write([]byte(strings.Join(lines, "\r\n")))
}

func (ws *Ws) write(data []byte) error {
	_, err := ws.bufrw.Write(data)
	if err != nil {
		return err
	}
	return ws.bufrw.Flush()
}

type Ws struct {
	conn    net.Conn
	bufrw   *bufio.ReadWriter
	request *http.Request
}
