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
	ContentLength  int64  `json:"contentlength"`
	HTML           string `json:"html"`
	Title          string `json:"title"`

	// Name of the screenshot file
	Filename string `json:"filename"`

	// Failed flag set if the result should be considered failed
	Failed       bool   `json:"failed"`
	FailedReason string `json:"failedreason"`

	Headers []Header     `json:"headers"`
	Network []NetworkLog `json:"network"`
	Console []ConsoleLog `json:"console"`
}

func (r *Result) AddHeader(key string, value string) {
	r.Headers = append(r.Headers, Header{
		Key:   key,
		Value: value,
	})
}

func (r *Result) AddNetworkLog(log NetworkLog) {
	r.Network = append(r.Network, log)
}

type Header struct {
	ID       uint `json:"-" gorm:"primarykey"`
	ResultID uint `json:"-"`

	Key   string `json:"key"`
	Value string `json:"value"`
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
