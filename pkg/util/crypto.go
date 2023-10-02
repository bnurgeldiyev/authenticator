package util

import (
	"crypto"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (hash string, err error) {
	var hashByte []byte
	hashByte, err = bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		hash = ""
		return
	}
	hash = string(hashByte)
	return
}

func VerifyPasswordFromHash(password string, hash string) error {
	bHash := []byte(hash)
	return bcrypt.CompareHashAndPassword(bHash, []byte(password))
}

func GenerateHmacSha256(key, message string) (signature string, err error) {
	mac := hmac.New(sha256.New, []byte(key))
	_, err = mac.Write([]byte(message))
	if err != nil {
		return "", err
	}
	msgHash := mac.Sum(nil)
	signature = base64.StdEncoding.EncodeToString(msgHash)
	return signature, nil
}

func VerifyHmacSha256(key, message, msgHash string) error {
	var expMsgHashByte []byte
	msgHashByte, err := base64.StdEncoding.DecodeString(msgHash)
	if err != nil {
		return errors.Wrap(err, "could not decode msgHash")
	}
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(message))
	expMsgHashByte = mac.Sum(nil)
	if hmac.Equal(msgHashByte, expMsgHashByte) {
		return nil
	}
	return errors.New("invalid hmac")
}

func GenerateRsaKeyPair(bits int) (*rsa.PrivateKey, *rsa.PublicKey) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, bits)
	return privateKey, &privateKey.PublicKey
}

func ParseRsaPrivateKeyFromPemStr(privPEM string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privPEM))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return priv, nil
}

func ExportRsaPublicKeyAsPemStr(publicKey *rsa.PublicKey) (string, error) {
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", err
	}
	pubKeyPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: pubKeyBytes,
		},
	)

	return string(pubKeyPem), nil
}

func ExportRsaPrivateKeyAsPemStr(privateKey *rsa.PrivateKey) (string, error) {
	privKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privKeyPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privKeyBytes,
		},
	)
	return string(privKeyPem), nil
}

func ParseRsaPublicKeyFromPemStr(pubPEM string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pubPEM))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	switch pub := pub.(type) {
	case *rsa.PublicKey:
		return pub, nil
	default:
		break // fall through
	}
	return nil, errors.New("Key type is not RSA")
}

func RsaSha256Verify(r *rsa.PublicKey, message string, signature string) error {
	signByte, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return errors.Wrap(err, "could not decode signature")
	}
	h := sha256.New()
	h.Write([]byte(message))
	d := h.Sum(nil)
	return rsa.VerifyPKCS1v15(r, crypto.SHA256, d, signByte)
}

func RsaSha256Sign(r *rsa.PrivateKey, message string) (signature string, err error) {
	h := sha256.New()
	h.Write([]byte(message))
	d := h.Sum(nil)
	signByte, err := rsa.SignPKCS1v15(rand.Reader, r, crypto.SHA256, d)
	if err != nil {
		return "", errors.Wrap(err, "could not sign message")
	}
	signature = base64.StdEncoding.EncodeToString(signByte)
	return
}

func RsaSha1Sign(r *rsa.PrivateKey, message string) (signature string, err error) {
	h := sha1.New()
	h.Write([]byte(message))
	d := h.Sum(nil)
	var signByte []byte
	signByte, err = rsa.SignPKCS1v15(rand.Reader, r, crypto.SHA1, d)
	if err != nil {
		return "", errors.Wrap(err, "could not sign message")
	}
	signature = base64.StdEncoding.EncodeToString(signByte)
	return
}

func RsaDecrypt(r *rsa.PrivateKey, message string) (result []byte, err error) {
	msgByte, err := base64.StdEncoding.DecodeString(message)
	if err != nil {
		result = nil
		err = errors.Wrap(err, "could not decode message")
		return
	}
	result, err = rsa.DecryptPKCS1v15(rand.Reader, r, msgByte)
	if err != nil {
		result = nil
		err = errors.Wrap(err, "could not decrypt message")
		return
	}
	return
}

func RsaEncrypt(r *rsa.PublicKey, message []byte) (result string, err error) {
	cipherText, err := rsa.EncryptPKCS1v15(rand.Reader, r, message)
	if err != nil {
		result = ""
		err = errors.Wrap(err, "could not encrypt message")
		return
	}
	result = base64.StdEncoding.EncodeToString(cipherText)
	return
}

func GenerateHmacSha1(key string, message string) (signature string, err error) {
	keyByte, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		signature = ""
		err = errors.Wrap(err, "could not decode key")
		return
	}
	mac := hmac.New(sha1.New, keyByte)
	mac.Write([]byte(message))
	msgHash := mac.Sum(nil)
	signature = base64.StdEncoding.EncodeToString(msgHash)
	return
}

func VerifyHmacSha1(key string, message string, msgHash string) error {
	keyByte, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return errors.Wrap(err, "could not decode key")
	}
	msgHashByte, err := base64.StdEncoding.DecodeString(msgHash)
	if err != nil {
		return errors.Wrap(err, "could not decode msgHash")
	}

	mac := hmac.New(sha1.New, keyByte)
	mac.Write([]byte(message))
	expectedMsgHash := mac.Sum(nil)
	if hmac.Equal(expectedMsgHash, msgHashByte) {
		return nil
	} else {
		return errors.New("invalid hmac")
	}
}
