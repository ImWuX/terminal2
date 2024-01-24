package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
)

type Connection struct {
	id   string
	proc *os.Process
	ws   *websocket.Conn
	pty  *os.File

	downloadPath string

	dead   bool
	lock   sync.Mutex
	waiter sync.WaitGroup
}

func (conn *Connection) close() {
	conn.lock.Lock()
	if !conn.dead {
		conn.dead = true
		if conn.ws != nil {
			conn.ws.Close()
		}
		if conn.proc != nil {
			conn.proc.Kill()
		}
		conn.waiter.Done()
	}
	conn.lock.Unlock()
}

func (conn *Connection) wsWrite(messageType byte, data []byte) {
	err := conn.ws.WriteMessage(websocket.BinaryMessage, append([]byte{messageType}, data...))
	if err != nil {
		fmt.Println(">> wsWrite() error.", err)
		conn.close()
	}
}

func (conn *Connection) wsWriteMessage(message string) {
	conn.wsWrite(1, []byte(message))
}

func (conn *Connection) wsWriteConnectionId(id string) {
	conn.wsWrite(2, []byte(id))
}

func (conn *Connection) wsWriteTheme(theme Theme) {
	j, err := json.Marshal(theme)
	if err != nil {
		fmt.Println(err)
		conn.close()
	}
	conn.wsWrite(3, j)
}

func (conn *Connection) wsWriteToaster(message string) {
	conn.wsWrite(4, []byte(message))
}

func (conn *Connection) wsWriteOpenWindow(url string) {
	conn.wsWrite(5, []byte(url))
}

func (conn *Connection) handlePTY() {
	defer fmt.Printf(">> HandlePTY(%s) closed\n", conn.id)
	buf := make([]byte, 2048)
	for {
		n, err := conn.pty.Read(buf)
		if err != nil {
			conn.wsWriteMessage("pty closed. exiting")
			conn.ws.Close()
			break
		}
		conn.wsWriteMessage(string(buf[:n]))
	}
}

func (conn *Connection) handleSocket() {
	defer fmt.Printf(">> HandleSocket(%s) closed\n", conn.id)
	for {
		mtype, buf, err := conn.ws.ReadMessage()
		if mtype == -1 || mtype == websocket.CloseMessage || mtype == websocket.CloseMessageTooBig {
			return
		}
		if mtype != websocket.BinaryMessage {
			fmt.Println(">> invalid message type:", mtype)
			return
		}
		if err != nil {
			fmt.Println(">> ws read error.", err)
			return
		}
		if len(buf) < 1 {
			continue
		}
		switch buf[0] {
		case 1: // UTF-8 Encoded Key
			if len(buf) < 2 {
				continue
			}
			chLen := buf[1]
			if len(buf) != 2+int(chLen) {
				continue
			}
			_, err = conn.pty.Write(buf[2 : 2+chLen])
			if err != nil {
				fmt.Println(">> pty write error.", err)
				return
			}
		case 2: // Term Resize
			if len(buf) < 5 {
				continue
			}
			cols := uint16(buf[1])<<8 | uint16(buf[2])
			rows := uint16(buf[3])<<8 | uint16(buf[4])
			pty.Setsize(conn.pty, &pty.Winsize{
				Rows: rows,
				Cols: cols,
			})
		}
	}
}
