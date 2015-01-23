package main

import (
	"github.com/realglobe-Inc/edo/util"
)

type consent struct {
	Scops util.StringSet `json:"scope,omitempty" bson:"scope,omitempty"`
	Clms  util.StringSet `json:"claim,omitempty" bson:"claim,omitempty"`
}

func (this *consent) copy() *consent {
	return &consent{this.Scops.Copy(), this.Clms.Copy()}
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
