package lnurl

type PayParams struct {
	Status          string         `json:"status"`
	Callback        string         `json:"callback"`
	Tag             string         `json:"tag"`
	MaxSendable     int64          `json:"maxSendable"`
	MinSendable     int64          `json:"minSendable"`
	EncodedMetadata string         `json:"metadata"`
	CommentAllowed  int64          `json:"commentAllowed"`
	PayerData       *PayerDataSpec `json:"payerData,omitempty"`

	Metadata Metadata `json:"-"`
}

type PayerDataSpec struct {
	FreeName         *PayerDataItem    `json:"name"`
	PubKey           *PayerDataItem    `json:"pubkey"`
	LightningAddress *PayerDataItem    `json:"identifier"`
	Email            *PayerDataItem    `json:"email"`
	KeyAuth          *PayerDataKeyAuth `json:"auth"`
}

type PayerDataItem struct {
	Mandatory bool `json:"mandatory"`
}

type PayerDataKeyAuth struct {
	Mandatory bool   `json:"mandatory"`
	K1        string `json:"k1"`
}

type Metadata struct {
	Description     string
	LongDescription string
	Image           struct {
		DataURI string
		Bytes   []byte
		Ext     string
	}
	LightningAddress string
	IsEmail          bool
}
