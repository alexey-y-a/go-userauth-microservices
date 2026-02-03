package logger

import (
	"io"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
)


var (
    globalLogger zerolog.Logger
    once sync.Once
)
func Init() {
    once.Do(func() {
        var output io.Writer = os.Stdout

        globalLogger = zerolog.New(output).With().Timestamp().Logger()

        zerolog.TimeFieldFormat = time.RFC3339
    })
}

func L() zerolog.Logger {
    Init()
    return globalLogger
}