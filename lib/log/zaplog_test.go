package log_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	itn "github.com/1set/starlet/internal"
	lg "github.com/1set/starlet/lib/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestLoadModule_Log_NoSet(t *testing.T) {
	script := "load('log', 'info')\ninfo('this is 1st info message only')"
	res, err := itn.ExecModuleWithErrorTest(t, lg.ModuleName, lg.LoadModule, script, "")
	if err != nil {
		t.Errorf("log.SetLog(nil) expects no error, actual error = '%v', result = %v", err, res)
		return
	}
}

func TestLoadModule_Log_SetLog_Nil(t *testing.T) {
	lg.SetLog(nil)
	script := "load('log', 'info')\ninfo('this is 2nd info message only')"
	res, err := itn.ExecModuleWithErrorTest(t, lg.ModuleName, lg.LoadModule, script, "")
	if err != nil {
		t.Errorf("log.SetLog(nil) expects no error, actual error = '%v', result = %v", err, res)
		return
	}
}

func TestLoadModule_Log(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		wantErr  string
		keywords []string
	}{
		{
			name: `debug message`,
			script: itn.HereDoc(`
				load('log', 'debug')
				debug('this is a debug message only')
			`),
			keywords: []string{"DEBUG", "this is a debug message only"},
		},
		{
			name: `debug with no args`,
			script: itn.HereDoc(`
				load('log', 'debug')
				debug()
			`),
			wantErr: "log.debug: expected at least 1 argument, got 0",
		},
		{
			name: `debug with invalid arg type`,
			script: itn.HereDoc(`
				load('log', 'debug')
				debug(520)
			`),
			wantErr: "log.debug: expected string as first argument, got int",
		},
		{
			name: `debug with incomplete args`,
			script: itn.HereDoc(`
				load('log', 'debug')
				debug('this is a broken message', "what")
			`),
			keywords: []string{"DEBUG", "this is a broken message", "ERROR", "Ignored key without a value."},
		},
		{
			name: `debug with key values`,
			script: itn.HereDoc(`
				load('log', 'debug')
				m = {"mm": "this is more"}
				l = [2, "LIST", 3.14, True]
				debug('this is a data message', "map", m, "list", l)
			`),
			keywords: []string{"DEBUG", "this is a data message", `{"map": {"mm":"this is more"}, "list": [2,"LIST",3.14,true]}`},
		},
		{
			name: `info message`,
			script: itn.HereDoc(`
				load('log', 'info')
				info('this is an info message', 1, 2, "hello", "world")
			`),
			keywords: []string{"INFO", "this is an info message", `{"1": 2, "hello": "world"}`},
		},
		{
			name: `info self args`,
			script: itn.HereDoc(`
				load('log', 'info')
				d = {"hello": "world"}
				d["a"] = d
				l = [1,2,3]
				l.append(l)
				s = set([4,5,6])
				info('this is complex info message', "self1", d, "self2", l, "self3", s)
			`),
			keywords: []string{"INFO", "this is complex info message", `{"self1": "{\"hello\": \"world\", \"a\": {...}}", "self2": "[1, 2, 3, [...]]", "self3": [4,5,6]}`},
		},
		{
			name: `warn message`,
			script: itn.HereDoc(`
				load('log', 'warn')
				warn('this is a warning message only')
			`),
			keywords: []string{"WARN", "this is a warning message only"},
		},
		{
			name: `error message`,
			script: itn.HereDoc(`
				load('log', 'error')
				error('this is an error message only')
			`),
			keywords: []string{"ERROR", "this is an error message only"},
		},
		{
			name: `fatal message`,
			script: itn.HereDoc(`
				load('log', 'fatal')
				fatal('this is a fatal message only')
			`),
			wantErr: `this is a fatal message only`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, b := buildCustomLogger()
			lg.SetLog(l)
			res, err := itn.ExecModuleWithErrorTest(t, lg.ModuleName, lg.LoadModule, tt.script, tt.wantErr)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("log(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
				return
			}
			if len(tt.wantErr) > 0 {
				return
			}
			if len(tt.keywords) > 0 {
				bs := b.String()
				for _, k := range tt.keywords {
					if !strings.Contains(bs, k) {
						t.Errorf("log(%q) expects keyword = '%v', actual log = '%v'", tt.name, k, bs)
						return
					}
				}
			} else {
				fmt.Println(b.String())
			}
		})
	}
}

func buildCustomLogger() (*zap.SugaredLogger, *bytes.Buffer) {
	buf := bytes.NewBufferString("")
	var al zap.LevelEnablerFunc = func(lvl zapcore.Level) bool {
		return true
	}
	ce := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})
	cr := zapcore.NewCore(ce, zapcore.AddSync(buf), al)
	op := []zap.Option{
		zap.AddCaller(),
		//zap.AddStacktrace(zap.ErrorLevel),
	}
	logger := zap.New(cr, op...)
	return logger.Sugar(), buf
}
