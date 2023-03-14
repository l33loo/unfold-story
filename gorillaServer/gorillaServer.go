package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-type", "text/html")
		f, err := os.Open("../public/index.html")
		if err != nil {
			fmt.Println(err.Error())
		}

		io.Copy(w, f)
	})

	http.HandleFunc("/public/styles.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-type", "text/css")
		f, err := os.Open("../public/styles.css")

		if err != nil {
			fmt.Println(err.Error())
		}

		io.Copy(w, f)
	})

	http.HandleFunc("/public/scripts.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-type", "text/javascript")
		f, err := os.Open("../public/scripts.js")

		if err != nil {
			fmt.Println(err.Error())
		}

		io.Copy(w, f)
	})

	http.HandleFunc("/public/game/styles.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-type", "text/css")
		f, err := os.Open("../public/game/styles.css")

		if err != nil {
			fmt.Println(err.Error())
		}

		io.Copy(w, f)
	})

	http.HandleFunc("/public/game/scripts.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-type", "text/javascript")
		f, err := os.Open("../public/game/scripts.js")

		if err != nil {
			fmt.Println(err.Error())
		}

		io.Copy(w, f)
	})

	http.HandleFunc("/game", func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("../public/game/index.html")
		if err != nil {
			fmt.Println(err.Error())
		}

		io.Copy(w, f)
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		// Just a little test with echo
		for {
			messageType, _, err := conn.ReadMessage()
			if err != nil {
				log.Println(err)
				return
			}

			// echo := p[0:]

			// echo = append(echo, []byte(" from server")...)

			largeMsg := "lllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll lllllllllllllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll llllllllllllllllllllllllllll"

			fmt.Println("largeMsg length", len(largeMsg))
			err = conn.WriteMessage(messageType, []byte(largeMsg))
			if err != nil {
				log.Panicln(err)
				return
			}
		}
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
