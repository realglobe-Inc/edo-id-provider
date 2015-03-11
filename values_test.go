package main

import (
	"crypto"
	"crypto/rsa"
	"github.com/realglobe-Inc/edo-lib/jwt"
	"time"
)

const (
	filePerm = 0644
)

const (
	testLabel = "edo-test"

	testIdLen = 5
	testUiUri = "/html"

	// 動作に影響は出ないはず。
	testSavDur    = 15 * time.Millisecond
	testStaleDur  = 10 * time.Millisecond
	testCaExpiDur = 5 * time.Millisecond

	// 動作に影響あり。
	// Go の GC が 10ms くらいは時間を使うと言っているので、それ以上に。
	testCodExpiDur   = 20 * time.Millisecond
	testTokExpiDur   = 20 * time.Millisecond
	testIdTokExpiDur = 20 * time.Millisecond
	testSessExpiDur  = 20 * time.Millisecond
	testTicDur       = 20 * time.Millisecond

	testSigAlg = "RS256"
)

var testIdpPriKey crypto.PrivateKey
var testIdpPubKey crypto.PublicKey

func init() {
	priKey, err := jwt.KeyFromJwkMap(map[string]interface{}{
		"d":   "Up0OHC99j80MEvW8BuP89zTwMXckk6qklvZr7-VAe9hHteE-DIGkV9fWIzaC1v5U5uNoSAa22GvsqbrvZs1-6yN1fyMjcmKBSFwIesUYwQY7Ya9z00sXNwQWUYg5yooxQht1yzHN8OSpiVGBWzRbW-GSusJWiXlqtax3-n5UhmE",
		"dp":  "Ug-DzGk80iuaFG3B50F1UcwKPx3IcjppfTJLIbZ8hVzj7NHtLgjZRz6LmRzpQjzukjMf0wPL3Ny-29pdalbmJQ",
		"dq":  "NpDXLWU4lou6m3vPvhpDu9_6M0-fgO6usD67DDIvr5RIOLs5VPrxTmSO7Oc1dk38y-uqyBCTV4e1KLclraWeEQ",
		"e":   "AQAB",
		"kty": "RSA",
		"n":   "mXyTgXHmo1Jnw7DUlZY7STfEiXKg-NTXyN3ccROppYHouEiLCGivWKYW8MlF-muf3Vdvzv3djP4R92tDaSvpZaLsc7hdcqREnBkXe1-RE-O-_y0ht5cJqIE-UXxThL6K53-VxMTpYYPeFhvk3JPyAQmI_zhLpN4hKTKnnPNX5qM",
		"p":   "yMbYrkuqtKK-MG868IfBpDmwF-58qiVim5bzNmVniqaTY4v0wqSRpzFI_q-brd3yHVZ5W-r2nv_6bJ69FkV3cw",
		"q":   "w7Psau2Mh0hSB67mmWD1iWO1WJvwmNlQVLHblnCb0CRsK13LzXTqUU-GFGZqlEymEfeMRpmfXL7EK5pES2soEQ",
		"qi":  "UcZv3PMXGV_BLyJvrxbVGRkSl_b3GqeQ5xkiwxAs3zM0pTyfM8m2dBcDwy73lKsaN_c-LdeEcK8qT446iuYycQ",
	})
	if err != nil {
		panic(err)
	}
	testIdpPriKey = priKey.(*rsa.PrivateKey)
	testIdpPubKey = &priKey.(*rsa.PrivateKey).PublicKey
}

const testTaKid = "testkey"

var testTaPriKey crypto.PrivateKey
var testTaPubKey crypto.PublicKey
var testTa *ta

func init() {
	priKey, err := jwt.KeyFromJwkMap(map[string]interface{}{
		"d":   "tSO9OVXVbtGsU9tlEl4hxp__E2U78Rvr8fdWbYySKlvI9Zbdiqdg9ipQhDeQ_AXgfEU_d10r8_67IH0opn_YFqpkWc7CEazoEVkP_goo--1MhkPV9luEeIHQCPD5WBpsyjOvA7HHahxAv6iuhcegirex0gRZsfc_yR7Py5VtAIE",
		"dp":  "O4lyRKCpjOOqUgHKyXE1aol0xXX0DZAbQAEZ-f5BRIfRJtzlELAP8KauCoHFGv1Rn2gh9Bvzk9F-1bjk_xxEuQ",
		"dq":  "NnF5QnRlD_9Is9RcAMh3LHu55C4itPQZKtIYXqbSu-FOM5p1gptlL-bEtBRCE-1uHJlehQLh__JCUyxWFB48UQ",
		"e":   "AQAB",
		"kty": "RSA",
		"n":   "yBfun8cJp8GIlAST7FINnM6-bqCXA96xf4ZAejYmqmwETbg4yno0n3EeXcxSMFkdLH-myY2BQi5QBWu34zaFh7G4T9Lsecr9m3xqUSx2rV3dJP3l9Qa7d186kkTFkd7-wS4pDmFlsaglU2hvnEEAoF4kHLpujo-zeITYSktFmsk",
		"p":   "7JPLgbLUJOYFFgbvXUUhsZkEn8rAvc1hdnOq7UTT0mIHxXvYyR37cptKSjjVpykAndot2t7bMcOJloqG8jCUmQ",
		"q":   "2IVY38ptluVmSKBxLJO0k8BIFKF72IXR2DAlS6CDR_FeDAjHbD7xvllF56AHFjkLexM77JJyqR4AjOl6cBLlsQ",
		"qi":  "2vPpArCCbWhp9NS5OKaUXnPlzYDFYkEarPdtI6EeTO4DkpefYpiaE08A2UkcH1mkEBOsJkYdEgzzyAQOxQblRA",
	})
	if err != nil {
		panic(err)
	}
	testTaPriKey = priKey.(*rsa.PrivateKey)
	testTaPubKey = &priKey.(*rsa.PrivateKey).PublicKey

	testTa = newTa(
		"testta",
		"testtaname",
		map[string]bool{
			"https://testta.example.org/":             true,
			"https://testta.example.org/redirect/uri": true,
		},
		map[string]interface{}{
			testTaKid: testTaPubKey,
		})
	testTa.Upd = testTa.Upd.Add(-(time.Duration(testTa.Upd.Nanosecond()) % time.Millisecond)) // mongodb の粒度がミリ秒のため。
}
