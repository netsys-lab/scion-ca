package main

import (
	"net/http"

	"github.com/netsys-lab/scion-step-proxy/api"
)

func main() {

	r := api.NewRouter()
	http.ListenAndServe(":3000", r)
}
