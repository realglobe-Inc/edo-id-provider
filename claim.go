package main

import (
	"encoding/json"
	"gopkg.in/mgo.v2/bson"
	"strings"
)

// claims パラメータや request オブジェクトの中身の claims 用。

// クレーム名のマップの下に言語タグのマップを持たせる。
// でも、JSON にしたときは claims パラメータのときと同じ、クレーム名#言語タグ のマップになる。
// *claimUnit は非 nil になるようにする。
type claimRequest map[string]map[string]*claimUnit

type claimUnit struct {
	Ess  bool          `json:"essential,omitempty"`
	Val  interface{}   `json:"value,omitempty"`
	Vals []interface{} `json:"values,omitempty"`
}

func (this *claimUnit) isNull() bool {
	return !this.Ess && this.Val == nil && len(this.Vals) == 0
}

// *claimUnit は nil を許す。
type claimRequestIntermediate map[string]*claimUnit

func (this claimRequest) toIntermediate() claimRequestIntermediate {
	buff := map[string]*claimUnit{}
	for clmName, clms := range this {
		for loc, clm := range clms {
			name := clmName
			if loc != "" {
				name += "#" + loc
			}
			if clm.isNull() {
				clm = nil
			}
			buff[name] = clm
		}
	}
	return claimRequestIntermediate(buff)
}

func (this *claimRequest) fromIntermediate(clms claimRequestIntermediate) {
	buff := claimRequest{}
	for clmName, clm := range clms {
		if clm == nil {
			clm = &claimUnit{}
		}
		var loc string
		pos := strings.Index(clmName, "#")
		if pos > 0 {
			loc = clmName[pos+1:]
			clmName = clmName[:pos]
		}
		m := buff[clmName]
		if m == nil {
			m = map[string]*claimUnit{}
			buff[clmName] = m
		}
		m[loc] = clm
	}
	*this = buff
	return
}

func (this claimRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(this.toIntermediate())
}

func (this *claimRequest) UnmarshalJSON(data []byte) error {
	var clms claimRequestIntermediate
	if err := json.Unmarshal(data, &clms); err != nil {
		return err
	}
	this.fromIntermediate(clms)
	return nil
}

func (this claimRequest) GetBSON() (interface{}, error) {
	return this.toIntermediate(), nil
}

func (this *claimRequest) SetBSON(raw bson.Raw) error {
	var clms claimRequestIntermediate
	if err := raw.Unmarshal(&clms); err != nil {
		return err
	}
	this.fromIntermediate(clms)
	return nil
}
