package main

import (
	"fmt"
	"net/http"
)

func main() {
	serveMux := http.NewServeMux()
	if err := http.ListenAndServe("localhost:8080", serveMux); err != nil {
		fmt.Println(err)
	}
}
