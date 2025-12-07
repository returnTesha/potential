package http

import (
	"github.com/gin-gonic/gin"
)

func corsMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 허용된 origin 체크
		for _, allowed := range allowedOrigins {
			if origin == allowed {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
				break
			}
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// SetupRouter는 HTTP 라우터를 설정합니다.
// 모든 엔드포인트(경로)를 정의하고 핸들러를 연결합니다.
//
// 파라미터:
// - handler: HTTP 핸들러 (위에서 만든 것)
//
// 반환값:
// - *gin.Engine: Gin 라우터 엔진
func SetupRouter(handler *Handler, allowedOrigins []string) *gin.Engine {
	// gin.Default()는 기본 미들웨어가 포함된 라우터를 생성합니다.
	//
	// 포함된 미들웨어:
	// - Logger: 요청 로그 출력 (예: GET /databases 200 15ms)
	// - Recovery: panic 발생 시 자동 복구 (500 에러 반환)
	//
	// gin.New()를 쓰면 미들웨어 없는 빈 라우터
	router := gin.Default()
	router.Use(corsMiddleware(allowedOrigins))

	// ==========================================
	// Health Check
	// ==========================================

	// GET /health
	// router.GET(경로, 핸들러함수)
	router.GET("/health", handler.HealthCheck)

	// ==========================================
	// Database Management
	// ==========================================

	v1 := router.Group("/api/dms/v1")
	{
		databases := v1.Group("/databases")
		{
			databases.GET("", handler.ListDatabases)
			databases.POST("", handler.RegisterDatabase)
			databases.GET("/:dbID", handler.GetDatabaseInfo)
			databases.DELETE("/:dbID", handler.DisconnectDatabase)
			databases.POST("/:dbID/query", handler.ExecuteQuery)
		}
	}
	// 등으로 변경됨

	return router
}

// 라우팅 예시:
//
// POST /databases
// → handler.RegisterDatabase()
//
// GET /databases
// → handler.ListDatabases()
//
// GET /databases/postgres-prod
// → handler.GetDatabaseInfo()
//    dbID = "postgres-prod"
//
// DELETE /databases/postgres-prod
// → handler.DisconnectDatabase()
//    dbID = "postgres-prod"
//
// POST /databases/postgres-prod/query
// → handler.ExecuteQuery()
//    dbID = "postgres-prod"
