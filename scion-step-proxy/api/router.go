package api

// This file is auto-generated, don't modify it manually

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/golang-jwt/jwt/v4"
	"github.com/netsys-lab/scion-step-proxy/models"
	"github.com/netsys-lab/scion-step-proxy/pkg/scioncrypto"
	"github.com/netsys-lab/scion-step-proxy/pkg/step"
	"github.com/sirupsen/logrus"
)

// NewRouter creates a new router for the spec and the given handlers.
// CA Service
//
// API for renewing SCION certificates.
//
// 0.1.0
//

type ApiRouter struct {
	TrcPath      string
	JwtSecret    string
	CertDuration string
	Router       http.Handler
}

func NewApiRouter(trcPath, jwtSecret, certDuration string) *ApiRouter {
	r := chi.NewRouter()
	ar := &ApiRouter{
		TrcPath:      trcPath,
		JwtSecret:    jwtSecret,
		CertDuration: certDuration,
		Router:       r,
	}

	r.Get("/healthcheck", healthCheck)
	r.Post("/auth/token", auth)
	r.Post("/ra/isds/{isdNumber}/ases/{asNumber}/certificates/renewal", ar.renewCert)

	return ar
}

func (ar *ApiRouter) renewCert(wr http.ResponseWriter, req *http.Request) {
	var renewRequest models.RenewalRequest
	if err := json.NewDecoder(req.Body).Decode(&renewRequest); err != nil {
		logrus.Error(err)
		sendProblem(wr, "/ra/isds/{isdNumber}/ases/{asNumber}/certificates/renewal", "Could not parse JSON request body", http.StatusBadRequest)
		return
	}
	isdNumber := chi.URLParam(req, "isdNumber")
	asNumber := chi.URLParam(req, "asNumber")
	logrus.Info("Got isd ", isdNumber, "and AS ", asNumber)

	file, err := os.CreateTemp("/tmp/", "*.csr")
	if err != nil {
		logrus.Error(err)
		sendProblem(wr, "/ra/isds/{isdNumber}/ases/{asNumber}/certificates/renewal", "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	bts, err := base64.StdEncoding.DecodeString(renewRequest.Csr)
	if err != nil {
		logrus.Error(err)
		sendProblem(wr, "/ra/isds/{isdNumber}/ases/{asNumber}/certificates/renewal", "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = scioncrypto.ExtractAndVerifyCsr(ar.TrcPath, bts, file)
	if err != nil {
		logrus.Error(err)
		sendProblem(wr, "/ra/isds/{isdNumber}/ases/{asNumber}/certificates/renewal", "Internal Server Error", http.StatusInternalServerError)
		return
	}

	stepCli := step.NewStepCliAdapter()

	certFile, err := os.CreateTemp("/tmp/", "*.crt")
	if err != nil {
		logrus.Error(err)
		sendProblem(wr, "/ra/isds/{isdNumber}/ases/{asNumber}/certificates/renewal", "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = stepCli.SignCert(file.Name(), certFile.Name(), ar.CertDuration)
	os.Remove(file.Name())
	if err != nil {
		logrus.Error(err)
		sendProblem(wr, "/ra/isds/{isdNumber}/ases/{asNumber}/certificates/renewal", "Internal Server Error", http.StatusInternalServerError)
		return
	}

	respCertChain, err := scioncrypto.ExtractCerts(certFile.Name())
	if err != nil {
		logrus.Error(err)
		sendProblem(wr, "/ra/isds/{isdNumber}/ases/{asNumber}/certificates/renewal", "Internal Server Error", http.StatusInternalServerError)
		return
	}
	os.Remove(certFile.Name())

	resp := models.RenewalResponse{
		CertificateChain: respCertChain,
	}
	err = json.NewEncoder(wr).Encode(&resp)
	if err != nil {
		logrus.Error(err)
		sendProblem(wr, "/ra/isds/{isdNumber}/ases/{asNumber}/certificates/renewal", "Could not write response", http.StatusInternalServerError)
		return
	}
}

func auth(wr http.ResponseWriter, req *http.Request) {
	var accessCredentials models.AccessCredentials
	if err := json.NewDecoder(req.Body).Decode(&accessCredentials); err != nil {
		sendProblem(wr, "auth/token", "Could not parse JSON request body", http.StatusBadRequest)
		return
	}

	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"clientId": accessCredentials.ClientId,
		"expires":  "3600",
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte("asdölkölkasd"))
	if err != nil {
		logrus.Error(err)
		sendProblem(wr, "auth/token", "Could not create token", http.StatusInternalServerError)
		return
	}

	accessToken := models.AccessToken{
		AccessToken: tokenString,
		ExpiresIn:   3600,
		TokenType:   "Bearer",
	}
	if err := json.NewEncoder(wr).Encode(&accessToken); err != nil {
		sendProblem(wr, "auth/token", "Could not write response", http.StatusInternalServerError)
		return
	}

}

func healthCheck(wr http.ResponseWriter, req *http.Request) {
	wr.Header().Add("Cache-Control", "no-store")
	healthCheckStatus := models.HealthCheckStatus{
		Status: "available",
	}
	if err := json.NewEncoder(wr).Encode(&healthCheckStatus); err != nil {
		sendProblem(wr, "healthCheck", "Could not write Response", http.StatusInternalServerError)
		return
	}
}

func sendProblem(wr http.ResponseWriter, errorType, title string, status int32) {
	wr.WriteHeader(int(status))
	problem := models.Problem{
		Status: status,
		Type:   errorType,
		Title:  title,
	}
	_ = json.NewEncoder(wr).Encode(&problem)
}
