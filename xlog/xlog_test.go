package xlog

import (
	"github.com/coder2z/d-telegram/config"
	"go.uber.org/zap/zapcore"
	"reflect"
	"testing"
)

func Test_getLogWriter(t *testing.T) {
	type args struct {
		cfg *config.Config
	}
	tests := []struct {
		name string
		args args
		want zapcore.WriteSyncer
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getLogWriter(tt.args.cfg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getLogWriter() = %v, want %v", got, tt.want)
			}
		})
	}
}
