package domain

import (
	"errors"
	"fmt"
	"time" // time은 표준 라이브러리입니다
)

// QueryResult는 쿼리 실행 결과를 담는 구조체입니다.
type QueryResult struct {
	// []string은 문자열 슬라이스(slice)입니다.
	// 슬라이스는 동적 배열로, Java의 ArrayList와 비슷합니다.
	Columns []string // 컬럼 이름들 (예: ["id", "name", "email"])

	// []map[string]interface{}는 복잡해 보이지만,
	// "각 row는 map이고, 여러 row를 슬라이스로 담는다"는 의미입니다.
	// interface{}는 Java의 Object와 비슷합니다 (모든 타입 가능)
	Rows []map[string]interface{} // 실제 데이터

	// int64는 64비트 정수입니다 (큰 숫자 지원)
	RowsAffected int64 // 영향받은 row 수 (INSERT/UPDATE/DELETE용)

	// time.Duration은 시간 간격을 나타냅니다
	ExecutionTime time.Duration // 쿼리 실행 시간
}

// IsEmpty는 결과가 비어있는지 확인합니다.
func (qr *QueryResult) IsEmpty() bool {
	// len()으로 슬라이스 길이를 확인합니다
	// 길이가 0이면 비어있는 것
	return len(qr.Rows) == 0
}

// RowCount는 결과의 row 개수를 반환합니다.
func (qr *QueryResult) RowCount() int {
	// int와 int64는 다른 타입입니다
	// len()은 int를 반환하므로 그대로 반환
	return len(qr.Rows)
}

// FirstRow는 첫 번째 row를 반환합니다.
// Go에서는 여러 값을 반환할 수 있습니다! (Java와 다른 점)
// (반환값1, 반환값2) 형태로 반환합니다
func (qr *QueryResult) FirstRow() (map[string]interface{}, error) {
	// 결과가 비어있으면 에러 반환
	if qr.IsEmpty() {
		// nil은 "값이 없음"을 의미합니다
		// 첫 번째 반환값은 nil, 두 번째는 error
		return nil, errors.New("no rows found")
	}

	// 슬라이스 인덱싱: [0]은 첫 번째 요소
	// 에러가 없으면 nil 반환
	return qr.Rows[0], nil
}

// GetColumn은 특정 컬럼의 모든 값을 추출합니다.
func (qr *QueryResult) GetColumn(columnName string) ([]interface{}, error) {
	// 결과가 비어있으면 에러
	if qr.IsEmpty() {
		return nil, errors.New("no rows found")
	}

	// make()는 슬라이스를 생성하는 내장 함수입니다
	// make([]타입, 길이, 용량) 형태입니다
	// 용량을 지정하면 메모리를 미리 할당해서 효율적입니다
	values := make([]interface{}, 0, len(qr.Rows))

	// for-range는 슬라이스를 순회하는 Go의 방식입니다
	// for 인덱스, 값 := range 슬라이스 { ... }
	// _는 "사용하지 않는 변수"를 의미합니다 (인덱스를 사용하지 않음)
	for _, row := range qr.Rows {
		// map에서 값을 가져올 때: map[키]
		// ok는 키가 존재하는지 여부를 나타냅니다 (true/false)
		// 이것을 "comma ok idiom"이라고 합니다
		value, ok := row[columnName]
		if !ok {
			// !는 NOT 연산자입니다
			return nil, fmt.Errorf("column %s not found", columnName)
		}

		// append()는 슬라이스에 요소를 추가하는 내장 함수입니다
		// Java의 list.add()와 비슷합니다
		values = append(values, value)
	}

	return values, nil
}

// FormatExecutionTime은 실행 시간을 사람이 읽기 쉬운 형태로 반환합니다.
func (qr *QueryResult) FormatExecutionTime() string {
	// time.Duration은 자동으로 적절한 단위로 변환됩니다
	// 예: 1500000000 nanoseconds -> "1.5s"
	return qr.ExecutionTime.String()
}

// Summary는 쿼리 결과 요약을 반환합니다.
func (qr *QueryResult) Summary() string {
	// 조건부 표현식은 if-else로 작성합니다
	// Go에는 삼항 연산자(? :)가 없습니다
	var resultType string
	if qr.RowsAffected > 0 {
		resultType = "affected" // INSERT/UPDATE/DELETE
	} else {
		resultType = "selected" // SELECT
	}

	return fmt.Sprintf(
		"Query %s: %d rows, executed in %s",
		resultType,
		qr.RowCount(),
		qr.FormatExecutionTime(),
	)
}
