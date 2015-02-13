package main

import (
	"net/http"
)

type selectRequest struct {
	*browserRequest

	tic     string
	accName string
	loc     string
}

func newSelectRequest(r *http.Request) *selectRequest {
	return &selectRequest{
		browserRequest: newBrowserRequest(r),
		tic:            r.FormValue(formSelTic),
		accName:        r.FormValue(formAccName),
		loc:            r.FormValue(formLoc),
	}
}

func (this *selectRequest) ticket() string {
	return this.tic
}

func (this *selectRequest) selectInfo() (accName string) {
	return this.accName
}

func (this *selectRequest) locale() string {
	return this.loc
}
