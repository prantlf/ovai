package auth

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"time"

	"github.com/prantlf/ovai/internal/log"
)

type account struct {
	ProjectId    string `json:"project_id"`
	PrivateKeyId string `json:"private_key_id"`
	PrivateKey   string `json:"private_key"`
	ClientEmail  string `json:"client_email"`
	Scope        string `json:"scope,omitempty"`
	AuthUri      string `json:"auth_uri"`
}

type header struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
	KID string `json:"kid,omitempty"`
}

type payload struct {
	Iat   int64  `json:"iat"`
	Exp   int64  `json:"exp"`
	Scope string `json:"scope,omitempty"`
	Aud   string `json:"aud"`
	Iss   string `json:"iss"`
}

var privateKey = decodeKey()

func encodePart[T any](part *T) (string, error) {
	bytes, err := json.Marshal(part)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func encodeToken(head *header, pay *payload) (string, error) {
	headEnc, err := encodePart(head)
	if err != nil {
		log.Dbg("encoding token header failed: %v", err)
		return "", errors.New("encoding token header failed")
	}
	payEnc, err := encodePart(pay)
	if err != nil {
		log.Dbg("encoding token payload failed: %v", err)
		return "", errors.New("encoding token payload failed")
	}
	tokenEnc := headEnc + "." + payEnc
	hasher := sha256.New()
	if _, err := hasher.Write([]byte(tokenEnc)); err != nil {
		log.Dbg("hashing token failed: %v", err)
		return "", errors.New("hashing token failed")
	}
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hasher.Sum(nil))
	if err != nil {
		log.Dbg("signing token failed: %v", err)
		return "", errors.New("signing token failed")
	}
	signatureEnc := base64.RawURLEncoding.EncodeToString(signature)
	return tokenEnc + "." + signatureEnc, nil
}

func decodeKey() *rsa.PrivateKey {
	keyDec, _ := pem.Decode([]byte(Account.PrivateKey))
	if keyDec == nil {
		log.Ftl("decoding private key failed")
	}
	keyBytes := keyDec.Bytes
	keyParsed, err := x509.ParsePKCS8PrivateKey(keyBytes)
	if err != nil {
		if keyParsed, err = x509.ParsePKCS1PrivateKey(keyBytes); err != nil {
			log.Ftl("parsing private key failed: %v", err)
		}
	}
	keyRSA, ok := keyParsed.(*rsa.PrivateKey)
	if !ok {
		log.Ftl("invalid private key")
	}
	log.Dbg("%d characters of private key parsed to %d bytes",
		len(Account.PrivateKey), keyRSA.Size())
	return keyRSA
}

func createToken() (string, error) {
	head := &header{
		Alg: "RS256",
		Typ: "JWT",
		KID: Account.PrivateKeyId,
	}
	iat := time.Now().Add(-10 * time.Second)
	exp := iat.Add(time.Hour)
	scope := Account.Scope
	if len(scope) == 0 {
		scope = "https://www.googleapis.com/auth/cloud-platform"
	}
	authUri := Account.AuthUri
	if len(authUri) == 0 {
		authUri = "https://www.googleapis.com/oauth2/v4/token"
	}
	pay := &payload{
		Iat:   iat.Unix(),
		Exp:   exp.Unix(),
		Scope: scope,
		Aud:   authUri,
		Iss:   Account.ClientEmail,
	}
	if log.IsDbg {
		headJson, errHead := json.MarshalIndent(head, " ", "  ")
		payJson, errPay := json.MarshalIndent(pay, " ", "  ")
		if errHead != nil || errPay != nil {
			log.Dbg("> create token with header %+v\n and payload %+v", head, pay)
		} else {
			log.Dbg("> create token with header %s\n and payload %s", headJson, payJson)
		}
	}
	return encodeToken(head, pay)
}
