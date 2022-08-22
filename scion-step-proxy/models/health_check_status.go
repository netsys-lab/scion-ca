// This file is auto-generated, DO NOT EDIT.
//
// Source:
//     Title: CA Service
//     Version: 0.1.0
package models

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// HealthCheckStatus is an object.
type HealthCheckStatus struct {
	// Status:
	// available
    // starting
    // stopping
    // unavailable
	Status string `json:"status"`
}

// Validate implements basic validation for this model
func (m HealthCheckStatus) Validate() error {
	return validation.Errors{}.Filter()
}

// GetStatus returns the Status property
func (m HealthCheckStatus) GetStatus() string {
	return m.Status
}

// SetStatus sets the Status property
func (m *HealthCheckStatus) SetStatus(val string) {
	m.Status = val
}
