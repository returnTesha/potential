// Package postgres는 PostgreSQL 전용 Adapter 구현을 제공합니다.
// 이 패키지는:
// 1. output.Adapter 인터페이스를 구현합니다
// 2. lib/pq 드라이버를 사용합니다 (PostgreSQL 전용 드라이버)
// 3. PostgreSQL 전용 쿼리와 기능을 처리합니다
package postgres

import (
	"context"
	"database/sql" // 표준 라이브러리: DB 인터페이스
	"fmt"
	"time"

	// PostgreSQL 드라이버를 import합니다.
	// _는 "blank identifier"로, 패키지를 직접 사용하지 않지만
	// init() 함수를 실행하기 위해 import합니다.
	//
	// lib/pq의 init()가 sql.Register()를 호출해서
	// "postgres" 드라이버를 등록합니다.
	_ "github.com/lib/pq"

	// Domain import
	"space/internal/domain"
)

// PostgresAdapter는 PostgreSQL 전용 구현체입니다.
// 빈 구조체 (struct{})로 선언했습니다.
//
// 왜 빈 구조체?
// → 이 Adapter는 상태(state)를 가질 필요가 없음
// → 메서드만 제공하면 됨
// → 메모리 0 bytes 사용!
type PostgresAdapter struct{}

// NewAdapter는 PostgresAdapter를 생성합니다.
//
// 반환 타입이 *PostgresAdapter인 이유:
// → ConnectionManager가 구체 타입을 알아야 하므로
// → (인터페이스 타입이 아닌 구체 타입 반환)
func NewAdapter() *PostgresAdapter {
	// &PostgresAdapter{}는 빈 구조체의 포인터를 생성
	return &PostgresAdapter{}
}

// Connect는 PostgreSQL 데이터베이스에 실제 연결을 생성합니다.
// 이 메서드는 output.Adapter 인터페이스를 구현합니다.
func (a *PostgresAdapter) Connect(ctx context.Context, db *domain.Database) (*sql.DB, error) {
	// ==========================================
	// 1단계: 연결 문자열(DSN) 생성
	// ==========================================

	// DSN = Data Source Name
	// PostgreSQL 형식: "postgres://user:password@host:port/dbname?options"
	//
	// db.ConnectionString()은 domain에서 이미 구현했습니다!
	// 재사용하는 거예요 (중복 제거!)
	dsn := db.ConnectionString()

	// 디버깅용 로그 (실제로는 password 노출 주의!)
	// fmt.Printf("[Postgres] Connecting to: %s (password hidden)\n", db.SafeString())

	// ==========================================
	// 2단계: 데이터베이스 연결 열기 🔥
	// ==========================================

	// sql.Open()은 DB 연결을 초기화합니다.
	//
	// 중요: sql.Open()은 즉시 연결하지 않습니다!
	// → 단지 연결 정보만 설정
	// → 실제 연결은 처음 쿼리할 때 또는 Ping할 때
	//
	// 파라미터:
	// - "postgres": 드라이버 이름 (lib/pq가 등록한 이름)
	// - dsn: 연결 문자열
	//
	// 반환값:
	// - *sql.DB: DB 연결 객체 (Connection Pool)
	// - error: 에러 (연결 정보 파싱 실패 등)
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		// sql.Open 실패
		// 보통 DSN 형식이 잘못된 경우
		return nil, fmt.Errorf("sql.Open failed: %w", err)
	}

	// ==========================================
	// 3단계: 타임아웃 설정
	// ==========================================

	// context.WithTimeout은 새로운 context를 만듭니다.
	// 원본 ctx에 타임아웃을 추가한 것
	//
	// 5*time.Second = 5초
	// → 5초 안에 연결 안 되면 자동으로 취소
	//
	// cancel은 타임아웃을 수동으로 취소하는 함수
	// defer cancel()로 함수 종료 시 자동 호출
	//
	// 왜 defer cancel()?
	// → context 리소스 누수 방지
	// → 타임아웃 고루틴을 정리
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// ==========================================
	// 4단계: 실제 연결 확인! 🔥
	// ==========================================

	// PingContext는 실제로 DB에 신호를 보냅니다.
	// "ping" = "너 살아있니?"
	//
	// 이 시점에 실제로:
	// 1. TCP 연결 생성
	// 2. PostgreSQL 프로토콜 handshake
	// 3. 인증 수행 (username/password)
	//
	// 만약 실패하면:
	// - 호스트에 도달 못함
	// - 포트가 닫혀있음
	// - 비밀번호 틀림
	// - DB가 없음
	// 등의 에러 발생
	if err := conn.PingContext(ctx); err != nil {
		// Ping 실패하면 연결 닫기
		// Close()는 모든 리소스를 정리합니다
		conn.Close()

		return nil, fmt.Errorf("ping failed: %w", err)
	}

	// ==========================================
	// 5단계: 연결 성공! ✅
	// ==========================================

	// fmt.Printf("[Postgres] Successfully connected to %s\n", db.ID)

	// *sql.DB 반환
	// 이것은 Connection Pool입니다!
	return conn, nil
}

// ExecuteQuery는 PostgreSQL에 쿼리를 실행하고 결과를 반환합니다.
func (a *PostgresAdapter) ExecuteQuery(ctx context.Context, conn *sql.DB, query string) (*domain.QueryResult, error) {
	// ==========================================
	// 1단계: 실행 시간 측정 시작
	// ==========================================

	// time.Now()는 현재 시각을 반환
	// 나중에 time.Since(start)로 경과 시간 계산
	start := time.Now()

	// ==========================================
	// 2단계: 쿼리 실행! 🔥
	// ==========================================

	// conn.QueryContext()는 SELECT 쿼리를 실행합니다.
	//
	// QueryContext vs Query:
	// - QueryContext: context를 받음 (타임아웃, 취소 가능) ✅
	// - Query: context 없음 (구식)
	//
	// 파라미터:
	// - ctx: context (타임아웃 설정 등)
	// - query: SQL 쿼리 문자열
	//
	// 반환값:
	// - *sql.Rows: 쿼리 결과 (여러 row)
	// - error: 쿼리 실패 시
	//
	// 주의: Rows는 반드시 Close()해야 합니다!
	rows, err := conn.QueryContext(ctx, query)
	if err != nil {
		// 쿼리 실패 (문법 에러, 테이블 없음 등)
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	// defer rows.Close()는 함수 종료 시 자동으로 Close
	// 왜 중요한가?
	// → Close하지 않으면 connection이 반환되지 않음
	// → Connection Pool이 고갈될 수 있음!
	defer rows.Close()

	// ==========================================
	// 3단계: 컬럼 정보 가져오기
	// ==========================================

	// rows.Columns()는 결과의 컬럼 이름들을 반환
	// 예: ["id", "name", "email"]
	//
	// 반환값:
	// - []string: 컬럼 이름 슬라이스
	// - error: 에러 (거의 발생 안 함)
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	// ==========================================
	// 4단계: Row 데이터 파싱
	// ==========================================

	// 결과를 담을 슬라이스
	// []map[string]interface{}는:
	// - 각 row는 map[string]interface{}
	// - 여러 row를 슬라이스로 담음
	//
	// map[string]interface{}의 의미:
	// - 키: 컬럼 이름 (string)
	// - 값: 컬럼 값 (interface{} = 모든 타입 가능)
	//
	// 예: {"id": 1, "name": "Alice", "email": "alice@example.com"}
	results := []map[string]interface{}{}

	// rows.Next()는 다음 row로 이동합니다.
	// 반환값: bool
	// - true: 다음 row가 있음
	// - false: 더 이상 row가 없음 (루프 종료)
	//
	// Go의 for 루프는 조건만 있으면 됩니다 (while과 비슷)
	for rows.Next() {
		// ==========================================
		// 4-1단계: Row의 값을 담을 공간 준비
		// ==========================================

		// make()로 슬라이스 생성
		// len(columns) = 컬럼 개수만큼
		//
		// values는 실제 값을 담을 슬라이스
		// 예: [1, "Alice", "alice@example.com"]
		values := make([]interface{}, len(columns))

		// valuePtrs는 각 값의 포인터를 담을 슬라이스
		// Scan()은 포인터를 요구하기 때문!
		//
		// 예: [&values[0], &values[1], &values[2]]
		valuePtrs := make([]interface{}, len(columns))

		// 각 값의 주소를 valuePtrs에 저장
		for i := range values {
			// &values[i]는 values[i]의 주소(포인터)
			valuePtrs[i] = &values[i]
		}

		// ==========================================
		// 4-2단계: Row 데이터 읽기
		// ==========================================

		// rows.Scan()은 현재 row의 데이터를 읽습니다.
		//
		// Scan(dest ...interface{})의 의미:
		// - dest: 가변 인자 (여러 개 전달 가능)
		// - ...은 슬라이스를 펼쳐서 전달 (spread operator)
		//
		// valuePtrs...는:
		// - valuePtrs[0], valuePtrs[1], valuePtrs[2], ... 로 펼쳐짐
		//
		// Scan은 각 포인터가 가리키는 곳에 값을 씁니다!
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// ==========================================
		// 4-3단계: map으로 변환
		// ==========================================

		// make()로 맵 생성
		row := make(map[string]interface{})

		// 컬럼 이름과 값을 매핑
		// for i, col := range columns:
		// - i: 인덱스 (0, 1, 2, ...)
		// - col: 컬럼 이름 ("id", "name", "email")
		for i, col := range columns {
			// 맵에 저장: row[컬럼이름] = 값
			//
			// values[i]는 interface{} 타입
			// 실제로는 int, string, time.Time 등 다양한 타입
			row[col] = values[i]
		}

		// ==========================================
		// 4-4단계: 결과에 추가
		// ==========================================

		// append()로 슬라이스에 row 추가
		results = append(results, row)
	}

	// ==========================================
	// 5단계: Row 순회 중 에러 체크
	// ==========================================

	// rows.Err()는 순회 중 발생한 에러를 반환
	// 예: 네트워크 끊김, context 취소 등
	//
	// 왜 필요한가?
	// → rows.Next()가 false를 반환해도 에러인지 정상 종료인지 모름
	// → rows.Err()로 확인 필수!
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	// ==========================================
	// 6단계: QueryResult 생성
	// ==========================================

	// time.Since(start)는 경과 시간 계산
	// start부터 지금까지의 Duration
	executionTime := time.Since(start)

	// domain.QueryResult 생성
	return &domain.QueryResult{
		Columns:       columns,             // 컬럼 이름들
		Rows:          results,             // 실제 데이터
		RowsAffected:  int64(len(results)), // SELECT는 row 개수
		ExecutionTime: executionTime,       // 실행 시간
	}, nil
}

// GetTables는 PostgreSQL의 모든 테이블 목록을 조회합니다.
// PostgreSQL 전용 쿼리를 사용합니다!
func (a *PostgresAdapter) GetTables(ctx context.Context, conn *sql.DB) ([]string, error) {
	// ==========================================
	// PostgreSQL 전용 쿼리! 🔥
	// ==========================================

	// pg_tables는 PostgreSQL 시스템 카탈로그
	// schemaname='public'은 public 스키마의 테이블만
	//
	// 다른 DB는 쿼리가 다름:
	// - Oracle: SELECT table_name FROM user_tables
	// - MySQL: SHOW TABLES
	// - SQL Server: SELECT name FROM sys.tables
	query := `
		SELECT tablename 
		FROM pg_tables 
		WHERE schemaname = 'public'
		ORDER BY tablename
	`

	// QueryContext로 쿼리 실행
	rows, err := conn.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	// 테이블 이름을 담을 슬라이스
	// []string 타입
	var tables []string

	// 각 row 순회
	for rows.Next() {
		// 테이블 이름을 담을 변수
		var tableName string

		// Scan으로 값 읽기
		// 컬럼이 하나뿐이므로 변수 하나만 전달
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("failed to scan table name: %w", err)
		}

		// 슬라이스에 추가
		tables = append(tables, tableName)
	}

	// 에러 체크
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during iteration: %w", err)
	}

	return tables, nil
}

// GetColumns는 특정 테이블의 컬럼 목록을 조회합니다.
func (a *PostgresAdapter) GetColumns(ctx context.Context, conn *sql.DB, tableName string) ([]string, error) {
	// ==========================================
	// PostgreSQL 전용 쿼리! 🔥
	// ==========================================

	// information_schema.columns는 표준 SQL 뷰
	// (대부분의 DB가 지원하지만 세부사항은 다름)
	//
	// $1은 파라미터 placeholder
	// → SQL Injection 방지!
	// → tableName이 직접 문자열로 들어가지 않음
	query := `
		SELECT column_name 
		FROM information_schema.columns 
		WHERE table_schema = 'public' 
		  AND table_name = $1
		ORDER BY ordinal_position
	`

	// QueryContext의 세 번째 파라미터부터는 쿼리 파라미터
	// $1 = tableName
	// $2가 있다면 네 번째 파라미터, ...
	//
	// 이렇게 하면:
	// - SQL Injection 안전
	// - 자동으로 escape 처리
	rows, err := conn.QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns: %w", err)
	}
	defer rows.Close()

	var columns []string

	for rows.Next() {
		var columnName string

		if err := rows.Scan(&columnName); err != nil {
			return nil, fmt.Errorf("failed to scan column name: %w", err)
		}

		columns = append(columns, columnName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during iteration: %w", err)
	}

	return columns, nil
}
