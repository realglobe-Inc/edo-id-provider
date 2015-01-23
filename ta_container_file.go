package main

import (
	"crypto"
	"encoding/json"
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/edo/util"
	"github.com/realglobe-Inc/go-lib-rg/erro"
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
	rediUris := t.redirectUris()
	pubKeys := []map[string]interface{}{}
	for k, v := range t.keys() {
		if m := util.EncodePublicKeyToJwkMap(k, v); m != nil {
			pubKeys = append(pubKeys, m)
		}
	}
	return &taIntermediate{
		Id:       t.id(),
		Name:     t.name(),
		RediUris: rediUris,
		PubKeys:  pubKeys,
	}
}

func intermediateToTa(ti *taIntermediate) (*ta, error) {
	pubKeys := map[string]crypto.PublicKey{}
	for _, v := range ti.PubKeys {
		if kid, pubKey, err := util.ParsePublicKeyFromJwkMap(v); err != nil {
			return nil, erro.Wrap(err)
		} else if pubKey != nil {
			pubKeys[kid] = pubKey
		}
	}
	return newTa(ti.Id, ti.Name, ti.RediUris, pubKeys), nil
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
