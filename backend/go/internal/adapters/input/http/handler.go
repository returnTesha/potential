// Package http는 HTTP API를 제공하는 Input Adapter입니다.
// 이 패키지는:
// 1. HTTP 요청을 받습니다
// 2. JSON을 파싱합니다
// 3. Domain 객체로 변환합니다
// 4. Core Service를 호출합니다
// 5. HTTP 응답을 반환합니다
package http

import (
	"net/http" // HTTP 상태 코드 (200, 404 등)

	"github.com/gin-gonic/gin" // Gin 웹 프레임워크

	"space/internal/adapters/input/http/dto"
	"space/internal/domain"
	"space/internal/ports/input"
)

// Handler는 HTTP 요청을 처리하는 구조체입니다.
// 이것은 Gin의 핸들러 함수들을 제공합니다.
type Handler struct {
	// service는 Input Port 인터페이스입니다.
	// 실제 구현체(Core)를 모릅니다!
	// 그냥 "이 인터페이스를 만족하는 뭔가"만 알면 됩니다.
	service input.DatabaseService
}

// NewHandler는 Handler를 생성합니다.
//
// 의존성 주입(DI):
// - service를 외부에서 받아옴
// - Handler는 service의 구체 타입을 모름
// - 테스트할 때 Mock을 주입할 수 있음!
func NewHandler(service input.DatabaseService) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterDatabase는 새로운 데이터베이스를 등록합니다.
// HTTP: POST /databases
//
// Gin의 핸들러 함수 시그니처:
// - func(c *gin.Context)
// - c는 요청/응답을 다루는 컨텍스트
func (h *Handler) RegisterDatabase(c *gin.Context) {
	// ==========================================
	// 1단계: JSON 파싱
	// ==========================================

	// var로 변수 선언
	// RegisterDatabaseRequest 타입의 빈 구조체
	var req dto.RegisterDatabaseRequest

	// ShouldBindJSON은 요청 Body의 JSON을 req에 바인딩합니다.
	//
	// 자동으로:
	// 1. JSON 파싱
	// 2. 구조체 필드에 매핑 (json 태그 사용)
	// 3. 유효성 검사 (binding 태그 사용)
	//
	// 반환값:
	// - nil: 성공
	// - error: 실패 (JSON 형식 오류, 필수 필드 누락 등)
	if err := c.ShouldBindJSON(&req); err != nil {
		// JSON 파싱 실패
		// 400 Bad Request 반환
		//
		// c.JSON(상태코드, 데이터)
		// - 상태코드: http.StatusBadRequest = 400
		// - 데이터: gin.H는 map[string]interface{}의 단축형
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request",
			"details": err.Error(), // 구체적인 에러 메시지
		})
		return // 함수 종료 (더 이상 처리하지 않음)
	}

	// ==========================================
	// 2단계: DTO → Domain 변환
	// ==========================================

	// &domain.Database{...}는 구조체 리터럴로 객체 생성
	// req의 데이터를 domain.Database로 변환
	db := &domain.Database{
		ID:       req.ID,
		Name:     req.Name,
		Type:     domain.DatabaseType(req.Type), // string → DatabaseType 변환
		Host:     req.Host,
		Port:     req.Port,
		Schema:   req.Schema,
		Username: req.Username,
		Password: req.Password,
		Status:   domain.Disconnected, // 초기 상태
	}

	// ==========================================
	// 3단계: Service 호출 (Core)
	// ==========================================

	// c.Request.Context()는 HTTP 요청의 context를 가져옵니다.
	// 이 context는:
	// - HTTP 연결이 끊기면 자동으로 취소됨
	// - 타임아웃 설정이 있으면 자동으로 적용됨
	ctx := c.Request.Context()

	// service.RegisterDatabase() 호출
	// 이것은 Input Port 인터페이스 메서드!
	// 실제로는 Core의 구현체가 실행됨
	if err := h.service.RegisterDatabase(ctx, db); err != nil {
		// ==========================================
		// 에러 처리
		// ==========================================

		// 에러 타입별로 다른 HTTP 상태 코드 반환
		// errors.Is()로 에러 체크
		//
		// Go 1.13+ 에러 처리:
		// - errors.Is(err, target): err가 target인지 확인
		// - Wrapped 에러도 확인 가능

		// 에러 응답 생성
		errorResp := dto.ErrorResponse{
			Error:   "failed to register database",
			Message: err.Error(),
		}

		// 상태 코드 결정
		statusCode := http.StatusInternalServerError // 기본 500

		// Domain 에러 체크
		switch err {
		case domain.ErrAlreadyConnected:
			statusCode = http.StatusConflict // 409
			errorResp.Error = "database already exists"

		case domain.ErrInvalidDatabaseType:
			statusCode = http.StatusBadRequest // 400
			errorResp.Error = "invalid database type"

		case domain.ErrMissingCredentials:
			statusCode = http.StatusBadRequest // 400
			errorResp.Error = "missing credentials"
		}

		// 에러 응답 반환
		c.JSON(statusCode, errorResp)
		return
	}

	// ==========================================
	// 4단계: 성공 응답
	// ==========================================

	// Domain → DTO 변환
	response := dto.FromDomain(db)

	// 201 Created 반환
	// http.StatusCreated = 201
	c.JSON(http.StatusCreated, response)
}

// ExecuteQuery는 특정 데이터베이스에 쿼리를 실행합니다.
// HTTP: POST /databases/:dbID/query
//
// :dbID는 URL 파라미터 (path parameter)
// 예: POST /databases/postgres-prod/query
func (h *Handler) ExecuteQuery(c *gin.Context) {
	// ==========================================
	// 1단계: URL 파라미터 추출
	// ==========================================

	// c.Param("dbID")는 URL에서 :dbID 값을 가져옵니다.
	// 예: /databases/postgres-prod/query
	//     → dbID = "postgres-prod"
	dbID := c.Param("dbID")

	// ==========================================
	// 2단계: JSON 파싱
	// ==========================================

	var req dto.ExecuteQueryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request",
			"details": err.Error(),
		})
		return
	}

	// ==========================================
	// 3단계: Service 호출
	// ==========================================

	ctx := c.Request.Context()

	// service.ExecuteQuery() 호출
	result, err := h.service.ExecuteQuery(ctx, dbID, req.Query)
	if err != nil {
		// 에러 처리
		errorResp := dto.ErrorResponse{
			Error:   "query execution failed",
			Message: err.Error(),
		}

		statusCode := http.StatusInternalServerError

		switch err {
		case domain.ErrDatabaseNotFound:
			statusCode = http.StatusNotFound // 404
			errorResp.Error = "database not found"

		case domain.ErrDatabaseNotConnected:
			statusCode = http.StatusServiceUnavailable // 503
			errorResp.Error = "database not connected"

		case domain.ErrQueryTimeout:
			statusCode = http.StatusRequestTimeout // 408
			errorResp.Error = "query timeout"
		}

		c.JSON(statusCode, errorResp)
		return
	}

	// ==========================================
	// 4단계: 성공 응답
	// ==========================================

	// Domain → DTO 변환
	response := dto.FromDomainQueryResult(result)

	// 200 OK 반환
	c.JSON(http.StatusOK, response)
}

// ListDatabases는 모든 데이터베이스 목록을 반환합니다.
// HTTP: GET /databases
func (h *Handler) ListDatabases(c *gin.Context) {
	ctx := c.Request.Context()

	// Service 호출
	databases, err := h.service.ListDatabases(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "failed to list databases",
			Message: err.Error(),
		})
		return
	}

	// Domain → DTO 변환 (슬라이스 전체)
	response := dto.FromDomainList(databases)

	// 200 OK 반환
	c.JSON(http.StatusOK, gin.H{
		"databases": response,
		"count":     len(response),
	})
}

// GetDatabaseInfo는 특정 데이터베이스 정보를 조회합니다.
// HTTP: GET /databases/:dbID
func (h *Handler) GetDatabaseInfo(c *gin.Context) {
	dbID := c.Param("dbID")
	ctx := c.Request.Context()

	// Service 호출
	db, err := h.service.GetDatabaseInfo(ctx, dbID)
	if err != nil {
		statusCode := http.StatusInternalServerError

		if err == domain.ErrDatabaseNotFound {
			statusCode = http.StatusNotFound // 404
		}

		c.JSON(statusCode, dto.ErrorResponse{
			Error:   "failed to get database info",
			Message: err.Error(),
		})
		return
	}

	// Domain → DTO 변환
	response := dto.FromDomain(db)

	c.JSON(http.StatusOK, response)
}

// DisconnectDatabase는 데이터베이스 연결을 종료합니다.
// HTTP: DELETE /databases/:dbID
func (h *Handler) DisconnectDatabase(c *gin.Context) {
	dbID := c.Param("dbID")
	ctx := c.Request.Context()

	// Service 호출
	if err := h.service.DisconnectDatabase(ctx, dbID); err != nil {
		statusCode := http.StatusInternalServerError

		if err == domain.ErrDatabaseNotFound {
			statusCode = http.StatusNotFound // 404
		}

		c.JSON(statusCode, dto.ErrorResponse{
			Error:   "failed to disconnect database",
			Message: err.Error(),
		})
		return
	}

	// 204 No Content 반환 (성공, 응답 본문 없음)
	c.Status(http.StatusNoContent)
}

// HealthCheck는 서버 상태를 확인합니다.
// HTTP: GET /health
//
// 헬스체크는 Load Balancer나 Kubernetes가 서버 상태를 확인할 때 사용
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "DMS",
	})
}
