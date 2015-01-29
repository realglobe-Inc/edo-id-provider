package main

import ()

var scopToClms = map[string]map[string]bool{
	"profile": map[string]bool{
		"name":               true,
		"family_name":        true,
		"given_name":         true,
		"middle_name":        true,
		"nickname":           true,
		"preferred_username": true,
		"profile":            true,
		"picture":            true,
		"website":            true,
		"gender":             true,
		"birthdate":          true,
		"zoneinfo":           true,
		"locale":             true,
		"updated_at":         true,
	},
	"email": map[string]bool{
		"email":          true,
		"email_verified": true,
	},
	"address": map[string]bool{
		"address": true,
	},
	"phone": map[string]bool{
		"phone_number":          true,
		"phone_number_verified": true,
	},
}

// scope に対応するクレームを返す。
// 返り値は自由に書き換えて良い。
func scopesToClaims(scops map[string]bool) map[string]bool {
	clms := map[string]bool{}
	for scop, ok := range scops {
		if !ok {
			continue
		}
		for clm, ok := range scopToClms[scop] {
			if !ok {
				continue
			}
			clms[clm] = true
		}
	}
	return clms
}
