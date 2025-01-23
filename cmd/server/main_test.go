package main

import (
	"context"
	"github.com/stretchr/testify/assert"
	"logger/cmd/server/initconf"
	"logger/internal/handlers"
	"logger/internal/storage/memstorage"
	"reflect"
	"testing"
)

func Test_storeInit(t *testing.T) {
	type args struct {
		ctx   context.Context
		store handlers.Storager
		conf  *initconf.Config
	}
	var store handlers.Storager
	tests := []struct {
		name    string
		args    args
		want    memstorage.MemStorage
		wantErr bool
	}{
		{
			name: "Positive Test Test_storeInit",
			args: args{
				ctx:   context.Background(),
				store: store,
				conf: &initconf.Config{
					DatabaseDSN: "",
					Restore:     false,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := storeInit(tt.args.ctx, tt.args.store, tt.args.conf)
			if (err != nil) != tt.wantErr {
				t.Errorf("storeInit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// TypeOf in this case must be memstorage.MemStorage
			if !assert.Equal(t, reflect.TypeOf(got), reflect.TypeOf(tt.want)) {
				t.Errorf("storeInit() got = %v, want %v", got, tt.want)
			}
		})
	}
}
