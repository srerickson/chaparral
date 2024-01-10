package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var (
	dir    = "."
	doTest = false
)

func main() {
	flag.StringVar(&dir, "dir", ".", "directory for ouput pems")
	flag.BoolVar(&doTest, "text", false, "test genrated certs/keys with test server")
	flag.Parse()

	var (
		caCrtFile = filepath.Join(dir, "ca.crt")
		caKeyFile = filepath.Join(dir, "ca.key")
		clCrtFile = filepath.Join(dir, "client.crt")
		clKeyFile = filepath.Join(dir, "client.key")
		srCrtFile = filepath.Join(dir, "server.crt")
		srKeyFile = filepath.Join(dir, "server.key")
	)

	ca := rootCA("CA")
	caPubKey, caPrivKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		log.Fatal(err)
	}
	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, caPubKey, caPrivKey)
	if err != nil {
		log.Fatal(err)
	}
	if err := writeCertKey(caBytes, caPrivKey, caCrtFile, caKeyFile); err != nil {
		log.Fatal(err)
	}

	server := serverCertificate("localhost")
	serverPubKey, serverPrivKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		log.Fatal(err)
	}
	serverBytes, err := x509.CreateCertificate(rand.Reader, server, ca, serverPubKey, caPrivKey)
	if err != nil {
		log.Fatal(err)
	}
	if err := writeCertKey(serverBytes, serverPrivKey, srCrtFile, srKeyFile); err != nil {
		log.Fatal(err)
	}

	client := clientCertificate("client@testing.net")
	clientPubKey, clientPrivKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		log.Fatal(err)
	}
	clientBytes, err := x509.CreateCertificate(rand.Reader, client, ca, clientPubKey, caPrivKey)
	if err != nil {
		log.Fatal(err)
	}
	if err := writeCertKey(clientBytes, clientPrivKey, clCrtFile, clKeyFile); err != nil {
		log.Fatal(err)
	}

	if doTest {
		go startServer(srCrtFile, srKeyFile, caCrtFile)
		cli := httpClient(clCrtFile, clKeyFile, caCrtFile)
		resp, err := cli.Get("https://localhost:8080")
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			log.Fatal("not 200")
		}
		io.Copy(os.Stdout, resp.Body)
	}
}

func rootCA(cn string) *x509.Certificate {
	return &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   cn,
			Organization: []string{cn},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
	}
}

func serverCertificate(cn string) *x509.Certificate {
	return &x509.Certificate{
		SerialNumber:          big.NewInt(2),
		Subject:               pkix.Name{CommonName: cn},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		DNSNames:              []string{cn},
		IPAddresses:           []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		BasicConstraintsValid: true,
	}
}

func clientCertificate(email string) *x509.Certificate {
	return &x509.Certificate{
		SerialNumber:   big.NewInt(3),
		Subject:        pkix.Name{CommonName: email},
		NotBefore:      time.Now(),
		NotAfter:       time.Now().AddDate(1, 0, 0),
		EmailAddresses: []string{email},
		KeyUsage:       x509.KeyUsageDigitalSignature,
	}
}

func writeCertKey(caBytes []byte, key any, certName, keyName string) error {
	certWriter, err := os.Create(certName)
	if err != nil {
		return err
	}
	defer certWriter.Close()
	keyWriter, err := os.Create(keyName)
	if err != nil {
		return err
	}
	defer keyWriter.Close()
	if err := pem.Encode(certWriter, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	}); err != nil {
		return err
	}
	encPivKey, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return err
	}
	if err := pem.Encode(keyWriter, &pem.Block{
		Type:  "ED25519 PRIVATE KEY",
		Bytes: encPivKey,
	}); err != nil {
		return err
	}
	return nil
}

func startServer(serverCertFile, serverKeyFile, caCertFile string) error {
	srv := http.Server{
		Addr: ":8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, r.TLS.PeerCertificates[0].EmailAddresses[0])
		}),
	}
	pem, err := os.ReadFile(caCertFile)
	if err != nil {
		return err
	}
	clientCAs := x509.NewCertPool()
	clientCAs.AppendCertsFromPEM(pem)
	srv.TLSConfig = &tls.Config{
		ClientCAs:  clientCAs,
		ClientAuth: tls.RequireAndVerifyClientCert,
	}
	return srv.ListenAndServeTLS(serverCertFile, serverKeyFile)
}

func httpClient(clientCertFile, clientKeyFile, caCertFile string) *http.Client {
	var trans http.RoundTripper
	cert, err := tls.LoadX509KeyPair(clientCertFile, clientKeyFile)
	if err != nil {
		panic(err)
	}
	pem, err := os.ReadFile(caCertFile)
	if err != nil {
		panic(err)
	}
	rootCAs := x509.NewCertPool()
	rootCAs.AppendCertsFromPEM(pem)
	trans = &http.Transport{
		TLSClientConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      rootCAs,
		},
	}
	return &http.Client{
		Transport: trans,
	}
}
