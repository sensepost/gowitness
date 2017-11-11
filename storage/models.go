package storage

type HTTResponse struct {
	URL                string         `json:"url"`
	FinalURL           string         `json:"final_url"`
	ScreenshotFile     string         `json:"screenshot_file"`
	ResponseCode       int            `json:"response_code"`
	ResponseCodeString string         `json:"response_code_string"`
	Headers            []HTTPHeader   `json:"headers"`
	SSL                SSLCertificate `json:"ssl_certificate"`
}

type HTTPHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type SSLCertificate struct {
	PeerCertificates []SSLCertificateAttributes `json:"peer_certificates"`
	CipherSuite      uint16                     `json:"cipher_suite"`
}

type SSLCertificateAttributes struct {
	SubjectCommonName  string   `json:"subject_common_name"`
	IssuerCommonName   string   `json:"issuer_common_name"`
	SignatureAlgorithm string   `json:"signature_algorith"`
	DNSNames           []string `json:"dns_names"`
}
