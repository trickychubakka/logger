package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"testing"
)

func Test_initConfig(t *testing.T) {
	type wantType struct {
		wantErr bool
		Flag    bool
	}

	want := wantType{
		wantErr: false,
	}

	if err := initConfig(); (err != nil) != want.wantErr {
		t.Errorf("initConfig() error = %v, wantErr %v", err, want.wantErr)
	}
}

func Test_initConfigEnv(t *testing.T) {

	if err := os.Setenv("ERRCHECK_ENABLE", "true"); err != nil {
		t.Errorf("initConfig() error")
	}

	if err := initConfig(); err != nil {
		t.Errorf("initConfig() error = %v", err)
	}

	if envErrCheckEnable := os.Getenv("ERRCHECK_ENABLE"); envErrCheckEnable != "" {
		e, err := strconv.ParseBool(envErrCheckEnable)
		if err == nil {
			assert.Equal(t, e, flags.ErrCheckEnable)
		}
	}
}
