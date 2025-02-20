// Package encryption -- реализация шифрования тела HTTP запроса.
package encryption

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"hash"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// GenerateRSAKeyPair -- функция генерации пары ключей и сохранения их в файлы.
// Генерируются приватный и публичный ключи, записываются в соответствующие файлы.
// Возвращаются ссылки на объекты private и public ключей и ошибка.
func GenerateRSAKeyPair(privKeyFile, pubKeyFile string) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	// Генерация ключей RSA
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Println("rsa.GenerateKey error :", err)
		return nil, nil, fmt.Errorf("rsa.EncryptOAEP error: %v", err)
	}

	// Преобразование приватного ключа в PEM формат
	privatePEMBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	// Преобразование публичного ключа в PEM формат
	publicPEMBlock := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&key.PublicKey),
	}

	fullPath, err := filepath.Abs(privKeyFile)
	if err != nil {
		log.Println("GenerateRSAKeyPair filepath.Abs(privKeyFile) error :", err)
	}
	log.Println("START GenerateRSAKeyPair with privateKey file with fullPath", fullPath)

	err = CreateFile(privKeyFile, pem.EncodeToMemory(privatePEMBlock))
	if err != nil {
		log.Println("generateRSAKeyPair: create private key error", err)
		return nil, nil, fmt.Errorf("GenerateRSAKeyPair. CreateFile error: %v", err)
	}

	err = CreateFile(pubKeyFile, pem.EncodeToMemory(publicPEMBlock))
	if err != nil {
		log.Println("generateRSAKeyPair: create public key error", err)
		return nil, nil, fmt.Errorf("GenerateRSAKeyPair. CreateFile error: %v", err)
	}

	return key, &key.PublicKey, nil
}

// CreateFile создание файла.
func CreateFile(filename string, data []byte) error {
	err := os.WriteFile(filename, data, 0644)
	if err != nil {
		log.Println("createFile error:", err)
		return err
	}
	log.Println("createFile: file created", filename)
	return nil
}

// ReadFile чтение файла.
func ReadFile(filename string) ([]byte, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Println("ReadFile error :", err)
	}
	return data, err
}

// ReadPrivateKeyFile чтение RSA private key из файла.
func ReadPrivateKeyFile(filename string) (*rsa.PrivateKey, error) {
	fullPath, err := filepath.Abs(filename)
	if err != nil {
		log.Println("ReadPrivateKeyFile error :", err)
	}
	log.Println("START ReadPrivateKeyFile() with privateKey file with fullPath", fullPath)

	data, err := ReadFile(filename)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// ReadPublicKeyFile чтение RSA public key из файла.
func ReadPublicKeyFile(filename string) (*rsa.PublicKey, error) {
	fullPath, err := filepath.Abs(filename)
	if err != nil {
		log.Println("ReadPrivateKeyFile error :", err)
	}
	log.Println("START ReadPublicKeyFile() with privateKey file with fullPath", fullPath)

	data, err := ReadFile(filename)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}
	key, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// EncryptOAEP шифрование данных открытым ключом chunk-ами из-за большого для RSA шифрования размера body.
func EncryptOAEP(hash hash.Hash, random io.Reader, public *rsa.PublicKey, msg []byte, label []byte) ([]byte, error) {
	msgLen := len(msg)
	step := public.Size() - 2*hash.Size() - 2
	var encryptedBytes []byte

	for start := 0; start < msgLen; start += step {
		finish := start + step
		if finish > msgLen {
			finish = msgLen
		}

		encryptedBlockBytes, err := rsa.EncryptOAEP(hash, random, public, msg[start:finish], label)
		if err != nil {
			return nil, err
		}

		encryptedBytes = append(encryptedBytes, encryptedBlockBytes...)
	}

	return encryptedBytes, nil
}

// DecryptOAEP дешифрование данных открытым ключом chunk-ами из-за большого для RSA шифрования размера body.
func DecryptOAEP(hash hash.Hash, random io.Reader, private *rsa.PrivateKey, msg []byte, label []byte) ([]byte, error) {
	msgLen := len(msg)
	step := private.PublicKey.Size()
	var decryptedBytes []byte
	n := 0

	for start := 0; start < msgLen; start += step {
		finish := start + step
		if finish > msgLen {
			finish = msgLen
		}

		decryptedBlockBytes, err := rsa.DecryptOAEP(hash, random, private, msg[start:finish], label)
		if err != nil {
			return nil, err
		}
		n++
		decryptedBytes = append(decryptedBytes, decryptedBlockBytes...)
	}
	log.Println("DecryptOAEP chunks number is:", n)

	return decryptedBytes, nil
}

// EncryptData шифрование данных открытым ключом.
func EncryptData(data []byte, publicKey *rsa.PublicKey) ([]byte, error) {
	hash := sha512.New()
	encryptedMessage, err := EncryptOAEP(hash, rand.Reader, publicKey, data, nil)
	if err != nil {
		return nil, fmt.Errorf("rsa.EncryptOAEP error: %v", err)
	}
	return encryptedMessage, nil
}

// DecryptData расшифровка данных приватным ключом.
func DecryptData(data []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	hash := sha512.New()
	//encryptedMessage, err := rsa.DecryptOAEP(hash, rand.Reader, privateKey, data, nil)
	encryptedMessage, err := DecryptOAEP(hash, rand.Reader, privateKey, data, nil)
	if err != nil {
		return nil, fmt.Errorf("rsa.DecryptOAEP error: %v", err)
	}
	return encryptedMessage, nil
}

// Константы для определения шаблонов User-Agent из заголовков входящих пакетов, для которых не используется RSA шифрование.
const (
	Postman = "ostman" // for Postman
)

// DecryptRequestHandler хэндлер распаковки body входящего Request запроса.
func DecryptRequestHandler(_ context.Context, privateKey *rsa.PrivateKey) gin.HandlerFunc {
	return func(c *gin.Context) {
		userAgent := c.Request.UserAgent()
		//log.Println("User-Agent is", c.Request.Header.Get("User-Agent"))
		log.Println("User-Agent is", userAgent)
		if strings.Contains(userAgent, Postman) {
			log.Println("RSA encryption for this User-Agent =", userAgent, " type disabled")
			return
		}
		log.Println("DecryptRequestHandler START")
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Println("DecryptRequestHandler: ioutil.ReadAll body error", err)
			c.Status(http.StatusInternalServerError)
			return
		}

		decryptedBody, err := DecryptData(body, privateKey)
		if err != nil {
			log.Println("DecryptRequestHandler: DecryptData error", err)
			c.Set("DecryptRequestHandler", "error")
			c.Status(http.StatusInternalServerError)
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewReader(decryptedBody))
		c.Set("DecryptRequestHandler", "success")
		c.Next()
		log.Println("DecryptRequestHandler END")
	}
}
