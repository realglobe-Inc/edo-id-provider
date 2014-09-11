package main

import ()

// リクエストが悪い。
type invalidRequest struct {
	Msg string `json:"message"`
}

func (err *invalidRequest) Error() string {
	return err.Msg
}

func newInvalidRequest(msg string) *invalidRequest {
	return &invalidRequest{msg}
}

// セッションが有効でない。
type invalidSession struct{}

func (err *invalidSession) Error() string {
	return "no valid session."
}

func newInvalidSession() *invalidSession {
	return &invalidSession{}
}

// そんなユーザーいない。
type userNotFound struct {
	UsrName string
}

func (err *userNotFound) Error() string {
	return "user " + err.UsrName + " is not exist."
}

func newUserNotFound(usrName string) *userNotFound {
	return &userNotFound{usrName}
}

// パスワードが合ってない。
type invalidPassword struct {
	UsrName string
}

func (err *invalidPassword) Error() string {
	return "wrong password for user " + err.UsrName + "."
}

func newInvalidPassword(usrName string) *invalidPassword {
	return &invalidPassword{usrName}
}

// パスワードが合ってない。
type invalidRedirectUri struct {
	servUuid string
	rediUri  string
}

func (err *invalidRedirectUri) Error() string {
	return err.rediUri + " does not belong " + err.servUuid + "."
}

func newInvalidRedirectUri(servUuid, rediUri string) *invalidRedirectUri {
	return &invalidRedirectUri{servUuid, rediUri}
}
