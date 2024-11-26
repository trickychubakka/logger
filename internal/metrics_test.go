package internal

import (
	"github.com/stretchr/testify/assert"
	"logger/conf"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestMetricsPolling(t *testing.T) {
	type args struct {
		metrics MetricsStorage
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Positive Test MetricsPolling",
			args: args{
				metrics: NewMetricsStorageObj(),
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
		want MetricsStorage
	}{
		{
			name: "Positive Test NewMetricsObj",
			want: MetricsStorage{
				gaugeMap:   make(map[string]float64),
				counterMap: make(map[string]int64),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMetricsStorageObj(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMetricsObj() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSendMetrics(t *testing.T) {
	type args struct {
		metrics *MetricsStorage
		c       string
		config  *conf.AgentConfig
	}
	metrics := MetricsStorage{
		gaugeMap:   make(map[string]float64),
		counterMap: make(map[string]int64),
	}
	metrics.gaugeMap["metric1"] = 1

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

			if err := SendMetrics(tt.args.metrics, server.URL+tt.args.c, tt.args.config); (err != nil) != tt.wantErr {
				t.Errorf("SendMetrics() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSendRequest(t *testing.T) {
	type args struct {
		client *http.Client
		url    string
		config *conf.AgentConfig
	}
	type want struct {
		code        int
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
			// Start mock local HTTP server fo test
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
			res, err := SendRequest(tt.args.client, server.URL+tt.args.url, nil, "text/plain", tt.args.config)
			assert.Equal(t, tt.want.code, res.StatusCode)
			if (err != nil) != tt.wantErr {
				t.Errorf("SendRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
			res.Body.Close()
		})
	}
}
