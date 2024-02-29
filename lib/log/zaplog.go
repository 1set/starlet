package log

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	dc "github.com/1set/starlet/dataconv"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ModuleName defines the expected name for this Module when used
// in starlark's load() function, eg: load('log', 'info')
const ModuleName = "log"

// Initialized as global functions to be used as default
var (
	defaultModule = NewModule(NewDefaultLogger())
	// LoadModule loads the default log module. It is concurrency-safe and idempotent.
	LoadModule = defaultModule.LoadModule
	// SetLog sets the logger of the default log module from outside the package. If l is nil, a noop logger is used, which does nothing.
	SetLog = defaultModule.SetLog
)

// NewDefaultLogger creates a new logger as a default. It is used when no logger is provided to NewModule.
func NewDefaultLogger() *zap.SugaredLogger {
	cfg := zap.NewDevelopmentConfig()
	cfg.DisableCaller = true
	cfg.DisableStacktrace = true
	lg, _ := cfg.Build()
	return lg.Sugar()
}

// Module wraps the starlark module for the log package.
type Module struct {
	once      sync.Once
	logModule starlark.StringDict
	logger    *zap.SugaredLogger
}

// NewModule creates a new log module. If logger is nil, a new development logger is created.
func NewModule(lg *zap.SugaredLogger) *Module {
	if lg == nil {
		lg = NewDefaultLogger()
	}
	return &Module{logger: lg}
}

// LoadModule returns the log module loader. It is concurrency-safe and idempotent.
func (m *Module) LoadModule() (starlark.StringDict, error) {
	m.once.Do(func() {
		// If logger is nil, create a new development logger.
		if m.logger == nil {
			m.logger = NewDefaultLogger()
		}

		// Create the log module
		m.logModule = starlark.StringDict{
			ModuleName: &starlarkstruct.Module{
				Name: ModuleName,
				Members: starlark.StringDict{
					"debug": m.genLoggerBuiltin("debug", zap.DebugLevel),
					"info":  m.genLoggerBuiltin("info", zap.InfoLevel),
					"warn":  m.genLoggerBuiltin("warn", zap.WarnLevel),
					"error": m.genLoggerBuiltin("error", zap.ErrorLevel),
					"fatal": m.genLoggerBuiltin("fatal", zap.FatalLevel),
				},
			},
		}
	})
	return m.logModule, nil
}

// SetLog sets the logger of the log module from outside the package. If l is nil, a noop logger is used, which does nothing.
func (m *Module) SetLog(l *zap.SugaredLogger) {
	if l == nil {
		m.logger = zap.NewNop().Sugar()
		return
	}
	m.logger = l
}

// genLoggerBuiltin is a helper function to generate a starlark Builtin function that logs a message at a given level.
func (m *Module) genLoggerBuiltin(name string, level zapcore.Level) starlark.Callable {
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
			logFn = m.logger.Debugw
		case zap.InfoLevel:
			logFn = m.logger.Infow
		case zap.WarnLevel:
			logFn = m.logger.Warnw
		case zap.ErrorLevel:
			logFn = m.logger.Errorw
		case zap.FatalLevel:
			logFn = m.logger.Errorw
			retErr = true
		default:
			return nil, fmt.Errorf("unsupported log level: %v", level)
		}

		// append leftover arguments to message
		if len(args) > 1 {
			var ps []string
			for _, a := range args[1:] {
				ps = append(ps, dc.StarString(a))
			}
			msg += " " + strings.Join(ps, " ")
		}

		// convert args to key-value pairs
		var kvp []interface{}
		for _, pair := range kwargs {
			// for each key-value pair
			if pair.Len() != 2 {
				continue
			}
			key, val := pair[0], pair[1]

			// for keys, try to interpret as string, or use String() as fallback
			kvp = append(kvp, dc.StarString(key))

			// for values, try to unmarshal to Go types, or use String() as fallback
			if v, e := dc.Unmarshal(val); e == nil {
				kvp = append(kvp, v)
			} else {
				kvp = append(kvp, val.String())
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
