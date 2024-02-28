package log

import (
	"sync"

	"go.starlark.net/starlark"
	"go.uber.org/zap"
)

// ModuleName defines the expected name for this Module when used
// in starlark's load() function, eg: load('log', 'info')
const ModuleName = "log"

var (
	once      sync.Once
	logModule starlark.StringDict
	logger    *zap.SugaredLogger
)
