package main

import (
	"flag"
	"net/http"
	"time"

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

// Logrus setup function
func configureLogging() error {
	l, err := logrus.ParseLevel(*loglevel)
	if err != nil {
		return err
	}
	logrus.SetLevel(l)
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat:   time.RFC3339Nano,
		DisableHTMLEscape: true,
	})
	return nil
}

func main() {
	flag.Parse()

	err := configureLogging()
	if err != nil {
		logrus.Fatal(err)
	}

	if jwtSecrect == nil || *jwtSecrect == "" {
		logrus.Fatal("No jwtSecret provided")
	}

	r := api.NewApiRouter(*trcPath, *jwtSecrect, *certDuration)
	http.ListenAndServe(*local, r.Router)
}
