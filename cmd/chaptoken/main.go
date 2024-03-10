package main

// command to generate web token for auth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/srerickson/chaparral/server"
)

var (
	keyFile = flag.String("key", "key.pem", "rsa key pem")
	name    = flag.String("name", "", "user name")
	id      = flag.String("id", "", "user id")
	email   = flag.String("email", "", "uesr email")
	roles   = flag.String("roles", "", "roles separated by commas (chaparral:admin)")
	exp     = flag.Int("exp", 7, "days until the token expires")
)

func main() {
	flag.Parse()
	key, err := readKey(*keyFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "reading key %q: %v", *keyFile, err)
		os.Exit(1)
	}
	roles := strings.Split(*roles, ",")
	token, err := genToken(key, *id, *email, *name, *exp, roles...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "generating token: %v", err)
		os.Exit(1)
	}
	fmt.Println(token)
}

func genToken(key *rsa.PrivateKey, id, email, name string, exp int, roles ...string) (string, error) {
	user := server.AuthUser{
		ID:    id,
		Email: email,
		Name:  name,
		Roles: roles,
	}
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: key}, (&jose.SignerOptions{}).WithType("JWT"))
	if err != nil {
		return "", err
	}
	token := server.AuthToken{
		User: user,
		Claims: jwt.Claims{
			Subject:   email,
			Issuer:    "chaparral",
			Audience:  jwt.Audience{"chaparral"},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now().AddDate(0, 0, -1)),
			Expiry:    jwt.NewNumericDate(time.Now().AddDate(0, 0, exp)),
		},
	}
	return jwt.Signed(signer).Claims(token).Serialize()
}

func readKey(keyfile string) (*rsa.PrivateKey, error) {
	byts, err := os.ReadFile(keyfile)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(byts)
	if block == nil {
		return nil, errors.New("not PEM encoded key")
	}
	anyKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	}
	switch k := anyKey.(type) {
	case *rsa.PrivateKey:
		return k, nil
	default:
		return nil, errors.New("not an rsa key")
	}
}
