package main

import (
	"crypto/x509"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

func readHttp(url string) (*[]byte, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)
	body, err := ioutil.ReadAll(response.Body)
	return &body, nil
}

func GetCertPool(pemPath string) (*x509.CertPool, error) {
	var data []byte
	//goland:noinspection HttpUrlsUsage
	if strings.HasPrefix(pemPath, "http://") || strings.HasPrefix(pemPath, "https://") {
		received, err := readHttp(pemPath)
		if err != nil {
			return nil, err
		}
		data = *received
	} else {
		received, err := ioutil.ReadFile(pemPath)
		if err != nil {
			return nil, err
		}
		data = received
	}
	certs := x509.NewCertPool()
	if !certs.AppendCertsFromPEM(data) {
		return nil, errors.New("PEM Parse Failed")
	}
	return certs, nil
}
