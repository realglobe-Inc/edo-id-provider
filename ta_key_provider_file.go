package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/edo/util"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"net/url"
	"strings"
	"time"
)

func keyToEscapedPubPath(key string) string {
	return url.QueryEscape(key) + ".pub"
}

func escapedPubPathToKey(path string) string {
	if !strings.HasSuffix(path, ".pub") {
		return ""
	}
	key, _ := url.QueryUnescape(path[:len(path)-len(".pub")])
	return key
}

func publicKeyToPem(pubKey interface{}) (pemStr interface{}, err error) {
	block := &pem.Block{
		Type: "PUBLIC KEY",
	}
	block.Bytes, err = x509.MarshalPKIXPublicKey(pubKey.(*rsa.PublicKey))
	if err != nil {
		return nil, erro.Wrap(err)
	}

	return pem.EncodeToMemory(block), nil
}

func publicKeyMarshal(value interface{}) (data []byte, err error) {
	pemStr, err := publicKeyToPem(value)
	if err != nil {
		return nil, erro.Wrap(err)
	}
	return pemStr.([]byte), nil
}

// data を PEM 形式の文字列として、rsa.PublicKey にデコードする。
func publicKeyUnmarshal(data []byte) (interface{}, error) {
	pubKey, err := util.ParseRsaPublicKey(string(data))
	if err != nil {
		return nil, erro.Wrap(err)
	}
	return pubKey, nil
}

// スレッドセーフ。
func NewFileTaKeyProvider(path string, staleDur, expiDur time.Duration) TaKeyProvider {
	return newTaKeyProvider(driver.NewFileListedKeyValueStore(path, keyToEscapedPubPath, escapedPubPathToKey, publicKeyMarshal, publicKeyUnmarshal, staleDur, expiDur))
}
