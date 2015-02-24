package main

import (
	"encoding/json"
	"github.com/realglobe-Inc/go-lib/erro"
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

// json.Unmarshal したものから読み取る。
func claimRequestFromMap(raw interface{}) (claimRequest, error) {
	mClms, ok := raw.(map[string]interface{})
	if !ok {
		return nil, erro.New("not map")
	}

	clms := claimRequest{}
	for clmName, rawClm := range mClms {
		clm, err := claimUnitFromMap(rawClm)
		if err != nil {
			return nil, erro.Wrap(err)
		}
		if clm == nil {
			clm = &claimUnit{}
		}
		var loc string
		pos := strings.Index(clmName, "#")
		if pos > 0 {
			loc = clmName[pos+1:]
			clmName = clmName[:pos]
		}
		m := clms[clmName]
		if m == nil {
			m = map[string]*claimUnit{}
			clms[clmName] = m
		}
		m[loc] = clm
	}

	return clms, nil
}

// json.Unmarshal したものから読み取る。
func claimUnitFromMap(raw interface{}) (*claimUnit, error) {
	clm := &claimUnit{}

	if raw == nil {
		// null 指定。
		return clm, nil
	}

	m, ok := raw.(map[string]interface{})
	if !ok {
		return nil, erro.New("not map")
	}

	if raw := m["essential"]; raw != nil {
		b, ok := raw.(bool)
		if !ok {
			return nil, erro.New("essential is not boolean")
		}
		clm.Ess = b
	}
	if val := m["value"]; val != nil {
		clm.Val = val
	}
	if raw := m["values"]; raw != nil {
		vals, ok := raw.([]interface{})
		if !ok {
			return nil, erro.New("values is not array")
		}
		clm.Vals = vals
	}

	return clm, nil
}
