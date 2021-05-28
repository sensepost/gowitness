package storage

import (
	"encoding/json"
	"strconv"

	"gorm.io/gorm"
)

// URL contains information about a URL
type URL struct {
	gorm.Model

	URL            string
	FinalURL       string
	ResponseCode   int
	ResponseReason string
	Proto          string
	ContentLength  int64
	Title          string
	Filename       string
	PerceptionHash string

	Headers      []Header
	TLS          TLS
	Technologies []Technologie
}

// AddHeader adds a new header to a URL
func (url *URL) AddHeader(key string, value string) {
	url.Headers = append(url.Headers, Header{
		Key:   key,
		Value: value,
	})
}

// AddTechnlogies adds a new technologies to a URL
func (url *URL) AddTechnologie(value string) {
	url.Technologies = append(url.Technologies, Technologie{
		Value: value,
	})
}

// MarshallCSV returns values as a slice
func (url *URL) MarshallCSV() (res []string) {
	return []string{url.URL,
		url.FinalURL,
		strconv.Itoa(url.ResponseCode),
		url.ResponseReason,
		url.Proto,
		strconv.Itoa(int(url.ContentLength)),
		url.Title,
		url.Filename}
}

// MarshallJSON returns values as a slice
func (url *URL) MarshallJSON() ([]byte, error) {
	var tmp struct {
		URL            string `json:"url"`
		FinalURL       string `json:"final_url"`
		ResponseCode   int    `json:"response_code"`
		ResponseReason string `json:"response_reason"`
		Proto          string `json:"proto"`
		ContentLength  int64  `json:"content_length"`
		Title          string `json:"title"`
		Filename       string `json:"file_name"`
	}

	tmp.URL = url.URL
	tmp.FinalURL = url.FinalURL
	tmp.ResponseCode = url.ResponseCode
	tmp.ResponseReason = url.ResponseReason
	tmp.Proto = url.Proto
	tmp.ContentLength = url.ContentLength
	tmp.Title = url.Title
	tmp.Filename = url.Filename

	return json.Marshal(&tmp)
}

// Header contains an HTTP header
type Header struct {
	gorm.Model

	URLID uint
	Key   string
	Value string
}

// Technologie contains a technologie
type Technologie struct {
	gorm.Model

	URLID uint
	Value string
}

// TLS contains TLS information for a URL
type TLS struct {
	gorm.Model

	URLID           uint
	Version         uint16
	ServerName      string
	TLSCertificates []TLSCertificate
}

// TLSCertificate contain TLS Certificate information
type TLSCertificate struct {
	gorm.Model

	TLSID              uint
	Raw                []byte
	DNSNames           []TLSCertificateDNSName
	SubjectCommonName  string
	IssuerCommonName   string
	SignatureAlgorithm string
	PubkeyAlgorithm    string
}

// AddDNSName adds a new DNS Name to a Certificate
func (tlsCert *TLSCertificate) AddDNSName(name string) {
	tlsCert.DNSNames = append(tlsCert.DNSNames, TLSCertificateDNSName{Name: name})
}

// TLSCertificateDNSName has DNS names for a TLS certificate
type TLSCertificateDNSName struct {
	gorm.Model

	TLSCertificateID uint
	Name             string
}
