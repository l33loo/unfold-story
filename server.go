package main

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"log"
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
		// As per RFC6455:
		// The client includes the hostname in the |Host| header field of its
		// handshake as per [RFC2616], so that both the client and the server
		// can verify that they agree on which host is in use.
		host := r.Header.Get("Host")
		if host != "localhost" {
			w.WriteHeader(403)
			fmt.Printf("forbidden host: %s\n", host)
			return
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
		wsKey := r.Header.Get("Sec-WebSocket-Key")
		wsKeyConcat := strings.TrimSpace(wsKey) + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
		wsBytes := []byte(wsKeyConcat)
		hasher := sha1.New()
		hasher.Write(wsBytes)
		sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

		w.WriteHeader(101)
		w.Header().Add("Upgrade", "websocket")
		w.Header().Add("Connection", "Upgrade")
		w.Header().Add("Sec-WebSocket-Accept", sha)
		w.Header().Add("Sec-WebSocket-Protocol", "chat")
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
