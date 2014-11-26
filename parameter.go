package main

import (
	"flag"
	"fmt"
	"github.com/realglobe-Inc/edo/util"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"github.com/realglobe-Inc/go-lib-rg/file"
	"github.com/realglobe-Inc/go-lib-rg/rglog/level"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type parameters struct {
	// 画面表示ログ。
	consLv level.Level

	// 追加ログ。
	logType string
	logLv   level.Level

	// ファイルログ。
	logPath string

	// fluentd ログ。
	fluAddr string
	fluTag  string

	// サービス検索。
	taExpType string

	// ファイルベースサービス検索。
	taExpPath string

	// Web ベースサービス検索。
	taExpAddr string

	// mongo サービス検索。
	taExpUrl  string
	taExpDb   string
	taExpColl string

	// 公開鍵レジストリ。
	taKeyRegType string

	// ファイルベース公開鍵レジストリ。
	taKeyRegPath string

	// Web ベース公開鍵レジストリ。
	taKeyRegAddr string

	// mongo 公開鍵レジストリ。
	taKeyRegUrl  string
	taKeyRegDb   string
	taKeyRegColl string

	// ユーザー名索引。
	usrNameIdxType string

	// ファイルベースユーザー名索引。
	usrNameIdxPath string

	// mongo ユーザー名索引。
	usrNameIdxUrl  string
	usrNameIdxDb   string
	usrNameIdxColl string

	// ユーザー属性レジストリ。
	usrAttrRegType string

	// ファイルベースユーザー属性レジストリ。
	usrAttrRegPath string

	// mongo ユーザー属性レジストリ。
	usrAttrRegUrl  string
	usrAttrRegDb   string
	usrAttrRegColl string

	// セッション管理。
	sessContType string

	// ファイルベースセッション管理。
	sessContPath string

	// mongo セッション管理。
	sessContUrl  string
	sessContDb   string
	sessContColl string

	// アクセストークン発行用コード管理。
	codeContType string

	// ファイルベースアクセストークン発行用コード管理。
	codeContPath string

	// mongo アクセストークン発行用コード管理。
	codeContUrl  string
	codeContDb   string
	codeContColl string

	// アクセストークン管理。
	accTokenContType string

	// ファイルベースアクセストークン管理。
	accTokenContPath string

	// mongo アクセストークン管理。
	accTokenContUrl  string
	accTokenContDb   string
	accTokenContColl string

	// ソケット。
	socType string

	// UNIX ソケット。
	socPath string

	// TCP ソケット。
	socPort int

	// プロトコル。
	protType string

	// 無通信での認証済みセッションの有効期間。
	maxSessExpiDur time.Duration // デフォルトかつ最大。

	// アクセストークン発行用コードの有効期間。
	codeExpiDur time.Duration

	// アクセストークンの有効期間。
	accTokenExpiDur    time.Duration // デフォルト。
	maxAccTokenExpiDur time.Duration
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

	flags.StringVar(&param.taExpType, "taExpType", "web", "TA explorer type.")
	flags.StringVar(&param.taExpPath, "taExpPath", filepath.Join(filepath.Dir(os.Args[0]), "ta_uris"), "TA explorer directory.")
	flags.StringVar(&param.taExpAddr, "taExpAddr", "http://localhost:16034", "TA explorer address.")
	flags.StringVar(&param.taExpUrl, "taExpUrl", "localhost", "TA explorer address.")
	flags.StringVar(&param.taExpDb, "taExpDb", "edo", "TA explorer database name.")
	flags.StringVar(&param.taExpColl, "taExpColl", "ta_uris", "TA explorer collection name.")

	flags.StringVar(&param.taKeyRegType, "taKeyRegType", "web", "TA key provider type.")
	flags.StringVar(&param.taKeyRegPath, "taKeyRegPath", filepath.Join(filepath.Dir(os.Args[0]), "public_keys"), "TA key provider directory.")
	flags.StringVar(&param.taKeyRegAddr, "taKeyRegAddr", "http://localhost:16033", "TA key provider address.")
	flags.StringVar(&param.taKeyRegUrl, "taKeyRegUrl", "localhost", "TA key provider address.")
	flags.StringVar(&param.taKeyRegDb, "taKeyRegDb", "edo", "TA key provider database name.")
	flags.StringVar(&param.taKeyRegColl, "taKeyRegColl", "public_keys", "TA key provider collection name.")

	flags.StringVar(&param.usrNameIdxType, "usrNameIdxType", "mongo", "Username index type.")
	flags.StringVar(&param.usrNameIdxPath, "usrNameIdxPath", filepath.Join(filepath.Dir(os.Args[0]), "user_ids"), "Username index directory.")
	flags.StringVar(&param.usrNameIdxUrl, "usrNameIdxUrl", "localhost", "Username index address.")
	flags.StringVar(&param.usrNameIdxDb, "usrNameIdxDb", "edo", "Username index database name.")
	flags.StringVar(&param.usrNameIdxColl, "usrNameIdxColl", "user_ids", "Username index collection name.")

	flags.StringVar(&param.usrAttrRegType, "usrAttrRegType", "mongo", "User attribute registry type.")
	flags.StringVar(&param.usrAttrRegPath, "usrAttrRegPath", filepath.Join(filepath.Dir(os.Args[0]), "user_attributes"), "User attribute registry directory.")
	flags.StringVar(&param.usrAttrRegUrl, "usrAttrRegUrl", "localhost", "User attribute registry address.")
	flags.StringVar(&param.usrAttrRegDb, "usrAttrRegDb", "edo", "User attribute registry database name.")
	flags.StringVar(&param.usrAttrRegColl, "usrAttrRegColl", "user_attributes", "User attribute registry collection name.")

	flags.StringVar(&param.sessContType, "sessContType", "mongo", "Session container lister type.")
	flags.StringVar(&param.sessContPath, "sessContPath", filepath.Join(filepath.Dir(os.Args[0]), "sessions"), "Session container lister directory.")
	flags.StringVar(&param.sessContUrl, "sessContUrl", "localhost", "Session container lister address.")
	flags.StringVar(&param.sessContDb, "sessContDb", "edo", "Session container lister database name.")
	flags.StringVar(&param.sessContColl, "sessContColl", "sessions", "Session container lister collection name.")

	flags.StringVar(&param.codeContType, "codeContType", "mongo", "Code container lister type.")
	flags.StringVar(&param.codeContPath, "codeContPath", filepath.Join(filepath.Dir(os.Args[0]), "access_codes"), "Code container lister directory.")
	flags.StringVar(&param.codeContUrl, "codeContUrl", "localhost", "Code container lister address.")
	flags.StringVar(&param.codeContDb, "codeContDb", "edo", "Code container lister database name.")
	flags.StringVar(&param.codeContColl, "codeContColl", "access_codes", "Code container lister collection name.")

	flags.StringVar(&param.accTokenContType, "accTokenContType", "mongo", "Access token container lister type.")
	flags.StringVar(&param.accTokenContPath, "accTokenContPath", filepath.Join(filepath.Dir(os.Args[0]), "access_tokens"), "Access token container lister directory.")
	flags.StringVar(&param.accTokenContUrl, "accTokenContUrl", "localhost", "Access token container lister address.")
	flags.StringVar(&param.accTokenContDb, "accTokenContDb", "edo", "Access token container lister database name.")
	flags.StringVar(&param.accTokenContColl, "accTokenContColl", "access_tokens", "Access token container lister collection name.")

	flags.StringVar(&param.socType, "socType", "tcp", "Socket type.")
	flags.StringVar(&param.socPath, "socPath", filepath.Join(filepath.Dir(os.Args[0]), "run", label+".soc"), "UNIX socket path.")
	flags.IntVar(&param.socPort, "socPort", 16040, "TCP socket port.")

	flags.StringVar(&param.protType, "protType", "http", "Protocol type.")

	flags.DurationVar(&param.maxSessExpiDur, "maxSessExpiDur", 24*time.Hour, "Max session expiration duration.")
	flags.DurationVar(&param.codeExpiDur, "codeExpiDur", 10*time.Minute, "Code expiration duration.")
	flags.DurationVar(&param.accTokenExpiDur, "accTokenExpiDur", 24*time.Hour, "Default access token expiration duration.")
	flags.DurationVar(&param.maxAccTokenExpiDur, "maxAccTokenExpiDur", 30*24*time.Hour, "Max access token expiration duration.")

	var config string
	flags.StringVar(&config, "f", "", "Config file path.")

	// 実行引数を読んで、設定ファイルを指定させてから、
	// 設定ファイルを読んで、また実行引数を読む。
	flags.Parse(args[1:])
	if config != "" {
		if exist, err := file.IsExist(config); err != nil {
			return nil, erro.Wrap(err)
		} else if !exist {
			log.Warn("Config file " + config + " is not exist.")
		} else {
			buff, err := ioutil.ReadFile(config)
			if err != nil {
				return nil, erro.Wrap(err)
			}
			flags.CompleteParse(strings.Fields(string(buff)))
		}
	}
	flags.Parse(args[1:])

	if l := len(flags.Args()); l > 0 {
		log.Warn("Ignore extra parameters ", flags.Args(), ".")
	}

	return param, nil
}
