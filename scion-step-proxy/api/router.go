package api

// This file is auto-generated, don't modify it manually

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/golang-jwt/jwt/v4"
	"github.com/netsys-lab/scion-step-proxy/database"
	"github.com/netsys-lab/scion-step-proxy/models"
	"github.com/netsys-lab/scion-step-proxy/pkg/scioncrypto"
	"github.com/netsys-lab/scion-step-proxy/pkg/step"
	caconfig "github.com/scionproto/scion/private/ca/config"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
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
	JwtSecret    []byte
	CertDuration string
	DB           *gorm.DB
	Router       http.Handler
}

func NewApiRouter(trcPath, jwtSecret, certDuration string, db *gorm.DB) *ApiRouter {
	r := chi.NewRouter()

	secretKey := caconfig.NewPEMSymmetricKey(jwtSecret)

	secretValue, err := secretKey.Get()
	if err != nil {
		logrus.Fatal(err)
	}

	ar := &ApiRouter{
		TrcPath:      trcPath,
		JwtSecret:    secretValue,
		CertDuration: certDuration,
		DB:           db,
		Router:       r,
	}

	r.Get("/healthcheck", healthCheck)
	r.Post("/auth/token", ar.auth)
	r.Post("/ra/isds/{isdNumber}/ases/{asNumber}/certificates/renewal", ar.renewCert)

	return ar
}

func randomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[:length]
}

func (ar *ApiRouter) renewCert(wr http.ResponseWriter, req *http.Request) {

	_, err := authJwt(req.Header.Get("authorization"), ar.JwtSecret)
	if err != nil {
		logrus.Warn("JWT auth failed")
		logrus.Error(err)
		sendProblem(wr, "/ra/isds/{isdNumber}/ases/{asNumber}/certificates/renewal", "JWT auth failed", http.StatusUnauthorized)
		return
	}

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

	csr, err := scioncrypto.ExtractAndVerifyCsr(ar.TrcPath, bts, file)
	if err != nil {
		logrus.Error(err)
		sendProblem(wr, "/ra/isds/{isdNumber}/ases/{asNumber}/certificates/renewal", "Internal Server Error", http.StatusInternalServerError)
		return
	}

	stepCli := step.NewStepCliAdapter()

	// certFile, err := os.CreateTemp("/tmp/", "*.crt")
	certFileName := filepath.Join("/tmp", fmt.Sprintf("%s.cert", randomString(16)))
	if err != nil {
		logrus.Error(err)
		sendProblem(wr, "/ra/isds/{isdNumber}/ases/{asNumber}/certificates/renewal", "Internal Server Error", http.StatusInternalServerError)
		return
	}
	err = os.Chmod(file.Name(), 0777)
	if err != nil {
		logrus.Error(err)
		sendProblem(wr, "/ra/isds/{isdNumber}/ases/{asNumber}/certificates/renewal", "Internal Server Error", http.StatusInternalServerError)
		return
	}

	isdAS := fmt.Sprintf("%s-%s", isdNumber, asNumber)
	if len(csr.Subject.ExtraNames) > 0 {
		str, ok := csr.Subject.ExtraNames[0].Value.(string)
		if ok {
			isdAS = str
		}
	}
	logrus.Info("Got ISDAS ", isdAS)
	err = stepCli.SignCert(file.Name(), certFileName, ar.CertDuration, isdAS)
	// os.Remove(file.Name())
	if err != nil {
		logrus.Error(err)
		sendProblem(wr, "/ra/isds/{isdNumber}/ases/{asNumber}/certificates/renewal", "Internal Server Error", http.StatusInternalServerError)
		return
	}

	respCertChain, err := scioncrypto.ExtractCerts(certFileName)
	if err != nil {
		logrus.Error(err)
		sendProblem(wr, "/ra/isds/{isdNumber}/ases/{asNumber}/certificates/renewal", "Internal Server Error", http.StatusInternalServerError)
		return
	}
	// os.Remove(certFileName)

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

func authJwt(tokenStr string, secret []byte) (*jwt.Token, error) {

	// Remove bearer things
	realToken := strings.Replace(tokenStr, "Bearer ", "", 1)
	logrus.Info(realToken)
	logrus.Info(time.Now())

	timeOffset := ""
	timeOffsetEnv := os.Getenv("JWT_SUPPORTED_TIME_OFFSET_MINS")
	if timeOffsetEnv != "" {
		timeOffset = timeOffsetEnv
	}

	if timeOffset != "" {
		realTimeOffset, err := strconv.Atoi(timeOffset)
		if err != nil {
			logrus.Warn("Failed to pase JWT_SUPPORTED_TIME_OFFSET_MINS=", timeOffsetEnv)
		} else {
			logrus.Info("Adjusting time offset to ", realTimeOffset, " minutes")
			jwt.TimeFunc = func() time.Time {
				return time.Now().Add(time.Minute * time.Duration(realTimeOffset))
			}
		}

	}

	// Parse takes the token string and a function for looking up the key. The latter is especially
	// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
	// head of the token to identify which key to use, but the parsed token (head and claims) is provided
	// to the callback, providing flexibility.
	token, err := jwt.Parse(realToken, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return token, nil
	} else {
		return nil, fmt.Errorf("Token invalid")
	}
}

// Unused for now, the CS is issuing its own jwts
func (ar *ApiRouter) auth(wr http.ResponseWriter, req *http.Request) {
	var accessCredentials models.AccessCredentials
	if err := json.NewDecoder(req.Body).Decode(&accessCredentials); err != nil {
		sendProblem(wr, "auth/token", "Could not parse JSON request body", http.StatusBadRequest)
		return
	}

	var user database.User
	result := ar.DB.Where("clientId = ? AND clientSecret = ?", accessCredentials.ClientId, accessCredentials.ClientSecret).First(&user)
	if result.Error != nil || result.RowsAffected == 0 {
		if result.Error != nil {
			logrus.Error(result.Error)
		}
		sendProblem(wr, "/ra/isds/{isdNumber}/ases/{asNumber}/certificates/renewal", "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"clientId": accessCredentials.ClientId,
		"expires":  "3600",
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(ar.JwtSecret))
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
