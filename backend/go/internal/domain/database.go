// Package domain은 DMS의 핵심 비즈니스 모델을 정의합니다.
// 이 패키지는 외부 의존성이 전혀 없으며 (표준 라이브러리만 사용),
// 순수한 비즈니스 로직과 규칙만을 포함합니다.
package domain

import (
	"errors"
	"fmt"
	"log"
)

// DatabaseType은 지원하는 데이터베이스 종류를 나타내는 타입입니다.
// Go에서는 string을 기반으로 새로운 타입을 만들 수 있습니다.
// 이렇게 하면 타입 안전성이 높아집니다 (실수로 잘못된 문자열 전달 방지)
type DatabaseType string

// const 블록으로 상수들을 그룹화합니다.
// 이것은 Java의 enum과 비슷한 역할을 합니다.
const (
	PostgreSQL DatabaseType = "postgres16.3" // PostgreSQL 16.3
	Oracle11g  DatabaseType = "oracle11g"    // Oracle 11g
	Oracle19c  DatabaseType = "oracle19c"    // Oracle 19c
	MariaDB    DatabaseType = "mariadb10.11" // MariaDB 10.11
)

// ConnectionStatus는 데이터베이스 연결 상태를 나타냅니다.
type ConnectionStatus string

const (
	Connected    ConnectionStatus = "connected"    // 연결됨
	Disconnected ConnectionStatus = "disconnected" // 연결 끊김
	Connecting   ConnectionStatus = "connecting"   // 연결 중
	Error        ConnectionStatus = "error"        // 에러 상태
)

// Database는 데이터베이스 연결 정보를 담는 핵심 Entity입니다.
// Go에서 struct는 Java의 class와 비슷하지만, 상속이 없고 더 단순합니다.
// struct 필드는 대문자로 시작하면 public(exported), 소문자면 private(unexported)입니다.
type Database struct {
	ID       string           // 데이터베이스 고유 ID (예: "postgres-prod")
	Name     string           // 데이터베이스 이름
	Type     DatabaseType     // 데이터베이스 타입 (위에서 정의한 custom type)
	Host     string           // 호스트 주소 (예: "10.0.0.1")
	Port     int              // 포트 번호 (예: 5432)
	Schema   string           // 스키마 이름 (Oracle용, 선택사항)
	Username string           // 사용자명
	Password string           // 비밀번호
	Status   ConnectionStatus // 현재 연결 상태
}

// Validate는 Database 객체의 유효성을 검증합니다.
// Go에서 메서드는 func와 함수명 사이에 receiver를 적습니다.
// (db *Database)가 receiver이며, 이는 Java의 'this'와 비슷합니다.
// *Database는 포인터 receiver로, 원본 객체를 참조합니다.
func (db *Database) Validate() error {
	// Go의 에러 처리는 명시적입니다. 예외(exception)가 아닌 error 값을 반환합니다.

	// ID 검증
	if db.ID == "" {
		// errors.New()는 새로운 error를 생성합니다
		return errors.New("database ID is required")
	}

	// Name 검증
	if db.Name == "" {
		return errors.New("database name is required")
	}

	// Host 검증
	if db.Host == "" {
		return errors.New("host is required")
	}

	// Port 범위 검증 (비즈니스 규칙!)
	if db.Port < 1 || db.Port > 65535 {
		// fmt.Errorf는 형식화된 에러 메시지를 만듭니다 (printf 스타일)
		return fmt.Errorf("invalid port number: %d (must be 1-65535)", db.Port)
	}

	// Username 검증
	if db.Username == "" {
		return errors.New("username is required")
	}

	// Password 검증
	if db.Password == "" {
		return errors.New("password is required")
	}

	// DatabaseType 유효성 검증
	// 메서드를 호출할 때는 db.Type.IsValid() 이렇게 체이닝합니다
	if !db.Type.IsValid() {
		return fmt.Errorf("unsupported database type: %s", db.Type)
	}

	// Go에서 에러가 없으면 nil을 반환합니다
	// nil은 Java의 null과 비슷합니다
	return nil
}

// IsValid는 DatabaseType이 지원되는 타입인지 확인합니다.
// receiver가 (dt DatabaseType)로 값 타입입니다 (포인터 아님)
// 값을 변경할 필요가 없고, 크기가 작으면 값 receiver를 사용합니다
func (dt DatabaseType) IsValid() bool {
	// switch 문으로 여러 case를 한 번에 체크할 수 있습니다
	switch dt {
	case PostgreSQL, Oracle11g, Oracle19c, MariaDB:
		return true // 지원하는 타입
	default:
		return false // 미지원 타입
	}
}

// DefaultPort는 각 DB 타입의 기본 포트를 반환합니다.
// 이것은 도메인 지식(Domain Knowledge)입니다!
func (dt DatabaseType) DefaultPort() int {
	switch dt {
	case PostgreSQL:
		return 5432
	case Oracle11g, Oracle19c:
		return 1521
	case MariaDB:
		return 3306
	default:
		return 0 // 알 수 없는 타입
	}
}

// String은 DatabaseType을 사람이 읽기 쉬운 문자열로 변환합니다.
// Go의 Stringer 인터페이스를 구현하는 특별한 메서드입니다.
// fmt.Println()이나 fmt.Sprintf()에서 자동으로 호출됩니다.
func (dt DatabaseType) String() string {
	return string(dt) // DatabaseType을 string으로 변환
}

// ConnectionString은 데이터베이스 연결 문자열을 생성합니다.
// 각 DB마다 연결 문자열 형식이 다릅니다 (도메인 지식!)
func (db *Database) ConnectionString() string {
	// switch로 DB 타입별로 다른 형식 생성
	switch db.Type {
	case PostgreSQL:
		// PostgreSQL 연결 문자열 형식
		// fmt.Sprintf는 형식화된 문자열을 만듭니다 (Java의 String.format과 비슷)
		// %s는 문자열, %d는 정수를 나타냅니다
		return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			db.Username, db.Password, db.Host, db.Port, db.Name)

	case Oracle11g, Oracle19c:
		// Schema를 SID로 사용
		sid := db.Name // 기본값
		if db.Schema != "" {
			sid = db.Schema // Schema 있으면 우선
		}

		log.Printf(fmt.Sprintf("%s/%s@%s:%d/%s",
			db.Username, db.Password, db.Host, db.Port, sid))
		return fmt.Sprintf("%s/%s@%s:%d/%s",
			db.Username, db.Password, db.Host, db.Port, sid)

	case MariaDB:
		// MariaDB/MySQL 연결 문자열 형식
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
			db.Username, db.Password, db.Host, db.Port, db.Name)

	default:
		// 알 수 없는 타입은 빈 문자열
		return ""
	}
}

// IsConnected는 현재 데이터베이스가 연결되어 있는지 확인합니다.
// bool을 반환하는 간단한 헬퍼 메서드입니다.
func (db *Database) IsConnected() bool {
	// ==로 비교, Go는 자동으로 boolean을 반환합니다
	return db.Status == Connected
}

// CanConnect는 연결에 필요한 모든 정보가 있는지 확인합니다.
// 비즈니스 규칙: 연결하려면 최소한 Host, Port, Username, Password가 필요
func (db *Database) CanConnect() bool {
	// &&는 논리 AND 연산자입니다
	// Go는 short-circuit evaluation을 합니다 (앞이 false면 뒤를 평가 안 함)
	return db.Host != "" &&
		db.Port > 0 &&
		db.Username != "" &&
		db.Password != ""
}

// SafeString은 민감한 정보(비밀번호)를 숨긴 문자열을 반환합니다.
// 로깅이나 디버깅할 때 사용합니다.
func (db *Database) SafeString() string {
	// 여러 줄에 걸친 문자열을 만들 때는 이렇게 합니다
	return fmt.Sprintf(
		"Database{ID: %s, Name: %s, Type: %s, Host: %s, Port: %d, Username: %s, Status: %s}",
		db.ID,
		db.Name,
		db.Type,
		db.Host,
		db.Port,
		db.Username,
		// 비밀번호는 출력하지 않습니다! (보안)
		db.Status,
	)
}

// MaskedPassword는 마스킹된 비밀번호를 반환합니다.
// 비밀번호가 있으면 "****", 없으면 빈 문자열
func (db *Database) MaskedPassword() string {
	// len()은 문자열 길이를 반환하는 내장 함수입니다
	if len(db.Password) == 0 {
		return ""
	}
	return "****"
}

// Clone은 Database 객체의 복사본을 만듭니다.
// Go에서는 struct를 직접 복사하면 얕은 복사(shallow copy)가 됩니다.
// *를 사용해서 역참조(dereference)하면 값이 복사됩니다.
func (db *Database) Clone() *Database {
	// *db는 db 포인터가 가리키는 값을 의미합니다
	// 이렇게 하면 모든 필드가 복사된 새 struct가 만들어집니다
	copy := *db

	// &는 주소 연산자로, 값의 포인터를 반환합니다
	return &copy
}
