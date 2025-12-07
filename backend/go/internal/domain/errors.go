package domain

import (
	"errors"
	"fmt"
)

// Go에서 에러는 보통 변수로 미리 정의합니다.
// var 키워드로 패키지 레벨 변수를 선언합니다.
// 관례적으로 에러 변수명은 Err로 시작합니다.
var (
	// Connection 관련 에러
	// errors.New()는 간단한 에러를 만듭니다
	ErrDatabaseNotFound     = errors.New("database not found")
	ErrDatabaseNotConnected = errors.New("database is not connected")
	ErrAlreadyConnected     = errors.New("database already connected")
	ErrConnectionFailed     = errors.New("failed to connect to database")

	// Query 관련 에러
	ErrQueryTimeout = errors.New("query execution timeout")
	ErrInvalidQuery = errors.New("invalid query")
	ErrQueryFailed  = errors.New("query execution failed")

	// Validation 관련 에러
	ErrInvalidDatabaseType = errors.New("invalid database type")
	ErrInvalidPort         = errors.New("invalid port number")
	ErrMissingCredentials  = errors.New("missing credentials")

	// General 에러
	ErrInternal = errors.New("internal error")
)

// 에러를 이렇게 미리 정의하면 좋은 점:
// 1. 에러 비교가 쉬움: if err == domain.ErrDatabaseNotFound { ... }
// 2. 에러 메시지 일관성
// 3. 문서화 용이

// DomainError는 도메인 에러에 추가 정보를 담기 위한 커스텀 에러입니다.
// Go에서 커스텀 에러를 만들려면 error 인터페이스를 구현하면 됩니다.
type DomainError struct {
	Code    string // 에러 코드 (예: "DB_NOT_FOUND")
	Message string // 사람이 읽을 수 있는 메시지
	Err     error  // 원본 에러 (wrapping)
}

// Error() 메서드를 구현하면 error 인터페이스를 만족합니다.
// Go의 인터페이스는 명시적으로 선언하지 않아도 됩니다 (duck typing)
func (e *DomainError) Error() string {
	// 원본 에러가 있으면 함께 표시
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap은 원본 에러를 반환합니다.
// Go 1.13부터 에러 체이닝을 지원합니다.
// errors.Is()나 errors.As()와 함께 사용됩니다.
func (e *DomainError) Unwrap() error {
	return e.Err
}

// NewDomainError는 DomainError를 생성하는 헬퍼 함수입니다.
// Go에서는 생성자 대신 New로 시작하는 함수를 많이 사용합니다.
func NewDomainError(code, message string, err error) *DomainError {
	// &는 주소 연산자로, struct의 포인터를 반환합니다
	// Go에서는 struct 리터럴을 이렇게 만듭니다
	return &DomainError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}
