package main

import (
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/edo/util"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"time"
)

type sessionIntermediate struct {
	Id       string                                 `json:"id"`
	ExpiDate time.Time                              `json:"expires"`
	SelAccId string                                 `json:"selected_account,omitempty"`
	Accs     map[string]*sessionAccountIntermediate `json:"accounts"`
	SelCod   string                                 `json:"selection_code,omitempty"`
	ConsCod  string                                 `json:"consent_code,omitempty"`
}

type sessionAccountIntermediate struct {
	Auth     bool                `json:"authenticated"`
	AuthDate time.Time           `json:"authentication_date,omitempty"`
	TaConss  map[string][]string `json:'tas'`
}

func sessionToIntermediate(sess *session) *sessionIntermediate {
	accs := map[string]*sessionAccountIntermediate{}
	for accId, acc := range sess.accs {
		taConss := map[string][]string{}
		for taId, consSet := range acc.taConss {
			conss := []string{}
			for cons := range consSet {
				conss = append(conss, cons)
			}
			taConss[taId] = conss
		}
		accs[accId] = &sessionAccountIntermediate{
			Auth:     acc.auth,
			AuthDate: acc.authDate,
			TaConss:  taConss,
		}
	}
	return &sessionIntermediate{
		Id:       sess.sessId,
		ExpiDate: sess.expiDate,
		SelAccId: sess.selAccId,
		Accs:     accs,
		SelCod:   sess.selCod,
		ConsCod:  sess.consCod,
	}
}

func intermediateToSession(sessInter *sessionIntermediate) *session {
	accs := map[string]*sessionAccount{}
	for accId, acc := range sessInter.Accs {
		taConss := map[string]map[string]bool{}
		for taId, conss := range acc.TaConss {
			consSet := map[string]bool{}
			for _, cons := range conss {
				consSet[cons] = true
			}
			taConss[taId] = consSet
		}
		accs[accId] = &sessionAccount{
			auth:     acc.Auth,
			authDate: acc.AuthDate,
			taConss:  taConss,
		}
	}
	return &session{
		sessId:   sessInter.Id,
		expiDate: sessInter.ExpiDate,
		selAccId: sessInter.SelAccId,
		accs:     accs,
		selCod:   sessInter.SelCod,
		consCod:  sessInter.ConsCod,
	}
}

type sessionContainer interface {
	put(sess *session) error
	get(sessId string) (*session, error)
}

type sessionContainerImpl struct {
	base driver.TimeLimitedKeyValueStore
}

func (this *sessionContainerImpl) put(sess *session) error {
	if _, err := this.base.Put(sess.id(), sess, sess.expirationDate()); err != nil {
		return erro.Wrap(err)
	}

	return nil
}

func (this *sessionContainerImpl) get(sessId string) (*session, error) {
	val, _, err := this.base.Get(sessId, nil)
	if err != nil {
		return nil, erro.Wrap(err)
	} else if val == nil {
		return nil, nil
	}
	return val.(*session), nil
}

type sessionContainerWrapper struct {
	// セッション番号の文字数。
	idLen int
	// デフォルトの有効期限。
	expiDur time.Duration

	sessionContainer
}

func (this *sessionContainerWrapper) put(sess *session) error {
	if sess.id() == "" {
		sessId, err := util.SecureRandomString(this.idLen)
		if err != nil {
			return erro.Wrap(err)
		}
		sess.setId(sessId)

		// セッション ID が決まった。
		log.Debug("Session ID " + mosaic(sess.id()) + " was generated")
	}

	sess.setExpirationDate(time.Now().Add(this.expiDur))
	return this.sessionContainer.put(sess)
}
