// Package output은 "우리 시스템에서 외부로 나가는" 인터페이스를 정의합니다.
// Output Port = Driven Port = Secondary Port
// "우리가 외부를 호출한다" → DB, 외부 API 등에 요청을 보냄
package output

import (
	"context"

	// Domain만 import!
	"space/internal/domain"
)

// DatabaseRepository는 데이터베이스 연결과 쿼리 실행을 담당하는 인터페이스입니다.
// 이것은 Core가 "나는 이런 기능이 필요해!"라고 요구하는 계약서입니다.
//
// 중요: 이 인터페이스는 Core가 정의하고, Adapter가 구현합니다!
// (Input Port와 방향이 반대)
type DatabaseRepository interface {
	// Connect는 데이터베이스에 실제 연결을 생성합니다.
	//
	// 파라미터:
	//   - ctx: context.Context - 연결 타임아웃 설정
	//   - db: *domain.Database - 연결할 DB 정보
	//
	// 반환값:
	//   - error: 연결 실패 시 에러 (domain.ErrConnectionFailed 등)
	//
	// 구현 책임:
	//   - 실제 DB 드라이버 사용 (lib/pq, godror 등)
	//   - Connection Pool 생성
	//   - Ping으로 연결 확인
	//   - db.Status를 Connected로 변경
	Connect(ctx context.Context, db *domain.Database) error

	// Disconnect는 데이터베이스 연결을 종료합니다.
	//
	// 파라미터:
	//   - dbID: string - 종료할 DB의 ID
	//
	// 구현 책임:
	//   - Connection Pool 닫기
	//   - 리소스 정리
	//   - db.Status를 Disconnected로 변경
	Disconnect(ctx context.Context, dbID string) error

	// ExecuteQuery는 특정 DB에 쿼리를 실행합니다.
	//
	// 파라미터:
	//   - dbID: string - 대상 DB ID
	//   - query: string - 실행할 SQL
	//
	// 반환값:
	//   - *domain.QueryResult: 결과
	//   - error: 실행 실패 시
	//
	// 구현 책임:
	//   - 해당 dbID의 connection 찾기
	//   - conn.QueryContext() 실행
	//   - 결과를 domain.QueryResult로 변환
	//   - 실행 시간 측정
	ExecuteQuery(ctx context.Context, dbID string, query string) (*domain.QueryResult, error)

	// IsConnected는 특정 DB가 연결되어 있는지 확인합니다.
	//
	// 파라미터:
	//   - dbID: string - 확인할 DB ID
	//
	// 반환값:
	//   - bool: 연결되어 있으면 true
	//
	// 구현 책임:
	//   - Connection Pool에 해당 DB 있는지 확인
	//   - 실제 Ping으로 연결 상태 확인
	IsConnected(ctx context.Context, dbID string) bool

	// GetTables는 특정 DB의 테이블 목록을 조회합니다.
	//
	// 파라미터:
	//   - dbID: string - 대상 DB ID
	//
	// 반환값:
	//   - []string: 테이블 이름 목록
	//   - error: 조회 실패 시
	//
	// 구현 책임:
	//   - DB 타입별로 다른 쿼리 실행
	//   - Postgres: SELECT tablename FROM pg_tables WHERE schemaname='public'
	//   - Oracle: SELECT table_name FROM user_tables
	//   - MariaDB: SHOW TABLES
	GetTables(ctx context.Context, dbID string) ([]string, error)

	// GetColumns는 특정 테이블의 컬럼 정보를 조회합니다.
	//
	// 파라미터:
	//   - dbID: string - 대상 DB ID
	//   - tableName: string - 테이블 이름
	//
	// 반환값:
	//   - []string: 컬럼 이름 목록
	//   - error: 조회 실패 시
	GetColumns(ctx context.Context, dbID string, tableName string) ([]string, error)

	// ListConnections는 현재 관리 중인 모든 DB 연결 목록을 반환합니다.
	//
	// 반환값:
	//   - []*domain.Database: 모든 DB 정보
	//   - error: 조회 실패 시
	//
	// 구현 책임:
	//   - Connection Manager의 모든 연결 반환
	ListConnections(ctx context.Context) ([]*domain.Database, error)
}

// Output Port 특징:
//
// 1. Core가 필요로 하는 기능을 정의
//    Core: "나는 DB에 연결하고 쿼리 실행할 수 있어야 해!"
//    Adapter: "알았어, 내가 구현해줄게!"
//
// 2. 구현 방법은 알 필요 없음
//    이 인터페이스는 "어떻게(How)"는 모름
//    "무엇을(What)" 해야 하는지만 정의
//
// 3. 교체 가능성
//    Postgres Adapter 구현 → Oracle Adapter로 교체 가능
//    인터페이스만 만족하면 됨!

// 예시: Core에서 어떻게 사용할까?
//
// type DatabaseService struct {
//     repo output.DatabaseRepository  // 인터페이스!
// }
//
// func (s *DatabaseService) RegisterDatabase(ctx, db) error {
//     // 비즈니스 로직
//     if err := db.Validate(); err != nil {
//         return err
//     }
//
//     // Output Port 호출
//     return s.repo.Connect(ctx, db)  // 어떻게 연결되는지 몰라도 됨!
// }
