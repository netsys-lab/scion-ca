// This file is auto-generated, DO NOT EDIT.
//
// Source:
//     Title: CA Service
//     Version: 0.1.0
package models

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"regexp"
)

// postCertificateRenewalQueryParametersAsNumberPattern is the validation pattern for PostCertificateRenewalQueryParameters.AsNumber
var postCertificateRenewalQueryParametersAsNumberPattern = regexp.MustCompile(`^([a-f0-9]{1,4}:){2}([a-f0-9]{1,4})|\d+$`)

// PostCertificateRenewalQueryParameters is an object.
type PostCertificateRenewalQueryParameters struct {
	// IsdNumber: ISD number of the Autonomous System requesting the certificate chain renewal.
	IsdNumber int32 `json:"isd-number"`
	// AsNumber: AS Number of the Autonomous System requesting the certificate chain renewal.
	AsNumber AS `json:"as-number"`
}

// Validate implements basic validation for this model
func (m PostCertificateRenewalQueryParameters) Validate() error {
	return validation.Errors{
		"asNumber": validation.Validate(
			m.AsNumber, validation.NotNil, validation.Match(postCertificateRenewalQueryParametersAsNumberPattern),
		),
	}.Filter()
}

// GetIsdNumber returns the IsdNumber property
func (m PostCertificateRenewalQueryParameters) GetIsdNumber() int32 {
	return m.IsdNumber
}

// SetIsdNumber sets the IsdNumber property
func (m *PostCertificateRenewalQueryParameters) SetIsdNumber(val int32) {
	m.IsdNumber = val
}

// GetAsNumber returns the AsNumber property
func (m PostCertificateRenewalQueryParameters) GetAsNumber() AS {
	return m.AsNumber
}

// SetAsNumber sets the AsNumber property
func (m *PostCertificateRenewalQueryParameters) SetAsNumber(val AS) {
	m.AsNumber = val
}
