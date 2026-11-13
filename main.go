// A tiny web server: something real to build, attest and record.
package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, "Hello from the in-toto GitHub demo")
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
