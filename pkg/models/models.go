package models

import (
	"time"
)

// RequestType are network log types
type RequestType int

const (
	HTTP RequestType = 0
	WS
)

// Result is a Gowitness result
type Result struct {
	ID uint `json:"-" gorm:"primarykey"`

	URL            string `json:"url"`
	FinalURL       string `json:"finalurl"`
	ResponseCode   int    `json:"responsecode"`
	ResponseReason string `json:"responsereason"`
	Protocol       string `json:"protocol"`
	ContentLength  int64  `json:"contentlength"`
	HTML           string `json:"html"`
	Title          string `json:"title"`
	PerceptionHash string `json:"perceptionhash"`

	// Name of the screenshot file
	Filename string `json:"filename"`
	IsPDF    bool   `json:"ispdf"`

	// Failed flag set if the result should be considered failed
	Failed       bool   `json:"failed"`
	FailedReason string `json:"failedreason"`

	TLS          TLS          `json:"tls"`
	Technologies []Technology `json:"technologies"`

	Headers []Header     `json:"headers"`
	Network []NetworkLog `json:"network"`
	Console []ConsoleLog `json:"console"`
}

func (r *Result) HeaderMap() map[string][]string {
	headersMap := make(map[string][]string)

	for _, header := range r.Headers {
		var values []string
		for _, headerValue := range header.Values {
			values = append(values, headerValue.Value)
		}
		// Append the values to the map for the given header key
		headersMap[header.Key] = append(headersMap[header.Key], values...)
	}

	return headersMap
}

type TLS struct {
	ID       uint `json:"-" gorm:"primarykey"`
	ResultID uint `json:"-"`

	Protocol                 string       `json:"protocol"`
	KeyExchange              string       `json:"keyexchange"`
	Cipher                   string       `json:"cipher"`
	SubjectName              string       `json:"subjectname"`
	SanList                  []TLSSanList `json:"sanlist"`
	Issuer                   string       `json:"issuer"`
	ValidFrom                float64      `json:"validfrom"`
	ValidTo                  float64      `json:"validto"`
	ServerSignatureAlgorithm *int         `json:"serversignaturealgorithm"`
	EncryptedClientHello     bool         `json:"encryptedclienthello"`
}

type TLSSanList struct {
	ID    uint `json:"-" gorm:"primarykey"`
	TLSID uint `json:"-"`

	Value string `json:"value"`
}

type Technology struct {
	ID       uint `json:"-" gorm:"primarykey"`
	ResultID uint `json:"-"`

	Value string `json:"value"`
}

type Header struct {
	ID       uint `json:"-" gorm:"primarykey"`
	ResultID uint `json:"-"`

	Key    string        `json:"key"`
	Values []HeaderValue `json:"value"`
}

type HeaderValue struct {
	ID       uint `json:"-" gorm:"primarykey"`
	HeaderID uint `json:"-"`

	Value string `json:"string"`
}

type NetworkLog struct {
	ID       uint `json:"-" gorm:"primarykey"`
	ResultID uint `json:"-"`

	RequestType RequestType `json:"requesttype"`
	StatusCode  int         `json:"statuscode"`
	URL         string      `json:"url"`
	RemoteIP    string      `json:"remoteip"`
	MIMEType    string      `json:"mimetype"`
	Time        time.Time   `json:"time"`
	Error       string      `json:"error"`
}

type ConsoleLog struct {
	ID       uint `json:"-" gorm:"primarykey"`
	ResultID uint `json:"-"`

	Type  string `json:"type"`
	Value string `json:"value"`
}
