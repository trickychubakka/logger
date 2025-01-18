package memstorage

import (
	"context"
	"github.com/stretchr/testify/assert"
	"log"
	"logger/internal/storage"
	"reflect"
	"testing"
)

func createTestStor() MemStorage {
	return MemStorage{
		gaugeMap:   map[string]float64{"Gauge1": 1.1, "Gauge2": 2.2, "Gauge3": 3.3},
		counterMap: map[string]int64{"Counter1": 1, "Counter2": 2, "Counter3": 3},
	}
}

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
				gaugeMap:   tt.fields.gaugeMap,
				counterMap: tt.fields.counterMap,
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
				gaugeMap:   tt.fields.gaugeMap,
				counterMap: tt.fields.counterMap,
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
		want MemStorage
	}{
		{
			name: "Test positive New()",
			want: MemStorage{
				gaugeMap:   make(map[string]float64),
				counterMap: make(map[string]int64),
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

func TestMarshal(t *testing.T) {
	type args struct {
		stor MemStorage
	}

	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "Test Marshal() positive",
			args: args{
				stor: createTestStor(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Marshal(tt.args.stor)
			log.Println("got :", string(got))
			if (err != nil) != tt.wantErr {
				t.Errorf("Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !assert.Equal(t, reflect.TypeOf(got), reflect.TypeOf(tt.want)) {
				t.Errorf("Marshal() got = %v, want %v", reflect.TypeOf(got), reflect.TypeOf(tt.want))
			}
		})
	}
}

func TestUnmarshal(t *testing.T) {
	type args struct {
		data []byte
		stor *MemStorage
	}
	stor := createTestStor()
	d, _ := Marshal(stor)

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test Unmarshal() positive",
			args: args{
				data: d,
				stor: &MemStorage{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Unmarshal(tt.args.data, tt.args.stor); (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMemStorage_GetAllCountersMap(t *testing.T) {
	type args struct {
		store MemStorage
		ctx   context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]int64
		wantErr bool
	}{

		{
			name: "Test GetAllCountersMap positive",
			args: args{
				ctx:   context.Background(),
				store: createTestStor(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.args.store.GetAllCountersMap(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllCountersMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !assert.Equal(t, reflect.TypeOf(got), reflect.TypeOf(tt.want)) {
				t.Errorf("GetAllCountersMap() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemStorage_GetAllGaugesMap(t *testing.T) {
	type args struct {
		store MemStorage
		ctx   context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]float64
		wantErr bool
	}{

		{
			name: "Test GetAllGaugesMap positive",
			args: args{
				ctx:   context.Background(),
				store: createTestStor(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.args.store.GetAllGaugesMap(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllGaugesMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !assert.Equal(t, reflect.TypeOf(got), reflect.TypeOf(tt.want)) {
				t.Errorf("GetAllGaugesMap() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemStorage_GetAllMetrics(t *testing.T) {
	type args struct {
		store MemStorage
		ctx   context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    MemStorage
		wantErr bool
	}{

		{
			name: "Test GetAllMetrics positive",
			args: args{
				ctx:   context.Background(),
				store: createTestStor(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.args.store.GetAllMetrics(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !assert.Equal(t, reflect.TypeOf(got), reflect.TypeOf(tt.want)) {
				t.Errorf("GetAllMetrics() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemStorage_GetCounter(t *testing.T) {
	type args struct {
		store MemStorage
		ctx   context.Context
		key   string
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{

		{
			name: "Test GetCounter positive",
			args: args{
				ctx:   context.Background(),
				store: createTestStor(),
				key:   "Counter2",
			},
			want:    2,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.args.store.GetCounter(tt.args.ctx, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCounter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !assert.Equal(t, got, tt.want) {
				t.Errorf("GetCounter() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemStorage_GetGauge(t *testing.T) {
	type args struct {
		store MemStorage
		ctx   context.Context
		key   string
	}
	tests := []struct {
		name    string
		args    args
		want    float64
		wantErr bool
	}{

		{
			name: "Test GetGauge positive",
			args: args{
				ctx:   context.Background(),
				store: createTestStor(),
				key:   "Gauge2",
			},
			want:    2.2,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.args.store.GetGauge(tt.args.ctx, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetGauge() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !assert.Equal(t, got, tt.want) {
				t.Errorf("GetGauge() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemStorage_GetValue(t *testing.T) {
	type args struct {
		store MemStorage
		ctx   context.Context
		t     string
		key   string
	}
	tests := []struct {
		name    string
		args    args
		want    any
		wantErr bool
	}{

		{
			name: "Test GetValue Gauge positive",
			args: args{
				ctx:   context.Background(),
				store: createTestStor(),
				t:     "gauge",
				key:   "Gauge2",
			},
			want:    2.2,
			wantErr: false,
		},
		{
			name: "Test GetValue Counter positive",
			args: args{
				ctx:   context.Background(),
				store: createTestStor(),
				t:     "counter",
				key:   "Counter2",
			},
			want:    int64(2), // из-за типа any в tests необходимо явно кастовать в int64
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.args.store.GetValue(tt.args.ctx, tt.args.t, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !assert.Equal(t, got, tt.want) {
				t.Errorf("GetValue() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemStorage_UpdateBatch(t *testing.T) {
	type args struct {
		store   MemStorage
		ctx     context.Context
		metrics []storage.Metrics
		key     string
	}
	var g = 100.1
	var c int64 = 100
	tests := []struct {
		name    string
		args    args
		want    any
		wantErr bool
	}{

		{
			name: "Test TestMemStorage_UpdateBatch positive",
			args: args{
				ctx:     context.Background(),
				store:   createTestStor(),
				metrics: []storage.Metrics{{ID: "Gauge100", MType: "gauge", Value: &g}, {ID: "Counter100", MType: "counter", Delta: &c}},
				key:     "Gauge100",
			},
			want:    100.1,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.args.store.UpdateBatch(tt.args.ctx, tt.args.metrics)
			got := tt.args.store.gaugeMap[tt.args.key]
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateBatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !assert.Equal(t, tt.args.store.gaugeMap[tt.args.key], tt.want) {
				t.Errorf("UpdateBatch() got = %v, want %v", got, tt.want)
			}
		})
	}
}
