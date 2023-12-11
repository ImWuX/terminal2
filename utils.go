package main

import (
	"math/rand"
	"os"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

func RandStr(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func (ctx *Context) SettingGet(key string, defaultValue string) string {
	row := ctx.db.QueryRow("SELECT value FROM settings WHERE key = ?;", key)
	var value string
	if err := row.Scan(&value); err != nil {
		return defaultValue
	}
	return value
}

func (ctx *Context) SettingSet(key string, value string) error {
	_, err := ctx.db.Exec("INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?);", key, value)
	return err
}

func (ctx *Context) SetupSettingsTable() error {
	_, err := ctx.db.Exec(`
	CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT
	);`,
	)
	return err
}
