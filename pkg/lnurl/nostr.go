package lnurl

type PayParamsWithNostr struct {
	PayParams
	AllowsNostr bool   `json:"allowsNostr"`
	NostrPubkey string `json:"nostrPubkey"`
}
