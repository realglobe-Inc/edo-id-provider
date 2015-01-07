package main

import (
	"crypto"
	"time"
)

func newRedisTokenContainer(idLen int, selfId string, key crypto.PrivateKey, kid, alg string, idTokExpiDur time.Duration,
	url, tag string, caStaleDur, caExpiDur time.Duration) tokenContainer {
	panic("not yet implemented")
}
