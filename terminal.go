package main

import (
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"net/rpc"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/creack/pty"
	_ "github.com/mattn/go-sqlite3"
	"github.com/prometheus/procfs"
	"golang.org/x/net/websocket"
)

const ENV_ID_NAME = "TERMINAL2_ID"
const ENV_BIND_NAME = "TERMINAL2_BIND"

type Context struct {
	bindAddress string
	connections []*Connection
	lock        sync.Mutex
	procFs      procfs.FS
	db          *sql.DB
}

func (ctx *Context) getConnection(id string) *Connection {
	for _, conn := range ctx.connections {
		if conn.id == id {
			return conn
		}
	}
	return nil
}

func (ctx *Context) getConnectionLock(id string) *Connection {
	ctx.lock.Lock()
	conn := ctx.getConnection(id)
	ctx.lock.Unlock()
	return conn
}

func (ctx *Context) setupDatabase() error {
	if err := ctx.SetupThemeTable(); err != nil {
		return err
	}
	if err := ctx.SetupSettingsTable(); err != nil {
		return err
	}
	return nil
}

func main() {
	id, idPresent := os.LookupEnv(ENV_ID_NAME)
	bind, bindPresent := os.LookupEnv(ENV_BIND_NAME)
	if idPresent && bindPresent {
		cli(bind, id)
		return
	}

	bindAddress := flag.String("bind", "127.0.0.1:4100", "Bind address")
	databasePath := flag.String("database", "terminal.db", "Path to the database")
	webPath := flag.String("webdir", "./web", "Path to the static web folder")
	flag.Parse()
	bindEnv, bindEnvPresent := os.LookupEnv("BIND")
	if bindEnvPresent {
		*bindAddress = bindEnv
	}
	dbEnv, dbEnvPresent := os.LookupEnv("DATABASE")
	if dbEnvPresent {
		*databasePath = dbEnv
	}
	webEnv, webEnvPresent := os.LookupEnv("WEBDIR")
	if webEnvPresent {
		*webPath = webEnv
	}

	ctx := &Context{
		bindAddress: *bindAddress,
		connections: make([]*Connection, 0),
	}

	db, err := sql.Open("sqlite3", *databasePath)
	if err != nil {
		panic(err)
	}
	ctx.db = db
	if err := ctx.setupDatabase(); err != nil {
		panic(err)
	}

	rpc.Register(ctx)
	rpc.HandleHTTP() // CRITICAL: Need to figure out how exposed this endpoint is, at the end of the day it is protected by the reverse proxy

	procFs, err := procfs.NewFS("/proc")
	if err != nil {
		panic(err)
	}
	ctx.procFs = procFs

	http.Handle("/", http.FileServer(http.Dir(*webPath)))
	http.HandleFunc("/download", ctx.fileDownload)
	http.HandleFunc("/upload", ctx.fileUpload)
	http.Handle("/ws", websocket.Handler(func(c *websocket.Conn) {
		conn := Connection{ws: c, downloadPath: "", dead: false}
		conn.waiter.Add(1)

		ctx.lock.Lock()
		for {
			id := RandStr(64)
			if c := ctx.getConnection(id); c == nil {
				conn.id = id
				break
			}
		}
		ctx.connections = append(ctx.connections, &conn)
		ctx.lock.Unlock()

		cwd, err := os.Getwd()
		if err != nil {
			conn.close()
			return
		}

		cmd := exec.Command("/bin/bash")
		if dir, err := os.UserHomeDir(); err != nil {
			conn.close()
			return
		} else {
			cmd.Dir = dir
		}

		// TODO: Work on building out env vars better
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", ENV_ID_NAME, conn.id))
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", ENV_BIND_NAME, ctx.bindAddress))
		cmd.Env = append(cmd.Env, "TERM=xterm-256color")
		for _, env := range os.Environ() {
			key := strings.ToLower(strings.Split(env, "=")[0])
			if key == "path" {
				cmd.Env = append(cmd.Env, fmt.Sprintf("%s:%s", env, cwd))
			}
			if key == "home" || key == "shell" || key == "logname" || key == "user" || key == "mail" || key == "lang" {
				cmd.Env = append(cmd.Env, env)
			}
		}

		conn.pty, err = pty.Start(cmd)
		if err != nil {
			conn.close()
			return
		}
		conn.proc = cmd.Process
		conn.wsWriteConnectionId(conn.id)

		theme, err := ctx.ThemeQuery(ctx.SettingGet("theme", "default"))
		if err == nil {
			conn.wsWriteTheme(theme)
		}

		go conn.handlePTY()
		go conn.handleSocket()
		conn.waiter.Wait()
	}))

	fmt.Printf("Serving at %s\n", ctx.bindAddress)
	http.ListenAndServe(ctx.bindAddress, nil)
}
