package cfg

import (
	"crypto/x509"
	"encoding/pem"
	"github.com/pkg/errors"
	"os"
)

func openCertificate(file string) (*x509.Certificate, error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return nil, errors.Wrap(err, "cannot read the file")
	}

	block, _ := pem.Decode(content)
	if block == nil {
		return nil, errors.New("certitifate is not found in file")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse certificate")
	}
	return cert, nil
}

func checkCertificates(tlsConf *TLS) error {
	if tlsConf != nil {
		contentCa, err := os.ReadFile(tlsConf.CaFile)
		if err != nil {
			return errors.Wrap(err, "cannot read the CA file")
		}
		_, err = os.ReadFile(tlsConf.KeyFile)
		if err != nil {
			return errors.Wrap(err, "cannot read the secret file")
		}

		leaf, err := openCertificate(tlsConf.CrtFile)
		if err != nil {
			return errors.Wrap(err, "incorrect certificate")
		}

		roots := x509.NewCertPool()
		ok := roots.AppendCertsFromPEM(contentCa)
		if !ok {
			return errors.New("failed to parse root certificate")
		}

		if _, err := leaf.Verify(x509.VerifyOptions{
			Roots: roots,
		}); err == nil {
			return err
		}
		return nil

	}
	return nil
}
