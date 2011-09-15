/*
   Copyright (C) 2003-2011 Institute for Systems Biology
                           Seattle, Washington, USA.

   This library is free software; you can redistribute it and/or
   modify it under the terms of the GNU Lesser General Public
   License as published by the Free Software Foundation; either
   version 2.1 of the License, or (at your option) any later version.

   This library is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
   Lesser General Public License for more details.

   You should have received a copy of the GNU Lesser General Public
   License along with this library; if not, write to the Free Software
   Foundation, Inc., 59 Temple Place, Suite 330, Boston, MA 02111-1307  USA

*/

package main

import (
	"os"
	"websocket"
	"json"
	"bufio"
	"time"
)

// represents one end of a web socket and has facilities for sending and receiving messages via chans
type Connection struct {
	Socket   *websocket.Conn    //the socket that the connection wraps
	OutChan  chan WorkerMessage // the out box. send messages with c.OutChan<-msg
	InChan   chan WorkerMessage // the in box. getmsg:=<-c.InChan
	ReConChan chan WorkerMessage
	DiedChan chan int           // send died message out on this
	isWorker bool               // indicates if this connection is for a worker node
}

// Wraps a web socket in a connection starts routines that receive and send messages
func NewConnection(Socket *websocket.Conn, isWorker bool) *Connection {
	n := Connection{Socket: Socket,
		OutChan:  make(chan WorkerMessage, conbuffersize),
		InChan:   make(chan WorkerMessage, conbuffersize),
		ReConChan: make(chan WorkerMessage, 0),
		DiedChan: make(chan int, 1),
		isWorker: isWorker}
	go n.GetMsgs()
	go n.SendMsgs()
	return &n
}

// monitor OutChan and sends any messages through web socket usually started in NewConnection
func (con Connection) SendMsgs() {
	for {
		msg := <-con.OutChan

		msgjson, err := json.Marshal(msg)
		if err != nil {
			logger.Warn(err)
			return
		}

		//try to send over and over.
		sent := false
		for ( sent ) {
			if _, err := con.Socket.Write(msgjson); err != nil {
				logger.Warn(err)
			} else {
				sent= true
			}
			<-time.After(25*second)
		}
		
	}
}

// monitor web socket and put messages in the InChan usually started in NewConnection
func (con Connection) GetMsgs() {
	for {
		decoder := json.NewDecoder(con.Socket)
		var msg WorkerMessage
		err := decoder.Decode(&msg)
		if err != nil {
			logger.Warn(err)
		}

		switch {
		case err == os.EOF:
			remote := con.Socket.RemoteAddr().String()
			con.Socket.Close()
			if con.isWorker {
				
			
				con.DiedChan <- 1
				msg :=<- con.ReConChan
				msgjson, err := json.Marshal(msg)
					if err != nil {
						logger.Warn(err)
				}
				
				reconnected := false
				
				for t:=1;t<16;t=t*2 {
				
					logger.Printf("Atempting reconnect in %v seconds.",t)
					<-time.After(int64(t)*second)
					logger.Printf("Atempting reconnect now.")
					
					ws, err := DialWebSocket(remote) 
					if err != nil {
						logger.Warn(err)
						continue
					}

					if _, err = ws.Write(msgjson); err != nil {
						logger.Warn(err)
					} else {
						reconnected=true
						con.Socket=ws
						break
					}
					
				}
				if reconnected != true {
					DieIn(10)
				}
				
			} 

			return //TODO: recover
		case err == bufio.ErrBufferFull:
			decoder = json.NewDecoder(con.Socket)
		case err != nil:
			continue
		}
		con.InChan <- msg
	}
}
