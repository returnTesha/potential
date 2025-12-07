package dto

import (
	"space/internal/domain"
)

// DatabaseResponse는 데이터베이스 정보를 반환하는 응답 구조체입니다.
type DatabaseResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Schema   string `json:"schema,omitempty"` // 비어있으면 JSON에서 제외
	Username string `json:"username"`
	Status   string `json:"status"`

	// 비밀번호는 응답에 포함하지 않습니다! (보안)
}

// QueryResultResponse는 쿼리 실행 결과를 반환하는 응답 구조체입니다.
type QueryResultResponse struct {
	Columns       []string                 `json:"columns"`
	Rows          []map[string]interface{} `json:"rows"`
	RowCount      int                      `json:"row_count"`
	ExecutionTime string                   `json:"execution_time"` // "15ms" 형태
}

// ErrorResponse는 에러를 반환하는 응답 구조체입니다.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// SuccessResponse는 성공 메시지를 반환하는 응답 구조체입니다.
type SuccessResponse struct {
	Message string `json:"message"`
}

// FromDomain은 domain.Database를 DatabaseResponse로 변환합니다.
//
// 함수 시그니처 설명:
// - FromDomain: 함수 이름
// - (db *domain.Database): 파라미터 (Domain 객체)
// - *DatabaseResponse: 반환 타입 (DTO)
//
// 이것은 "변환 함수" 또는 "매퍼(Mapper)"라고 부릅니다.
func FromDomain(db *domain.Database) *DatabaseResponse {
	// &DatabaseResponse{...}는 구조체 리터럴 + 포인터 생성
	return &DatabaseResponse{
		ID:       db.ID,
		Name:     db.Name,
		Type:     string(db.Type), // DatabaseType → string 변환
		Host:     db.Host,
		Port:     db.Port,
		Schema:   db.Schema,
		Username: db.Username,
		Status:   string(db.Status), // ConnectionStatus → string 변환
		// Password는 의도적으로 제외! (보안)
	}
}

// FromDomainQueryResult는 domain.QueryResult를 QueryResultResponse로 변환합니다.
func FromDomainQueryResult(result *domain.QueryResult) *QueryResultResponse {
	return &QueryResultResponse{
		Columns:       result.Columns,
		Rows:          result.Rows,
		RowCount:      result.RowCount(),
		ExecutionTime: result.FormatExecutionTime(),
	}
}

// FromDomainList는 domain.Database 슬라이스를 DatabaseResponse 슬라이스로 변환합니다.
//
// []*domain.Database는 포인터 슬라이스를 의미합니다.
// "슬라이스의 각 요소가 포인터"
func FromDomainList(databases []*domain.Database) []*DatabaseResponse {
	// make()로 슬라이스 생성
	// len(databases): 초기 길이 (0으로 시작)
	// len(databases): 용량 (메모리 미리 할당)
	responses := make([]*DatabaseResponse, 0, len(databases))

	// for-range로 순회
	for _, db := range databases {
		// FromDomain으로 변환 후 append
		responses = append(responses, FromDomain(db))
	}

	return responses
}

// 예시 JSON 응답:
// GET /databases/postgres-prod
// {
//   "id": "postgres-prod",
//   "name": "production",
//   "type": "postgres16.3",
//   "host": "10.0.0.1",
//   "port": 5432,
//   "username": "admin",
//   "status": "connected"
// }
//
// POST /databases/postgres-prod/query
// {
//   "columns": ["id", "name", "email"],
//   "rows": [
//     {"id": 1, "name": "Alice", "email": "alice@example.com"},
//     {"id": 2, "name": "Bob", "email": "bob@example.com"}
//   ],
//   "row_count": 2,
//   "execution_time": "15ms"
// }
