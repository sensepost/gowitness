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
	ID uint `json:"id" gorm:"primarykey"`

	URL            string `json:"url"`
	FinalURL       string `json:"final_url"`
	ResponseCode   int    `json:"response_code"`
	ResponseReason string `json:"response_reason"`
	Protocol       string `json:"protocol"`
	ContentLength  int64  `json:"content_length"`
	HTML           string `json:"html" gorm:"index"`
	Title          string `json:"title" gorm:"index"`
	PerceptionHash string `json:"perception_hash" gorm:"index"`

	// Name of the screenshot file
	Filename string `json:"file_name"`
	IsPDF    bool   `json:"is_pdf"`

	// Failed flag set if the result should be considered failed
	Failed       bool   `json:"failed"`
	FailedReason string `json:"failed_reason"`

	TLS          TLS          `json:"tls"`
	Technologies []Technology `json:"technologies"`

	Headers []Header     `json:"headers"`
	Network []NetworkLog `json:"network"`
	Console []ConsoleLog `json:"console"`
}

func (r *Result) HeaderMap() map[string][]string {
	headersMap := make(map[string][]string)

	for _, header := range r.Headers {
		headersMap[header.Key] = []string{header.Value}
	}

	return headersMap
}

type TLS struct {
	ID       uint `json:"id" gorm:"primarykey"`
	ResultID uint `json:"resultid"`

	Protocol                 string       `json:"protocol"`
	KeyExchange              string       `json:"key_exchange"`
	Cipher                   string       `json:"cipher"`
	SubjectName              string       `json:"subject_name"`
	SanList                  []TLSSanList `json:"san_list"`
	Issuer                   string       `json:"issuer"`
	ValidFrom                float64      `json:"valid_from"`
	ValidTo                  float64      `json:"valid_to"`
	ServerSignatureAlgorithm *int         `json:"server_signature_algorithm"`
	EncryptedClientHello     bool         `json:"encrypted_client_hello"`
}

type TLSSanList struct {
	ID    uint `json:"id" gorm:"primarykey"`
	TLSID uint `json:"tls_id"`

	Value string `json:"value"`
}

type Technology struct {
	ID       uint `json:"id" gorm:"primarykey"`
	ResultID uint `json:"result_id"`

	Value string `json:"value" gorm:"index"`
}

type Header struct {
	ID       uint `json:"id" gorm:"primarykey"`
	ResultID uint `json:"result_id"`

	Key   string `json:"key"`
	Value string `json:"value" gorm:"index"`
}

type NetworkLog struct {
	ID       uint `json:"id" gorm:"primarykey"`
	ResultID uint `json:"result_id"`

	RequestType RequestType `json:"request_type"`
	StatusCode  int         `json:"status_code"`
	URL         string      `json:"url"`
	RemoteIP    string      `json:"remote_ip"`
	MIMEType    string      `json:"mime_type"`
	Time        time.Time   `json:"time"`
	Error       string      `json:"error"`
}

type ConsoleLog struct {
	ID       uint `json:"id" gorm:"primarykey"`
	ResultID uint `json:"result_id"`

	Type  string `json:"type"`
	Value string `json:"value" gorm:"index"`
}
