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
/*Connection represents one end of a websocket and has facilities 
For sending and recieving messages via channels
TODO:Needs code to clean up after it is done...*/

package main

import (
	"os"
	"websocket"
	"json"
	"bufio"
)

type Connection struct {
	Socket  *websocket.Conn //the socket that the connection wraps
	OutChan chan clientMsg  // the out box. send messages with c.OutChan<-msg
	InChan  chan clientMsg  // the in box. getmsg:=<-c.InChan
}

//Wraps a websocket in a connection starts the goroutines that recieve and send messages
func NewConnection(Socket *websocket.Conn) *Connection {
	n := Connection{Socket: Socket, OutChan: make(chan clientMsg, 10), InChan: make(chan clientMsg, 10)}
	go n.GetMsgs()
	go n.SendMsgs()
	return &n
}

//goroutine to monitor the OutChan and send any messages through the websocket
//usually started in NewConnection
func (con Connection) SendMsgs() {

	for {
		msg := <-con.OutChan

		msgjson, err := json.Marshal(msg)
		if err != nil {
			log("error json.Marshaling msg: %v", err)
			return
		}

		//log("sending:%v\n", string(msgjson))
		if _, err := con.Socket.Write(msgjson); err != nil {
			log("Error sending msg: %v", err)
		}
		//log("msg sent\n")
	}

}

//goroutine to monitor websocket and put the messages in the InChan
//usually started in NewConnection
func (con Connection) GetMsgs() {

	for {
		decoder := json.NewDecoder(con.Socket)
		var msg clientMsg
		err := decoder.Decode(&msg)
		switch {
		case err == os.EOF:
			log("EOF recieved on websocket.")
			con.Socket.Close()
			if isMaster == true {
				DieIn(10000000000)
			}

			return //TODO: recover
		case err == bufio.ErrBufferFull:
			log("buffer full, restarting json decoder: %v", err)
			decoder = json.NewDecoder(con.Socket)
		case err != nil:
			log("error parseing client msg json: %v", err)
			continue
		}
		con.InChan <- msg

	}

}
