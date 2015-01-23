package main

import (
	"crypto"
	"github.com/realglobe-Inc/edo/util"
	"reflect"
	"testing"
)

var testTa *ta

func init() {
	pubKey, err := util.ParseRsaPublicKey(`-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAzq6SrcupWm+rwSKsIVeO
yoRfUc0uPvoxPAEGF5uxZ9y+oqoPnTFCUAhDTX1lXDv4eHDPVsLuh8q75Bk0uDsY
zMNtrODC/neW6U3aKXHOXpdmzrDihk8/elgIxnZR5Dm5Vl5EeyCENwQIdxg+knU9
dkX7XwFib/RZAK2SXv1Xtgj4x6q7VgPl3zghdr67jCjZo3zgl0SxbZcOe4Yu4YGi
79+UZ14/tD9EghQGmFtFRac2xIH5iAYDhvDVi6zJgRjevSdds1xqKI3hkQJNT3zj
u3wa4HhSwmMLXwPTUXeTukTU1gU57++SWzrUogi71aQPcv8Y1k78Li5bS/VN1WTN
5QIDAQAB
-----END PUBLIC KEY-----`)
	if err != nil {
		panic(err)
	}
	testTa = newTa("abcde", "aaaaa", map[string]bool{"https://example.com/": true, "https://example.com/a/b/c": true}, map[string]crypto.PublicKey{"": pubKey})
}

func testTaContainer(t *testing.T, taCont taContainer) {
	if ta, err := taCont.get(testTa.id()); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(ta, testTa) {
		t.Error(ta)
	}

	if ta, err := taCont.get(testTa.id() + "a"); err != nil {
		t.Fatal(err)
	} else if ta != nil {
		t.Error(ta)
	}
}
