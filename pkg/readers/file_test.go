package readers

import (
	"reflect"
	"testing"
)

func TestUrlsFor(t *testing.T) {
	fr := FileReader{
		Options: &FileReaderOptions{},
	}

	tests := []struct {
		name      string
		candidate string
		ports     []int
		want      []string
	}{
		{
			name:      "Test with IP",
			candidate: "192.168.1.1",
			ports:     []int{80, 443, 8443},
			want: []string{
				"http://192.168.1.1:80",
				"http://192.168.1.1:443",
				"http://192.168.1.1:8443",
				"https://192.168.1.1:80",
				"https://192.168.1.1:443",
				"https://192.168.1.1:8443",
			},
		},
		{
			name:      "Test with IP and port",
			candidate: "192.168.1.1:8080",
			ports:     []int{80, 443, 8443},
			want: []string{
				"http://192.168.1.1:8080",
				"https://192.168.1.1:8080",
			},
		},
		{
			name:      "Test with IP and port with spaces",
			candidate: "   192.168.1.1:8080   ",
			ports:     []int{80, 443, 8443},
			want: []string{
				"http://192.168.1.1:8080",
				"https://192.168.1.1:8080",
			},
		},
		{
			name:      "Test with scheme, IP and port",
			candidate: "http://192.168.1.1:8080",
			ports:     []int{80, 443, 8443},
			want: []string{
				"http://192.168.1.1:8080",
			},
		},
		{
			name:      "Test with scheme and IP",
			candidate: "https://192.168.1.1",
			ports:     []int{80, 443, 8443},
			want: []string{
				"https://192.168.1.1:80",
				"https://192.168.1.1:443",
				"https://192.168.1.1:8443",
			},
		},
		{
			name:      "Test with IP and path",
			candidate: "192.168.1.1/path",
			ports:     []int{80, 443, 8443},
			want: []string{
				"http://192.168.1.1:80/path",
				"http://192.168.1.1:443/path",
				"http://192.168.1.1:8443/path",
				"https://192.168.1.1:80/path",
				"https://192.168.1.1:443/path",
				"https://192.168.1.1:8443/path",
			},
		},
		{
			name:      "Test with scheme, IP, port and path",
			candidate: "http://192.168.1.1:8080/path",
			ports:     []int{80, 443, 8443},
			want: []string{
				"http://192.168.1.1:8080/path",
			},
		},
		{
			name:      "Test with IP, port and path",
			candidate: "192.168.1.1:8080/path",
			ports:     []int{80, 443, 8443},
			want: []string{
				"http://192.168.1.1:8080/path",
				"https://192.168.1.1:8080/path",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fr.urlsFor(tt.candidate, tt.ports)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("urlsFor() =>\n\nhave: %v\nwant %v", got, tt.want)
			}
		})
	}
}
