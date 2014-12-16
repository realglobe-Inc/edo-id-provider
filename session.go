package main

import (
	"time"
)

type session struct {
	sessId   string
	expiDate time.Time

	// 選択中のアカウント ID。
	selAccId string
	// 認証したことのあるアカウント。
	accs map[string]*sessionAccount

	// 最後に紐付けられた選択コード。
	selCod string
	// 最後に紐付けられた同意コード。
	consCod string
}

type sessionAccount struct {
	// 現在認証されているか。
	auth bool
	// 最後に認証した日時。
	authDate time.Time
	// TA ごとの同意。
	taConss map[string]map[string]bool
}

func (this *session) copy() *session {
	c := *this
	c.accs = map[string]*sessionAccount{}
	for accId, acc := range this.accs {
		c.accs[accId] = acc.copy()
	}
	return &c
}

func (this *sessionAccount) copy() *sessionAccount {
	c := *this
	c.taConss = map[string]map[string]bool{}
	for taId, conss := range this.taConss {
		cConss := map[string]bool{}
		for k, v := range conss {
			cConss[k] = v
		}
		c.taConss[taId] = cConss
	}
	return &c
}

// 白紙のセッションをつくる。
func newSession() *session {
	return &session{
		accs: map[string]*sessionAccount{},
	}
}

// ID を返す。
func (this *session) id() string {
	return this.sessId
}

// ID を設定する。
func (this *session) setId(id string) {
	this.sessId = id
}

// 有効期限を返す。
func (this *session) expirationDate() time.Time {
	return this.expiDate
}

// 有効期限を設定する。
func (this *session) setExpirationDate(expiDate time.Time) {
	this.expiDate = expiDate
}

// 選択されているアカウントの ID を返す。
func (this *session) account() string {
	return this.selAccId
}

// アカウントを認証済みアカウントとして加え、選択する。
// 選択コードも解除する。
// 状態が変わったときのみ true を返す。
func (this *session) setAccount(accId string) bool {
	if acc := this.accs[accId]; acc != nil {
		if acc.auth {
			return false
		} else {
			acc.auth = true
			return true
		}
	} else {
		acc = &sessionAccount{
			auth:    true,
			taConss: map[string]map[string]bool{},
		}
		this.accs[accId] = acc
		return true
	}
}

// 現在紐付けられている選択コードを返す。
func (this *session) selectionCode() string {
	return this.selCod
}

// 選択コードを紐付ける。
func (this *session) setSelectionCode(selCod string) {
	this.selCod = selCod
}

// 一度認証したことのあるアカウントの中から 1 つを選択する。
// 選択コードも解除する。
// 状態が変わったときのみ true を返す。
func (this *session) selectAccount(accId string) bool {
	if accId == this.selAccId {
		return false
	} else if _, ok := this.accs[accId]; ok {
		this.selAccId = accId
		return true
	} else {
		return false
	}
}

// アカウントが選択されていて認証済みの場合のみ true。
func (this *session) isAuthenticated() bool {
	if this.selAccId == "" {
		return false
	} else {
		return this.accs[this.selAccId].auth
	}
}

// アカウントが選択されていたら、そのアカウントを認証済みでなくする。
// 状態が変わったときのみ true を返す。
func (this *session) clearAuthenticated() bool {
	if this.selAccId == "" {
		return false
	} else {
		acc := this.accs[this.selAccId]
		if acc.auth {
			return false
		} else {
			return true
		}
	}
}

// 現在紐付けられている同意コードを返す。
func (this *session) consentCode() string {
	return this.consCod
}

// 同意コードを紐付ける。
func (this *session) setConsentCode(consCod string) {
	this.consCod = consCod
}

// アカウントの同意が得られていないクレームがあるかどうか。
func (this *session) hasNotConsented(accId, taId string, clms map[string]bool) bool {
	if len(clms) == 0 {
		return false
	} else if acc := this.accs[accId]; acc == nil {
		return true
	} else if conss := acc.taConss[taId]; conss == nil {
		return true
	} else {
		for clm := range clms {
			if !conss[clm] {
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
	} else if acc := this.accs[accId]; acc == nil {
		for clm := range clms {
			rems[clm] = true
		}
	} else if conss := acc.taConss[taId]; conss == nil {
		for clm := range clms {
			rems[clm] = true
		}
	} else {
		for clm := range clms {
			if !conss[clm] {
				rems[clm] = true
			}
		}
	}
	return rems
}

// アカウントで同意する。
// 同意コードも解除する。
// 状態が変わったときのみ true を返す。
func (this *session) consent(accId, taId string, clms map[string]bool) bool {
	mod := false
	acc := this.accs[accId]
	if acc == nil {
		mod = true
		acc = &sessionAccount{
			auth:    true,
			taConss: map[string]map[string]bool{},
		}
		this.accs[accId] = acc
	}
	conss := acc.taConss[taId]
	if conss == nil {
		mod = true
		conss = map[string]bool{}
	}
	for clm := range clms {
		if !conss[clm] {
			mod = true
		}
		conss[clm] = true
	}
	return mod
}
