package logger

import (
	"fmt"
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Warn(msg string, keysAndValues ...any)
	Info(msg string, keysAndValues ...any)
	Error(msg string, keysAndValues ...any)
	Debug(msg string, keysAndValues ...any)
	With(keysAndValues ...any) Logger
	WithGroup(name string) Logger
}

// Проверка на уровне компилятора реализует ли ZapLogger интерфейс Logger
var _ Logger = (*ZapLogger)(nil)

type ZapLogger struct {
	logger *zap.SugaredLogger
}

func NewZapLogger(logLevel string) (*ZapLogger, error) {
	logLvl, err := zapcore.ParseLevel(logLevel)
	if err != nil {
		return nil, fmt.Errorf("parsing log level error from zap: %w", err)
	}
	consoleEncoderConfig := zapcore.EncoderConfig{
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      "C",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder, // Цветные уровней логирования
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	consoleEncoder := zapcore.NewConsoleEncoder(consoleEncoderConfig)
	consoleWriter := os.Stdout
	consoleCore := newZapCore(consoleEncoder, consoleWriter, logLvl)

	return &ZapLogger{
		logger: zap.New(consoleCore).Sugar(),
	}, nil

}

func newZapCore(encoder zapcore.Encoder, writer io.Writer, logLevel zapcore.Level) zapcore.Core {
	return zapcore.NewCore(
		encoder,
		zapcore.AddSync(writer),
		logLevel,
	)
}

func (zap *ZapLogger) Info(msg string, keysAndValues ...any) {
	zap.logger.Infow(msg, keysAndValues...)
}

func (zap *ZapLogger) Error(msg string, keysAndValues ...any) {
	zap.logger.Errorw(msg, keysAndValues...)
}

func (zap *ZapLogger) Debug(msg string, keysAndValues ...any) {
	zap.logger.Debugw(msg, keysAndValues...)
}

func (zap *ZapLogger) Warn(msg string, keysAndValues ...any) {
	zap.logger.Warnw(msg, keysAndValues...)
}

func (z *ZapLogger) With(keysAndValues ...any) Logger {
	return &ZapLogger{
		logger: z.logger.With(keysAndValues...),
	}
}

func (z *ZapLogger) WithGroup(name string) Logger {
	return &ZapLogger{
		logger: z.logger.Desugar().
			With(zap.Namespace(name)).
			Sugar(),
	}
}
