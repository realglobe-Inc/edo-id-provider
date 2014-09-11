package main

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"math/big"
	"time"
)

// 便宜的に集めただけ。
type system struct {
	driver.ServiceExplorer
	driver.UserNameIndex
	driver.UserAttributeRegistry
	sessCont driver.TimeLimitedKeyValueStore
	codeCont driver.TimeLimitedKeyValueStore

	cookieMaxAge   int
	maxSessExpiDur time.Duration
	codeExpiDur    time.Duration
}

// ユーザー認証済みを示すセッション。
type Session struct {
	Id       string    `json:"id"`
	UsrUuid  string    `json:"user_uuid"`
	ExpiDate time.Time `json:"expiration_date"`
}

const sessIdLen int = 20 // セッション ID の文字数。

func (sys *system) NewSession(usrUuid string, expiDur time.Duration) (*Session, error) {
	var sessId string
	for {
		sessIdBitLen := sessIdLen * 6
		maxVal := big.NewInt(0).Lsh(big.NewInt(1), uint(sessIdBitLen))
		val, err := rand.Int(rand.Reader, maxVal)
		if err != nil {
			return nil, erro.Wrap(err)
		}
		buff := val.Bytes()

		for len(buff) < (sessIdBitLen+5)/8 {
			buff = append(buff, 0)
		}

		sessId = base64.StdEncoding.EncodeToString(buff)
		if len(sessId) > sessIdLen {
			sessId = sessId[:sessIdLen]
		}

		if sess, err := sys.sessCont.Get(sessId); err != nil {
			return nil, erro.Wrap(err)
		} else if sess == nil {
			break
		}
	}

	// セッション ID が決まった。
	log.Debug("Session ID was published.")

	if expiDur > sys.maxSessExpiDur {
		expiDur = sys.maxSessExpiDur
	}
	sess := &Session{sessId, usrUuid, time.Now().Add(expiDur)}

	if err := sys.sessCont.Put(sessId, sess, sess.ExpiDate); err != nil {
		return nil, erro.Wrap(err)
	}

	log.Debug("Session was published.")
	return sess, nil
}

// 有効なセッションを取り出す。
func (sys *system) Session(sessId string) (*Session, error) {
	val, err := sys.sessCont.Get(sessId)
	var sess *Session
	if val != nil {
		m := val.(map[string]interface{})
		expiDate, err := time.Parse(time.RFC3339, m["expiration_date"].(string))
		if err != nil {
			return nil, erro.Wrap(err)
		}
		sess = &Session{
			m["id"].(string),
			m["user_uuid"].(string),
			expiDate,
		}
	}
	return sess, err
}

// アクセストークン発行許可証。
type Code struct {
	Id       string    `json:"id"`
	ServUuid string    `json:"service_uuid"`
	ExpiDate time.Time `json:"expiration_date"`
}

const codeIdLen int = 20 // コード ID の文字数。

func (sys *system) NewCode(servUuid string) (*Code, error) {
	var codeId string
	for {
		codeIdBitLen := codeIdLen * 6
		maxVal := big.NewInt(0).Lsh(big.NewInt(1), uint(codeIdBitLen))
		val, err := rand.Int(rand.Reader, maxVal)
		if err != nil {
			return nil, erro.Wrap(err)
		}
		buff := val.Bytes()

		for len(buff) < (codeIdBitLen+5)/8 {
			buff = append(buff, 0)
		}

		codeId = base64.StdEncoding.EncodeToString(buff)
		if len(codeId) > codeIdLen {
			codeId = codeId[:codeIdLen]
		}

		if code, err := sys.codeCont.Get(codeId); err != nil {
			return nil, erro.Wrap(err)
		} else if code == nil {
			break
		}
	}

	// セッション ID が決まった。
	log.Debug("Code ID was published.")

	code := &Code{codeId, servUuid, time.Now().Add(sys.codeExpiDur)}

	if err := sys.codeCont.Put(codeId, code, code.ExpiDate); err != nil {
		return nil, erro.Wrap(err)
	}

	log.Debug("Code was published.")
	return code, nil
}

// 有効なアクセストークン取得用コードを取り出す。
func (sys *system) Code(codeId string) (*Code, error) {
	val, err := sys.codeCont.Get(codeId)
	var code *Code
	if val != nil {
		m := val.(map[string]interface{})
		expiDate, err := time.Parse(time.RFC3339, m["expiration_date"].(string))
		if err != nil {
			return nil, erro.Wrap(err)
		}
		code = &Code{
			m["id"].(string),
			m["service_uuid"].(string),
			expiDate,
		}
	}
	return code, err
}

func (sys *system) UserPassword(usrUuid string) (passwd string, err error) {
	attr, err := sys.UserAttribute(usrUuid, "login_password")
	if attr != nil && attr != "" {
		passwd = attr.(string)
	}
	return passwd, err
}
