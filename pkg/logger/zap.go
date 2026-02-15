package logger

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/Roh-Bot/blog-api/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const RequestIDKey = "request_id"

type AsyncZapLogger struct {
	logger     *zap.Logger
	queue      chan logEntry
	quit       chan struct{}
	dropped    uint64
	batchSize  int
	flushDelay time.Duration
}

type logEntry struct {
	level  zapcore.Level
	msg    string
	fields []zap.Field
	ctx    context.Context
}

// NewAsyncZapLogger creates a non-blocking, async Zap logger.
func ZapNew(cfg config.Logger, cores ...zapcore.Core) (*AsyncZapLogger, error) {
	//Wrapping stdout core in buffered sink for bulk writes to stdout
	bufferedSink := &zapcore.BufferedWriteSyncer{
		WS:            StdoutSink(),
		Size:          256 * 1024,
		FlushInterval: time.Second * 2,
	}

	stdoutCore := NewZapCore(bufferedSink, cfg.Level)

	z := zap.New(
		stdoutCore,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	l := &AsyncZapLogger{
		logger:     z,
		queue:      make(chan logEntry, cfg.BufferSize),
		quit:       make(chan struct{}),
		batchSize:  cfg.BatchSize,  // configurable for high throughput
		flushDelay: cfg.FlushDelay, // flush interval for batching
	}

	go l.worker()
	return l, nil
}

// NewZapCore helper for standardized encoding and level parsing
func NewZapCore(syncer zapcore.WriteSyncer, level string) zapcore.Core {
	return zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig()),
		zapcore.Lock(syncer),
		parseLevel(level),
	)
}

func encoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		MessageKey:     "Msg",
		LevelKey:       "Level",
		TimeKey:        "Timestamp",
		NameKey:        "logger",
		CallerKey:      "Caller",
		StacktraceKey:  "stacktrace",
		FunctionKey:    zapcore.OmitKey,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

func parseLevel(lvl string) zapcore.Level {
	switch lvl {
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	case "panic":
		return zap.PanicLevel
	case "fatal":
		return zap.FatalLevel
	default:
		return zap.InfoLevel
	}
}

// --- Worker and async internals ---

func (l *AsyncZapLogger) worker() {
	defer func() {
		if r := recover(); r != nil {
			l.logger.Error("logger worker panic", zap.Any("recover", r))
		}
	}()

	ticker := time.NewTicker(l.flushDelay)
	defer ticker.Stop()

	batch := make([]logEntry, 0, l.batchSize)

	for {
		select {
		case entry := <-l.queue:
			batch = append(batch, entry)
			if len(batch) >= l.batchSize {
				l.writeBatch(batch)
				batch = batch[:0]
			}

			// Optional: log queue saturation
			if len(l.queue) > cap(l.queue)*9/10 {
				l.logger.Warn("log queue nearing capacity",
					zap.Int("current", len(l.queue)),
					zap.Int("capacity", cap(l.queue)))
			}

		case <-ticker.C:
			if len(batch) > 0 {
				l.writeBatch(batch)
				batch = nil
			}

		case <-l.quit:
			// drain remaining logs
			for {
				select {
				case entry := <-l.queue:
					batch = append(batch, entry)
				default:
					if len(batch) > 0 {
						l.writeBatch(batch)
					}
					return
				}
			}
		}
	}
}

func (l *AsyncZapLogger) writeBatch(batch []logEntry) {
	for _, entry := range batch {
		l.writeEntry(entry)
	}
}

func (l *AsyncZapLogger) writeEntry(entry logEntry) {
	switch entry.level {
	case zapcore.DebugLevel:
		l.logger.Debug(entry.msg, entry.fields...)
	case zapcore.InfoLevel:
		l.logger.Info(entry.msg, entry.fields...)
	case zapcore.WarnLevel:
		l.logger.Warn(entry.msg, entry.fields...)
	case zapcore.ErrorLevel:
		l.logger.Error(entry.msg, entry.fields...)
	default:
		l.logger.Info(entry.msg, entry.fields...)
	}
}

func (l *AsyncZapLogger) enqueue(entry logEntry) {
	select {
	case l.queue <- entry:
	default:
		atomic.AddUint64(&l.dropped, 1)
	}
}

// --- Public Logging API ---

func (l *AsyncZapLogger) log(ctx context.Context, level zapcore.Level, msg string, fields []zap.Field) {
	// Add RequestID without map allocation
	if reqID, ok := ctx.Value(RequestIDKey).(string); ok {
		fields = append(fields, zap.String(RequestIDKey, reqID))
	}

	entry := logEntry{
		level:  level,
		msg:    msg,
		fields: fields,
	}

	select {
	case l.queue <- entry:
	default:
		atomic.AddUint64(&l.dropped, 1)
	}
}

func (l *AsyncZapLogger) Info(ctx context.Context, msg string, fields ...zapcore.Field) {
	l.log(ctx, zapcore.InfoLevel, msg, fields)
}

func (l *AsyncZapLogger) Error(ctx context.Context, msg string, fields ...zapcore.Field) {
	l.log(ctx, zapcore.ErrorLevel, msg, fields)
}

func (l *AsyncZapLogger) Warn(ctx context.Context, msg string, fields ...zapcore.Field) {
	l.log(ctx, zapcore.WarnLevel, msg, fields)
}

func (l *AsyncZapLogger) Debug(ctx context.Context, msg string, fields ...zapcore.Field) {
	l.log(ctx, zapcore.DebugLevel, msg, fields)
}

// Flush gracefully stops the logger and flushes the remaining entries.
func (l *AsyncZapLogger) Flush() error {
	close(l.quit)
	return l.logger.Sync()
}

// DroppedCount returns the number of dropped log entries.
func (l *AsyncZapLogger) DroppedCount() uint64 {
	return atomic.LoadUint64(&l.dropped)
}

// With allows you to add default fields to the logger.
func (l *AsyncZapLogger) With(fields ...zapcore.Field) *AsyncZapLogger {
	newLogger := *l
	newLogger.logger = l.logger.With(fields...)
	return &newLogger
}
