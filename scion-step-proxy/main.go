package main

import (
	"flag"
	"net/http"

	"github.com/netsys-lab/scion-step-proxy/api"
	"github.com/sirupsen/logrus"
)

var (
	loglevel     = flag.String("loglevel", "INFO", "Log level (ERROR|WARN|INFO|DEBUG|TRACE)")
	local        = flag.String("local", ":8088", "Local listen address (default: :3000)")
	jwtSecrect   = flag.String("jwtSecrect", "", "Secret to generate JWT's (default: unset)")
	certDuration = flag.String("certDuration", "1d", "Expiration Time of certs (default: 1d)")
	trcPath      = flag.String("trcPath", "", "Path to trc files with the format $ISD-$base-$serial.trc (default: '')")
)

func main() {
	flag.Parse()

	if jwtSecrect == nil || *jwtSecrect == "" {
		logrus.Fatal("No jwtSecret provided")
	}

	r := api.NewApiRouter(*trcPath, *jwtSecrect, *certDuration)
	http.ListenAndServe(*local, r.Router)
}
