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
	// 乱数文字数の長さ。
	randLen int
	// インスタンス内での ID 被りを防ぐための通し番号。
	ser int64
}

func newIdGenerator(randLen int) idGenerator {
	return idGenerator{
		randLen: randLen,
		ser:     rand.New(rand.NewSource(time.Now().UnixNano())).Int63(),
	}
}

func (this *idGenerator) newId() (id string, err error) {
	id, err = util.SecureRandomString(this.randLen)
	if err != nil {
		return "", erro.Wrap(err)
	}

	const bLen = 64 / 8       // int64
	const sLen = bLen * 8 / 6 // BASE64 にして文字を使い切れない上位ビットは捨てる。

	v := big.NewInt(atomic.AddInt64(&this.ser, 1))
	v = v.Lsh(v, bLen*8-sLen*6) // BASE64 の 6 ビット区切りと最下位ビットの位置を揃える。
	buff := v.Bytes()
	if len(buff) < bLen {
		// 上位を 0 詰め。
		buff = append(make([]byte, bLen-len(buff)), buff...)
	} else if len(buff) > bLen {
		// 上位を捨てる。
		buff = buff[len(buff)-bLen:]
	}
	id += base64.URLEncoding.EncodeToString(buff)[:sLen]

	return id, nil
}
