// internal/adapters/output/oracle/adapter.go

package oracle19c

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/sijms/go-ora/v2"

	"space/internal/domain"
)

type OracleAdapter struct{}

func NewAdapter() *OracleAdapter {
	return &OracleAdapter{}
}

func (a *OracleAdapter) Connect(ctx context.Context, db *domain.Database) (*sql.DB, error) {
	// ==========================================
	// go-ora DSN 형식
	// ==========================================

	// Schema에 SID가 있으면 사용, 없으면 Name 사용
	sid := db.Name
	if db.Schema != "" {
		sid = db.Schema
	}

	// go-ora DSN: oracle://user:password@host:port/service_name
	dsn := fmt.Sprintf("oracle://%s:%s@%s:%d/%s",
		db.Username,
		db.Password,
		db.Host,
		db.Port,
		sid,
	)

	fmt.Printf("[Oracle] Connecting with go-ora: %s\n", dsn)

	// ==========================================
	// 드라이버명: "oracle" (go-ora)
	// ==========================================

	conn, err := sql.Open("oracle", dsn) // "godror" → "oracle"
	if err != nil {
		return nil, fmt.Errorf("sql.Open failed: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := conn.PingContext(ctx); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ping failed: %w", err)
	}

	fmt.Printf("[Oracle] Successfully connected to %s\n", db.ID)

	return conn, nil
}

func (a *OracleAdapter) ExecuteQuery(ctx context.Context, conn *sql.DB, query string) (*domain.QueryResult, error) {
	start := time.Now()

	rows, err := conn.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	results := []map[string]interface{}{}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))

		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			row[col] = values[i]
		}

		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	executionTime := time.Since(start)

	return &domain.QueryResult{
		Columns:       columns,
		Rows:          results,
		RowsAffected:  int64(len(results)),
		ExecutionTime: executionTime,
	}, nil
}

func (a *OracleAdapter) GetTables(ctx context.Context, conn *sql.DB) ([]string, error) {
	query := `
		SELECT table_name 
		FROM user_tables 
		ORDER BY table_name
	`

	rows, err := conn.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []string

	for rows.Next() {
		var tableName string

		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("failed to scan table name: %w", err)
		}

		tables = append(tables, tableName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during iteration: %w", err)
	}

	return tables, nil
}

func (a *OracleAdapter) GetColumns(ctx context.Context, conn *sql.DB, tableName string) ([]string, error) {
	query := `
		SELECT column_name 
		FROM user_tab_columns 
		WHERE table_name = :1
		ORDER BY column_id
	`

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
