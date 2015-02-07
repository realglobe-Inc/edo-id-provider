package main

import (
	"encoding/base64"
	"github.com/realglobe-Inc/edo/util/secrand"
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
	// インスタンスごとの ID 被りを防ぐために与えられる文字列。
	suf string
}

func newIdGenerator(randLen int, suf string) idGenerator {
	return idGenerator{
		randLen: randLen,
		ser:     rand.New(rand.NewSource(time.Now().UnixNano())).Int63(),
		suf:     suf,
	}
}

func (this *idGenerator) newId() (id string, err error) {
	return this.id(this.randLen)
}

// 乱数部分の長さを指定して ID を発行させる。
// randLen が 0 なら err は必ず nil。
func (this *idGenerator) id(randLen int) (id string, err error) {
	id, err = secrand.String(randLen)
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

	return id + this.suf, nil
}
