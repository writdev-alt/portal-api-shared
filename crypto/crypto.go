package crypto

import (
	"golang.org/x/crypto/bcrypt"

	"github.com/writdev-alt/portal-api-shared/logger"
)

func HashAndSalt(plainPassword []byte) string {
	hash, err := bcrypt.GenerateFromPassword(plainPassword, bcrypt.MinCost)
	if err != nil {
		logger.Errorf("Failed to HashAndSalt: %v", err)
	}
	return string(hash)
}

func ComparePassword(hashedPassword string, plainPassword []byte) bool {
	byteHash := []byte(hashedPassword)
	err := bcrypt.CompareHashAndPassword(byteHash, plainPassword)
	return err == nil
}
