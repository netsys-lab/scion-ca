// This file is auto-generated, DO NOT EDIT.
//
// Source:
//     Title: CA Service
//     Version: 0.1.0
package models

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

// RenewalRequest is an object.
type RenewalRequest struct {
	// Csr: Base64 encoded renewal request as described below.
	// The renewal requests consists of a CMS SignedData structure that
	// contains a PKCS#10 defining the parameters of the requested
	// certificate.
	// The following must hold for the CMS structure:
	// - The `certificates` field in `SignedData` MUST contain an existing
	//   and verifiable certificate chain that authenticates the private
	//   key that was used to sign the CMS structure. It MUST NOT contain
	//   any other certificates.
	// - The `eContentType` is set to `id-data`. The contents of `eContent`
	//   is the ASN.1 DER encoded PKCS#10. This ensures backwards
	//   compatibility with PKCS#7, as described in
	//   [RFC5652](https://tools.ietf.org/html/rfc5652#section-5.2.1)
	// - The `SignerIdentifier` MUST be the choice `IssuerAndSerialNumber`,
	//   thus, `version` in `SignerInfo` must be 1, as required by
	//   [RFC5652](https://tools.ietf.org/html/rfc5652#section-5.3)
	Csr string `json:"csr"`
}

// Validate implements basic validation for this model
func (m RenewalRequest) Validate() error {
	return validation.Errors{
		"csr": validation.Validate(
			m.Csr, validation.Required, is.Base64,
		),
	}.Filter()
}

// GetCsr returns the Csr property
func (m RenewalRequest) GetCsr() string {
	return m.Csr
}

// SetCsr sets the Csr property
func (m *RenewalRequest) SetCsr(val string) {
	m.Csr = val
}
