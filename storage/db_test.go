package storage

import (
	"net/url"
	"reflect"
	"testing"
)

func Test_parseDBLocation(t *testing.T) {
	userInfo := url.UserPassword("user", "pass")
	type args struct {
		dbLocation string
	}
	tests := []struct {
		name    string
		args    args
		want    *url.URL
		want1   string
		wantErr bool
	}{
		{name: "default db location", args: args{dbLocation: "sqlite://gowitness.sqlite3"}, want: &url.URL{
			Scheme: "sqlite",
			Host:   "gowitness.sqlite3",
		}, want1: "gowitness.sqlite3", wantErr: false},
		{name: "absolute path db location", args: args{dbLocation: "sqlite:///tmp/testing/gowitness.sqlite3"}, want: &url.URL{
			Scheme: "sqlite",
			Path:   "/tmp/testing/gowitness.sqlite3",
		}, want1: "/tmp/testing/gowitness.sqlite3", wantErr: false},
		{name: "non default relative db location", args: args{dbLocation: "sqlite://bonkers.sqlite3"}, want: &url.URL{
			Scheme: "sqlite",
			Host:   "bonkers.sqlite3",
		}, want1: "bonkers.sqlite3", wantErr: false},
		{name: "empty incorrect dbLocation should revert to default path", args: args{dbLocation: "sqlite://"}, want: &url.URL{
			Scheme: "sqlite",
			Host:   "",
			Path:   "",
		}, want1: "gowitness.sqlite3", wantErr: false},
		{name: "postgres db path", args: args{dbLocation: "postgres://user:pass@host:5432/db"}, want: &url.URL{
			Scheme: "postgres",
			Host:   "host:5432",
			Path:   "/db",
			User:   userInfo,
		}, want1: "postgres://user:pass@host:5432/db", wantErr: false},
		{name: "empty string test", args: args{dbLocation: ""}, want: &url.URL{}, want1: "", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := parseDBLocation(tt.args.dbLocation)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDBLocation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseDBLocation() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("parseDBLocation() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
