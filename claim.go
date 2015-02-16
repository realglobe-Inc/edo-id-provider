package main

import (
	"encoding/json"
	"gopkg.in/mgo.v2/bson"
	"strings"
)

// claims パラメータや request オブジェクトの中身の claims 用。

// 言語タグはマップのキーに入らない。
// でも、JSON にしたときは claims パラメータのときと同じになる。
type claimRequest struct {
	accInf map[string]*claimUnit
	idTok  map[string]*claimUnit
}

type claimUnit struct {
	Ess  bool          `json:"essential,omitempty"`
	Val  interface{}   `json:"value,omitempty"`
	Vals []interface{} `json:"values,omitempty"`

	loc string
}

type claimRequestIntermediate struct {
	AccInf map[string]*claimUnit `json:"userinfo,omitempty"`
	IdTok  map[string]*claimUnit `json:"id_token,omitempty"`
}

func (this *claimRequest) toIntermediate() *claimRequestIntermediate {
	accInf := map[string]*claimUnit{}
	idTok := map[string]*claimUnit{}
	for clmName, clm := range this.accInf {
		if clm.loc != "" {
			clmName += "#" + clm.loc
		}
		accInf[clmName] = clm
	}
	for clmName, clm := range this.idTok {
		if clm.loc != "" {
			clmName += "#" + clm.loc
		}
		idTok[clmName] = clm
	}
	return &claimRequestIntermediate{accInf, idTok}
}

func (this *claimRequest) fromIntermediate(clms *claimRequestIntermediate) {
	accInf := map[string]*claimUnit{}
	idTok := map[string]*claimUnit{}
	for clmName, clm := range clms.AccInf {
		pos := strings.Index(clmName, "#")
		if pos > 0 {
			clm.loc = clmName[pos+1:]
			clmName = clmName[:pos]
		}
		accInf[clmName] = clm
	}
	for clmName, clm := range clms.IdTok {
		pos := strings.Index(clmName, "#")
		if pos > 0 {
			clm.loc = clmName[pos+1:]
			clmName = clmName[:pos]
		}
		idTok[clmName] = clm
	}
	this.accInf = accInf
	this.idTok = idTok
	return
}

func (this *claimRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(this.toIntermediate())
}

func (this *claimRequest) UnmarshalJSON(data []byte) error {
	var clms claimRequestIntermediate
	if err := json.Unmarshal(data, &clms); err != nil {
		return err
	}
	this.fromIntermediate(&clms)
	return nil
}

func (this *claimRequest) GetBSON() (interface{}, error) {
	return this.toIntermediate(), nil
}

func (this *claimRequest) SetBSON(raw bson.Raw) error {
	var clms claimRequestIntermediate
	if err := raw.Unmarshal(&clms); err != nil {
		return err
	}
	this.fromIntermediate(&clms)
	return nil
}
