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

// Problem is an object. Error message encoded as specified in
// [RFC7807](https://tools.ietf.org/html/rfc7807)
type Problem struct {
	// CorrelationId: Identifier to correlate multiple error messages to the same case.
	CorrelationId string `json:"correlation_id,omitempty"`
	// Detail: A human readable explanation specific to this occurrence of the problem that is helpful to locate the problem and give advice on how to proceed. Written in English and readable for engineers, usually not suited for non technical stakeholders and not localized.
	Detail string `json:"detail,omitempty"`
	// Instance: A URI reference that identifies the specific occurrence of the problem, e.g. by adding a fragment identifier or sub-path to the problem type.
	Instance string `json:"instance,omitempty"`
	// Status: The HTTP status code generated by the server for this occurrence of the problem.
	Status int32 `json:"status"`
	// Title: A short summary of the problem type. Written in English and readable for engineers, usually not suited for non technical stakeholders and not localized.
	Title string `json:"title"`
	// Type: A URI reference that uniquely identifies the problem type in the context of the provided API.
	Type string `json:"type"`
}

// Validate implements basic validation for this model
func (m Problem) Validate() error {
	return validation.Errors{
		"correlationId": validation.Validate(
			m.CorrelationId, is.UUID,
		),
		"status": validation.Validate(
			m.Status, validation.Required, validation.Min(int32(100)), validation.Max(int32(599)),
		),
	}.Filter()
}

// GetCorrelationId returns the CorrelationId property
func (m Problem) GetCorrelationId() string {
	return m.CorrelationId
}

// SetCorrelationId sets the CorrelationId property
func (m *Problem) SetCorrelationId(val string) {
	m.CorrelationId = val
}

// GetDetail returns the Detail property
func (m Problem) GetDetail() string {
	return m.Detail
}

// SetDetail sets the Detail property
func (m *Problem) SetDetail(val string) {
	m.Detail = val
}

// GetInstance returns the Instance property
func (m Problem) GetInstance() string {
	return m.Instance
}

// SetInstance sets the Instance property
func (m *Problem) SetInstance(val string) {
	m.Instance = val
}

// GetStatus returns the Status property
func (m Problem) GetStatus() int32 {
	return m.Status
}

// SetStatus sets the Status property
func (m *Problem) SetStatus(val int32) {
	m.Status = val
}

// GetTitle returns the Title property
func (m Problem) GetTitle() string {
	return m.Title
}

// SetTitle sets the Title property
func (m *Problem) SetTitle(val string) {
	m.Title = val
}

// GetType returns the Type property
func (m Problem) GetType() string {
	return m.Type
}

// SetType sets the Type property
func (m *Problem) SetType(val string) {
	m.Type = val
}
