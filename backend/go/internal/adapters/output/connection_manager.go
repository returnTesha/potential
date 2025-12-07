// Package output은 외부 시스템(DB)과의 실제 연결을 구현합니다.
// 이 패키지는:
// 1. Output Port 인터페이스를 구현합니다
// 2. 실제 DB 드라이버를 사용합니다 (lib/pq, godror 등)
// 3. 여러 DB 연결을 동시에 관리합니다
package output

import (
	"context"
	"database/sql" // 표준 라이브러리: DB 인터페이스
	"fmt"
	"space/internal/adapters/output/oracle19c"
	"space/internal/adapters/output/postgres"
	"sync" // 동시성 제어를 위한 패키지
	"time"

	// Domain import
	"space/internal/domain"

	// Output Port import (구현할 인터페이스)
	"space/internal/ports/output"
)

// ConnectionManager는 여러 데이터베이스 연결을 관리하는 구조체입니다.
// 이것은 output.DatabaseRepository 인터페이스를 구현합니다.
//
// 핵심 책임:
// 1. 여러 DB 연결을 동시에 관리 (Connection Pool)
// 2. DB 타입별 Adapter 선택
// 3. 동시성 안전 보장 (여러 고루틴이 동시 접근 가능)
type ConnectionManager struct {
	// connections는 dbID를 키로, Connection을 값으로 하는 맵입니다.
	// map[키타입]값타입 형태로 선언합니다.
	//
	// 예: map["postgres-prod"] = &Connection{...}
	connections map[string]*Connection

	// mu는 Mutex(뮤텍스)로, 동시성 제어를 위한 잠금 장치입니다.
	// sync.RWMutex는 읽기/쓰기 잠금을 분리해서 성능을 높입니다.
	//
	// RWMutex 특징:
	// - RLock(): 여러 고루틴이 동시에 읽기 가능
	// - Lock(): 쓰기는 한 번에 하나만 가능
	//
	// 왜 필요한가?
	// → 여러 HTTP 요청이 동시에 connections map을 읽거나 쓸 수 있기 때문!
	mu sync.RWMutex
}

// Connection은 하나의 데이터베이스 연결 정보를 담습니다.
type Connection struct {
	// DB는 Domain 정보 (ID, Type, Host 등)
	DB *domain.Database

	// ConnPool은 실제 데이터베이스 연결 풀입니다.
	// *sql.DB는 Go 표준 라이브러리의 DB 연결 객체
	//
	// Connection Pool이란?
	// → 미리 여러 개의 DB 연결을 만들어두고 재사용하는 것
	// → 매번 새로 연결하는 것보다 훨씬 빠름!
	ConnPool *sql.DB

	// Adapter는 DB 타입별 전용 구현체입니다.
	// Postgres는 PostgresAdapter, Oracle은 OracleAdapter 등
	Adapter Adapter
}

// Adapter는 DB별 전용 기능을 정의하는 인터페이스입니다.
// 각 DB(Postgres, Oracle, MariaDB)마다 연결 방법과 쿼리 방식이 다르므로
// 이렇게 추상화합니다.
type Adapter interface {
	// Connect는 실제 DB에 연결을 생성합니다.
	Connect(ctx context.Context, db *domain.Database) (*sql.DB, error)

	// ExecuteQuery는 쿼리를 실행하고 결과를 반환합니다.
	ExecuteQuery(ctx context.Context, conn *sql.DB, query string) (*domain.QueryResult, error)

	// GetTables는 테이블 목록을 조회합니다.
	// (DB마다 쿼리가 다름!)
	GetTables(ctx context.Context, conn *sql.DB) ([]string, error)

	// GetColumns는 특정 테이블의 컬럼 목록을 조회합니다.
	GetColumns(ctx context.Context, conn *sql.DB, tableName string) ([]string, error)
}

// NewConnectionManager는 ConnectionManager를 생성합니다.
//
// Go 관례:
// - 생성자 함수는 New로 시작
// - 인터페이스 타입을 반환 (output.DatabaseRepository)
//
// 반환 타입이 인터페이스인 이유:
// → 사용하는 쪽(Core)이 구체 타입을 알 필요 없게 하기 위해!
func NewConnectionManager() output.DatabaseRepository {
	return &ConnectionManager{
		// make()는 맵을 초기화하는 내장 함수입니다.
		// make(map[키타입]값타입) 형태
		//
		// 왜 make가 필요한가?
		// → 맵은 반드시 초기화해야 사용 가능
		// → 초기화 없이 사용하면 panic(런타임 에러) 발생!
		connections: make(map[string]*Connection),
	}
}

// Connect는 새로운 데이터베이스에 연결합니다.
// 이 메서드는 output.DatabaseRepository 인터페이스를 구현합니다.
func (cm *ConnectionManager) Connect(ctx context.Context, db *domain.Database) error {
	// ==========================================
	// 1단계: 쓰기 잠금 (Lock)
	// ==========================================

	// cm.mu.Lock()은 다른 고루틴이 접근하지 못하게 잠급니다.
	// 쓰기 작업이므로 독점 잠금 필요!
	cm.mu.Lock()

	// defer는 함수가 종료될 때 자동으로 실행됩니다.
	// defer cm.mu.Unlock()은 "함수 끝날 때 잠금 해제"를 보장
	//
	// 왜 defer를 쓰나?
	// → 중간에 return해도 자동으로 Unlock 됨!
	// → 잠금 해제를 까먹을 일이 없음!
	defer cm.mu.Unlock()

	// ==========================================
	// 2단계: 중복 연결 체크
	// ==========================================

	// 맵에서 값 확인: value, ok := map[key]
	// ok는 키가 존재하는지 여부 (true/false)
	// _는 value를 사용하지 않는다는 의미
	//
	// "comma ok idiom"이라고 부릅니다.
	if _, exists := cm.connections[db.ID]; exists {
		// 이미 존재하면 에러 반환
		return domain.ErrAlreadyConnected
	}

	// ==========================================
	// 3단계: DB 타입별 Adapter 선택
	// ==========================================

	// createAdapter()는 DB 타입에 맞는 Adapter를 생성합니다.
	// (아래에서 구현)
	adapter, err := cm.createAdapter(db.Type)
	if err != nil {
		return err
	}

	// ==========================================
	// 4단계: 실제 DB 연결! 🔥
	// ==========================================

	// adapter.Connect()가 실제로 DB에 연결합니다!
	// 예: PostgreSQL이면 "postgres://user:pass@host:port/db" 형태로 연결
	connPool, err := adapter.Connect(ctx, db)
	if err != nil {
		// 연결 실패
		// %w는 원본 에러를 포함(wrap)
		return fmt.Errorf("failed to connect to %s: %w", db.ID, err)
	}

	// ==========================================
	// 5단계: Connection Pool 설정
	// ==========================================

	// SetMaxOpenConns는 최대 동시 연결 수를 설정합니다.
	// 25개 = 동시에 최대 25개의 쿼리 실행 가능
	connPool.SetMaxOpenConns(25)

	// SetMaxIdleConns는 유휴(idle) 연결을 최대 몇 개 유지할지 설정
	// 5개 = 사용하지 않는 연결을 5개까지 유지 (재사용 위해)
	connPool.SetMaxIdleConns(5)

	// SetConnMaxLifetime은 연결의 최대 수명을 설정
	// 5분 = 5분 후 연결을 닫고 새로 만듦 (오래된 연결 방지)
	connPool.SetConnMaxLifetime(5 * time.Minute)

	// ==========================================
	// 6단계: Ping으로 실제 연결 확인! 🔥
	// ==========================================

	// PingContext는 실제로 DB에 신호를 보내서 연결을 확인합니다.
	// 연결은 됐지만 실제로 통신이 안 될 수도 있기 때문!
	if err := connPool.PingContext(ctx); err != nil {
		// Ping 실패하면 연결 닫기
		connPool.Close()
		return fmt.Errorf("ping failed for %s: %w", db.ID, err)
	}

	// ==========================================
	// 7단계: 연결 정보 저장
	// ==========================================

	// &Connection{...}는 Connection 구조체 포인터 생성
	// 맵에 저장: map[키] = 값
	cm.connections[db.ID] = &Connection{
		DB:       db,
		ConnPool: connPool,
		Adapter:  adapter,
	}

	// ==========================================
	// 8단계: 상태 업데이트! 🔥
	// ==========================================

	// db는 포인터이므로, 여기서 변경하면 원본도 변경됩니다!
	db.Status = domain.Connected

	// 성공!
	return nil
}

// createAdapter는 DB 타입에 맞는 Adapter를 생성합니다.
// private 메서드 (소문자 시작) - 외부에서 호출 불가
func (cm *ConnectionManager) createAdapter(dbType domain.DatabaseType) (Adapter, error) {
	// switch로 DB 타입별 분기
	switch dbType {
	case domain.PostgreSQL:
		// PostgresAdapter 생성
		// postgres 패키지의 NewAdapter() 함수 호출
		return postgres.NewAdapter(), nil

	case domain.Oracle19c:
		// Oracle11g와 Oracle19c는 같은 Adapter 사용
		// 콤마로 여러 case를 한 번에 처리 가능!
		return oracle19c.NewAdapter(), nil

	case domain.Oracle11g:
		return oracle19c.NewAdapter(), nil
	//case domain.MariaDB:
	//	// MariaDB Adapter 생성
	//	return mariadb.NewAdapter(), nil

	default:
		// 지원하지 않는 타입
		return nil, domain.ErrInvalidDatabaseType
	}
}

// Disconnect는 데이터베이스 연결을 종료합니다.
func (cm *ConnectionManager) Disconnect(ctx context.Context, dbID string) error {
	// 쓰기 작업이므로 Lock (독점)
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 연결 찾기
	conn, exists := cm.connections[dbID]
	if !exists {
		// 없으면 에러
		return domain.ErrDatabaseNotFound
	}

	// Connection Pool 닫기
	// Close()는 모든 연결을 정리하고 종료합니다.
	if err := conn.ConnPool.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}

	// 상태 업데이트
	conn.DB.Status = domain.Disconnected

	// 맵에서 제거
	// delete()는 맵에서 키를 삭제하는 내장 함수
	delete(cm.connections, dbID)

	return nil
}

// ExecuteQuery는 특정 DB에 쿼리를 실행합니다.
func (cm *ConnectionManager) ExecuteQuery(ctx context.Context, dbID string, query string) (*domain.QueryResult, error) {
	// ==========================================
	// 1단계: 읽기 잠금 (RLock)
	// ==========================================

	// 읽기 작업이므로 RLock (여러 고루틴 동시 읽기 가능)
	cm.mu.RLock()

	// 연결 찾기
	conn, exists := cm.connections[dbID]

	// 읽기 끝나면 잠금 해제
	// 여기서 RUnlock하는 이유:
	// → 쿼리 실행은 오래 걸릴 수 있으므로, 빨리 잠금 해제
	// → conn은 복사했으므로 안전
	cm.mu.RUnlock()

	// ==========================================
	// 2단계: 연결 존재 확인
	// ==========================================

	if !exists {
		return nil, domain.ErrDatabaseNotFound
	}

	// ==========================================
	// 3단계: 쿼리 실행! 🔥
	// ==========================================

	// Adapter의 ExecuteQuery() 호출
	// 실제로 DB에 쿼리를 보냅니다!
	result, err := conn.Adapter.ExecuteQuery(ctx, conn.ConnPool, query)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	return result, nil
}

// IsConnected는 특정 DB가 연결되어 있는지 확인합니다.
func (cm *ConnectionManager) IsConnected(ctx context.Context, dbID string) bool {
	// 읽기 잠금
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// 연결 찾기
	conn, exists := cm.connections[dbID]
	if !exists {
		return false
	}

	// 실제 Ping으로 확인! 🔥
	// 맵에는 있지만 실제 연결이 끊어졌을 수도 있음
	if err := conn.ConnPool.PingContext(ctx); err != nil {
		// Ping 실패하면 상태 업데이트
		conn.DB.Status = domain.Disconnected
		return false
	}

	// 연결 상태 확인
	return conn.DB.Status == domain.Connected
}

// GetTables는 특정 DB의 테이블 목록을 조회합니다.
func (cm *ConnectionManager) GetTables(ctx context.Context, dbID string) ([]string, error) {
	// 읽기 잠금
	cm.mu.RLock()
	conn, exists := cm.connections[dbID]
	cm.mu.RUnlock()

	if !exists {
		return nil, domain.ErrDatabaseNotFound
	}

	// Adapter의 GetTables() 호출
	// DB 타입별로 다른 쿼리가 실행됨!
	tables, err := conn.Adapter.GetTables(ctx, conn.ConnPool)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}

	return tables, nil
}

// GetColumns는 특정 테이블의 컬럼 목록을 조회합니다.
func (cm *ConnectionManager) GetColumns(ctx context.Context, dbID string, tableName string) ([]string, error) {
	cm.mu.RLock()
	conn, exists := cm.connections[dbID]
	cm.mu.RUnlock()

	if !exists {
		return nil, domain.ErrDatabaseNotFound
	}

	// Adapter의 GetColumns() 호출
	columns, err := conn.Adapter.GetColumns(ctx, conn.ConnPool, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	return columns, nil
}

// ListConnections는 현재 관리 중인 모든 연결 목록을 반환합니다.
func (cm *ConnectionManager) ListConnections(ctx context.Context) ([]*domain.Database, error) {
	// 읽기 잠금
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// make()로 슬라이스 생성
	// make([]타입, 초기길이, 용량)
	// 0은 초기 길이, len(cm.connections)는 용량
	//
	// 왜 용량을 미리 지정?
	// → append할 때 메모리 재할당을 줄여서 성능 향상!
	databases := make([]*domain.Database, 0, len(cm.connections))

	// 맵 순회: for 키, 값 := range 맵
	for _, conn := range cm.connections {
		// append()로 슬라이스에 추가
		databases = append(databases, conn.DB)
	}

	return databases, nil
}

// DisconnectAll은 모든 연결을 종료합니다.
// 서버 종료 시 호출하면 좋습니다.
func (cm *ConnectionManager) DisconnectAll(ctx context.Context) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 에러를 모아둘 슬라이스
	// 일부 연결이 실패해도 나머지는 계속 종료
	var errors []error

	// 모든 연결 순회
	for dbID, conn := range cm.connections {
		// 연결 닫기
		if err := conn.ConnPool.Close(); err != nil {
			// 에러 발생해도 계속 진행
			// append()로 에러 추가
			errors = append(errors, fmt.Errorf("failed to close %s: %w", dbID, err))
		}

		// 상태 업데이트
		conn.DB.Status = domain.Disconnected
	}

	// 맵 초기화 (모든 항목 삭제)
	cm.connections = make(map[string]*Connection)

	// 에러가 있었다면 반환
	if len(errors) > 0 {
		return fmt.Errorf("errors during disconnect: %v", errors)
	}

	return nil
}
