// Package service는 DMS의 핵심 비즈니스 로직(Use Cases)을 구현합니다.
// 이 패키지는:
// 1. Input Port(인터페이스)를 구현합니다
// 2. Output Port(인터페이스)를 사용합니다
// 3. Domain 모델을 사용합니다
// 4. 실제 구현(Adapter)은 모릅니다!
package service

import (
	"context"
	"fmt"

	// Domain import (안쪽)
	"space/internal/domain"

	// Ports import (인터페이스만)
	"space/internal/ports/input"
	"space/internal/ports/output"
	// ❌ adapters는 절대 import 하지 않습니다!
	// Core는 구체적인 구현(HTTP, Postgres 등)을 알면 안 됩니다.
)

// databaseService는 DatabaseService 인터페이스의 실제 구현체입니다.
//
// Go 관례:
// - 구현 struct는 소문자로 시작 (private/unexported)
// - 인터페이스는 대문자로 시작 (public/exported)
// - 외부에는 인터페이스만 노출하고, 구현체는 숨김
type databaseService struct {
	// repo는 Output Port 인터페이스입니다.
	// 실제로 Postgres인지 Oracle인지 MongoDB인지 모릅니다!
	// 그냥 "이 인터페이스를 만족하는 뭔가"만 알면 됩니다.
	repo output.DatabaseRepository
}

// NewDatabaseService는 databaseService의 생성자 함수입니다.
//
// Go 관례:
// - 생성자 함수는 New로 시작 (NewXxx 형태)
// - 인터페이스 타입을 반환 (구현체가 아님!)
//
// 파라미터:
//   - repo: output.DatabaseRepository - 의존성 주입(DI)
//
// 반환값:
//   - input.DatabaseService - 인터페이스 타입! (구현체 아님)
//
// 왜 인터페이스를 반환할까?
// → 사용하는 쪽(HTTP Handler)도 구현체를 알 필요가 없게 하기 위해!
func NewDatabaseService(repo output.DatabaseRepository) input.DatabaseService {
	// &databaseService{...}는 struct 포인터를 생성합니다.
	// { } 안에 필드 값을 초기화합니다.
	return &databaseService{
		repo: repo, // repo 필드에 파라미터 repo 할당
	}
}

// RegisterDatabase는 새로운 데이터베이스 연결을 등록합니다.
// 이 메서드는 input.DatabaseService 인터페이스를 구현합니다.
//
// Go의 인터페이스 구현:
// - "implements DatabaseService" 같은 선언이 없음!
// - 메서드 시그니처만 일치하면 자동으로 인터페이스 구현
// - (s *databaseService)가 메서드를 가지면 자동으로 input.DatabaseService 만족
func (s *databaseService) RegisterDatabase(ctx context.Context, db *domain.Database) error {
	// ==========================================
	// 1단계: Domain 검증 (비즈니스 규칙)
	// ==========================================

	// Validate()는 domain.Database의 메서드입니다.
	// ID, Name, Host, Port 등이 유효한지 검사합니다.
	if err := db.Validate(); err != nil {
		// err != nil은 "에러가 있다"는 의미입니다.
		// fmt.Errorf는 에러 메시지에 추가 정보를 붙입니다.
		// %w는 원본 에러를 포함(wrap)합니다 (Go 1.13+)
		// 이렇게 하면 errors.Is()나 errors.As()로 에러 체크 가능
		return fmt.Errorf("validation failed: %w", err)
	}

	// ==========================================
	// 2단계: 타입 검증
	// ==========================================

	// db.Type.IsValid()는 DatabaseType의 메서드입니다.
	// PostgreSQL, Oracle, MariaDB 등 지원하는 타입인지 확인
	if !db.Type.IsValid() {
		// !는 NOT 연산자 (false를 true로, true를 false로)
		// domain.ErrInvalidDatabaseType은 미리 정의된 에러 변수
		return domain.ErrInvalidDatabaseType
	}

	// ==========================================
	// 3단계: 연결 가능 여부 확인
	// ==========================================

	// CanConnect()는 최소 필수 정보가 있는지 확인
	// Host, Port, Username, Password 등
	if !db.CanConnect() {
		return domain.ErrMissingCredentials
	}

	// ==========================================
	// 4단계: 실제 DB 연결 시도 (Output Port 호출!)
	// ==========================================

	// 🔥 여기가 핵심!
	// s.repo.Connect()를 호출하지만,
	// 실제로 Postgres에 연결되는지, Oracle에 연결되는지 모릅니다!
	// s.repo는 인터페이스이므로, 런타임에 실제 구현체가 결정됩니다.
	//
	// 의존성 주입(DI) 덕분에:
	// - 테스트할 때: MockRepository.Connect() 실행
	// - 실제 운영: PostgresRepository.Connect() 실행
	if err := s.repo.Connect(ctx, db); err != nil {
		// 연결 실패 시 에러 반환
		// 원본 에러를 wrap해서 추가 컨텍스트 제공
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// ==========================================
	// 5단계: 성공!
	// ==========================================

	// 모든 단계를 통과하면 nil 반환 (에러 없음)
	// 이 시점에 db.Status는 Connected로 변경되어 있음
	// (Output Adapter가 변경함)
	return nil
}

// ExecuteQuery는 특정 데이터베이스에 쿼리를 실행합니다.
func (s *databaseService) ExecuteQuery(ctx context.Context, dbID string, query string) (*domain.QueryResult, error) {
	// ==========================================
	// 1단계: 입력값 검증
	// ==========================================

	// len()은 문자열 길이를 반환하는 내장 함수
	// 길이가 0이면 빈 문자열
	if len(dbID) == 0 {
		// errors.New()로 새 에러 생성
		return nil, fmt.Errorf("dbID is required")
	}

	if len(query) == 0 {
		return nil, fmt.Errorf("query is required")
	}

	// ==========================================
	// 2단계: 연결 상태 확인
	// ==========================================

	// IsConnected()는 Output Port의 메서드
	// 해당 DB가 실제로 연결되어 있는지 확인
	if !s.repo.IsConnected(ctx, dbID) {
		// 연결되어 있지 않으면 미리 정의된 에러 반환
		return nil, domain.ErrDatabaseNotConnected
	}

	// ==========================================
	// 3단계: 쿼리 실행 (Output Port 호출!)
	// ==========================================

	// 🔥 실제 쿼리 실행
	// s.repo.ExecuteQuery()가 실제 DB에 쿼리를 보냅니다.
	// 하지만 Core는 어떻게 실행되는지 모릅니다!
	result, err := s.repo.ExecuteQuery(ctx, dbID, query)
	if err != nil {
		// 쿼리 실패 시
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	// ==========================================
	// 4단계: 결과 반환
	// ==========================================

	// domain.QueryResult를 그대로 반환
	// Core는 결과 변환이나 가공을 하지 않음
	// (필요하다면 여기서 비즈니스 로직 추가 가능)
	return result, nil
}

// ListDatabases는 현재 연결된 모든 데이터베이스 목록을 반환합니다.
func (s *databaseService) ListDatabases(ctx context.Context) ([]*domain.Database, error) {
	// Output Port의 ListConnections() 호출
	// 단순히 Repository에게 위임(delegate)
	databases, err := s.repo.ListConnections(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %w", err)
	}

	// 🔒 보안: 비밀번호 마스킹
	// 반환하기 전에 민감한 정보를 숨깁니다.
	//
	// for-range로 슬라이스 순회
	// range는 복사본을 반환하므로, 원본을 수정하려면 인덱스 사용
	for i := range databases {
		// databases[i]는 *domain.Database 포인터
		// 원본을 직접 수정
		databases[i].Password = databases[i].MaskedPassword()
	}

	return databases, nil
}

// DisconnectDatabase는 특정 데이터베이스 연결을 종료합니다.
func (s *databaseService) DisconnectDatabase(ctx context.Context, dbID string) error {
	// ==========================================
	// 1단계: 입력값 검증
	// ==========================================

	if len(dbID) == 0 {
		return fmt.Errorf("dbID is required")
	}

	// ==========================================
	// 2단계: 연결 존재 확인
	// ==========================================

	// 연결되어 있지 않으면 의미 없음
	if !s.repo.IsConnected(ctx, dbID) {
		return domain.ErrDatabaseNotFound
	}

	// ==========================================
	// 3단계: 연결 종료 (Output Port 호출!)
	// ==========================================

	if err := s.repo.Disconnect(ctx, dbID); err != nil {
		return fmt.Errorf("failed to disconnect: %w", err)
	}

	return nil
}

// GetDatabaseInfo는 특정 데이터베이스의 정보를 조회합니다.
func (s *databaseService) GetDatabaseInfo(ctx context.Context, dbID string) (*domain.Database, error) {
	// ==========================================
	// 1단계: 입력값 검증
	// ==========================================

	if len(dbID) == 0 {
		return nil, fmt.Errorf("dbID is required")
	}

	// ==========================================
	// 2단계: 전체 목록 조회
	// ==========================================

	// ListConnections()로 모든 DB 가져옴
	databases, err := s.repo.ListConnections(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get database info: %w", err)
	}

	// ==========================================
	// 3단계: 해당 DB 찾기
	// ==========================================

	// for-range로 슬라이스 순회
	// _, db := range databases에서:
	// - _는 인덱스 (사용 안 함)
	// - db는 현재 요소 (*domain.Database)
	for _, db := range databases {
		// 문자열 비교는 == 연산자
		if db.ID == dbID {
			// 찾았다!

			// 🔒 보안: 비밀번호 마스킹 후 반환
			db.Password = db.MaskedPassword()

			return db, nil
		}
	}

	// ==========================================
	// 4단계: 못 찾음
	// ==========================================

	// 루프를 다 돌았는데 못 찾으면 에러
	return nil, domain.ErrDatabaseNotFound
}

// 추가 헬퍼 메서드들 (선택사항)

// GetTables는 특정 데이터베이스의 테이블 목록을 조회합니다.
// 이 메서드는 input.DatabaseService 인터페이스에는 없지만,
// 추가 기능으로 제공할 수 있습니다.
func (s *databaseService) GetTables(ctx context.Context, dbID string) ([]string, error) {
	// 연결 확인
	if !s.repo.IsConnected(ctx, dbID) {
		return nil, domain.ErrDatabaseNotConnected
	}

	// Output Port 호출
	tables, err := s.repo.GetTables(ctx, dbID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}

	return tables, nil
}

// ValidateQuery는 쿼리의 기본적인 유효성을 검사합니다.
// (실제 구문 분석은 하지 않고, 위험한 키워드만 체크)
//
// private 메서드 (소문자 시작) - 외부에서 호출 불가
func (s *databaseService) validateQuery(query string) error {
	// 간단한 검증 예시
	// 실제로는 더 정교한 검증이 필요할 수 있음

	// strings 패키지를 import해야 함 (위에 추가 필요)
	// import "strings"

	// 빈 쿼리 체크
	if len(query) == 0 {
		return domain.ErrInvalidQuery
	}

	// 너무 긴 쿼리 체크 (예: 10000자 제한)
	if len(query) > 10000 {
		return fmt.Errorf("query too long (max 10000 characters)")
	}

	// 여기에 추가 검증 로직 가능:
	// - SQL Injection 방지
	// - 위험한 키워드 체크 (DROP, TRUNCATE 등)
	// - etc.

	return nil
}
