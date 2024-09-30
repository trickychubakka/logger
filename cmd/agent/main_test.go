package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestMetricsPolling(t *testing.T) {
	type args struct {
		metrics Metrics
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Positive Test MetricsPolling",
			args: args{
				metrics: NewMetricsObj(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := MetricsPolling(&tt.args.metrics); (err != nil) != tt.wantErr {
				t.Errorf("MetricsPolling() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewMetricsObj(t *testing.T) {
	tests := []struct {
		name string
		want Metrics
	}{
		{
			name: "Positive Test NewMetricsObj",
			want: Metrics{
				gaugeMap:   make(map[string]float64),
				counterMap: make(map[string]int64),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMetricsObj(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMetricsObj() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSendMetrics(t *testing.T) {
	type args struct {
		metrics *Metrics
		c       string
	}
	metrics := Metrics{
		gaugeMap:   make(map[string]float64),
		counterMap: make(map[string]int64),
	}
	metrics.gaugeMap["metric1"] = 1
	//metrics.counterMap["metric2"] = 2

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Positive Test SendMetrics",
			args: args{
				metrics: &metrics,
				c:       "/update",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/update/gauge/metric1/1" {
					t.Errorf("Expected to request '/update/gauge/metric1/1', got: %s", r.URL.Path)
				}
				if r.Header.Get("Content-Type") != "text/plain" {
					t.Errorf("Expected Content-Type: text/plain header, got: %s", r.Header.Get("Content-Type"))
				}
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			if err := SendMetrics(tt.args.metrics, server.URL+tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("SendMetrics() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSendRequest(t *testing.T) {
	type args struct {
		client *http.Client
		url    string
	}
	type want struct {
		code int
		//response    string
		contentType string
	}
	tests := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "Positive Test SendRequest",
			args: args{
				client: &http.Client{},
				url:    "/update/gauge/metric1/1",
			},
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Start a local HTTP server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/update/gauge/metric1/1" {
					t.Errorf("Expected to request '/update/gauge/metric1/1', got: %s", r.URL.Path)
				}
				if r.Header.Get("Content-Type") != "text/plain" {
					t.Errorf("Expected Content-Type: text/plain header, got: %s", r.Header.Get("Content-Type"))
				}
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()
			res, err := SendRequest(tt.args.client, server.URL+tt.args.url)
			assert.Equal(t, tt.want.code, res.StatusCode)
			//assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
			if (err != nil) != tt.wantErr {
				t.Errorf("SendRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
			res.Body.Close()
		})
	}
}
