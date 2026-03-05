package logging

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Init 初始化全局 logger
func Init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

// Logger 返回全局 logger，方便替换实现
func Logger() zerolog.Logger {
	return log.Logger
}
