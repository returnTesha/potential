// Package input은 "외부에서 우리 시스템으로 들어오는" 인터페이스를 정의합니다.
// Input Port = Driving Port = Primary Port
// "누군가 우리를 호출한다" → HTTP, CLI, gRPC 등이 이 인터페이스를 사용
package input

import (
	"context"

	// Domain만 import! (안쪽만 의존)
	"space/internal/domain"
)

// DatabaseService는 DMS의 핵심 Use Case를 정의하는 인터페이스입니다.
// 이것은 "계약서"입니다. "이런 기능을 제공할게!"라고 약속하는 것.
//
// Go의 인터페이스 특징:
// 1. 메서드 시그니처만 정의 (구현 없음)
// 2. 명시적으로 구현 선언 안 해도 됨 (duck typing)
// 3. 관례: 인터페이스 이름은 -er로 끝남 (Reader, Writer, Service...)
type DatabaseService interface {
	// RegisterDatabase는 새로운 데이터베이스 연결을 등록합니다.
	//
	// 파라미터:
	//   - ctx: context.Context - 타임아웃, 취소 등을 위한 컨텍스트
	//   - db: *domain.Database - 등록할 데이터베이스 정보
	//
	// 반환값:
	//   - error: 성공하면 nil, 실패하면 에러
	//
	// 비즈니스 규칙:
	//   - db.Validate()가 성공해야 함
	//   - 중복된 ID는 허용하지 않음
	//   - 실제 DB 연결까지 성공해야 함
	RegisterDatabase(ctx context.Context, db *domain.Database) error

	// ExecuteQuery는 특정 데이터베이스에 쿼리를 실행합니다.
	//
	// 파라미터:
	//   - ctx: context.Context - 쿼리 타임아웃 설정 가능
	//   - dbID: string - 데이터베이스 고유 ID (예: "postgres-prod")
	//   - query: string - 실행할 SQL 쿼리
	//
	// 반환값:
	//   - *domain.QueryResult: 쿼리 실행 결과
	//   - error: 에러 발생 시
	//
	// 주의사항:
	//   - dbID에 해당하는 DB가 연결되어 있어야 함
	//   - 악의적인 쿼리 방지는 어댑터에서 처리 (여기는 계약만)
	ExecuteQuery(ctx context.Context, dbID string, query string) (*domain.QueryResult, error)

	// ListDatabases는 현재 연결된 모든 데이터베이스 목록을 반환합니다.
	//
	// 반환값:
	//   - []*domain.Database: 연결된 DB 목록 (슬라이스)
	//   - error: 에러 발생 시
	ListDatabases(ctx context.Context) ([]*domain.Database, error)

	// DisconnectDatabase는 특정 데이터베이스 연결을 종료합니다.
	//
	// 파라미터:
	//   - dbID: string - 종료할 데이터베이스 ID
	//
	// 반환값:
	//   - error: 성공하면 nil
	DisconnectDatabase(ctx context.Context, dbID string) error

	// GetDatabaseInfo는 특정 데이터베이스의 정보를 조회합니다.
	//
	// 파라미터:
	//   - dbID: string - 조회할 데이터베이스 ID
	//
	// 반환값:
	//   - *domain.Database: DB 정보 (비밀번호는 마스킹됨)
	//   - error: DB를 찾을 수 없으면 domain.ErrDatabaseNotFound
	GetDatabaseInfo(ctx context.Context, dbID string) (*domain.Database, error)
}

// Go 인터페이스 핵심 개념:
//
// 1. 암묵적 구현 (Implicit Implementation)
//    Java처럼 "implements DatabaseService"라고 명시하지 않아도 됨!
//    메서드만 똑같이 구현하면 자동으로 인터페이스를 만족함
//
// 2. 작은 인터페이스 선호
//    Go 관례: 인터페이스는 작을수록 좋음
//    이상적: 1-3개 메서드
//    우리는 5개 (좀 크긴 하지만 관련된 것들이라 OK)
//
// 3. 사용하는 쪽에서 정의
//    인터페이스는 "사용하는 쪽(consumer)"이 정의함
//    HTTP Handler가 이 인터페이스를 사용할 것!

// 예시: 이 인터페이스를 어떻게 사용할까?
//
// type HTTPHandler struct {
//     service input.DatabaseService  // 인터페이스!
// }
//
// func (h *HTTPHandler) HandleRegister(c *gin.Context) {
//     var db domain.Database
//     // ... JSON 파싱 ...
//     err := h.service.RegisterDatabase(ctx, &db)  // 인터페이스 메서드 호출
//     // ...
// }
