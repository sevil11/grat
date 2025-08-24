package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"strings"
)

// Sanitize cleans input strings
func Sanitize(s string) string {
	return strings.TrimSpace(s)
}

// GenerateRandomBytes creates random bytes of the specified length
func GenerateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	return bytes, err
}

// GenerateRandomString creates a random hex string
func GenerateRandomString(length int) (string, error) {
	bytes, err := GenerateRandomBytes(length / 2) // 2 hex chars per byte
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// EncryptAES encrypts plaintext with AES-256
func EncryptAES(plaintext []byte, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)
	
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptAES decrypts AES-256 encrypted ciphertext
func DecryptAES(ciphertext string, key []byte) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, err
	}
	
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	
	if len(data) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	
	iv := data[:aes.BlockSize]
	ciphertext = data[aes.BlockSize:]
	
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(data[aes.BlockSize:], data[aes.BlockSize:])
	
	return data[aes.BlockSize:], nil
}

// GetLocalIP returns the non-loopback local IP of the host
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}