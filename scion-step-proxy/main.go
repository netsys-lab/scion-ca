package main

import (
	"net/http"

	"github.com/netsys-lab/scion-step-proxy/api"
)

func main() {

	r := api.NewApiRouter("", "lkasjdlkasjdlkasjdklsaj", "3d")
	http.ListenAndServe(":3000", r.Router)
}
