package main

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"github.com/realglobe-Inc/edo/driver"
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
	RediUris []string                 `json:"redirect_uris" bson:"redirect_uris"`
	PubKeys  []map[string]interface{} `json:"keys"          bson:"keys"`

	Date   time.Time `json:"-" bson:"date"`
	Digest string    `json:"-" bson:"digest"`
}

func taToIntermediate(t *ta) *taIntermediate {
	rediUris := []string{}
	for k := range t.rediUris {
		rediUris = append(rediUris, k)
	}
	pubKeys := []map[string]interface{}{}
	for k, v := range t.pubKeys {
		pubKeys = append(pubKeys, map[string]interface{}{
			"kty": "RSA",
			"n":   base64.URLEncoding.EncodeToString(v.N.Bytes()),
			"e":   base64.URLEncoding.EncodeToString(big.NewInt(int64(v.E)).Bytes()),
			"kid": k,
		})
	}
	return &taIntermediate{
		Id:       t.id,
		Name:     t.name,
		RediUris: rediUris,
		PubKeys:  pubKeys,
	}
}

func intermediateToTa(ti *taIntermediate) (*ta, error) {
	rediUris := map[string]bool{}
	for _, v := range ti.RediUris {
		rediUris[v] = true
	}
	pubKeys := map[string]*rsa.PublicKey{}
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
		rediUris: rediUris,
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
func mapToPublicKey(m map[string]interface{}) (kid string, pubKey *rsa.PublicKey, err error) {
	var buff rsa.PublicKey

	if v := m["kty"]; v == nil {
		return "", nil, nil
	} else if s, ok := v.(string); !ok {
		return "", nil, erro.New("value of kty is not string")
	} else if s != "RSA" {
		return "", nil, nil
	}

	if v := m["n"]; v == nil {
		return "", nil, nil
	} else if s, ok := v.(string); !ok {
		return "", nil, erro.New("value of n is not string")
	} else if b, err := base64.URLEncoding.DecodeString(s); err != nil {
		return "", nil, erro.Wrap(err)
	} else {
		buff.N = (&big.Int{}).SetBytes(b)
	}

	if v := m["e"]; v == nil {
		return "", nil, nil
	} else if s, ok := v.(string); !ok {
		return "", nil, erro.New("value of e is not string")
	} else if b, err := base64.URLEncoding.DecodeString(s); err != nil {
		return "", nil, erro.Wrap(err)
	} else {
		buff.E = int((&big.Int{}).SetBytes(b).Int64())
	}

	if v := m["kid"]; v == nil {
		// 無ければ無いで良い。
	} else if s, ok := v.(string); !ok {
		return "", nil, erro.New("value of kid is not string")
	} else {
		kid = s
	}

	return kid, &buff, nil
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
