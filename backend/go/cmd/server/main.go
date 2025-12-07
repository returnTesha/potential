package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"space/internal/config"

	"space/internal/adapters/input/http"
	"space/internal/adapters/output"
	"space/internal/core/service"
)

func main() {
	// ==========================================
	// 1단계: 커맨드 라인 플래그 파싱 (새로 추가!)
	// ==========================================

	env := getEnv("ENV", "dev")
	defaultConfigPath := fmt.Sprintf("config/config.%s.toml", env)

	configPath := flag.String(
		"config",
		defaultConfigPath,
		"Path to configuration file",
	)
	flag.Parse()

	log.Printf("Environment: %s", env)
	log.Printf("Config file: %s", *configPath)

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// ==========================================
	// 2단계: 로거 설정
	// ==========================================

	log.SetPrefix(cfg.Logging.Prefix)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("Starting DMS (Data Management System)...")
	log.Printf("Server port: %s", cfg.Server.Port)
	log.Printf("Allowed origins: %v", cfg.Server.AllowedOrigins) // ← 추가!

	// ==========================================
	// 4단계: Context 생성
	// ==========================================

	ctx := context.Background()

	// ==========================================
	// 5단계: 의존성 주입 (DI) - 변경 없음
	// ==========================================

	log.Println("Creating Connection Manager...")
	connManager := output.NewConnectionManager()

	log.Println("Creating Database Service...")
	dbService := service.NewDatabaseService(connManager)

	log.Println("Creating HTTP Handler...")
	handler := http.NewHandler(dbService)

	// ==========================================
	// 6단계: 라우터 설정 - 변경 없음
	// ==========================================

	log.Println("Setting up routes...")
	router := http.SetupRouter(handler, cfg.Server.AllowedOrigins)

	// ==========================================
	// 7단계: 초기 DB 연결 (TOML 기반으로 완전 변경!)
	// ==========================================

	log.Println("Registering initial databases...")

	// ❌ 삭제: 하드코딩된 initialDBs
	// initialDBs := []*domain.Database{...}

	// ✅ 추가: TOML에서 읽은 설정 사용
	for _, dbCfg := range cfg.Databases {
		// ConnectOnStartup이 false면 스킵
		if !dbCfg.ConnectOnStartup {
			log.Printf("Skipping %s (connect_on_startup=false)", dbCfg.ID)
			continue
		}

		// DatabaseConfig → domain.Database 변환
		db, err := dbCfg.ToDomain()
		if err != nil {
			log.Printf("Invalid database config %s: %v", dbCfg.ID, err)
			continue
		}

		// TOML에서 읽은 타임아웃 사용
		connectCtx, cancel := context.WithTimeout(ctx, dbCfg.GetConnectionTimeout())

		if err := dbService.RegisterDatabase(connectCtx, db); err != nil {
			log.Printf("Failed to connect to %s: %v", db.ID, err)
		} else {
			log.Printf("Successfully connected to %s", db.ID)
		}

		cancel()
	}

	// ==========================================
	// 8단계: 서버 포트 설정 (TOML 기반으로 변경!)
	// ==========================================

	// ❌ 삭제: 환경 변수에서 읽기
	// port := os.Getenv("PORT")
	// if port == "" {
	//     port = "8080"
	// }

	// ✅ 추가: TOML에서 읽기
	addr := fmt.Sprintf(":%s", cfg.Server.Port)

	log.Printf("Server will listen on %s", addr)

	// ==========================================
	// 9단계: Graceful Shutdown - 변경 없음
	// ==========================================

	go func() {
		log.Printf("Starting HTTP server on %s", addr)
		if err := router.Run(addr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// ==========================================
	// 10단계: 종료 시그널 대기 - 변경 없음
	// ==========================================

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// ==========================================
	// 11단계: DB 연결 종료 (TOML 타임아웃 사용!)
	// ==========================================

	// ❌ 삭제: 하드코딩된 5초
	// shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)

	// ✅ 추가: TOML에서 읽은 타임아웃
	shutdownCtx, cancel := context.WithTimeout(ctx, cfg.Server.GetShutdownTimeout())
	defer cancel()

	databases, err := dbService.ListDatabases(shutdownCtx)
	if err != nil {
		log.Printf("Failed to list databases: %v", err)
	} else {
		for _, db := range databases {
			if err := dbService.DisconnectDatabase(shutdownCtx, db.ID); err != nil {
				log.Printf("Failed to disconnect %s: %v", db.ID, err)
			} else {
				log.Printf("Disconnected from %s", db.ID)
			}
		}
	}

	log.Println("Server stopped gracefully")
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
