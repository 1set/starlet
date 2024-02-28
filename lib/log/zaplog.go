package log

import (
	"errors"
	"fmt"
	"sync"

	"github.com/1set/starlet/dataconv"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ModuleName defines the expected name for this Module when used
// in starlark's load() function, eg: load('log', 'info')
const ModuleName = "log"

var (
	once      sync.Once
	logModule starlark.StringDict
	logger    *zap.SugaredLogger
)

// LoadModule loads the time module. It is concurrency-safe and idempotent.
func LoadModule() (starlark.StringDict, error) {
	once.Do(func() {
		// If logger is nil, create a new development logger.
		if logger == nil {
			lg, _ := zap.NewDevelopment()
			logger = lg.Sugar()
		}

		// Create the log module
		logModule = starlark.StringDict{
			ModuleName: &starlarkstruct.Module{
				Name: ModuleName,
				Members: starlark.StringDict{
					"debug": genLoggerBuiltin("debug", zap.DebugLevel),
					"info":  genLoggerBuiltin("info", zap.InfoLevel),
					"warn":  genLoggerBuiltin("warn", zap.WarnLevel),
					"error": genLoggerBuiltin("error", zap.ErrorLevel),
					"fatal": genLoggerBuiltin("fatal", zap.FatalLevel),
				},
			},
		}
	})
	return logModule, nil
}

// SetLog sets the logger from outside the package.
func SetLog(l *zap.SugaredLogger) {
	if l == nil {
		logger = zap.NewNop().Sugar()
		return
	}
	logger = l
}

// genLoggerBuiltin is a helper function to generate a starlark Builtin function that logs a message at a given level.
func genLoggerBuiltin(name string, level zapcore.Level) starlark.Callable {
	return starlark.NewBuiltin(ModuleName+"."+name, func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var msg string
		if len(args) <= 0 {
			return nil, fmt.Errorf("%s: expected at least 1 argument, got 0", fn.Name())
		} else if s, ok := args[0].(starlark.String); ok {
			msg = string(s)
		} else {
			return nil, fmt.Errorf("%s: expected string as first argument, got %s", fn.Name(), args[0].Type())
		}

		// find the correct log function
		var (
			logFn  func(msg string, keysAndValues ...interface{})
			retErr bool
		)
		switch level {
		case zap.DebugLevel:
			logFn = logger.Debugw
		case zap.InfoLevel:
			logFn = logger.Infow
		case zap.WarnLevel:
			logFn = logger.Warnw
		case zap.ErrorLevel:
			logFn = logger.Errorw
		case zap.FatalLevel:
			logFn = logger.Errorw
			retErr = true
		default:
			return nil, fmt.Errorf("unsupported log level: %v", level)
		}

		// convert args to key-value pairs
		var kvp []interface{}
		for i := range args {
			if i == 0 {
				continue
			}
			if i%2 == 1 {
				// for keys, try to interpret as string, or use String() as fallback
				if s, ok := args[i].(starlark.String); ok {
					kvp = append(kvp, s.GoString())
				} else {
					kvp = append(kvp, args[i].String())
				}
			} else {
				// for values, try to unmarshal to Go types, or use String() as fallback
				if v, e := dataconv.Unmarshal(args[i]); e == nil {
					kvp = append(kvp, v)
				} else {
					kvp = append(kvp, args[i].String())
				}
			}
		}

		// log the message
		logFn(msg, kvp...)
		if retErr {
			return starlark.None, errors.New(msg)
		}
		return starlark.None, nil
	})
}
