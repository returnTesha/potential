// Package dto는 Data Transfer Object를 정의합니다.
// DTO는 외부(HTTP)와 내부(Domain) 사이의 데이터 변환을 담당합니다.
//
// 왜 DTO가 필요한가?
// → HTTP JSON 구조와 Domain 구조가 다를 수 있음
// → 외부 API 변경이 Domain에 영향을 주지 않게
// → 유효성 검사 태그 등 HTTP 전용 기능 사용
package dto

// RegisterDatabaseRequest는 DB 등록 API의 요청 구조체입니다.
// JSON으로 받은 데이터를 이 구조체로 파싱합니다.
type RegisterDatabaseRequest struct {
	// ID는 데이터베이스 고유 식별자입니다.
	//
	// `json:"id"` 태그의 의미:
	// - JSON 파싱할 때 "id" 필드와 매핑
	// - Go 구조체 필드명(ID)과 JSON 필드명(id)을 연결
	//
	// `binding:"required"` 태그의 의미:
	// - gin 프레임워크의 유효성 검사
	// - required: 필수 항목 (없으면 400 에러)
	//
	// 태그 문법: `key:"value" key2:"value2"`
	ID string `json:"id" binding:"required"`

	// Name은 데이터베이스 이름입니다.
	Name string `json:"name" binding:"required"`

	// Type은 DB 종류입니다 (postgres16.3, oracle19c 등)
	Type string `json:"type" binding:"required"`

	// Host는 호스트 주소입니다.
	Host string `json:"host" binding:"required"`

	// Port는 포트 번호입니다.
	//
	// `binding:"required,min=1,max=65535"`의 의미:
	// - required: 필수
	// - min=1: 최소값 1
	// - max=65535: 최대값 65535
	Port int `json:"port" binding:"required,min=1,max=65535"`

	// Schema는 스키마 이름입니다 (Oracle용, 선택사항)
	// omitempty: JSON에 이 필드가 없어도 OK
	Schema string `json:"schema,omitempty"`

	// Username은 사용자명입니다.
	Username string `json:"username" binding:"required"`

	// Password는 비밀번호입니다.
	Password string `json:"password" binding:"required"`
}

// ExecuteQueryRequest는 쿼리 실행 API의 요청 구조체입니다.
type ExecuteQueryRequest struct {
	// Query는 실행할 SQL 쿼리입니다.
	Query string `json:"query" binding:"required"`
}

// 예시 JSON:
// POST /databases
// {
//   "id": "postgres-prod",
//   "name": "production",
//   "type": "postgres16.3",
//   "host": "10.0.0.1",
//   "port": 5432,
//   "username": "admin",
//   "password": "secret"
// }
//
// POST /databases/postgres-prod/query
// {
//   "query": "SELECT * FROM users LIMIT 10"
// }
