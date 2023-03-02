package database

import (
	"encoding/json"
	"os"

	caconfig "github.com/scionproto/scion/private/ca/config"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Seeds struct {
	Users []User
}

// Load seeds from a json file and seed the database if it is empty
// Used for easy bootstrapping
func RunSeeds(db *gorm.DB, seedFilePath string) error {

	var users []User
	result := db.Find(&users)
	if result.Error != nil {
		return result.Error
	}

	if len(users) > 0 {
		logrus.Info("Users already exist, skipping all seeds")
		return nil
	}

	seedFile, err := os.Open(seedFilePath)
	if err != nil {
		return err
	}
	defer seedFile.Close()
	var data Seeds
	if err := json.NewDecoder(seedFile).Decode(&data); err != nil {
		return err
	}

	for _, u := range data.Users {
		secretKey := caconfig.NewPEMSymmetricKey(u.Clientsecret)
		secretValue, err := secretKey.Get()
		u.Clientsecret = string(secretValue)
		if res := db.Create(&u); res.Error != nil {
			return err
		}
	}

	return nil
}
