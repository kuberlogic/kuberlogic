/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package compose

import (
	"bytes"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/plugin/commons"
	"testing"
)

func TestIsSecret(t *testing.T) {
	vd := newViewData(&commons.PluginRequest{})

	for _, v := range []struct {
		value  string
		result bool
	}{
		{"Not Secret", false},
		{"Secret", false},
		{"| Secret", false},
		{"| PersistentSecret", true},
		{"Chain | chain | PersistentSecret \"secretId\" | Chain", true},
	} {
		result := vd.isSecret(v.value)
		if result != v.result {
			t.Errorf("actual vs expected: %v vs %v", result, v.result)
		}
	}
}

func TestEndpoint(t *testing.T) {
	vd := newViewData(&commons.PluginRequest{
		TLSEnabled: true,
	})

	result, expected := vd.Endpoint("localhost"), "https://localhost"
	if result != expected {
		t.Errorf("actual vs expected: %v vs %v", result, expected)
	}

	vd = newViewData(&commons.PluginRequest{
		TLSEnabled: true,
		Host:       "example.com",
	})

	result, expected = vd.Endpoint("localhost"), "https://example.com"
	if result != expected {
		t.Errorf("actual vs expected: %v vs %v", result, expected)
	}

	vd = newViewData(&commons.PluginRequest{
		Host: "example.com",
	})

	result, expected = vd.Endpoint("localhost"), "http://example.com"
	if result != expected {
		t.Errorf("actual vs expected: %v vs %v", result, expected)
	}

	vd = newViewData(&commons.PluginRequest{})

	result, expected = vd.Endpoint("localhost"), "http://localhost"
	if result != expected {
		t.Errorf("actual vs expected: %v vs %v", result, expected)
	}
}

func TestGenerateKeyLen(t *testing.T) {
	vd := newViewData(&commons.PluginRequest{})
	result := vd.GenerateKey(10)
	if len(result) != 10 {
		t.Errorf("actual len vs expected len: %v vs %v", len(result), 10)
	}
}

func TestGenerateKeyLetters(t *testing.T) {
	letterBytes := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	vd := newViewData(&commons.PluginRequest{})

	for _, v := range []int{10, 1024, 512, 50} {
		str := vd.GenerateKey(v)
		for _, char := range []byte(str) {
			if bytes.IndexByte(letterBytes, char) < 0 {
				t.Errorf("letterBytes '%s' is not contain '%s'", letterBytes, string(char))
			}
		}
	}
}

func TestFromCache(t *testing.T) {
	vd := newViewData(&commons.PluginRequest{})
	repeated := "repeated-value"
	val := vd.FromCache("A", repeated)
	if val != repeated {
		t.Errorf("actual vs expected: %v vs %v", val, repeated)
	}
	val = vd.FromCache("A", "new-value")
	if val != repeated {
		t.Errorf("actual vs expected: %v vs %v", val, repeated)
	}
	val = vd.FromCache("B", "new-value")
	if val != "new-value" {
		t.Errorf("actual vs expected: %v vs %v", val, "new-value")
	}
}

func TestGenerateRSA(t *testing.T) {
	vd := newViewData(&commons.PluginRequest{})
	for _, bits := range []int{500, 1024, 2048, 4096} {
		result, err := vd.GenerateRSA(bits)
		if err != nil {
			t.Errorf("error generating RSA key: %v", err)
		}

		block, _ := pem.Decode([]byte(result))
		if block == nil {
			t.Errorf("block is nil for %v", result)
			return
		}
		pk, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			t.Errorf("parse rsa key error: %v", err)
			return
		}
		if err := pk.Validate(); err != nil {
			t.Errorf("pk validate error: %v", err)
		}
	}
}

func TestBase64(t *testing.T) {
	vd := newViewData(&commons.PluginRequest{})

	for _, s := range []string{
		"abc", "", "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
	} {
		result := vd.Base64(s)
		b, err := base64.StdEncoding.DecodeString(result)
		if err != nil {
			t.Errorf("decode base64 error: %v", err)
			return
		}

		if s != string(b) {
			t.Errorf("actual vs expected: %v vs %v", string(b), s)
		}
	}
}
