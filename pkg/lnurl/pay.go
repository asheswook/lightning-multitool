package lnurl

import "net/url"

type Response struct {
	Status string `json:"status,omitempty"`
	Reason string `json:"reason,omitempty"`
}

type ErrorResponse struct {
	Status string   `json:"status,omitempty"`
	Reason string   `json:"reason,omitempty"`
	URL    *url.URL `json:"-"`
}

type PayParams struct {
	Response
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

type PayResponse struct {
	Response
	SuccessAction map[string]string `json:"successAction"`
	Routes        []interface{}     `json:"routes"`
	PR            string            `json:"pr"`
	Disposable    bool              `json:"disposable"`
}

type SuccessActionType string

func (s SuccessActionType) String() string {
	return string(s)
}

const (
	SuccessActionMessage SuccessActionType = "message"
	SuccessActionURL                       = "url"
)

type SuccessAction struct {
	Tag     SuccessActionType `json:"tag"`
	Message string            `json:"message"`
	URL     string            `json:"url"`
}
