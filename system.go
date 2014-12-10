package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/edo/util"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"gopkg.in/mgo.v2"
	"math/big"
	"strings"
	"time"
)

// 便宜的に集めただけ。
type system struct {
	TaExplorer
	TaKeyProvider
	UserNameIndex
	UserAttributeRegistry
	sessCont     driver.TimeLimitedKeyValueStore
	codeCont     driver.TimeLimitedKeyValueStore
	accTokenCont driver.TimeLimitedKeyValueStore

	maxSessExpiDur     time.Duration // デフォルトかつ最大。
	codeExpiDur        time.Duration
	accTokenExpiDur    time.Duration // デフォルト。
	maxAccTokenExpiDur time.Duration
}

const attrLoginPasswd = "login_password"

func (sys *system) UserPassword(usrUuid string, caStmp *driver.Stamp) (passwd string, newCaStmp *driver.Stamp, err error) {
	value, newCaStmp, err := sys.UserAttributeRegistry.UserAttribute(usrUuid, attrLoginPasswd, caStmp)
	if err != nil {
		return "", nil, erro.Wrap(err)
	} else if value == nil {
		return "", newCaStmp, nil
	}
	return value.(string), newCaStmp, err
}

func (sys *system) UserAttribute(usrUuid, attrName string, caStmp *driver.Stamp) (attr interface{}, newCaStmp *driver.Stamp, err error) {
	// TODO フィルタをハードコードしちゃってる。
	if strings.HasSuffix(attrName, "password") {
		return nil, nil, nil
	}
	return sys.UserAttributeRegistry.UserAttribute(usrUuid, attrName, caStmp)
}

// ユーザー認証済みを示すセッション。
type Session struct {
	Id       string    `json:"id"`
	UsrUuid  string    `json:"user_uuid"`
	ExpiDate time.Time `json:"expiration_date"`
}

const sessIdLen int = 20 // セッション ID の文字数。

func (sys *system) NewSession(usrUuid string, expiDur time.Duration) (sess *Session, newCaStmp *driver.Stamp, err error) {
	var sessId string
	for {
		buff, err := util.SecureRandomBytes(sessIdLen * 6 / 8)
		if err != nil {
			return nil, nil, erro.Wrap(err)
		}

		sessId = base64.StdEncoding.EncodeToString(buff)

		if sess, _, err := sys.sessCont.Get(sessId, nil); err != nil {
			return nil, nil, erro.Wrap(err)
		} else if sess == nil {
			break
		}
	}

	// セッション ID が決まった。
	log.Debug("Session ID was generated.")

	if expiDur == 0 || expiDur > sys.maxSessExpiDur {
		expiDur = sys.maxSessExpiDur
	}

	sess = &Session{sessId, usrUuid, time.Now().Add(expiDur)}
	newCaStmp, err = sys.sessCont.Put(sessId, sess, sess.ExpiDate)
	if err != nil {
		return nil, nil, erro.Wrap(err)
	}

	log.Debug("Session was published.")
	return sess, newCaStmp, nil
}

func sessionUnmarshal(data []byte) (value interface{}, err error) {
	var res Session
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, erro.Wrap(err)
	}
	return &res, nil
}

func sessionMongoTake(query *mgo.Query) (value interface{}, stmp *driver.Stamp, err error) {
	var res struct {
		Value *Session
		Stamp *driver.Stamp
	}
	if err := query.One(&res); err != nil {
		return nil, nil, erro.Wrap(err)
	}
	return res.Value, res.Stamp, nil
}

// 有効なセッションを取り出す。
func (sys *system) Session(sessId string, caStmp *driver.Stamp) (sess *Session, newCaStmp *driver.Stamp, err error) {
	value, newCaStmp, err := sys.sessCont.Get(sessId, caStmp)
	if err != nil {
		return nil, nil, erro.Wrap(err)
	} else if value == nil {
		return nil, newCaStmp, nil
	}
	return value.(*Session), newCaStmp, nil
}

// セッションの有効期限を更新。
func (sys *system) UpdateSession(sess *Session) (newCaStmp *driver.Stamp, err error) {
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

func (sys *system) NewCode(usrUuid, servUuid string) (code *Code, newCaStmp *driver.Stamp, err error) {
	var codeId string
	for {
		codeIdBitLen := codeIdLen * 6
		maxVal := big.NewInt(0).Lsh(big.NewInt(1), uint(codeIdBitLen))
		value, err := rand.Int(rand.Reader, maxVal)
		if err != nil {
			return nil, nil, erro.Wrap(err)
		}
		buff := value.Bytes()

		for len(buff) < (codeIdBitLen+5)/8 {
			buff = append(buff, 0)
		}

		codeId = base64.StdEncoding.EncodeToString(buff)
		if len(codeId) > codeIdLen {
			codeId = codeId[:codeIdLen]
		}

		if code, _, err := sys.codeCont.Get(codeId, nil); err != nil {
			return nil, nil, erro.Wrap(err)
		} else if code == nil {
			break
		}
	}

	// コードが決まった。
	log.Debug("Code was generated.")

	code = &Code{codeId, usrUuid, servUuid, time.Now().Add(sys.codeExpiDur)}
	newCaStmp, err = sys.codeCont.Put(codeId, code, code.ExpiDate)
	if err != nil {
		return nil, nil, erro.Wrap(err)
	}

	log.Debug("Code was published.")
	return code, newCaStmp, nil
}

func codeUnmarshal(data []byte) (value interface{}, err error) {
	var res Code
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, erro.Wrap(err)
	}
	return &res, nil
}

func codeMongoTake(query *mgo.Query) (value interface{}, stmp *driver.Stamp, err error) {
	var res struct {
		Value *Code
		Stamp *driver.Stamp
	}
	if err := query.One(&res); err != nil {
		return nil, nil, erro.Wrap(err)
	}
	return res.Value, res.Stamp, nil
}

// 有効なアクセストークン取得用コードを取り出す。
func (sys *system) Code(codeId string, caStmp *driver.Stamp) (code *Code, newCaStmp *driver.Stamp, err error) {
	value, newCaStmp, err := sys.codeCont.Get(codeId, caStmp)
	if err != nil {
		return nil, nil, erro.Wrap(err)
	} else if value == nil {
		return nil, newCaStmp, nil
	}
	return value.(*Code), newCaStmp, nil
}

// アクセストークン。
type AccessToken struct {
	Id       string    `json:"id"`
	UsrUuid  string    `json:"user_uuid"`
	ExpiDate time.Time `json:"expiration_date"`
}

const accTokenIdLen int = 20 // アクセストークンの文字数。

func (sys *system) NewAccessToken(usrUuid string, expiDur time.Duration) (accToken *AccessToken, newCaStmp *driver.Stamp, err error) {
	var accTokenId string
	for {
		accTokenIdBitLen := accTokenIdLen * 6
		maxVal := big.NewInt(0).Lsh(big.NewInt(1), uint(accTokenIdBitLen))
		value, err := rand.Int(rand.Reader, maxVal)
		if err != nil {
			return nil, nil, erro.Wrap(err)
		}
		buff := value.Bytes()

		for len(buff) < (accTokenIdBitLen+5)/8 {
			buff = append(buff, 0)
		}

		accTokenId = base64.StdEncoding.EncodeToString(buff)
		if len(accTokenId) > accTokenIdLen {
			accTokenId = accTokenId[:accTokenIdLen]
		}

		if accToken, _, err := sys.accTokenCont.Get(accTokenId, nil); err != nil {
			return nil, nil, erro.Wrap(err)
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

	accToken = &AccessToken{accTokenId, usrUuid, time.Now().Add(expiDur)}
	newCaStmp, err = sys.accTokenCont.Put(accTokenId, accToken, accToken.ExpiDate)
	if err != nil {
		return nil, nil, erro.Wrap(err)
	}

	log.Debug("Access token was published.")
	return accToken, newCaStmp, nil
}

func accessTokenUnmarshal(data []byte) (value interface{}, err error) {
	var res AccessToken
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, erro.Wrap(err)
	}
	return &res, nil
}

func accessTokenMongoTake(query *mgo.Query) (value interface{}, stmp *driver.Stamp, err error) {
	var res struct {
		Value *AccessToken
		Stamp *driver.Stamp
	}
	if err := query.One(&res); err != nil {
		return nil, nil, erro.Wrap(err)
	}
	return res.Value, res.Stamp, nil
}

// 有効なアクセストークンを取り出す。
func (sys *system) AccessToken(accTokenId string, caStmp *driver.Stamp) (accToken *AccessToken, newCaStmp *driver.Stamp, err error) {
	value, newCaStmp, err := sys.accTokenCont.Get(accTokenId, caStmp)
	if err != nil {
		return nil, nil, erro.Wrap(err)
	} else if value == nil {
		return nil, newCaStmp, nil
	}
	return value.(*AccessToken), newCaStmp, nil
}
