package logger

import (
	"go.uber.org/zap"
)

// global logger, remember call `defer Logger.Sync()` in `main` function
var (
	Logger, _ = zap.NewProduction()
	Sugar     = Logger.Sugar()
)
