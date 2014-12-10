package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/edo/util"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"time"
)

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

func pemKeyGen(before string) string {
	return before + ".pub"
}

// スレッドセーフ。
func NewFileTaKeyProvider(path string, expiDur time.Duration) TaKeyProvider {
	return newTaKeyProvider(driver.NewFileKeyValueStore(path, pemKeyGen, publicKeyMarshal, publicKeyUnmarshal, expiDur))
}
