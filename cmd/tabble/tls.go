package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

// lanHosts lists the addresses other screens on this network can reach this machine at,
// falling back to localhost when there are none.
func lanHosts() []string {
	var out []string
	addrs, _ := net.InterfaceAddrs()
	for _, a := range addrs {
		if ipn, ok := a.(*net.IPNet); ok && !ipn.IP.IsLoopback() && ipn.IP.To4() != nil {
			out = append(out, ipn.IP.String())
		}
	}
	if len(out) == 0 {
		out = []string{"localhost"}
	}
	return out
}

func certDir(flagged string) string {
	if flagged != "" {
		return flagged
	}
	if d, err := os.UserCacheDir(); err == nil {
		return filepath.Join(d, "tabble")
	}
	return "."
}

// certificate returns this machine's self-made certificate, making a new one when there is none,
// when it is within a week of expiring, or when it does not name one of today's addresses.
// It is kept on disk so a phone that accepted it once is not asked again on every start.
func certificate(dir string, hosts []string) (tls.Certificate, error) {
	certFile, keyFile := filepath.Join(dir, "tabble-cert.pem"), filepath.Join(dir, "tabble-key.pem")
	if cert, err := tls.LoadX509KeyPair(certFile, keyFile); err == nil {
		if leaf, err := x509.ParseCertificate(cert.Certificate[0]); err == nil && time.Until(leaf.NotAfter) > 7*24*time.Hour && names(leaf, hosts) {
			return cert, nil
		}
	}
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 120))
	if err != nil {
		return tls.Certificate{}, err
	}
	tmpl := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: "Tabble on this network"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
		IPAddresses:           []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
	}
	if name, err := os.Hostname(); err == nil && name != "" {
		tmpl.DNSNames = append(tmpl.DNSNames, name, name+".local")
	}
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			tmpl.IPAddresses = append(tmpl.IPAddresses, ip)
		}
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, err
	}
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return tls.Certificate{}, err
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	if err := os.MkdirAll(dir, 0o700); err == nil { // keeping it is a convenience; serving does not depend on it
		os.WriteFile(certFile, certPEM, 0o644)
		os.WriteFile(keyFile, keyPEM, 0o600)
	}
	return tls.X509KeyPair(certPEM, keyPEM)
}

// names reports whether the certificate covers every address in hosts.
func names(leaf *x509.Certificate, hosts []string) bool {
	for _, h := range hosts {
		if leaf.VerifyHostname(h) != nil {
			return false
		}
	}
	return true
}
