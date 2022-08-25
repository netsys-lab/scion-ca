package scioncrypto

import (
	"crypto/x509"
	"encoding/base64"
	"io/ioutil"

	"github.com/netsys-lab/scion-step-proxy/models"
)

func ExtractCerts(path string) (*models.CertificateChain, error) {
	chain, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	certChain := decodePem(chain)
	caCert, err := x509.ParseCertificate(certChain.Certificate[1])
	if err != nil {
		return nil, err
	}
	asCert, err := x509.ParseCertificate(certChain.Certificate[0])
	if err != nil {
		return nil, err
	}

	asCertStr := base64.StdEncoding.EncodeToString(asCert.Raw)
	caCertStr := base64.StdEncoding.EncodeToString(caCert.Raw)

	respCertChain := &models.CertificateChain{
		AsCertificate: asCertStr,
		CaCertificate: caCertStr,
	}
	return respCertChain, nil
}
