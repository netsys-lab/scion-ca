package database

import (
	"gorm.io/gorm"
)

// User represents a node in the network identified by its public key
type User struct {
	gorm.Model
	Name         string
	Clientid     string
	Clientsecret string
}
