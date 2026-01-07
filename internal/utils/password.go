package utils

import "github.com/alexedwards/argon2id"

//Hash Password - Mã hóa Password
func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}

	return hash,nil
}

//Verify Password - Kiểm tra Password
func VerifyPassword(password string, hash string) (bool, error){
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}

	return match, nil
}	
