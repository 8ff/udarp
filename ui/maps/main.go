package main

import (
	"crypto/ecdsa"
	"crypto/x509"
	"embed"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

//go:embed www
var wwwContent embed.FS

func ParsePKCS8PrivateKeyFromPEM(key []byte) (*ecdsa.PrivateKey, error) {
	var err error

	// Parse PEM block
	var block *pem.Block
	if block, _ = pem.Decode(key); block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the key")
	}

	// Parse the key
	var parsedKey interface{}
	if parsedKey, err = x509.ParsePKCS8PrivateKey(block.Bytes); err != nil {
		return nil, err
	}

	var pkey *ecdsa.PrivateKey
	var ok bool
	if pkey, ok = parsedKey.(*ecdsa.PrivateKey); !ok {
		return nil, fmt.Errorf("key type is not ECDSA")
	}

	return pkey, nil
}

func loadKey() (*ecdsa.PrivateKey, error) {
	// Open file
	file, err := os.Open("./AuthKey_DLTPY97T3X.p8")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read file
	key, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	pkey, err := ParsePKCS8PrivateKeyFromPEM(key)
	if err != nil {
		return nil, err
	}

	return pkey, nil
}

// Rewrite above javascript code in go
func generateJWT() (string, error) {
	// Load key
	key, err := loadKey()
	if err != nil {
		return "", err
	}

	// TeamID 4B58947B7M
	// Key ID DLTPY97T3X

	// Create token json web token and use privateKey to sign it
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"iss": "4B58947B7M",
		"iat": fmt.Sprintf("%d", time.Now().Unix()),          // Current time in seconds since epoch
		"exp": fmt.Sprintf("%d", time.Now().Unix()+15778800), // Current time in seconds since epoch + 15778800
	})

	token.Header["kid"] = "DLTPY97T3X"

	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// Serve above json web token on /services/jwt
func serveJWT() {
	// Generate token
	token, err := generateJWT()
	if err != nil {
		panic(err)
	}

	// Serve token
	http.HandleFunc("/services/jwt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Return following data, token: token
		w.Write([]byte(fmt.Sprintf(`{"token": "%s"}`, token)))
	})
}

// Server index.html on /
func serveIndex() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		// Return index.html from wwwContent
		index, err := wwwContent.ReadFile("www/index.html")
		if err != nil {
			panic(err)
		}
		w.Write(index)
	})
}

// Start web server
func main() {
	serveJWT()
	serveIndex()
	http.ListenAndServe(":8080", nil)
}
