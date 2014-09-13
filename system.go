package main

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"math/big"
	"strings"
	"time"
)

// 便宜的に集めただけ。
type system struct {
	driver.ServiceExplorer
	driver.ServiceKeyRegistry
	driver.UserNameIndex
	driver.UserAttributeRegistry
	sessCont     driver.TimeLimitedKeyValueStore
	codeCont     driver.TimeLimitedKeyValueStore
	accTokenCont driver.TimeLimitedKeyValueStore

	maxSessExpiDur     time.Duration // デフォルトかつ最大。
	codeExpiDur        time.Duration
	accTokenExpiDur    time.Duration // デフォルト。
	maxAccTokenExpiDur time.Duration
}

const attrLoginPasswd = "login_password"

func (sys *system) UserPassword(usrUuid string) (passwd string, err error) {
	attr, err := sys.UserAttribute(usrUuid, attrLoginPasswd)
	if err != nil {
		return "", erro.Wrap(err)
	} else if attr == nil || attr == "" {
		return "", nil
	}
	return attr.(string), err
}

func (sys *system) UserAttribute(usrUuid, attrName string) (attr interface{}, err error) {
	// TODO フィルタをハードコードしちゃってる。
	if strings.HasSuffix(attrName, "password") {
		return nil, nil
	}
	attr, err = sys.UserAttributeRegistry.UserAttribute(usrUuid, attrName)
	if err != nil {
		return nil, erro.Wrap(err)
	} else if attr == nil || attr == "" {
		return nil, nil
	}
	return attr, err
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
	log.Debug("Session ID was generated.")

	if expiDur == 0 || expiDur > sys.maxSessExpiDur {
		expiDur = sys.maxSessExpiDur
	}
	sess := &Session{sessId, usrUuid, time.Now().Add(expiDur)}

	if err := sys.sessCont.Put(sessId, sess, sess.ExpiDate); err != nil {
		return nil, erro.Wrap(err)
	}

	log.Debug("Session was published.")
	return sess, nil
}

func recognizeSession(sess interface{}) (*Session, error) {
	switch s := sess.(type) {
	case *Session:
		return s, nil
	case map[string]interface{}:
		expiDate, err := time.Parse(time.RFC3339, s["expiration_date"].(string))
		if err != nil {
			return nil, erro.Wrap(err)
		}
		return &Session{
			s["id"].(string),
			s["user_uuid"].(string),
			expiDate,
		}, nil
	default:
		return nil, erro.New("unknown type ", sess, ".")
	}
}

// 有効なセッションを取り出す。
func (sys *system) Session(sessId string) (*Session, error) {
	val, err := sys.sessCont.Get(sessId)
	if err != nil {
		return nil, erro.Wrap(err)
	} else if val == nil {
		return nil, nil
	}
	return recognizeSession(val)
}

// セッションの有効期限を更新。
func (sys *system) UpdateSession(sess *Session) error {
	return sys.sessCont.Put(sess.Id, sess, sess.ExpiDate)
}

// アクセストークン発行用コード。
type Code struct {
	Id       string    `json:"id"`
	UsrUuid  string    `json:"user_uuid"`
	ServUuid string    `json:"service_uuid"`
	ExpiDate time.Time `json:"expiration_date"`
}

const codeIdLen int = 20 // アクセストークン発行用コードの文字数。

func (sys *system) NewCode(usrUuid, servUuid string) (*Code, error) {
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

	// コードが決まった。
	log.Debug("Code was generated.")

	code := &Code{codeId, usrUuid, servUuid, time.Now().Add(sys.codeExpiDur)}

	if err := sys.codeCont.Put(codeId, code, code.ExpiDate); err != nil {
		return nil, erro.Wrap(err)
	}

	log.Debug("Code was published.")
	return code, nil
}

func recognizeCode(code interface{}) (*Code, error) {
	switch s := code.(type) {
	case *Code:
		return s, nil
	case map[string]interface{}:
		expiDate, err := time.Parse(time.RFC3339, s["expiration_date"].(string))
		if err != nil {
			return nil, erro.Wrap(err)
		}
		return &Code{
			s["id"].(string),
			s["user_uuid"].(string),
			s["service_uuid"].(string),
			expiDate,
		}, nil
	default:
		return nil, erro.New("unknown type ", code, ".")
	}
}

// 有効なアクセストークン取得用コードを取り出す。
func (sys *system) Code(codeId string) (*Code, error) {
	val, err := sys.codeCont.Get(codeId)
	if err != nil {
		return nil, erro.Wrap(err)
	} else if val == nil {
		return nil, nil
	}
	return recognizeCode(val)
}

// アクセストークン。
type AccessToken struct {
	Id       string    `json:"id"`
	UsrUuid  string    `json:"user_uuid"`
	ExpiDate time.Time `json:"expiration_date"`
}

const accTokenIdLen int = 20 // アクセストークンの文字数。

func (sys *system) NewAccessToken(usrUuid string, expiDur time.Duration) (*AccessToken, error) {
	var accTokenId string
	for {
		accTokenIdBitLen := accTokenIdLen * 6
		maxVal := big.NewInt(0).Lsh(big.NewInt(1), uint(accTokenIdBitLen))
		val, err := rand.Int(rand.Reader, maxVal)
		if err != nil {
			return nil, erro.Wrap(err)
		}
		buff := val.Bytes()

		for len(buff) < (accTokenIdBitLen+5)/8 {
			buff = append(buff, 0)
		}

		accTokenId = base64.StdEncoding.EncodeToString(buff)
		if len(accTokenId) > accTokenIdLen {
			accTokenId = accTokenId[:accTokenIdLen]
		}

		if accToken, err := sys.accTokenCont.Get(accTokenId); err != nil {
			return nil, erro.Wrap(err)
		} else if accToken == nil {
			break
		}
	}

	// アクセストークンが決まった。
	log.Debug("Access token was generated.")

	if expiDur <= 0 {
		expiDur = sys.accTokenExpiDur
	} else if expiDur > sys.maxAccTokenExpiDur {
		expiDur = sys.maxAccTokenExpiDur
	}
	accToken := &AccessToken{accTokenId, usrUuid, time.Now().Add(expiDur)}

	if err := sys.accTokenCont.Put(accTokenId, accToken, accToken.ExpiDate); err != nil {
		return nil, erro.Wrap(err)
	}

	log.Debug("Access token was published.")
	return accToken, nil
}

func recognizeAccessToken(accToken interface{}) (*AccessToken, error) {
	switch s := accToken.(type) {
	case *AccessToken:
		return s, nil
	case map[string]interface{}:
		expiDate, err := time.Parse(time.RFC3339, s["expiration_date"].(string))
		if err != nil {
			return nil, erro.Wrap(err)
		}
		return &AccessToken{
			s["id"].(string),
			s["user_uuid"].(string),
			expiDate,
		}, nil
	default:
		return nil, erro.New("unknown type ", accToken, ".")
	}
}

// 有効なアクセストークンを取り出す。
func (sys *system) AccessToken(accTokenId string) (*AccessToken, error) {
	val, err := sys.accTokenCont.Get(accTokenId)
	if err != nil {
		return nil, erro.Wrap(err)
	} else if val == nil {
		return nil, nil
	}
	return recognizeAccessToken(val)
}
