// Work Engine Core
// Copyright (C) 2025 KitWork
// Author: Huỳnh Nhân Quốc | GitHub: https://github.com/huynhnhanquoc
// Licensed under GNU AGPL v3.0 — see <https://www.gnu.org/licenses/agpl-3.0.html>

package work

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func (cfg *Config) Database(source ...string) *Config {
	cfg.database = NewSource("database", source...)
	return cfg
}

type Database struct {
	defaultDB *gorm.DB
	list      map[string]*gorm.DB
}

// Add thêm một DB vào list, nếu isDefault = true hoặc list rỗng sẽ đặt làm default
func (dbs *Database) Add(key string, gormDB *gorm.DB, isDefault bool) {
	if dbs.list == nil {
		dbs.list = make(map[string]*gorm.DB)
	}

	// Nếu là DB mặc định hoặc list rỗng, gán default
	if isDefault || dbs.defaultDB == nil {
		dbs.defaultDB = gormDB
	}

	// Thêm vào map
	dbs.list[key] = gormDB
}

// Get lấy DB theo key, fallback về default nếu key không tồn tại
func (dbs *Database) Get(keys ...string) *gorm.DB {
	if len(keys) == 0 {
		return dbs.defaultDB
	}

	for _, key := range keys {
		if db, ok := dbs.list[key]; ok && db != nil {
			return db
		}
	}

	return dbs.defaultDB
}

type DBConnect struct {
	SQL       *sql.DB
	Gorm      *gorm.DB
	Key       string // tên key từ file config, ví dụ "postgresql"
	IsDefault bool
}

type DBConfig struct {
	Type     string `json:"type"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	SSL      string `json:"ssl"`
	Timezone string `json:"timezone"`
	Timeout  int    `json:"timeout"`
	MaxOpen  int    `json:"max_open"`
	MaxIdle  int    `json:"max_idle"`
	Lifetime int    `json:"lifetime"`
	MaxLimit int    `json:"max_limit"`
	Default  bool   `json:"default"`
}

// DSN tạo connection string PostgreSQL
func (c *DBConfig) DSN() string {
	ssl := c.SSL
	if ssl == "" {
		ssl = "disable"
	}
	return fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s port=%d sslmode=%s TimeZone=%s",
		c.User, c.Password, c.Name, c.Host, c.Port, ssl, c.Timezone,
	)
}

// ConnectDB kết nối tất cả database từ file config
func (cfg *Config) ConnectDB(accepts ...string) ([]DBConnect, error) {
	listDB, err := cfg.database.Scanner(accepts...)
	if err != nil {
		return nil, err
	}

	var dbs []DBConnect

	for key, raw := range listDB {
		// raw là map[string]interface{} với các field config
		data, err := json.Marshal(raw)
		if err != nil {
			return nil, fmt.Errorf("marshal db %s: %w", key, err)
		}

		var dbCfg DBConfig
		if err := json.Unmarshal(data, &dbCfg); err != nil {
			return nil, fmt.Errorf("unmarshal db %s: %w", key, err)
		}

		// Mở kết nối GORM
		gormDB, err := gorm.Open(postgres.Open(dbCfg.DSN()), &gorm.Config{})
		if err != nil {
			return nil, fmt.Errorf("connect db %s: %w", key, err)
		}

		// Lấy *sql.DB để quản lý connection pool và đóng
		sqlDB, err := gormDB.DB()
		if err != nil {
			return nil, fmt.Errorf("get sql.DB for db %s: %w", key, err)
		}

		sqlDB.SetMaxOpenConns(dbCfg.MaxOpen)
		sqlDB.SetMaxIdleConns(dbCfg.MaxIdle)
		sqlDB.SetConnMaxLifetime(time.Duration(dbCfg.Lifetime) * time.Minute)

		dbs = append(dbs, DBConnect{
			Key:       key,
			Gorm:      gormDB,
			SQL:       sqlDB,
			IsDefault: dbCfg.Default,
		})
	}

	return dbs, nil
}
