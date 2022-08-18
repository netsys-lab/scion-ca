package api

// This file is auto-generated, don't modify it manually

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	"github.com/golang-jwt/jwt/v4"
	"github.com/scionproto/scion/pkg/private/serrors"
	"github.com/scionproto/scion/pkg/scrypto/cms/protocol"
	"github.com/scionproto/scion/pkg/scrypto/cppki"
	"github.com/scionproto/scion/private/ca/renewal"
	"github.com/sirupsen/logrus"
)

// NewRouter creates a new router for the spec and the given handlers.
// CA Service
//
// API for renewing SCION certificates.
//
// 0.1.0
//

// DecodeSignedTRC parses the signed TRC.
func DecodeSignedTRC(raw []byte) (cppki.SignedTRC, error) {
	ci, err := protocol.ParseContentInfo(raw)
	if err != nil {
		return cppki.SignedTRC{}, serrors.WrapStr("error parsing ContentInfo", err)
	}
	sd, err := ci.SignedDataContent()
	if err != nil {
		return cppki.SignedTRC{}, serrors.WrapStr("error parsing SignedData", err)
	}
	if sd.Version != 1 {
		return cppki.SignedTRC{}, serrors.New("unsupported SignedData version", "version", 1)
	}
	if !sd.EncapContentInfo.IsTypeData() {
		return cppki.SignedTRC{}, serrors.WrapStr("unsupported EncapContentInfo type", err,
			"type", sd.EncapContentInfo.EContentType)
	}
	praw, err := sd.EncapContentInfo.EContentValue()
	if err != nil {
		return cppki.SignedTRC{}, serrors.WrapStr("error reading raw payload", err)
	}
	trc, err := cppki.DecodeTRC(praw)
	if err != nil {
		return cppki.SignedTRC{}, serrors.WrapStr("error parsing TRC payload", err)
	}
	return cppki.SignedTRC{Raw: raw, TRC: trc, SignerInfos: sd.SignerInfos}, nil
}

type LocalFetcher struct {
}

func (lf *LocalFetcher) SignedTRC(ctx context.Context, id cppki.TRCID) (cppki.SignedTRC, error) {
	trc := cppki.SignedTRC{}
	logrus.Warn("Reading TRC ", id.String())
	bts, err := os.ReadFile(id.String())
	if err != nil {
		return trc, nil
	}

	block, _ := pem.Decode(bts)
	logrus.Warn("READ TRC")
	sTrc, err := DecodeSignedTRC(block.Bytes)
	if err != nil {
		return trc, err
	}
	return sTrc, nil
}

type TRCFetcher interface {
	// SignedTRC fetches the signed TRC for a given ID.
	// The latest TRC can be requested by setting the serial and base number
	// to scrypto.LatestVer.
	SignedTRC(ctx context.Context, id cppki.TRCID) (cppki.SignedTRC, error)
}

type RequestVerifier struct {
	TRCFetcher TRCFetcher
}

// VerifyCMSSignedRenewalRequest verifies a renewal request that is encapsulated in a CMS
// envelop. It checks that the contained CSR is valid and correctly self-signed, and
// that the signature is valid and can be verified by the chain included in the CMS envelop.
func (r RequestVerifier) VerifyCMSSignedRenewalRequest(ctx context.Context,
	req []byte) (*x509.CertificateRequest, error) {

	ci, err := protocol.ParseContentInfo(req)
	if err != nil {
		return nil, serrors.WrapStr("parsing ContentInfo", err)
	}
	sd, err := ci.SignedDataContent()
	if err != nil {
		return nil, serrors.WrapStr("parsing SignedData", err)
	}

	chain, err := ExtractChain(sd)
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

	return r.processCSR(csr, chain[0])
}

// VerifySignature verifies the signature on the signed data with the provided
// chain. It is checked that the certificate chain is verifiable with an
// active TRC, and that the signature can be verified with the chain.
func (r RequestVerifier) VerifySignature(
	ctx context.Context,
	sd *protocol.SignedData,
	chain []*x509.Certificate,
) error {

	if sd.Version != 1 {
		return serrors.New("unsupported SignedData version", "actual", sd.Version, "supported", 1)
	}
	if c := len(sd.SignerInfos); c != 1 {
		return serrors.New("unexpected number of SignerInfos", "count", c)
	}
	si := sd.SignerInfos[0]
	signer, err := si.FindCertificate(chain)
	if err != nil {
		return serrors.WrapStr("selecting client certificate", err)
	}
	if signer != chain[0] {
		return serrors.New("not signed with AS certificate",
			"common_name", signer.Subject.CommonName)
	}
	if err := r.verifyClientChain(ctx, chain); err != nil {
		return serrors.WrapStr("verifying client chain", err)
	}

	if !sd.EncapContentInfo.IsTypeData() {
		return serrors.New("unsupported EncapContentInfo type",
			"type", sd.EncapContentInfo.EContentType)
	}
	pld, err := sd.EncapContentInfo.EContentValue()
	if err != nil {
		return serrors.WrapStr("reading payload", err)
	}

	if err := verifySignerInfo(pld, chain[0], si); err != nil {
		return serrors.WrapStr("verifying signer info", err)
	}

	return nil
}

func (r RequestVerifier) verifyClientChain(ctx context.Context, chain []*x509.Certificate) error {
	ia, err := cppki.ExtractIA(chain[0].Subject)
	if err != nil {
		return err
	}
	tid := cppki.TRCID{
		ISD:    ia.ISD(),
		Serial: 1,
		Base:   1,
	}
	trc, err := r.TRCFetcher.SignedTRC(ctx, tid)
	if err != nil {
		return serrors.WrapStr("loading TRC to verify client chain", err)
	}
	if trc.IsZero() {
		return serrors.New("TRC not found", "isd", ia.ISD())
	}
	now := time.Now()
	if val := trc.TRC.Validity; !val.Contains(now) {
		return serrors.New("latest TRC currently not active", "validity", val, "current_time", now)
	}
	opts := cppki.VerifyOptions{TRC: []*cppki.TRC{&trc.TRC}}
	if err := cppki.VerifyChain(chain, opts); err != nil {
		// If the the previous TRC is in grace period the CA certificate of the chain might
		// have been issued with a previous Root. Try verifying with the TRC in grace period.
		if now.After(trc.TRC.GracePeriodEnd()) {
			return serrors.WrapStr("verifying client chain", err)
		}
		graceID := trc.TRC.ID
		graceID.Serial--
		if err := r.verifyWithGraceTRC(ctx, now, graceID, chain); err != nil {
			return serrors.WrapStr("verifying client chain with TRC in grace period "+
				"after verification failure with latest TRC", err,
				"trc_id", trc.TRC.ID,
				"grace_trc_id", graceID,
			)
		}

	}
	return nil
}

func (r RequestVerifier) verifyWithGraceTRC(
	ctx context.Context,
	now time.Time,
	id cppki.TRCID,
	chain []*x509.Certificate,
) error {

	trc, err := r.TRCFetcher.SignedTRC(ctx, id)
	if err != nil {
		return serrors.WrapStr("loading TRC in grace period", err)
	}
	if trc.IsZero() {
		return serrors.New("TRC in grace period not found")
	}
	if val := trc.TRC.Validity; !val.Contains(now) {
		return serrors.New("TRC in grace period not active",
			"validity", val,
			"current_time", now,
		)
	}
	verifyOptions := cppki.VerifyOptions{TRC: []*cppki.TRC{&trc.TRC}}
	if err := cppki.VerifyChain(chain, verifyOptions); err != nil {
		return serrors.WrapStr("verifying client chain", err)
	}
	return nil
}

func verifySignerInfo(pld []byte, cert *x509.Certificate, si protocol.SignerInfo) error {
	hash, err := si.Hash()
	if err != nil {
		return err
	}
	attrDigest, err := si.GetMessageDigestAttribute()
	if err != nil {
		return err
	}
	actualDigest := hash.New()
	actualDigest.Write(pld)
	if !bytes.Equal(attrDigest, actualDigest.Sum(nil)) {
		return serrors.New("message digest does not match")
	}
	sigInput, err := si.SignedAttrs.MarshaledForVerifying()
	if err != nil {
		return err
	}
	algo := si.X509SignatureAlgorithm()
	return cert.CheckSignature(algo, sigInput, si.Signature)
}

func (r RequestVerifier) processCSR(csr *x509.CertificateRequest,
	cert *x509.Certificate) (*x509.CertificateRequest, error) {

	csrIA, err := cppki.ExtractIA(csr.Subject)
	if err != nil {
		return nil, serrors.WrapStr("extracting ISD-AS from CSR", err)
	}
	chainIA, err := cppki.ExtractIA(cert.Subject)
	if err != nil {
		return nil, serrors.WrapStr("extracting ISD-AS from certificate chain", err)
	}
	if !csrIA.Equal(chainIA) {
		return nil, serrors.New("signing subject is different from CSR subject",
			"csr_isd_as", csrIA, "chain_isd_as", chainIA)
	}
	if err := csr.CheckSignature(); err != nil {
		return nil, serrors.WrapStr("invalid CSR signature", err)
	}
	return csr, nil
}

func ExtractChain(sd *protocol.SignedData) ([]*x509.Certificate, error) {
	certs, err := sd.X509Certificates()
	if err == nil {
		if len(certs) == 0 {
			err = protocol.ErrNoCertificate
		} else if len(certs) != 2 {
			err = serrors.New("unexpected number of certificates", "count", len(certs))
		}
	}
	if err != nil {
		return nil, serrors.WrapStr("parsing certificate chain", err)
	}

	certType, err := cppki.ValidateCert(certs[0])
	if err != nil {
		return nil, serrors.WrapStr("checking certificate type", err)
	}
	if certType == cppki.CA {
		certs[0], certs[1] = certs[1], certs[0]
	}
	if err := cppki.ValidateChain(certs); err != nil {
		return nil, serrors.WrapStr("validating chain", err)
	}
	return certs, nil
}

func NewRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/healthcheck", healthCheck)
	r.Post("/auth/token", auth)
	r.Post("/ra/isds/{isdNumber}/ases/{asNumber}/certificates/renewal", renewCert)
	return r
}

func extractChain(raw []byte) ([]*x509.Certificate, error) {
	ci, err := protocol.ParseContentInfo(raw)
	if err != nil {
		return nil, serrors.WrapStr("parsing ContentInfo", err)
	}
	sd, err := ci.SignedDataContent()
	if err != nil {
		return nil, serrors.WrapStr("parsing SignedData", err)
	}
	return renewal.ExtractChain(sd)
}

func renewCert(wr http.ResponseWriter, req *http.Request) {
	var renewRequest RenewalRequest
	if err := json.NewDecoder(req.Body).Decode(&renewRequest); err != nil {
		wr.WriteHeader(http.StatusBadRequest)
		problem := Problem{
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

	file, err := os.Create("tmp")
	if err != nil {
		logrus.Error(err)
	}

	bts, err := base64.StdEncoding.DecodeString(renewRequest.Csr)
	if err != nil {
		logrus.Error(err)
	}

	r := RequestVerifier{
		TRCFetcher: &LocalFetcher{},
	}

	// csr, err := s.Verifier.VerifyCMSSignedRenewalRequest(ctx, req.CmsSignedRequest)
	csr, err := VerifyCMSSignedRenewalRequest(context.Background(), bts, &r)
	if err != nil {
		logrus.Error(err)
		return
	}

	// logrus.Warn(csr)

	err = pem.Encode(file, &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csr.Raw})
	// _, err = file.WriteString(renewRequest.Csr)
	if err != nil {
		logrus.Error(err)
	}

	// Fix chain validation
	// Pass csr to step ca
	// step ca sign --not-after=1440h switch.csr switch-new.crt
	// Implement proper token auth
}

/*
func ExtractChain(sd *protocol.SignedData) ([]*x509.Certificate, error) {
	certs, err := sd.X509Certificates()
	if err == nil {
		if len(certs) == 0 {
			err = protocol.ErrNoCertificate
		} else if len(certs) != 2 {
			err = serrors.New("unexpected number of certificates", "count", len(certs))
		}
	}
	if err != nil {
		return nil, serrors.WrapStr("parsing certificate chain", err)
	}

	certType, err := cppki.ValidateCert(certs[0])
	if err != nil {
		return nil, serrors.WrapStr("checking certificate type", err)
	}
	if certType == cppki.CA {
		certs[0], certs[1] = certs[1], certs[0]
	}
	if err := cppki.ValidateChain(certs); err != nil {
		return nil, serrors.WrapStr("validating chain", err)
	}
	return certs, nil
}*/

func VerifyCMSSignedRenewalRequest(ctx context.Context,
	req []byte, r *RequestVerifier) (*x509.CertificateRequest, error) {

	ci, err := protocol.ParseContentInfo(req)
	if err != nil {
		return nil, serrors.WrapStr("parsing ContentInfo", err)
	}
	sd, err := ci.SignedDataContent()
	if err != nil {
		return nil, serrors.WrapStr("parsing SignedData", err)
	}

	chain, err := ExtractChain(sd)
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
	var accessCredentials AccessCredentials
	if err := json.NewDecoder(req.Body).Decode(&accessCredentials); err != nil {
		wr.WriteHeader(http.StatusBadRequest)
		problem := Problem{
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
		problem := Problem{
			Status: http.StatusInternalServerError,
			Type:   "auth/token",
			Title:  "Could not create token",
		}
		_ = json.NewEncoder(wr).Encode(&problem)
		return
	}
	fmt.Println(tokenString)
	accessToken := AccessToken{
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
	healthCheckStatus := HealthCheckStatus{
		Status: "available",
	}
	if err := json.NewEncoder(wr).Encode(&healthCheckStatus); err != nil {
		wr.WriteHeader(http.StatusInternalServerError)
		return
	}

	// wr.WriteHeader(http.StatusOK)
}

/*func optionsHandlerFunc(allowedMethods ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Allow", strings.Join(allowedMethods, ", "))
	}
}*/
