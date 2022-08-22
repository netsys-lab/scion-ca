// This file is auto-generated, DO NOT EDIT.
//
// Source:
//     Title: CA Service
//     Version: 0.1.0
package models

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// AccessToken is an object.
type AccessToken struct {
	// AccessToken: The encoded JWT token
	AccessToken string `json:"access_token"`
	// ExpiresIn: Validity duration of this token in seconds.
	ExpiresIn int32 `json:"expires_in"`
	// TokenType: Type of returned access token. Currently always Bearer.
	TokenType string `json:"token_type"`
}

// Validate implements basic validation for this model
func (m AccessToken) Validate() error {
	return validation.Errors{}.Filter()
}

// GetAccessToken returns the AccessToken property
func (m AccessToken) GetAccessToken() string {
	return m.AccessToken
}

// SetAccessToken sets the AccessToken property
func (m *AccessToken) SetAccessToken(val string) {
	m.AccessToken = val
}

// GetExpiresIn returns the ExpiresIn property
func (m AccessToken) GetExpiresIn() int32 {
	return m.ExpiresIn
}

// SetExpiresIn sets the ExpiresIn property
func (m *AccessToken) SetExpiresIn(val int32) {
	m.ExpiresIn = val
}

// GetTokenType returns the TokenType property
func (m AccessToken) GetTokenType() string {
	return m.TokenType
}

// SetTokenType sets the TokenType property
func (m *AccessToken) SetTokenType(val string) {
	m.TokenType = val
}
