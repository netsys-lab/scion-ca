package api

// This file is auto-generated, don't modify it manually

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/golang-jwt/jwt/v4"
	"github.com/netsys-lab/scion-step-proxy/models"
	"github.com/netsys-lab/scion-step-proxy/pkg/scioncrypto"
	"github.com/netsys-lab/scion-step-proxy/pkg/step"
	"github.com/scionproto/scion/pkg/private/serrors"
	"github.com/scionproto/scion/pkg/scrypto/cms/protocol"
	"github.com/sirupsen/logrus"
)

// NewRouter creates a new router for the spec and the given handlers.
// CA Service
//
// API for renewing SCION certificates.
//
// 0.1.0
//

func NewRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/healthcheck", healthCheck)
	r.Post("/auth/token", auth)
	r.Post("/ra/isds/{isdNumber}/ases/{asNumber}/certificates/renewal", renewCert)
	return r
}

func renewCert(wr http.ResponseWriter, req *http.Request) {
	var renewRequest models.RenewalRequest
	if err := json.NewDecoder(req.Body).Decode(&renewRequest); err != nil {
		wr.WriteHeader(http.StatusBadRequest)
		problem := models.Problem{
			Status: http.StatusBadRequest,
			Type:   "/ra/isds/{isdNumber}/ases/{asNumber}/certificates/renewal",
			Title:  "Could not parse JSON request body",
		}
		_ = json.NewEncoder(wr).Encode(&problem)
		return
	}
	isdNumber := chi.URLParam(req, "isdNumber")
	asNumber := chi.URLParam(req, "asNumber")
	logrus.Info("Got isd ", isdNumber, "and AS ", asNumber)
	// renewRequest.Csr
	logrus.Info(renewRequest.Csr)

	file, err := os.CreateTemp("", "")
	if err != nil {
		logrus.Error(err)
		return
	}

	bts, err := base64.StdEncoding.DecodeString(renewRequest.Csr)
	if err != nil {
		logrus.Error(err)
		return
	}

	r := scioncrypto.RequestVerifier{
		TRCFetcher: &scioncrypto.LocalFetcher{},
	}

	// csr, err := s.Verifier.VerifyCMSSignedRenewalRequest(ctx, req.CmsSignedRequest)
	csr, err := VerifyCMSSignedRenewalRequest(context.Background(), bts, &r)
	if err != nil {
		logrus.Error(err)
		return
	}

	err = pem.Encode(file, &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csr.Raw})
	if err != nil {
		logrus.Error(err)
		return
	}

	stepCli := step.NewStepCliAdapter()
	err = stepCli.SignCert(file.Name(), "./issued.cert", "8h")
	if err != nil {
		logrus.Error(err)
		return
	}

	/*outFile, err := os.Open("./issued.cert")
	if err != nil {
		logrus.Error(err)
		return
	}*/

	chain, err := ioutil.ReadFile("./issued.cert")
	if err != nil {
		logrus.Error(err)
		return
	}

	certChain := decodePem(chain)
	caCert, err := x509.ParseCertificate(certChain.Certificate[1])
	if err != nil {
		logrus.Error(err)
		return
	}
	asCert, err := x509.ParseCertificate(certChain.Certificate[0])
	if err != nil {
		logrus.Error(err)
		return
	}

	asCertStr := base64.StdEncoding.EncodeToString(asCert.Raw)
	caCertStr := base64.StdEncoding.EncodeToString(caCert.Raw)

	respCertChain := models.CertificateChain{
		AsCertificate: asCertStr,
		CaCertificate: caCertStr,
	}

	resp := models.RenewalResponse{
		CertificateChain: respCertChain,
	}
	err = json.NewEncoder(wr).Encode(&resp)
	if err != nil {
		logrus.Error(err)
		return
	}

	/*ci := protocol.ContentInfo{
		ContentType: oid.ContentTypeSignedData,
		Content:     asn1.RawValue{Bytes: cert},
	}
	cert, err = asn1.Marshal(ci)
	if err != nil {
		logrus.Error(err)
		return
	}
	certStr := base64.StdEncoding.EncodeToString(cert)
	bts = []byte(certStr)

	wr.WriteHeader(http.StatusOK)
	resp := models.RenewalResponse{
		CertificateChain: bts,
	}
	err = json.NewEncoder(wr).Encode(&resp)
	if err != nil {
		logrus.Error(err)
		return
	}*/

	// wr.Write(bts)

	// Fix chain validation
	// Pass csr to step ca
	// step ca sign --not-after=1440h switch.csr switch-new.crt
	// Implement proper token auth
}

func decodePem(certInput []byte) tls.Certificate {
	var cert tls.Certificate
	certPEMBlock := []byte(certInput)
	var certDERBlock *pem.Block
	for {
		certDERBlock, certPEMBlock = pem.Decode(certPEMBlock)
		if certDERBlock == nil {
			break
		}
		if certDERBlock.Type == "CERTIFICATE" {
			cert.Certificate = append(cert.Certificate, certDERBlock.Bytes)
		}
	}
	return cert
}

func VerifyCMSSignedRenewalRequest(ctx context.Context,
	req []byte, r *scioncrypto.RequestVerifier) (*x509.CertificateRequest, error) {

	ci, err := protocol.ParseContentInfo(req)
	if err != nil {
		return nil, serrors.WrapStr("parsing ContentInfo", err)
	}
	sd, err := ci.SignedDataContent()
	if err != nil {
		return nil, serrors.WrapStr("parsing SignedData", err)
	}

	chain, err := scioncrypto.ExtractChain(sd)
	if err != nil {
		return nil, serrors.WrapStr("extracting signing certificate chain", err)
	}

	if err := r.VerifySignature(ctx, sd, chain); err != nil {
		return nil, err
	}

	pld, err := sd.EncapContentInfo.EContentValue()
	if err != nil {
		return nil, serrors.WrapStr("reading payload", err)
	}

	csr, err := x509.ParseCertificateRequest(pld)
	if err != nil {
		return nil, serrors.WrapStr("parsing CSR", err)
	}

	return csr, nil // r.processCSR(csr, chain[0])
}

func auth(wr http.ResponseWriter, req *http.Request) {
	var accessCredentials models.AccessCredentials
	if err := json.NewDecoder(req.Body).Decode(&accessCredentials); err != nil {
		wr.WriteHeader(http.StatusBadRequest)
		problem := models.Problem{
			Status: http.StatusBadRequest,
			Type:   "auth/token",
			Title:  "Could not parse JSON request body",
		}
		_ = json.NewEncoder(wr).Encode(&problem)
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
		fmt.Println(err)
		wr.WriteHeader(http.StatusInternalServerError)
		problem := models.Problem{
			Status: http.StatusInternalServerError,
			Type:   "auth/token",
			Title:  "Could not create token",
		}
		_ = json.NewEncoder(wr).Encode(&problem)
		return
	}
	fmt.Println(tokenString)
	accessToken := models.AccessToken{
		AccessToken: tokenString,
		ExpiresIn:   3600,
		TokenType:   "Bearer",
	}
	if err := json.NewEncoder(wr).Encode(&accessToken); err != nil {
		wr.WriteHeader(http.StatusInternalServerError)
		return
	}

}

func healthCheck(wr http.ResponseWriter, req *http.Request) {
	wr.Header().Add("Cache-Control", "no-store")
	healthCheckStatus := models.HealthCheckStatus{
		Status: "available",
	}
	if err := json.NewEncoder(wr).Encode(&healthCheckStatus); err != nil {
		wr.WriteHeader(http.StatusInternalServerError)
		return
	}
}
