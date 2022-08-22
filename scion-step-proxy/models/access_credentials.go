// This file is auto-generated, DO NOT EDIT.
//
// Source:
//     Title: CA Service
//     Version: 0.1.0
package models

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// AccessCredentials is an object.
type AccessCredentials struct {
	// ClientId: ID of the control service requesting authentication.
	ClientId string `json:"client_id"`
	// ClientSecret: Secret that authenticates the control service.
	ClientSecret string `json:"client_secret"`
}

// Validate implements basic validation for this model
func (m AccessCredentials) Validate() error {
	return validation.Errors{}.Filter()
}

// GetClientId returns the ClientId property
func (m AccessCredentials) GetClientId() string {
	return m.ClientId
}

// SetClientId sets the ClientId property
func (m *AccessCredentials) SetClientId(val string) {
	m.ClientId = val
}

// GetClientSecret returns the ClientSecret property
func (m AccessCredentials) GetClientSecret() string {
	return m.ClientSecret
}

// SetClientSecret sets the ClientSecret property
func (m *AccessCredentials) SetClientSecret(val string) {
	m.ClientSecret = val
}
