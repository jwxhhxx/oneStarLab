package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	_ "modernc.org/sqlite"

	"onestarlab/backend/internal/auth"
)

// Open 打开（必要时创建）SQLite 数据库，并初始化表结构与默认数据。
func Open(path string) (*sql.DB, error) {
	if path == "" {
		path = "data/onestar.db"
	}
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create database directory: %w", err)
		}
	}

	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// 单用户本地应用：单连接串行化所有读写，避免 SQLITE_BUSY 与写冲突。
	conn.SetMaxOpenConns(1)
	conn.SetMaxIdleConns(1)
	conn.SetConnMaxLifetime(0)

	pragmas := []string{
		"PRAGMA journal_mode = WAL",
		"PRAGMA synchronous = NORMAL",
		"PRAGMA busy_timeout = 5000",
		"PRAGMA foreign_keys = OFF",
	}
	for _, pragma := range pragmas {
		if _, err := conn.Exec(pragma); err != nil {
			conn.Close()
			return nil, fmt.Errorf("apply %s: %w", pragma, err)
		}
	}

	if err := migrate(conn); err != nil {
		conn.Close()
		return nil, err
	}
	if err := seed(conn); err != nil {
		conn.Close()
		return nil, err
	}

	return conn, nil
}

func migrate(conn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS products (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			sku TEXT NOT NULL UNIQUE,
			category TEXT NOT NULL DEFAULT '',
			supplier TEXT NOT NULL DEFAULT '',
			purchase_cost REAL NOT NULL DEFAULT 0,
			packaging_cost REAL NOT NULL DEFAULT 0,
			sale_price REAL NOT NULL DEFAULT 0,
			stock INTEGER NOT NULL DEFAULT 0,
			safe_stock INTEGER NOT NULL DEFAULT 0,
			note TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_products_category ON products(category)`,
		`CREATE INDEX IF NOT EXISTS idx_products_created ON products(created_at)`,
		`CREATE TABLE IF NOT EXISTS categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			created_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS orders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			order_no TEXT NOT NULL,
			channel TEXT NOT NULL DEFAULT '',
			customer_name TEXT NOT NULL DEFAULT '',
			items_json TEXT NOT NULL,
			total_amount REAL NOT NULL DEFAULT 0,
			discount_amount REAL NOT NULL DEFAULT 0,
			shipping_cost REAL NOT NULL DEFAULT 0,
			platform_fee REAL NOT NULL DEFAULT 0,
			goods_cost REAL NOT NULL DEFAULT 0,
			net_profit REAL NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'paid'
		)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_created ON orders(created_at)`,
		`CREATE TABLE IF NOT EXISTS pricing_rules (
			id INTEGER PRIMARY KEY,
			target_margin REAL NOT NULL,
			channel_fee_rate REAL NOT NULL,
			extra_cost REAL NOT NULL,
			rounding_mode TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS lab_inspirations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			tag TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL,
			summary TEXT NOT NULL DEFAULT '',
			keywords TEXT NOT NULL DEFAULT '',
			image TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_lab_inspirations_updated ON lab_inspirations(updated_at)`,
		`CREATE TABLE IF NOT EXISTS lab_projects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			category TEXT NOT NULL DEFAULT '',
			stage TEXT NOT NULL,
			progress INTEGER NOT NULL DEFAULT 0,
			launch_date TEXT NOT NULL DEFAULT '',
			note TEXT NOT NULL DEFAULT '',
			image TEXT NOT NULL DEFAULT '',
			sample_records_json TEXT NOT NULL DEFAULT '[]',
			updated_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_lab_projects_updated ON lab_projects(updated_at)`,
		`CREATE TABLE IF NOT EXISTS drawing_inspirations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			items_json TEXT NOT NULL DEFAULT '[]',
			styles_json TEXT NOT NULL DEFAULT '[]',
			created_at TEXT NOT NULL,
			completed INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE INDEX IF NOT EXISTS idx_drawing_created ON drawing_inspirations(created_at)`,
		`CREATE TABLE IF NOT EXISTS expenses (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			purpose TEXT NOT NULL,
			amount REAL NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_expenses_created ON expenses(created_at)`,
		`CREATE TABLE IF NOT EXISTS inventory_transactions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			product_id INTEGER NOT NULL,
			barcode TEXT NOT NULL DEFAULT '',
			type TEXT NOT NULL,
			quantity INTEGER NOT NULL DEFAULT 0,
			note TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_inventory_created ON inventory_transactions(created_at)`,
		`CREATE TABLE IF NOT EXISTS markets (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			date TEXT NOT NULL,
			location TEXT NOT NULL DEFAULT '',
			income REAL NOT NULL DEFAULT 0,
			note TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_markets_date ON markets(date)`,
		`CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL DEFAULT '',
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS planner_templates (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			size_key TEXT NOT NULL,
			width_mm REAL NOT NULL,
			height_mm REAL NOT NULL,
			padding_mm REAL NOT NULL,
			accent TEXT NOT NULL DEFAULT '',
			paper TEXT NOT NULL DEFAULT '',
			modules_json TEXT NOT NULL DEFAULT '[]',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_planner_updated ON planner_templates(updated_at)`,
	}

	for _, stmt := range stmts {
		if _, err := conn.Exec(stmt); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	return nil
}

func seed(conn *sql.DB) error {
	// 默认定价规则（与前端 defaultPricingRule 保持一致）。
	if _, err := conn.Exec(
		`INSERT OR IGNORE INTO pricing_rules (id, target_margin, channel_fee_rate, extra_cost, rounding_mode)
		 VALUES (1, 0.45, 0.05, 1.2, 'x.9')`,
	); err != nil {
		return fmt.Errorf("seed pricing rule: %w", err)
	}

	now := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")

	// “未分类” 固定 id = 0，其余默认分类仅在不存在时补建。
	if _, err := conn.Exec(
		`INSERT OR IGNORE INTO categories (id, name, created_at) VALUES (0, '未分类', ?)`, now,
	); err != nil {
		return fmt.Errorf("seed category 未分类: %w", err)
	}
	for _, name := range []string{"胶带", "贴纸", "便签", "套装"} {
		if _, err := conn.Exec(
			`INSERT OR IGNORE INTO categories (name, created_at) VALUES (?, ?)`, name, now,
		); err != nil {
			return fmt.Errorf("seed category %s: %w", name, err)
		}
	}
	return seedAuth(conn, now)
}

// seedAuth 补建登录口令与会话密钥，已有值时不覆盖。
func seedAuth(conn *sql.DB, now string) error {
	var passcodeRows int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM settings WHERE key = 'auth.passcodeHash'`).Scan(&passcodeRows); err != nil {
		return fmt.Errorf("check passcode: %w", err)
	}
	if passcodeRows == 0 {
		credentials, err := auth.HashPasscode(auth.DefaultPasscode)
		if err != nil {
			return fmt.Errorf("hash default passcode: %w", err)
		}
		values := []struct {
			key   string
			value string
		}{
			{"auth.passcodeSalt", credentials.Salt},
			{"auth.passcodeHash", credentials.Hash},
			{"auth.passcodeIterations", strconv.Itoa(credentials.Iterations)},
			{"auth.passcodeIsDefault", "1"},
		}
		for _, item := range values {
			if _, err := conn.Exec(
				`INSERT OR IGNORE INTO settings (key, value, updated_at) VALUES (?, ?, ?)`,
				item.key, item.value, now,
			); err != nil {
				return fmt.Errorf("seed setting %s: %w", item.key, err)
			}
		}
	}

	var secretRows int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM settings WHERE key = 'auth.sessionSecret'`).Scan(&secretRows); err != nil {
		return fmt.Errorf("check session secret: %w", err)
	}
	if secretRows == 0 {
		secret, err := auth.RandomToken(32)
		if err != nil {
			return fmt.Errorf("generate session secret: %w", err)
		}
		if _, err := conn.Exec(
			`INSERT OR IGNORE INTO settings (key, value, updated_at) VALUES ('auth.sessionSecret', ?, ?)`, secret, now,
		); err != nil {
			return fmt.Errorf("seed session secret: %w", err)
		}
	}
	return nil
}
