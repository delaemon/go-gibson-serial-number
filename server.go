package main
import (
	"net/http"
	"fmt"
)

func main () {
	Server()
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "gibson")
}

func Server() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
