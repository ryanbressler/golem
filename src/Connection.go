
/*
   Copyright (C) 2003-2010 Institute for Systems Biology
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
For sending and recieving connections
TODO:Needs code to clean up after it is done...*/

package main

import (
	"os"
	"websocket"
	"json"
	"fmt"
	
)

type Connection struct {
	Socket  *websocket.Conn
	OutChan chan clientMsg
	InChan  chan clientMsg
}

func NewConnection(Socket *websocket.Conn) *Connection {
	n := Connection{Socket: Socket, OutChan: make(chan clientMsg, 10), InChan: make(chan clientMsg, 10)}
	go n.GetMsgs()
	go n.SendMsgs()
	return &n
}

func (con Connection) SendMsgs() {

	for {
		msg := <-con.OutChan

		msgjson, err := json.Marshal(msg)
		if err != nil {
			fmt.Printf("error json.Marshaling msg: %v\n", err)
			return
		}

		//fmt.Printf("sending:%v\n", string(msgjson))
		if _, err := con.Socket.Write(msgjson); err != nil {
			fmt.Printf("Error sending msg: %v\n", err)
		}
		//fmt.Printf("msg sent\n")
	}

}

func (con Connection) GetMsgs() {
	decoder := json.NewDecoder(con.Socket)
	for {
		var msg clientMsg
		err := decoder.Decode(&msg)
		switch {
		case err == os.EOF:
			fmt.Printf("EOF recieved on websocket.")
			con.Socket.Close()
			return //TODO: recover
		case err != nil:
			fmt.Printf("error parseing client msg json: %v\n", err)
			continue
		}
		con.InChan <- msg

	}

}