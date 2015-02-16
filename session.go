package main

import (
	"encoding/json"
	"github.com/realglobe-Inc/edo/util/strset"
	"time"
)

type session struct {
	Id string `json:"id"`
	// 発行日時。
	Date time.Time `json:"date"`
	// 有効期限。
	ExpiDate time.Time `json:"expires"`
	// 更新日時。
	Upd time.Time `json:"update_at"`

	// 最後に選択・ログインしたアカウントの ID。
	CurAcc string `json:"current_account,omitempty"`
	// ログインしたことのあるアカウント。
	Accs sessionAccountMap `json:"accounts,omitempty"`
	// 最後に選択した言語。
	Loc string `json:"locale,omitempty"`

	// 進行中のユーザー認証・認可リクエスト。
	Req *authRequest `json:"request,omitempty"`
	// 進行中のリクエストにて発行された選択券。
	SelTic string `json:"select_ticket,omitempty"`
	// 進行中のリクエストにて発行されたログイン券。
	LoginTic string `json:"login_ticket,omitempty"`
	// 進行中のリクエストにて発行された同意券。
	ConsTic string `json:"consent_ticket,omitempty"`
	// 進行中のリクエストにて選択またはログインしたアカウント。
	Acc *sessionAccount `json:"accout,omitempty"`
	// 進行中のリクエストにてなされた同意。
	ConsScops strset.StringSet `json:"consented_scopes,omitempty"`
	ConsClms  strset.StringSet `json:"consented_claims,omitempty"`
	// 進行中のリクエストにてなされた拒否。
	DenyScops strset.StringSet `json:"denied_scopes,omitempty"`
	DenyClms  strset.StringSet `json:"denied_claims,omitempty"`
}

type sessionAccount struct {
	// 認証済みか。
	Auth bool `json:"authenticated,omitempty"`
	// ID 。
	Id string `json:"id"`
	// ログイン名。
	Name string `json:"username"`
	// 最後に認証した日時。
	AuthDate time.Time `json:"auth_time"`
}

// アカウント ID から sessionAccount へのマップ。
// JSON では sessionAccount の配列にする。
type sessionAccountMap map[string]*sessionAccount

func (this sessionAccountMap) MarshalJSON() ([]byte, error) {
	a := []*sessionAccount{}
	for _, v := range this {
		a = append(a, v)
	}
	return json.Marshal(a)
}

func (this *sessionAccountMap) UnmarshalJSON(data []byte) error {
	var a []*sessionAccount
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	m := map[string]*sessionAccount{}
	for _, v := range a {
		m[v.Id] = v
	}
	*this = sessionAccountMap(m)
	return nil
}

func (this *session) copy() *session {
	c := *this
	if this.Accs != nil {
		c.Accs = sessionAccountMap{}
		for accId, acc := range this.Accs {
			c.Accs[accId] = acc
		}
	}
	if this.ConsScops != nil {
		c.ConsScops = strset.New(this.ConsScops)
	}
	if this.ConsClms != nil {
		c.ConsClms = strset.New(this.ConsClms)
	}
	if this.DenyScops != nil {
		c.DenyScops = strset.New(this.DenyScops)
	}
	if this.DenyClms != nil {
		c.DenyClms = strset.New(this.DenyClms)
	}
	return &c
}

// 白紙のセッションをつくる。
func newSession() *session {
	return &session{
		Date: time.Now(),
	}
}

// ID を返す。
func (this *session) id() string {
	return this.Id
}

// ID を変更する。
func (this *session) setId(id string) {
	this.Id = id
}

// 発行日時を返す。
func (this *session) date() time.Time {
	return this.Date
}

// 有効期限を返す。
func (this *session) expirationDate() time.Time {
	return this.ExpiDate
}

// 有効期限を変更する。
func (this *session) setExpirationDate(expiDate time.Time) {
	this.ExpiDate = expiDate
	this.Upd = time.Now()
}

// 有効かどうかを返す。
func (this *session) valid() bool {
	return !this.ExpiDate.Before(time.Now())
}

// 更新日時を返す。
func (this *session) updateDate() time.Time {
	return this.Upd
}

// ユーザー認証・認可リクエストを始める。
func (this *session) startRequest(req *authRequest) {
	this.abort()
	this.Req = req
	return
}

// ユーザー認証・認可リクエストの結果を反映させる。
func (this *session) commit() (consScops, consClms, denyScops, denyClms map[string]bool) {
	if this.Acc != nil {
		this.CurAcc = this.Acc.Id
		if this.Accs == nil {
			this.Accs = sessionAccountMap{}
		}
		this.Accs[this.Acc.Id] = this.Acc
	}
	consScops = this.ConsScops
	consClms = this.ConsClms
	denyScops = this.DenyScops
	denyScops = this.DenyClms

	this.abort()
	return consScops, consClms, denyScops, denyClms
}

// ユーザー認証・認可リクエストを破棄する。
func (this *session) abort() {
	this.Upd = time.Now()
	this.Req = nil
	this.SelTic = ""
	this.LoginTic = ""
	this.ConsTic = ""
	this.Acc = nil
	this.ConsScops = nil
	this.ConsClms = nil
	this.DenyScops = nil
	this.DenyClms = nil
}

// 進行中のユーザー認証・認可リクエストを返す。
func (this *session) request() *authRequest {
	return this.Req
}

// 現在のアカウントの ID を返す。
func (this *session) currentAccount() string {
	if this.Acc != nil {
		return this.Acc.Id
	} else {
		return this.CurAcc
	}
}

// 現在のアカウントのログイン名を返す。
func (this *session) currentAccountName() string {
	if this.Acc != nil {
		return this.Acc.Name
	} else if this.CurAcc != "" {
		return this.Accs[this.CurAcc].Name
	} else {
		return ""
	}
}

// 現在のアカウントが認証されているかどうか。
func (this *session) currentAccountAuthenticated() bool {
	if this.Acc != nil {
		return this.Acc.Auth
	} else if this.CurAcc != "" {
		return this.Accs[this.CurAcc].Auth
	} else {
		return false
	}
}

// 現在のアカウントの認証日時を返す。
func (this *session) currentAccountDate() time.Time {
	if this.Acc != nil {
		return this.Acc.AuthDate
	} else if this.CurAcc != "" {
		return this.Accs[this.CurAcc].AuthDate
	} else {
		return time.Time{}
	}
}

// 関係するアカウント名を列挙する。
func (this *session) accountNames() map[string]bool {
	m := map[string]bool{}
	for _, v := range this.Accs {
		m[v.Name] = true
	}
	if this.Acc != nil {
		m[this.Acc.Name] = true
	}
	return m
}

// 選択したアカウントを設定する。
func (this *session) selectAccount(acc *account) {
	if this.Acc != nil && this.Acc.Id == acc.id() {
		// 既に選択している。
		return
	}

	var sessAcc *sessionAccount
	if this.Accs != nil {
		sessAcc = this.Accs[acc.id()]
	}
	if sessAcc == nil {
		// ログインしたこと無し。
		sessAcc = &sessionAccount{
			Id:   acc.id(),
			Name: acc.name(),
		}
	}
	this.Acc = sessAcc
	this.SelTic = ""
	return
}

// ログインしたアカウントを設定する。
func (this *session) loginAccount(acc *account) {
	this.Acc = &sessionAccount{
		Auth:     true,
		Id:       acc.id(),
		Name:     acc.name(),
		AuthDate: time.Now(),
	}
	this.LoginTic = ""
	return
}

// 同意を設定する。
func (this *session) consent(consScops, consClms, denyScops, denyClms map[string]bool) {
	this.ConsScops = consScops
	this.ConsClms = consClms
	this.DenyScops = denyScops
	this.DenyClms = denyClms
	this.ConsTic = ""
	return
}

// 選択券を返す。
func (this *session) selectTicket() string {
	return this.SelTic
}

// 選択券を紐付ける。
func (this *session) setSelectTicket(tic string) {
	this.SelTic = tic
}

// ログイン券を返す。
func (this *session) loginTicket() string {
	return this.LoginTic
}

// ログイン券を紐付ける。
func (this *session) setLoginTicket(tic string) {
	this.LoginTic = tic
}

// 同意券を返す。
func (this *session) consentTicket() string {
	return this.ConsTic
}

// 同意券を紐付ける。
func (this *session) setConsentTicket(tic string) {
	this.ConsTic = tic
}

// 必要な同意の中で拒否されたものを返す。
func (this *session) unconsentedEssentials() (scops, clms map[string]bool) {
	uncons := map[string]bool{}

	accInfClms, idTokClms := this.Req.claims()
	for _, clms := range []map[string]*claimUnit{accInfClms, idTokClms} {
		for clmName, req := range clms {
			if req != nil && req.Ess {
				if !this.ConsClms[clmName] {
					uncons[clmName] = true
				}
			}
		}
	}
	return nil, uncons
}

// 直近に選択した言語を返す。
func (this *session) locale() string {
	return this.Loc
}

// 言語を選択する。
func (this *session) setLocale(loc string) {
	this.Loc = loc
}
