package lndrest

type InvoiceState string

const (
	InvoiceState_OPEN     InvoiceState = "OPEN"
	InvoiceState_SETTLED  InvoiceState = "SETTLED"
	InvoiceState_CANCELED InvoiceState = "CANCELED"
	InvoiceState_ACCEPTED InvoiceState = "ACCEPTED"
)

type Invoice struct {
	Memo            string       `json:"memo,omitempty"`
	RPreimage       []byte       `json:"r_preimage,omitempty"`
	RHash           []byte       `json:"r_hash,omitempty"`
	Value           int64        `json:"value,string,omitempty"`
	ValueMsat       int64        `json:"value_msat,string,omitempty"`
	CreationDate    int64        `json:"creation_date,string,omitempty"`
	SettleDate      int64        `json:"settle_date,string,omitempty"`
	PaymentRequest  string       `json:"payment_request,omitempty"`
	DescriptionHash []byte       `json:"description_hash,omitempty"`
	Expiry          int64        `json:"expiry,string,omitempty"`
	FallbackAddr    string       `json:"fallback_addr,omitempty"`
	CltvExpiry      uint64       `json:"cltv_expiry,string,omitempty"`
	Private         bool         `json:"private,omitempty"`
	AddIndex        uint64       `json:"add_index,string,omitempty"`
	SettleIndex     uint64       `json:"settle_index,string,omitempty"`
	AmtPaidSat      int64        `json:"amt_paid_sat,string,omitempty"`
	AmtPaidMsat     int64        `json:"amt_paid_msat,string,omitempty"`
	State           InvoiceState `json:"state,omitempty"`
	IsKeysend       bool         `json:"is_keysend,omitempty"`
	PaymentAddr     []byte       `json:"payment_addr,omitempty"`
	IsAmp           bool         `json:"is_amp,omitempty"`
	IsBlinded       bool         `json:"is_blinded,omitempty"`
}
