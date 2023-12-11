package main

import (
	"fmt"
)

type RPCArgs struct {
	ConnectionId string
}

type SetThemeArgs struct {
	RPCArgs
	ThemeId string
}

func (ctx *Context) SetTheme(args *SetThemeArgs, reply *string) error {
	theme, err := ctx.ThemeQuery(args.ThemeId)
	if err != nil {
		*reply = "Invalid theme"
		return nil
	}

	conn := ctx.getConnectionLock(args.ConnectionId)
	if conn == nil {
		*reply = "Invalid connection id"
		return nil
	}
	conn.wsWriteTheme(theme)
	ctx.SettingSet("theme", theme.Id)

	*reply = fmt.Sprintf("Theme updated to %s", args.ThemeId)
	return nil
}

type PopToasterArgs struct {
	RPCArgs
	Text string
}

func (ctx *Context) PopToaster(args *PopToasterArgs, reply *bool) error {
	*reply = false
	conn := ctx.getConnectionLock(args.ConnectionId)
	if conn == nil {
		return nil
	}

	conn.wsWriteToaster(args.Text)
	*reply = true
	return nil
}

type DownloadArgs struct {
	RPCArgs
	Path string
}

func (ctx *Context) Download(args *DownloadArgs, reply *string) error {
	conn := ctx.getConnectionLock(args.ConnectionId)
	if conn == nil {
		*reply = "Invalid connection id"
		return nil
	}

	if conn.downloadPath != "" {
		*reply = fmt.Sprintf("Already downloading %s", conn.downloadPath)
		return nil
	}
	conn.downloadPath = args.Path
	conn.wsWriteOpenWindow("/download")
	*reply = fmt.Sprintf("Downloading %s", args.Path)
	return nil
}
