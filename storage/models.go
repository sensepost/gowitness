package storage

// HTTResponse contains an HTTP response
type HTTResponse struct {
	URL                string         `json:"url"`
	FinalURL           string         `json:"final_url"`
	ScreenshotFile     string         `json:"screenshot_file"`
	ResponseCode       int            `json:"response_code"`
	ResponseCodeString string         `json:"response_code_string"`
	Headers            []HTTPHeader   `json:"headers"`
	SSL                SSLCertificate `json:"ssl_certificate"`
	Title              string         `json:"page_title"`
	Hash               uint64         `json:"hash"`
}

// HTTPHeader contains an HTTP header key value pair
type HTTPHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// SSLCertificate contains an SSL certificate presented by URL
type SSLCertificate struct {
	PeerCertificates []SSLCertificateAttributes `json:"peer_certificates"`
	CipherSuite      uint16                     `json:"cipher_suite"`
}

// SSLCertificateAttributes contains the attributes of a certificate
type SSLCertificateAttributes struct {
	SubjectCommonName  string   `json:"subject_common_name"`
	IssuerCommonName   string   `json:"issuer_common_name"`
	SignatureAlgorithm string   `json:"signature_algorith"`
	DNSNames           []string `json:"dns_names"`
}
