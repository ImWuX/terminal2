package main

import (
	"database/sql"
	"encoding/json"
)

type Theme struct {
	Id              string     `json:"id"`
	Font            string     `json:"font"`
	CustomFont      json.Token `json:"customFont"`
	BackgroundImage json.Token `json:"bgImg"`
	XTerm           struct {
		Foreground    string `json:"foreground"`
		Background    string `json:"background"`
		Cursor        string `json:"cursor"`
		Selection     string `json:"selectionBackgroundTransparent"`
		Black         string `json:"black"`
		BrightBlack   string `json:"brightBlack"`
		Red           string `json:"red"`
		BrightRed     string `json:"brightRed"`
		Green         string `json:"green"`
		BrightGreen   string `json:"brightGreen"`
		Yellow        string `json:"yellow"`
		BrightYellow  string `json:"brightYellow"`
		Blue          string `json:"blue"`
		BrightBlue    string `json:"brightBlue"`
		Magenta       string `json:"magenta"`
		BrightMagenta string `json:"brightMagenta"`
		Cyan          string `json:"cyan"`
		BrightCyan    string `json:"brightCyan"`
		White         string `json:"white"`
		BrightWhite   string `json:"brightWhite"`
	} `json:"xTerm"`
}

func (ctx *Context) ThemeQuery(id string) (Theme, error) {
	row := ctx.db.QueryRow("SELECT * FROM themes WHERE id = ?", id)
	theme := Theme{}
	var backgroundImage sql.NullString
	var customFont sql.NullString
	if err := row.Scan(
		&theme.Id,
		&theme.Font,
		&customFont,
		&theme.XTerm.Foreground,
		&theme.XTerm.Background,
		&backgroundImage,
		&theme.XTerm.Cursor,
		&theme.XTerm.Selection,
		&theme.XTerm.Black,
		&theme.XTerm.BrightBlack,
		&theme.XTerm.Red,
		&theme.XTerm.BrightRed,
		&theme.XTerm.Green,
		&theme.XTerm.BrightGreen,
		&theme.XTerm.Yellow,
		&theme.XTerm.BrightYellow,
		&theme.XTerm.Blue,
		&theme.XTerm.BrightBlue,
		&theme.XTerm.Magenta,
		&theme.XTerm.BrightMagenta,
		&theme.XTerm.Cyan,
		&theme.XTerm.BrightCyan,
		&theme.XTerm.White,
		&theme.XTerm.BrightWhite,
	); err != nil {
		return theme, err
	}
	theme.BackgroundImage = nil
	if backgroundImage.Valid {
		theme.BackgroundImage = backgroundImage.String
	}
	theme.CustomFont = nil
	if customFont.Valid {
		theme.CustomFont = customFont.String
	}
	return theme, nil
}

func (ctx *Context) SetupThemeTable() error {
	if _, err := ctx.db.Exec(`
	CREATE TABLE IF NOT EXISTS themes (
		id TEXT PRIMARY KEY,
		font TEXT,
		customFont TEXT,
		foreground TEXT,
		background TEXT,
		backgroundImage TEXT,
		cursor TEXT,
		selection TEXT,
		black TEXT,
		brightBlack TEXT,
		red TEXT,
		brightRed TEXT,
		green TEXT,
		brightGreen TEXT,
		yellow TEXT,
		brightYellow TEXT,
		blue TEXT,
		brightBlue TEXT,
		magenta TEXT,
		brightMagenta TEXT,
		cyan TEXT,
		brightCyan TEXT,
		white TEXT,
		brightWhite TEXT
	);`,
	); err != nil {
		return err
	}
	_, err := ctx.db.Exec(`
	INSERT OR IGNORE INTO themes (
		id, font, customFont, foreground, background, backgroundImage, cursor, selection,
		black, brightBlack,
		red, brightRed,
		green, brightGreen,
		yellow, brightYellow,
		blue, brightBlue,
		magenta, brightMagenta,
		cyan, brightCyan,
		white, brightWhite
	) VALUES (
		'default', 'monospace', NULL, '#fff', '#000', NULL, '#f81ce5', 'rgba(255, 255, 255, 0.1)',
		'#000000', '#686868',
		'#C51E14', '#FD6F6B',
		'#1DC121', '#67F86F',
		'#C7C329', '#FFFA72',
		'#0A2FC4', '#6A76FB',
		'#C839C5', '#FD7CFC',
		'#20C5C6', '#68FDFE',
		'#C7C7C7', '#FFFFFF'
	);`)
	return err
}
