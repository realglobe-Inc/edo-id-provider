package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func TestClaimRequest(t *testing.T) {
	req := claimRequest{
		"name":     {"ja": {true, nil, nil}},
		"nickname": {"": {false, nil, nil}},
	}

	buff, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	var req2 claimRequest
	if err := json.Unmarshal(buff, &req2); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(req2, req) {
		t.Error(fmt.Sprintf("%#v", req2))
		t.Error(fmt.Sprintf("%#v", req))
		t.Error(string(buff))
	}
}
