// Copyright 2015 realglobe, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"github.com/realglobe-Inc/go-lib/erro"
	"github.com/realglobe-Inc/go-lib/rglog/level"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type parameters struct {
	// 画面ログ。
	consLv level.Level
	// 追加ログ。
	logType string
	logLv   level.Level
	// ファイルログ。
	logPath string
	logSize int64
	logNum  int
	// fluentd ログ。
	logAddr string
	logTag  string

	// ソケット。
	socType string
	// UNIX ソケット。
	socPath string
	// TCP ソケット。
	socPort int
	// プロトコル。
	protType string

	// IdP としての ID。
	selfId string
	// 署名方式。
	sigAlg string
	// 署名鍵の ID。
	sigKid string

	// URI
	pathOk     string
	pathAuth   string
	pathSel    string
	pathLgin   string
	pathCons   string
	pathTa     string
	pathTok    string
	pathAcnt   string
	pathCoopFr string
	pathCoopTo string
	// UI 用 HTML を提供する URI。
	pathUi     string
	pathSelUi  string
	pathLginUi string
	pathConsUi string
	// UI 用 HTML を置くディレクトリパス。
	uiDir string

	tmplErr string

	// セクタ固有のアカウント ID の計算に使う情報。
	pwSaltLen int
	// セッション。
	sessLabel    string
	sessLen      int
	sessExpIn    time.Duration
	sessRefDelay time.Duration
	sessDbExpIn  time.Duration
	// 認可コード。
	acodLen     int
	acodExpIn   time.Duration
	acodDbExpIn time.Duration
	// アクセストークン。
	tokLen     int
	tokExpIn   time.Duration
	tokDbExpIn time.Duration
	// 仲介コード。
	ccodLen     int
	ccodExpIn   time.Duration
	ccodDbExpIn time.Duration
	// JWT の ID (jti)。
	jtiLen     int
	jtiExpIn   time.Duration
	jtiDbExpIn time.Duration
	// チケット。
	ticLen   int
	ticExpIn time.Duration

	// バックエンドの指定。

	// redis
	redTimeout   time.Duration
	redPoolSize  int
	redPoolExpIn time.Duration
	// mongodb
	monTimeout time.Duration

	// 鍵 DB。
	keyDbType  string
	keyDbPath  string
	keyDbAddr  string
	keyDbTag   string
	keyDbExpIn time.Duration

	// web データ DB。
	webDbType  string
	webDbAddr  string
	webDbTag   string
	webDbExpIn time.Duration

	// アカウント情報 DB。
	acntDbType string
	acntDbAddr string
	acntDbTag  string
	acntDbTag2 string

	// スコープ・属性許可情報 DB。
	consDbType string
	consDbAddr string
	consDbTag  string
	consDbTag2 string

	// TA 情報 DB。
	taDbType string
	taDbAddr string
	taDbTag  string
	taDbTag2 string

	// セクタ固有のアカウント ID の計算に使う情報の DB。
	sectDbType string
	sectDbAddr string
	sectDbTag  string
	sectDbTag2 string

	// セクタ固有のアカウント ID 情報 DB。
	pwDbType string
	pwDbAddr string
	pwDbTag  string
	pwDbTag2 string

	// IdP 情報 DB。
	idpDbType string
	idpDbAddr string
	idpDbTag  string
	idpDbTag2 string

	// セッション DB。
	sessDbType string
	sessDbAddr string
	sessDbTag  string

	// 認可コード DB。
	acodDbType string
	acodDbAddr string
	acodDbTag  string

	// アクセストークン DB。
	tokDbType string
	tokDbAddr string
	tokDbTag  string

	// 仲介コード DB。
	ccodDbType string
	ccodDbAddr string
	ccodDbTag  string

	// JWT の ID の DB。
	jtiDbType string
	jtiDbAddr string
	jtiDbTag  string

	// その他のオプション。

	// Set-Cookie の Path。
	cookPath string
	// Set-Cookie を Secure にするか。
	cookSec bool

	debug bool
	// テスト用。
	shutCh chan struct{}
}

func parseParameters(args ...string) (param *parameters, err error) {

	const label = "edo-id-provider"

	flags := flag.NewFlagSet(label+" parameters", flag.ExitOnError)
	flags.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  "+args[0]+" [{FLAG}...]")
		fmt.Fprintln(os.Stderr, "FLAG:")
		flags.PrintDefaults()
	}

	param = &parameters{}

	flags.Var(level.Var(&param.consLv, level.INFO), "consLv", "Console log level")
	flags.StringVar(&param.logType, "logType", "", "Extra log: Type")
	flags.Var(level.Var(&param.logLv, level.ALL), "logLv", "Extra log: Level")
	flags.StringVar(&param.logPath, "logPath", filepath.Join(filepath.Dir(os.Args[0]), "log", label+".log"), "Extra log: File path")
	flags.Int64Var(&param.logSize, "logSize", 10*(1<<20) /* 10 MB */, "Extra log: File size limit")
	flags.IntVar(&param.logNum, "logNum", 10, "Extra log: File number limit")
	flags.StringVar(&param.logAddr, "logAddr", "localhost:24224", "Extra log: Fluentd address")
	flags.StringVar(&param.logTag, "logTag", label, "Extra log: Fluentd tag")

	flags.StringVar(&param.socType, "socType", "tcp", "Socket type")
	flags.StringVar(&param.socPath, "socPath", filepath.Join(filepath.Dir(os.Args[0]), "run", label+".soc"), "Unix socket path")
	flags.IntVar(&param.socPort, "socPort", 1604, "TCP socket port")
	flags.StringVar(&param.protType, "protType", "http", "Protocol type")

	flags.StringVar(&param.selfId, "selfId", "https://idp.example.org", "IdP ID")
	flags.StringVar(&param.sigAlg, "sigAlg", "RS256", "Signature algorithm")
	flags.StringVar(&param.sigKid, "sigKid", "", "Signature key ID")

	flags.StringVar(&param.pathOk, "pathOk", "/ok", "OK URI")
	flags.StringVar(&param.pathAuth, "pathAuth", "/auth", "Authentication URI")
	flags.StringVar(&param.pathSel, "pathSel", "/auth/select", "Account select URI")
	flags.StringVar(&param.pathLgin, "pathLgin", "/auth/login", "Login URI")
	flags.StringVar(&param.pathCons, "pathCons", "/auth/consent", "Consent URI")
	flags.StringVar(&param.pathTa, "pathTa", "/api/info/ta", "TA info URI")
	flags.StringVar(&param.pathTok, "pathTok", "/api/token", "Token URI")
	flags.StringVar(&param.pathAcnt, "pathAcnt", "/api/info/account", "Account info URI")
	flags.StringVar(&param.pathCoopFr, "pathCoopFr", "/api/coop/from", "Cooperation from URI")
	flags.StringVar(&param.pathCoopTo, "pathCoopTo", "/api/coop/to", "Cooperation to URI")
	flags.StringVar(&param.pathUi, "pathUi", "/ui", "UI URI")
	flags.StringVar(&param.pathSelUi, "pathSelUi", "/ui/select.html", "Account selection UI URI")
	flags.StringVar(&param.pathLginUi, "pathLginUi", "/ui/login.html", "Login UI URI")
	flags.StringVar(&param.pathConsUi, "pathConsUi", "/ui/consent.html", "Consent UI URI")
	flags.StringVar(&param.uiDir, "uiDir", "", "UI file directory")

	flags.StringVar(&param.tmplErr, "tmplErr", "", "Error UI template")

	flags.IntVar(&param.pwSaltLen, "pwSaltLen", 20, "Pairwise account ID calculation salt length")
	flags.StringVar(&param.sessLabel, "sessLabel", "Id-Provider", "Session ID label")
	flags.IntVar(&param.sessLen, "sessLen", 30, "Session ID length")
	flags.DurationVar(&param.sessExpIn, "sessExpIn", 7*24*time.Hour, "Session expiration duration")
	flags.DurationVar(&param.sessRefDelay, "sessRefDelay", 24*time.Hour, "Session refresh delay")
	flags.DurationVar(&param.sessDbExpIn, "sessDbExpIn", 14*24*time.Hour, "Session keep duration")
	flags.IntVar(&param.acodLen, "acodLen", 30, "Authorization code length")
	flags.DurationVar(&param.acodExpIn, "acodExpIn", 3*time.Minute, "Authorization code expiration duration")
	flags.DurationVar(&param.acodDbExpIn, "acodDbExpIn", time.Hour, "Authorization code keep duration")
	flags.IntVar(&param.tokLen, "tokLen", 30, "Access token length")
	flags.DurationVar(&param.tokExpIn, "tokExpIn", time.Hour, "Access token expiration duration")
	flags.DurationVar(&param.tokDbExpIn, "tokDbExpIn", 24*time.Hour, "Access token keep duration")
	flags.IntVar(&param.ccodLen, "ccodLen", 30, "Cooperation code length")
	flags.DurationVar(&param.ccodExpIn, "ccodExpIn", 10*time.Minute, "Cooperation code expiration duration")
	flags.DurationVar(&param.ccodDbExpIn, "ccodDbExpIn", time.Hour, "Cooperation code keep duration")
	flags.IntVar(&param.jtiLen, "jtiLen", 20, "JWT ID length")
	flags.DurationVar(&param.jtiExpIn, "jtiExpIn", 6*time.Hour, "JWT expiration duration")
	flags.DurationVar(&param.jtiDbExpIn, "jtiDbExpIn", 24*time.Hour, "JWT ID default keep duration")
	flags.IntVar(&param.ticLen, "ticLen", 10, "Ticket length")
	flags.DurationVar(&param.ticExpIn, "ticExpIn", 30*time.Minute, "Ticket expiration duration")

	flags.DurationVar(&param.redTimeout, "redTimeout", 30*time.Second, "redis timeout duration")
	flags.IntVar(&param.redPoolSize, "redPoolSize", 10, "redis pool size")
	flags.DurationVar(&param.redPoolExpIn, "redPoolExpIn", time.Minute, "redis connection keep duration")
	flags.DurationVar(&param.monTimeout, "monTimeout", 30*time.Second, "mongodb timeout duration")

	flags.StringVar(&param.keyDbType, "keyDbType", "redis", "Key DB type")
	flags.StringVar(&param.keyDbPath, "keyDbPath", filepath.Join(filepath.Dir(os.Args[0]), "key"), "Key DB directory")
	flags.StringVar(&param.keyDbAddr, "keyDbAddr", "localhost:6379", "Key DB address")
	flags.StringVar(&param.keyDbTag, "keyDbTag", "key", "Key DB tag")
	flags.DurationVar(&param.keyDbExpIn, "keyDbExpIn", 5*time.Minute, "Key DB expiration duration")

	flags.StringVar(&param.webDbType, "webDbType", "redis", "Web data DB type")
	flags.StringVar(&param.webDbAddr, "webDbAddr", "localhost:6379", "Web data DB address")
	flags.StringVar(&param.webDbTag, "webDbTag", "web", "Web data DB tag")
	flags.DurationVar(&param.webDbExpIn, "webDbExpIn", 7*24*time.Hour, "Web data keep duration")

	flags.StringVar(&param.acntDbType, "acntDbType", "mongo", "Account DB type")
	flags.StringVar(&param.acntDbAddr, "acntDbAddr", "localhost", "Account DB address")
	flags.StringVar(&param.acntDbTag, "acntDbTag", "edo", "Account DB tag")
	flags.StringVar(&param.acntDbTag2, "acntDbTag2", "account", "Account DB sub tag")

	flags.StringVar(&param.consDbType, "consDbType", "mongo", "Consent DB type")
	flags.StringVar(&param.consDbAddr, "consDbAddr", "localhost", "Consent DB address")
	flags.StringVar(&param.consDbTag, "consDbTag", "edo", "Consent DB tag")
	flags.StringVar(&param.consDbTag2, "consDbTag2", "consent", "Consent DB sub tag")

	flags.StringVar(&param.taDbType, "taDbType", "mongo", "TA DB type")
	flags.StringVar(&param.taDbAddr, "taDbAddr", "localhost", "TA DB address")
	flags.StringVar(&param.taDbTag, "taDbTag", "edo", "TA DB tag")
	flags.StringVar(&param.taDbTag2, "taDbTag2", "ta", "TA DB sub tag")

	flags.StringVar(&param.sectDbType, "sectDbType", "mongo", "Pairwise account ID calculation info DB type")
	flags.StringVar(&param.sectDbAddr, "sectDbAddr", "localhost", "Pairwise account ID calculation info DB address")
	flags.StringVar(&param.sectDbTag, "sectDbTag", "edo", "Pairwise account ID calculation info DB tag")
	flags.StringVar(&param.sectDbTag2, "sectDbTag2", "sector", "Pairwise account ID calculation info DB sub tag")

	flags.StringVar(&param.pwDbType, "pwDbType", "mongo", "Pairwise account ID DB type")
	flags.StringVar(&param.pwDbAddr, "pwDbAddr", "localhost", "Pairwise account ID DB address")
	flags.StringVar(&param.pwDbTag, "pwDbTag", "edo", "Pairwise account ID DB tag")
	flags.StringVar(&param.pwDbTag2, "pwDbTag2", "pairwise", "Pairwise account ID DB sub tag")

	flags.StringVar(&param.idpDbType, "idpDbType", "mongo", "IdP DB type")
	flags.StringVar(&param.idpDbAddr, "idpDbAddr", "localhost", "IdP DB address")
	flags.StringVar(&param.idpDbTag, "idpDbTag", "edo", "IdP DB tag")
	flags.StringVar(&param.idpDbTag2, "idpDbTag2", "idp", "IdP DB sub tag")

	flags.StringVar(&param.sessDbType, "sessDbType", "redis", "Session DB type")
	flags.StringVar(&param.sessDbAddr, "sessDbAddr", "localhost:6379", "Session DB address")
	flags.StringVar(&param.sessDbTag, "sessDbTag", "session", "Session DB tag")

	flags.StringVar(&param.acodDbType, "acodDbType", "redis", "Authorization code DB type")
	flags.StringVar(&param.acodDbAddr, "acodDbAddr", "localhost:6379", "Authorization code DB address")
	flags.StringVar(&param.acodDbTag, "acodDbTag", "authcode", "Authorization code DB tag")

	flags.StringVar(&param.tokDbType, "tokDbType", "redis", "Access token DB type")
	flags.StringVar(&param.tokDbAddr, "tokDbAddr", "localhost:6379", "Access token DB address")
	flags.StringVar(&param.tokDbTag, "tokDbTag", "token", "Access token DB tag")

	flags.StringVar(&param.ccodDbType, "ccodDbType", "redis", "Cooperation code DB type")
	flags.StringVar(&param.ccodDbAddr, "ccodDbAddr", "localhost:6379", "Cooperation code DB address")
	flags.StringVar(&param.ccodDbTag, "ccodDbTag", "coopcode", "Cooperation code DB tag")

	flags.StringVar(&param.jtiDbType, "jtiDbType", "redis", "JWT ID DB type")
	flags.StringVar(&param.jtiDbAddr, "jtiDbAddr", "localhost:6379", "JWT ID DB address")
	flags.StringVar(&param.jtiDbTag, "jtiDbTag", "jti", "JWT ID DB tag")

	flags.StringVar(&param.cookPath, "cookPath", "/", "Path in Set-Cookie")
	flags.BoolVar(&param.cookSec, "cookSec", true, "Secure flag in Set-Cookie")
	flags.BoolVar(&param.debug, "debug", false, "Debug mode")

	var config string
	flags.StringVar(&config, "c", "", "Config file path")

	// 実行引数を読んで、設定ファイルを指定させてから、
	// 設定ファイルを読んで、また実行引数を読む。
	flags.Parse(args[1:])
	if config != "" {
		if buff, err := ioutil.ReadFile(config); err != nil {
			if !os.IsNotExist(err) {
				return nil, erro.Wrap(err)
			}
			log.Warn("Config file " + config + " is not exist")
		} else {
			flags.Parse(strings.Fields(string(buff)))
		}
	}
	flags.Parse(args[1:])

	if l := len(flags.Args()); l > 0 {
		log.Warn("Ignore extra parameters ", flags.Args())
	}

	return param, nil
}

func (param *parameters) LogFilePath() string       { return param.logPath }
func (param *parameters) LogFileLimit() int64       { return param.logSize }
func (param *parameters) LogFileNumber() int        { return param.logNum }
func (param *parameters) LogFluentdAddress() string { return param.logAddr }
func (param *parameters) LogFluentdTag() string     { return param.logTag }

func (param *parameters) SocketType() string   { return param.socType }
func (param *parameters) SocketPort() int      { return param.socPort }
func (param *parameters) SocketPath() string   { return param.socPath }
func (param *parameters) ProtocolType() string { return param.protType }

// テスト用。
// 使うときは手動で param.shutCh = make(chan struct{}, 5) する。
func (param *parameters) ShutdownChannel() chan struct{} { return param.shutCh }
