// Package apperr は外部へ詳細を出さないアプリケーションエラーを定義します。
package apperr

import "errors"

// Codeはログと失敗分類に使う安全なエラー識別子です。
type Code uint8

const (
	CodeInvalidConfiguration Code = iota + 1
	CodeInvalidRequest
	CodeRejected
	CodeUnavailable
	CodeInternal
)

// Errorは原因の詳細を保持・出力しない分類済みエラーです。
type Error struct {
	code Code
}

func (e *Error) Error() string {
	return e.code.String()
}

// Newは分類済みエラーを作成します。
func New(code Code) error {
	return &Error{
		code: code,
	}
}

// CodeOfはエラーの安全な分類コードを返します。
func CodeOf(err error) Code {
	var appError *Error
	if errors.As(err, &appError) {
		return appError.code
	}

	return CodeUnavailable
}

// Stringは固定された安全な分類名を返します。
func (c Code) String() string {
	switch c {
	case CodeInvalidConfiguration:
		return "invalid_configuration"
	case CodeInvalidRequest:
		return "invalid_request"
	case CodeRejected:
		return "rejected"
	case CodeUnavailable:
		return "unavailable"
	case CodeInternal:
		return "internal"
	default:
		return "internal"
	}
}
