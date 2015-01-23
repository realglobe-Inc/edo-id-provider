package main

import (
	"net/http"
)

type loginRequest struct {
	browserRequest

	tic     string
	accName string
	passwd  string
}

func newLoginRequest(r *http.Request) *loginRequest {
	return &loginRequest{browserRequest: browserRequest{r: r}}
}

func (this *loginRequest) ticket() string {
	if this.tic == "" {
		this.tic = this.r.FormValue(formLoginTic)
	}
	return this.tic
}

func (this *loginRequest) loginInfo() (accName, passwd string) {
	if this.accName == "" && this.passwd == "" {
		this.accName = this.r.FormValue(formAccName)
		this.passwd = this.r.FormValue(formPasswd)
	}
	return this.accName, this.passwd
}
