package main

import (
	"github.com/realglobe-Inc/edo/util"
	"time"
)

type session struct {
	SessId   string    `json:"id"`
	ExpiDate time.Time `json:"expires"`

	// 選択中のアカウント ID。
	SelAccId string `json:"selected_account,omitempty"`
	// 認証したことのあるアカウント。
	Accs map[string]*sessionAccount `json:"accounts"`

	// 最後に紐付けられた選択コード。
	SelCod string `json:"selection_code,omitempty"`
	// 最後に紐付けられた同意コード。
	ConsCod string `json:"consent_code,omitempty"`
}

type sessionAccount struct {
	// 現在認証されているか。
	Auth bool `json:"authenticated"`
	// ログイン名。
	Name string
	// 最後に認証した日時。
	AuthDate time.Time `json:"authentication_date,omitempty"`
	// TA ごとの同意。
	TaConss map[string]*util.StringSet `json:'tas'`
}

func (this *session) copy() *session {
	c := *this
	c.Accs = map[string]*sessionAccount{}
	for accId, acc := range this.Accs {
		c.Accs[accId] = acc.copy()
	}
	return &c
}

func (this *sessionAccount) copy() *sessionAccount {
	c := *this
	c.TaConss = map[string]*util.StringSet{}
	for taId, conss := range this.TaConss {
		c.TaConss[taId] = util.NewStringSet(conss.Elements())
	}
	return &c
}

// 白紙のセッションをつくる。
func newSession() *session {
	return &session{
		Accs: map[string]*sessionAccount{},
	}
}

// ID を返す。
func (this *session) id() string {
	return this.SessId
}

// ID を設定する。
func (this *session) setId(id string) {
	this.SessId = id
}

// 有効期限を返す。
func (this *session) expirationDate() time.Time {
	return this.ExpiDate
}

// 有効期限を設定する。
func (this *session) setExpirationDate(expiDate time.Time) {
	this.ExpiDate = expiDate
}

// 選択されているアカウントの ID を返す。
func (this *session) account() string {
	return this.SelAccId
}

// 選択されているアカウントのログイン名を返す。
func (this *session) accountName() string {
	if this.SelAccId == "" {
		return ""
	}
	return this.Accs[this.SelAccId].Name
}

// アカウントを認証済みアカウントとして加え、選択する。
// 選択コードも解除する。
// 状態が変わったときのみ true を返す。
func (this *session) setAccount(acc *account) bool {
	if sessAcc := this.Accs[acc.id()]; sessAcc != nil {
		if sessAcc.Auth {
			return false
		} else {
			sessAcc.Auth = true
			return true
		}
	} else {
		sessAcc = &sessionAccount{
			Auth:    true,
			Name:    acc.name(),
			TaConss: map[string]*util.StringSet{},
		}
		this.Accs[acc.id()] = sessAcc
		return true
	}
}

// 現在紐付けられている選択コードを返す。
func (this *session) selectionCode() string {
	return this.SelCod
}

// 選択コードを紐付ける。
func (this *session) setSelectionCode(selCod string) {
	this.SelCod = selCod
}

// 一度認証したことのあるアカウントの中から 1 つを選択する。
// 選択コードも解除する。
// 状態が変わったときのみ true を返す。
func (this *session) selectAccount(accId string) bool {
	if accId == this.SelAccId {
		return false
	} else if _, ok := this.Accs[accId]; ok {
		this.SelAccId = accId
		return true
	} else {
		return false
	}
}
func (this *session) selectAccountByName(accName string) bool {
	var accId string
	for id, acc := range this.Accs {
		if acc.Name == accName {
			accId = id
			break
		}
	}
	if accId == this.SelAccId {
		return false
	} else if _, ok := this.Accs[accId]; ok {
		this.SelAccId = accId
		return true
	} else {
		return false
	}
}

// アカウントが選択されていて認証済みの場合のみ true。
func (this *session) isAuthenticated() bool {
	if this.SelAccId == "" {
		return false
	} else {
		return this.Accs[this.SelAccId].Auth
	}
}

// アカウントが選択されていたら、そのアカウントを認証済みでなくする。
// 状態が変わったときのみ true を返す。
func (this *session) clearAuthenticated() bool {
	if this.SelAccId == "" {
		return false
	} else {
		acc := this.Accs[this.SelAccId]
		if acc.Auth {
			return false
		} else {
			return true
		}
	}
}

// 現在紐付けられている同意コードを返す。
func (this *session) consentCode() string {
	return this.ConsCod
}

// 同意コードを紐付ける。
func (this *session) setConsentCode(consCod string) {
	this.ConsCod = consCod
}

// アカウントの同意が得られていないクレームがあるかどうか。
func (this *session) hasNotConsented(accId, taId string, clms map[string]bool) bool {
	if len(clms) == 0 {
		return false
	} else if acc := this.Accs[accId]; acc == nil {
		return true
	} else if conss := acc.TaConss[taId]; conss == nil {
		return true
	} else {
		for clm := range clms {
			if !conss.Contains(clm) {
				return true
			}
		}
		return false
	}
}

// アカウントの同意が得られていないクレームを返す。
func (this *session) notConsented(accId, taId string, clms map[string]bool) map[string]bool {
	rems := map[string]bool{}
	if len(clms) == 0 {
	} else if acc := this.Accs[accId]; acc == nil {
		for clm := range clms {
			rems[clm] = true
		}
	} else if conss := acc.TaConss[taId]; conss == nil {
		for clm := range clms {
			rems[clm] = true
		}
	} else {
		for clm := range clms {
			if !conss.Contains(clm) {
				rems[clm] = true
			}
		}
	}
	return rems
}

// アカウントで同意する。
// 同意コードも解除する。
// 状態が変わったときのみ true を返す。
func (this *session) consent(accId, accName, taId string, clms map[string]bool) bool {
	mod := false
	sessAcc := this.Accs[accId]
	if sessAcc == nil {
		mod = true
		sessAcc = &sessionAccount{
			Auth:    true,
			Name:    accName,
			TaConss: map[string]*util.StringSet{},
		}
		this.Accs[accId] = sessAcc
	}
	conss := sessAcc.TaConss[taId]
	if conss == nil {
		mod = true
		conss = util.NewStringSet(nil)
	}
	for clm := range clms {
		if !conss.Contains(clm) {
			mod = true
		}
		conss.Put(clm)
	}
	return mod
}
