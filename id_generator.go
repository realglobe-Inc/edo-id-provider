package main

import (
	"encoding/base64"
	"github.com/realglobe-Inc/edo/util"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"math/big"
	"math/rand"
	"sync/atomic"
	"time"
)

type idGenerator struct {
	// 文字数の下界。
	minIdLen int
	// インスタンス内での ID 被りを防ぐための通し番号。
	// 別インスタンスでは保証できない。
	ser int64
}

func newIdGenerator(minIdLen int) idGenerator {
	return idGenerator{
		minIdLen: minIdLen,
		ser:      rand.New(rand.NewSource(time.Now().UnixNano())).Int63(),
	}
}

func (this *idGenerator) newId() (id string, err error) {
	id, err = util.SecureRandomString(this.minIdLen)
	if err != nil {
		return "", erro.Wrap(err)
	}
	buff := big.NewInt(atomic.AddInt64(&this.ser, 1)).Bytes()
	id += base64.URLEncoding.EncodeToString(buff)[:(len(buff)*8+5)/6]
	return id, nil
}
