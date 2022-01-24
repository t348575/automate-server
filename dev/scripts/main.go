package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/argon2"
)

func GenerateRandomBytes(size uint32) []byte {
	token := make([]byte, size)
	rand.Read(token)
	return token
}

func main() {
	salt := GenerateRandomBytes(16)
	
	hash := argon2.IDKey([]byte("test"), salt, 3, 64*1024, 2, 32)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	fmt.Println("$argon2id$v=19$m=65536,t=3,p=2$" + b64Salt + "$" + b64Hash)
}