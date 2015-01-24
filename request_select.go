package main

import (
	"net/http"
)

const (
	formSelTic = "ticket"
)

type selectRequest struct {
	*browserRequest

	tic     string
	accName string
}

func newSelectRequest(r *http.Request) *selectRequest {
	return &selectRequest{
		browserRequest: newBrowserRequest(r),
		tic:            r.FormValue(formSelTic),
		accName:        r.FormValue(formAccName),
	}
}

func (this *selectRequest) ticket() string {
	return this.tic
}

func (this *selectRequest) selectInfo() (accName string) {
	return this.accName
}
