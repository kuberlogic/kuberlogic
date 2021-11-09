package cfg

import (
	"crypto/x509"
	"github.com/pkg/errors"
	"os"
)

func open(caFile, crtFile, keyFile string) error{
	if caFile != "" && crtFile != "" &&  keyFile != "" {
		ca, err := os.ReadFile(caFile)
		if err != nil {
			return errors.Wrap(err, "cannot read the ca file")
		}

		leafCrt, err := os.ReadFile(crtFile)
		if err != nil {
			return errors.Wrap(err, "cannot read the certificate file")
		}

		caCert, err := x509.ParseCertificate(ca)
		leaf, err := x509.ParseCertificate(leafCrt)

		intCerts := []*x509.Certificate{caCert}
		caCheck(leaf, intCerts)

	}
	return nil
}

func caCheck(leafCert *x509.Certificate, intCerts []*x509.Certificate) bool {
	caPool := x509.NewCertPool()
	for _, intCert := range intCerts {
		caPool.AddCert(intCert)
	}

	if _, err := leafCert.Verify(x509.VerifyOptions{
		Roots: caPool,
	}); err == nil {
		return true
	}
	return false
}