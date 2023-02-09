package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("./public/index.html")
		if err != nil {
			fmt.Println(err.Error())
		}

		io.Copy(w, f)
	})

	http.HandleFunc("/new", func(w http.ResponseWriter, r *http.Request) {
		// uuid, _ := exec.Command("uuidgen").Output()
		// redirectURL := fmt.Sprintf("%s/game/%s", r.URL.Host, uuid)
		// http.Redirect(w, r, redirectURL, http.StatusSeeOther)

		redirectURL := fmt.Sprintf("%s/game", r.URL.Host)
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
	})

	http.HandleFunc("/game", func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("./public/game/index.html")
		if err != nil {
			fmt.Println(err.Error())
		}

		io.Copy(w, f)
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
