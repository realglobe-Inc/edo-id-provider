package main

import (
	"net/http"
)

const (
	formSelTic = "ticket"
)

type selectRequest struct {
	browserRequest

	tic     string
	accName string
}

func newSelectRequest(r *http.Request) *selectRequest {
	return &selectRequest{browserRequest: browserRequest{r: r}}
}

func (this *selectRequest) ticket() string {
	if this.tic == "" {
		this.tic = this.r.FormValue(formSelTic)
	}
	return this.tic
}

func (this *selectRequest) selectInfo() (accName string) {
	if this.accName == "" {
		this.accName = this.r.FormValue(formAccName)
	}
	return this.accName
}
