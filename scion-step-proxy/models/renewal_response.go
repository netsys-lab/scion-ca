// This file is auto-generated, DO NOT EDIT.
//
// Source:
//     Title: CA Service
//     Version: 0.1.0
package models

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// RenewalResponse is an object.
type RenewalResponse struct {
	// CertificateChain:
	CertificateChain interface{} `json:"certificate_chain"`
}

// Validate implements basic validation for this model
func (m RenewalResponse) Validate() error {
	return validation.Errors{}.Filter()
}

// GetCertificateChain returns the CertificateChain property
func (m RenewalResponse) GetCertificateChain() interface{} {
	return m.CertificateChain
}

// SetCertificateChain sets the CertificateChain property
func (m *RenewalResponse) SetCertificateChain(val interface{}) {
	m.CertificateChain = val
}
