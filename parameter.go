package main

import (
	"flag"
	"fmt"
	"github.com/realglobe-Inc/edo/util"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"github.com/realglobe-Inc/go-lib-rg/rglog/level"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type parameters struct {
	// 画面ログ表示重要度。
	consLv level.Level

	// 追加ログ種別。
	logType string
	// 追加ログ表示重要度。
	logLv level.Level
	// ログファイルパス。
	logPath string
	// fluentd アドレス。
	fluAddr string
	// fluentd 用タグ。
	fluTag string

	// ソケット種別。
	socType string
	// UNIX ソケット。
	socPath string
	// TCP ソケット。
	socPort int

	// プロトコル種別。
	protType string

	// Set-Cookie に Secure をつけるか。
	secCook bool

	// キャッシュを最新とみなす期間。
	caStaleDur time.Duration
	// キャッシュを廃棄するまでの期間。
	caExpiDur time.Duration

	// IdP としての ID。
	selfId string
	// 署名用秘密鍵のパス。
	keyPath string
	// 署名用秘密鍵の ID。
	kid string
	// 署名用方式。
	sigAlg string

	// UI 用 HTML を提供する URI。
	uiUri string
	// UI 用 HTML を置くディレクトリパス。
	uiPath string

	// TA 格納庫種別。
	taContType string
	// TA 格納庫ディレクトリパス。
	taContPath string
	// TA 格納庫 mongodb アドレス。
	taContUrl string
	// TA 格納庫 mongodb データベース名。
	taContDb string
	// TA 格納庫 mongodb コレクション名。
	taContColl string

	// アカウント格納庫種別。
	accContType string
	// アカウント格納庫ディレクトリパス。
	accContPath string
	// 名前引きアカウント格納庫ディレクトリパス。
	accNameContPath string
	// アカウント格納庫 mongodb アドレス。
	accContUrl string
	// アカウント格納庫 mongodb データベース名。
	accContDb string
	// アカウント格納庫 mongodb コレクション名。
	accContColl string

	// 同意格納庫種別。
	consContType string
	// 同意格納庫ディレクトリパス。
	consContPath string
	// 名前引き同意格納庫ディレクトリパス。
	consNameContPath string
	// 同意格納庫 mongodb アドレス。
	consContUrl string
	// 同意格納庫 mongodb データベース名。
	consContDb string
	// 同意格納庫 mongodb コレクション名。
	consContColl string

	// セッション番号の文字数。
	sessIdLen int
	// セッションの有効期間。
	sessExpiDur time.Duration
	// セッション格納庫種別。
	sessContType string
	// セッション格納庫ディレクトリパス。
	sessContPath string
	// セッション期限格納庫ディレクトリパス。
	sessExpiContPath string
	// セッション格納庫 redis アドレス。
	sessContUrl string
	// セッション格納庫 redis キー接頭辞。
	sessContPrefix string

	// 認可コードの文字数。
	codIdLen int
	// 認可コードの有効期間。
	codExpiDur time.Duration
	// 認可コード格納庫種別。
	codContType string
	// 認可コード格納庫ディレクトリパス。
	codContPath string
	// 認可コード期限格納庫ディレクトリパス。
	codExpiContPath string
	// 認可コード格納庫 redis アドレス。
	codContUrl string
	// 認可コード格納庫 redis キー接頭辞。
	codContPrefix string

	// アクセストークンの文字数。
	tokIdLen int
	// アクセストークンの有効期間。
	tokExpiDur time.Duration
	// アクセストークン格納庫種別。
	tokContType string
	// アクセストークン格納庫ディレクトリパス。
	tokContPath string
	// アクセストークン期限格納庫ディレクトリパス。
	tokExpiContPath string
	// アクセストークン格納庫 redis アドレス。
	tokContUrl string
	// アクセストークン格納庫 redis キー接頭辞。
	tokContPrefix string

	// ID トークンの有効期間。
	idTokExpiDur time.Duration
}

func parseParameters(args ...string) (param *parameters, err error) {

	const label = "edo-id-provider"

	flags := util.NewFlagSet(label+" parameters", flag.ExitOnError)
	flags.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  "+args[0]+" [{FLAG}...]")
		fmt.Fprintln(os.Stderr, "FLAG:")
		flags.PrintDefaults()
	}

	param = &parameters{}

	flags.Var(level.Var(&param.consLv, level.INFO), "consLv", "Console log level.")
	flags.StringVar(&param.logType, "logType", "", "Extra log type.")
	flags.Var(level.Var(&param.logLv, level.ALL), "logLv", "Extra log level.")
	flags.StringVar(&param.logPath, "logPath", filepath.Join(filepath.Dir(os.Args[0]), "log", label+".log"), "File log path.")
	flags.StringVar(&param.fluAddr, "fluAddr", "localhost:24224", "fluentd address.")
	flags.StringVar(&param.fluTag, "fluTag", "edo."+label, "fluentd tag.")

	flags.StringVar(&param.socType, "socType", "tcp", "Socket type.")
	flags.StringVar(&param.socPath, "socPath", filepath.Join(filepath.Dir(os.Args[0]), "run", label+".soc"), "UNIX socket path.")
	flags.IntVar(&param.socPort, "socPort", 16040, "TCP socket port.")

	flags.StringVar(&param.protType, "protType", "http", "Protocol type.")

	flags.BoolVar(&param.secCook, "secCook", true, "Add Secure in Set-Cookie.")

	flags.DurationVar(&param.caStaleDur, "caStaleDur", 5*time.Minute, "Cache fresh duration.")
	flags.DurationVar(&param.caExpiDur, "caExpiDur", 30*time.Minute, "Cache expiration duration.")

	flags.StringVar(&param.selfId, "selfId", "https://example.com", "IdP ID.")
	flags.StringVar(&param.keyPath, "keyPath", filepath.Join(filepath.Dir(os.Args[0]), "private.key"), "Private key for signature.")
	flags.StringVar(&param.kid, "kid", "", "Private key ID.")
	flags.StringVar(&param.sigAlg, "sigAlg", "RS256", "Signature algorithm.")

	flags.StringVar(&param.uiUri, "uiUri", "/html", "UI uri.")
	flags.StringVar(&param.uiPath, "uiPath", filepath.Join(filepath.Dir(os.Args[0]), "html"), "Protocol type. http/fcgi")

	flags.StringVar(&param.taContType, "taContType", "file", "TA container type.")
	flags.StringVar(&param.taContPath, "taContPath", filepath.Join(filepath.Dir(os.Args[0]), "tas"), "TA container directory.")
	flags.StringVar(&param.taContUrl, "taContUrl", "localhost", "TA container address.")
	flags.StringVar(&param.taContDb, "taContDb", "edo", "TA container database name.")
	flags.StringVar(&param.taContColl, "taContColl", "tas", "TA container collection name.")

	flags.StringVar(&param.accContType, "accContType", "file", "Account container type.")
	flags.StringVar(&param.accContPath, "accContPath", filepath.Join(filepath.Dir(os.Args[0]), "accounts"), "Account container directory.")
	flags.StringVar(&param.accNameContPath, "accNameContPath", filepath.Join(filepath.Dir(os.Args[0]), "account_names"), "Name indexed account container directory.")
	flags.StringVar(&param.accContUrl, "accContUrl", "localhost", "Account container address.")
	flags.StringVar(&param.accContDb, "accContDb", "edo", "Account container database name.")
	flags.StringVar(&param.accContColl, "accContColl", "accounts", "Account container collection name.")

	flags.StringVar(&param.consContType, "consContType", "file", "Consent container type.")
	flags.StringVar(&param.consContPath, "consContPath", filepath.Join(filepath.Dir(os.Args[0]), "consents"), "Consent container directory.")
	flags.StringVar(&param.consContUrl, "consContUrl", "localhost", "Consent container address.")
	flags.StringVar(&param.consContDb, "consContDb", "edo", "Consent container database name.")
	flags.StringVar(&param.consContColl, "consContColl", "consents", "Consent container collection name.")

	flags.IntVar(&param.sessIdLen, "sessIdLen", 40, "Session ID length.")
	flags.DurationVar(&param.sessExpiDur, "sessExpiDur", 7*24*time.Hour, "Session expiration duration.")
	flags.StringVar(&param.sessContType, "sessContType", "memory", "Session container type.")
	flags.StringVar(&param.sessContPath, "sessContPath", filepath.Join(filepath.Dir(os.Args[0]), "sessions"), "Session container directory.")
	flags.StringVar(&param.sessExpiContPath, "sessExpiContPath", filepath.Join(filepath.Dir(os.Args[0]), "session_expires"), "Session container expiration date directory.")
	flags.StringVar(&param.sessContUrl, "sessContUrl", "localhost", "Session container address.")
	flags.StringVar(&param.sessContPrefix, "sessContPrefix", "edo.sessions", "Session container key prefix.")

	flags.IntVar(&param.codIdLen, "codIdLen", 40, "Code length.")
	flags.DurationVar(&param.codExpiDur, "codExpiDur", 3*time.Minute, "Code expiration duration.")
	flags.StringVar(&param.codContType, "codContType", "memory", "Code container type.")
	flags.StringVar(&param.codContPath, "codContPath", filepath.Join(filepath.Dir(os.Args[0]), "codes"), "Code container directory.")
	flags.StringVar(&param.codExpiContPath, "codExpiContPath", filepath.Join(filepath.Dir(os.Args[0]), "code_expires"), "Code container expiration date directory.")
	flags.StringVar(&param.codContUrl, "codContUrl", "localhost", "Code container address.")
	flags.StringVar(&param.codContPrefix, "codContPrefix", "edo.codes", "Code container key prefix.")

	flags.IntVar(&param.tokIdLen, "tokIdLen", 40, "Token length.")
	flags.DurationVar(&param.tokExpiDur, "tokExpiDur", 24*time.Hour, "Token expiration duration.")
	flags.StringVar(&param.tokContType, "tokContType", "memory", "Token container type.")
	flags.StringVar(&param.tokContPath, "tokContPath", filepath.Join(filepath.Dir(os.Args[0]), "tokens"), "Token container directory.")
	flags.StringVar(&param.tokExpiContPath, "tokExpiContPath", filepath.Join(filepath.Dir(os.Args[0]), "token_expires"), "Token container expiration date directory.")
	flags.StringVar(&param.tokContUrl, "tokContUrl", "localhost", "Token container address.")
	flags.StringVar(&param.tokContPrefix, "tokContPrefix", "edo.tokens", "Token container key prefix.")

	flags.DurationVar(&param.idTokExpiDur, "idTokExpiDur", 10*time.Minute, "ID token expiration duration.")

	var config string
	flags.StringVar(&config, "f", "", "Config file path.")

	// 実行引数を読んで、設定ファイルを指定させてから、
	// 設定ファイルを読んで、また実行引数を読む。
	flags.Parse(args[1:])
	if config != "" {
		if buff, err := ioutil.ReadFile(config); err != nil {
			if !os.IsNotExist(err) {
				return nil, erro.Wrap(err)
			}
			log.Warn("Config file " + config + " is not exist.")
		} else {
			flags.CompleteParse(strings.Fields(string(buff)))
		}
	}
	flags.Parse(args[1:])

	if l := len(flags.Args()); l > 0 {
		log.Warn("Ignore extra parameters ", flags.Args(), ".")
	}

	// uiUri を整形。
	uiUri := strings.TrimRight(param.uiUri, "/")
	uiUri = regexp.MustCompile("/+").ReplaceAllString(uiUri, "/")
	if uiUri == "" {
		uiUri = "/html"
	}
	if uiUri[0] != '/' {
		uiUri = "/" + uiUri
	}
	if param.uiUri != uiUri {
		log.Info("Use " + uiUri + " as UI uri")
		param.uiUri = uiUri
	}

	return param, nil
}
