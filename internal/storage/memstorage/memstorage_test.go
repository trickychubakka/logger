package memstorage

import (
	"context"
	//"logger/internal/storage"
	"reflect"
	"testing"
)

func TestMemStorage_UpdateCounter(t *testing.T) {
	type fields struct {
		gaugeMap   map[string]float64
		counterMap map[string]int64
	}
	type args struct {
		key   string
		value int64
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Test TestMemStorage_UpdateCounter positive",
			fields: fields{
				counterMap: make(map[string]int64),
			},
			args: args{
				key:   "metric1",
				value: 1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				GaugeMap:   tt.fields.gaugeMap,
				CounterMap: tt.fields.counterMap,
			}
			if err := ms.UpdateCounter(ctx, tt.args.key, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("UpdateCounter() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMemStorage_UpdateGauge(t *testing.T) {
	type fields struct {
		gaugeMap   map[string]float64
		counterMap map[string]int64
	}
	type args struct {
		key   string
		value float64
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Test TestMemStorage_UpdateGauge positive",
			fields: fields{
				gaugeMap: make(map[string]float64),
			},
			args: args{
				key:   "metric1",
				value: 1.0,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				GaugeMap:   tt.fields.gaugeMap,
				CounterMap: tt.fields.counterMap,
			}
			if err := ms.UpdateGauge(ctx, tt.args.key, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("UpdateGauge() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNew(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tests := []struct {
		name string
		//want storage.Storager
		want MemStorage
	}{
		{
			name: "Test positive New()",
			want: MemStorage{
				GaugeMap:   make(map[string]float64),
				CounterMap: make(map[string]int64),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := New(ctx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}
