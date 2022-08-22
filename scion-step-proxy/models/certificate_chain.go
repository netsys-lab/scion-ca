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

// CertificateChain is an object.
type CertificateChain struct {
	// AsCertificate: Base64 encoded AS certificate.
	AsCertificate string `json:"as_certificate"`
	// CaCertificate: Base64 encoded CA certificate.
	CaCertificate string `json:"ca_certificate"`
}

// Validate implements basic validation for this model
func (m CertificateChain) Validate() error {
	return validation.Errors{
		"asCertificate": validation.Validate(
			m.AsCertificate, validation.Required, is.Base64,
		),
		"caCertificate": validation.Validate(
			m.CaCertificate, validation.Required, is.Base64,
		),
	}.Filter()
}

// GetAsCertificate returns the AsCertificate property
func (m CertificateChain) GetAsCertificate() string {
	return m.AsCertificate
}

// SetAsCertificate sets the AsCertificate property
func (m *CertificateChain) SetAsCertificate(val string) {
	m.AsCertificate = val
}

// GetCaCertificate returns the CaCertificate property
func (m CertificateChain) GetCaCertificate() string {
	return m.CaCertificate
}

// SetCaCertificate sets the CaCertificate property
func (m *CertificateChain) SetCaCertificate(val string) {
	m.CaCertificate = val
}
