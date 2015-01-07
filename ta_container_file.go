package main

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/edo/util"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"math/big"
	"net/url"
	"strings"
	"time"
)

// ta の中間形式。
type taIntermediate struct {
	Id       string                   `json:"id"            bson:"id"`
	Name     string                   `json:"name"          bson:"name"`
	RediUris util.StringSet           `json:"redirect_uris" bson:"redirect_uris"`
	PubKeys  []map[string]interface{} `json:"keys"          bson:"keys"`

	Date   time.Time `json:"-" bson:"date"`
	Digest string    `json:"-" bson:"digest"`
}

func taToIntermediate(t *ta) *taIntermediate {
	rediUris := t.rediUris
	pubKeys := []map[string]interface{}{}
	for k, v := range t.pubKeys {
		switch pubKey := v.(type) {
		case *rsa.PublicKey:
			pubKeys = append(pubKeys, map[string]interface{}{
				"kty": "RSA",
				"n":   base64.URLEncoding.EncodeToString(pubKey.N.Bytes()),
				"e":   base64.URLEncoding.EncodeToString(big.NewInt(int64(pubKey.E)).Bytes()),
				"kid": k,
			})
		case *ecdsa.PublicKey:
			s := pubKey.Params().BitSize
			m := map[string]interface{}{}
			switch s {
			case 256:
				m["crv"] = "P-256"
			case 384:
				m["crv"] = "P-384"
			case 521:
				m["crv"] = "P-521"
			}
			if len(m) == 0 {
				break // switch pubKey
			}
			byteLen := (s + 7) / 8
			x, y := pubKey.X.Bytes(), pubKey.Y.Bytes()
			if len(x) < byteLen {
				buff := make([]byte, byteLen)
				copy(buff[byteLen-len(x):], x)
				x = buff
			}
			if len(y) < byteLen {
				buff := make([]byte, byteLen)
				copy(buff[byteLen-len(y):], y)
				y = buff
			}
			m["x"] = base64.URLEncoding.EncodeToString(x)
			m["y"] = base64.URLEncoding.EncodeToString(y)
			pubKeys = append(pubKeys, m)
		}
	}
	return &taIntermediate{
		Id:       t.id,
		Name:     t.name,
		RediUris: rediUris,
		PubKeys:  pubKeys,
	}
}

func intermediateToTa(ti *taIntermediate) (*ta, error) {
	pubKeys := map[string]crypto.PublicKey{}
	for _, v := range ti.PubKeys {
		if kid, pubKey, err := mapToPublicKey(v); err != nil {
			return nil, erro.Wrap(err)
		} else if pubKey != nil {
			pubKeys[kid] = pubKey
		}
	}
	return &ta{
		id:       ti.Id,
		name:     ti.Name,
		rediUris: ti.RediUris,
		pubKeys:  pubKeys,
	}, nil
}

func marshalTa(t interface{}) ([]byte, error) {
	return json.Marshal(taToIntermediate(t.(*ta)))
}

func unmarshalTa(data []byte) (interface{}, error) {
	var ti taIntermediate
	if err := json.Unmarshal(data, &ti); err != nil {
		return nil, erro.Wrap(err)
	}
	return intermediateToTa(&ti)
}

// 公開鍵を中間形式から戻す。
// 中間形式は以下のような JWK を json.Unmarshal したものだけに対応。
// {
//     "kty": "RSA",
//     "n": "XXX...",
//     "e": "YYY",
//     "kid": "ZZZ"
// }
func mapToPublicKey(m map[string]interface{}) (kid string, pubKey crypto.PublicKey, err error) {
	switch kty, _ := m["kty"].(string); kty {
	case "":
		return "", nil, nil
	case "RSA":
		var buff rsa.PublicKey
		if nStr, _ := m["n"].(string); nStr == "" {
			return "", nil, erro.New("no n")
		} else if nRaw, err := base64.URLEncoding.DecodeString(nStr); err != nil {
			return "", nil, erro.Wrap(err)
		} else {
			buff.N = (&big.Int{}).SetBytes(nRaw)
		}
		if eStr, _ := m["e"].(string); eStr == "" {
			return "", nil, erro.New("no e")
		} else if eRaw, err := base64.URLEncoding.DecodeString(eStr); err != nil {
			return "", nil, erro.Wrap(err)
		} else {
			buff.E = int((&big.Int{}).SetBytes(eRaw).Int64())
		}
		kid, _ = m["kid"].(string)
		return kid, &buff, nil
	case "EC":
		var buff ecdsa.PublicKey
		switch crv, _ := m["crv"].(string); crv {
		case "":
			return "", nil, erro.New("no crv")
		case "P-256":
			buff.Curve = elliptic.P256()
		case "P-384":
			buff.Curve = elliptic.P384()
		case "P-521":
			buff.Curve = elliptic.P521()
		default:
			return "", nil, erro.New("unsupported elliptic curve " + crv)
		}
		if xStr, _ := m["x"].(string); xStr == "" {
			return "", nil, erro.New("no x")
		} else if xRaw, err := base64.URLEncoding.DecodeString(xStr); err != nil {
			return "", nil, erro.Wrap(err)
		} else {
			buff.X = (&big.Int{}).SetBytes(xRaw)
		}
		if yStr, _ := m["y"].(string); yStr == "" {
			return "", nil, erro.New("no y")
		} else if yRaw, err := base64.URLEncoding.DecodeString(yStr); err != nil {
			return "", nil, erro.Wrap(err)
		} else {
			buff.Y = (&big.Int{}).SetBytes(yRaw)
		}
		kid, _ = m["kid"].(string)
		return kid, &buff, nil
	default:
		return "", nil, nil
	}
}

func keyToJsonPath(key string) string {
	return key + ".json"
}

func jsonPathToKey(path string) string {
	if !strings.HasSuffix(path, ".json") {
		return ""
	}
	return path[:len(path)-len(".json")]
}

func keyToEscapedJsonPath(key string) string {
	return keyToJsonPath(url.QueryEscape(key))
}

func escapedJsonPathToKey(path string) string {
	key, _ := url.QueryUnescape(jsonPathToKey(path))
	return key
}

// スレッドセーフ。
func newFileTaContainer(path string, staleDur, expiDur time.Duration) taContainer {
	return &taContainerImpl{driver.NewFileListedKeyValueStore(path,
		keyToEscapedJsonPath, escapedJsonPathToKey,
		marshalTa, unmarshalTa,
		staleDur, expiDur)}
}
