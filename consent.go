package main

import (
	"github.com/realglobe-Inc/edo/util"
	"time"
)

// 許可情報。
type consent struct {
	// 権利者アカウントの ID。
	AccId string `json:"account" bson:"account"`
	// 要求元 TA の ID。
	TaId string `json:"client_id" bson:"client_id"`

	// 更新日時。
	Upd time.Time `json:"update_at" bson:"update_at"`
	// 提供許可されている scope。
	Scops util.StringSet `json:"scope,omitempty" bson:"scope,omitempty"`
	// 提供許可されているクレーム。
	Clms util.StringSet `json:"claims,omitempty" bson:"claims,omitempty"`
}

func newConsent(accId, taId string, scops, clms map[string]bool) *consent {
	return &consent{
		AccId: accId,
		TaId:  taId,
		Upd:   time.Now(),
		Scops: scops,
		Clms:  clms,
	}
}

func (this *consent) updateDate() time.Time {
	return this.Upd
}

func (this *consent) scopes() map[string]bool {
	if this.Scops == nil {
		this.Scops = util.StringSet{}
	}
	return this.Scops
}

func (this *consent) claims() map[string]bool {
	if this.Clms == nil {
		this.Clms = util.StringSet{}
	}
	return this.Clms
}

// 同意 consScops, consClms で要求 reqScops, reqClms を満たせるかどうか調べる。
func satisfiable(consScops, consClms, reqScops, reqClms map[string]bool) bool {
	for scop := range reqScops {
		if !consScops[scop] {
			return false
		}
	}
	for clm := range reqClms {
		if !consClms[clm] {
			return false
		}
	}
	return true
}
