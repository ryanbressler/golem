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
	"fmt"
	"websocket"
	"os"
	"crypto/tls"
	"http"
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

	ws, err = websocket.Dial(url, "", origin)
	if err != nil {
		return nil, err
	}
	return ws, nil

}

func getCertFilePaths() (string, string) {
	certf := os.ShellExpand("$HOME/.golem/certificate.pem")
	keyf := os.ShellExpand("$HOME/.golem/key.pem")
	return certf, keyf
}

func getTlsConfig() (*tls.Config, os.Error) {
	certf, keyf := getCertFilePaths()
	cert, err := tls.LoadX509KeyPair(certf, keyf)
	if err != nil {
		vlog("Err loading tls keys from %v and %v: %v\n", certf, keyf, err)
		return nil, err
	}

	return &tls.Config{Certificates: []tls.Certificate{cert}, AuthenticateClient: true}, nil

}

func ConfigListenAndServeTLS(hostname string, handler http.Handler) os.Error {
	cfg, err := getTlsConfig()
	if err != nil {
		return err
	}
	listener, err := tls.Listen("tcp", hostname, cfg)
	if err != nil {
		vlog("Tls Listen Error : %v", err)
		return err
	}

	if err := http.Serve(listener, handler); err != nil {
		vlog("Tls Serve Error : %v", err)
		return err
	}
	return nil
}

//relys on global useTls being set
func ListenAndServeTLSorNot(hostname string, handler http.Handler) os.Error {
	if useTls {
		if err := ConfigListenAndServeTLS(hostname, nil); err != nil {
			vlog("ConfigListenAndServeTLS : %v", err)
			return err
		}

	} else {
		if err := http.ListenAndServe(hostname, nil); err != nil {
			vlog("ListenAndServe Error : %v", err)
			return err
		}

	}
	return nil
}
