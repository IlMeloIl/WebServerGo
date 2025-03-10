package main

import (
	"fmt"
	"net/http"
)

func main() {
	serveMux := http.NewServeMux()
	serveMux.Handle("/", http.FileServer(http.Dir(".")))
	serveMux.Handle("/assets/logo.png", http.FileServer(http.Dir(".")))
	if err := http.ListenAndServe(":8080", serveMux); err != nil {
		fmt.Println(err)
	}
}
