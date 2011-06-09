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
package main

import (
	"fmt"
	"websocket"
	"os"
	"crypto/tls"
)

//////////////////////
//tls

func wsDialToMaster(master string, useTls bool) (ws *websocket.Conn, err os.Error) {

	origin, err := os.Hostname()
	if err != nil {
		log("Error getting hostname")
	}
	prot := "ws"

	if useTls {
		prot = "wss"

	}
	url := fmt.Sprintf("%v://"+master+"/master/", prot)
	/*var client net.Conn
		parsedUrl, err := http.ParseURL(url)
		if err != nil {
			goto Error
		}

		switch prot {
		case "ws":
			client, err = net.Dial("tcp", parsedUrl.Host)

		case "wss":
			client, err = tls.Dial("tcp", parsedUrl.Host, getTlsConfig())


		}
		if err != nil {
				goto Error
	   	}

		ws, err = websocket.newClient(parsedUrl.RawPath, parsedUrl.Host, origin, url, "", client, websocket.handshake)
		if err != nil {
			goto Error
		}
		return

	   	Error:
			return nil, &websocket.DialError{url, "", origin, err}*/
	ws, err = websocket.Dial(url, "", origin)
	if err != nil {
		return nil, err
	}
	return ws, nil

}

func getCertFiles() (string, string) {
	certf := os.ShellExpand("$HOME/.golem/certificate.pem")
	keyf := os.ShellExpand("$HOME/.golem/key.pem")
	return certf, keyf
}

func getTlsConfig() *tls.Config {
	certf, keyf := getCertFiles()
	cert, err := tls.LoadX509KeyPair(certf, keyf)
	if err != nil {
		log("Err loading tls keys from %v and %v: %v\n", certf, keyf, err)
	}

	return &tls.Config{Certificates: []tls.Certificate{cert}, AuthenticateClient: true}

}
