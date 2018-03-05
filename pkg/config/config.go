package config

import (
	"crypto/tls"
	"encoding/base64"

	"github.com/pkg/errors"
	validator "gopkg.in/validator.v2"
)

// SyslogConfig syslog server configuration
type SyslogConfig struct {
	Debug        bool
	Proxy        bool
	Port         int `validate:"nonzero"`
	Region       string
	Profile      string
	Group        string `validate:"nonzero"`
	Stream       string `validate:"nonzero"`
	ClientCaCert string `validate:"nonzero"`
	Cert         string `validate:"nonzero"`
	Key          string `validate:"nonzero"`
}

// Validate validate the configuration
func (sc *SyslogConfig) Validate() error {
	return validator.Validate(sc)
}

// Certificate decode and return the certificate
func (sc *SyslogConfig) Certificate() (tls.Certificate, error) {
	certPEMBlock, err := base64.StdEncoding.DecodeString(sc.Cert)
	if err != nil {
		return tls.Certificate{}, errors.Wrap(err, "failed to decode certificate data from config")
	}

	keyPEMBlock, err := base64.StdEncoding.DecodeString(sc.Key)
	if err != nil {
		return tls.Certificate{}, errors.Wrap(err, "failed to decode key data from config")
	}

	return tls.X509KeyPair(certPEMBlock, keyPEMBlock)
}

// ClientCaCertificate decode and return the ca certificate
func (sc *SyslogConfig) ClientCaCertificate() ([]byte, error) {
	certPEMBlock, err := base64.StdEncoding.DecodeString(sc.ClientCaCert)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode ca certificate data from config")
	}

	return certPEMBlock, nil
}
