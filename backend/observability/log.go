// Package observability は秘密値を受け取らない構造化ログを提供します。
package observability

import (
	"context"
	"io"
	"log/slog"

	"github.com/yukihito-jokyu/topic2html/backend/apperr"
)

// Loggerは安全に固定したイベントだけを記録します。
type Logger struct {
	logger *slog.Logger
}

// EventLoggerは各層が安全なイベントを記録するための共通契約です。
type EventLogger interface {
	Info(context.Context, string)
	Error(context.Context, string, error)
	RequestCompleted(context.Context, string, string, int)
}

// NewLoggerはJSON形式の構造化ログを作成します。
func NewLogger(writer io.Writer) *Logger {
	return &Logger{
		logger: slog.New(slog.NewJSONHandler(writer, nil)),
	}
}

// NewDiscardLoggerは出力しないLoggerを作成します。
func NewDiscardLogger() *Logger {
	return NewLogger(io.Discard)
}

// Infoは値を伴わない成功イベントを記録します。
func (l *Logger) Info(ctx context.Context, event string) {
	l.logger.InfoContext(ctx, event)
}

// Errorはアプリケーションエラーの安全な分類コードだけを記録します。
func (l *Logger) Error(ctx context.Context, event string, err error) {
	l.logger.ErrorContext(ctx, event, slog.String("error_code", apperr.CodeOf(err).String()))
}

// RequestCompletedはquery、header、cookie、bodyを出力せずHTTP結果を記録します。
func (l *Logger) RequestCompleted(ctx context.Context, method, route string, status int) {
	l.logger.InfoContext(
		ctx,
		"http.request.completed",
		slog.String("method", method),
		slog.String("route", route),
		slog.Int("status", status),
	)
}
