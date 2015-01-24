package main

import (
	"net/http"
)

type loginRequest struct {
	*browserRequest

	tic     string
	accName string
	passwd  string
}

func newLoginRequest(r *http.Request) *loginRequest {
	return &loginRequest{
		browserRequest: newBrowserRequest(r),
		tic:            r.FormValue(formLoginTic),
		accName:        r.FormValue(formAccName),
		passwd:         r.FormValue(formPasswd),
	}
}

func (this *loginRequest) ticket() string {
	return this.tic
}

func (this *loginRequest) loginInfo() (accName, passwd string) {
	return this.accName, this.passwd
}
