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
	"big"
	"os"
	"crypto/tls"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"crypto/rand"
	"encoding/pem"
	"time"
	"http"
)

//connect a web socket to the master as a worker
func OpenWebSocketToMaster(master string) (ws *websocket.Conn) {
	logger.Debug("OpenWebSocketToMaster(%v)", master)
	

	prot := "ws"
	if useTls {
		prot = "wss"
	}

	url := fmt.Sprintf("%v://%v/master/", prot, master)
	ws = DialWebSocket(url)
	return
}

func DialWebSocket(url string) (ws *websocket.Conn){
	origin, err := os.Hostname()
	if err != nil {
		logger.Warn(err)
	}
	if ws, err = websocket.Dial(url, "", origin); err != nil {
		logger.Warn(err)
	}
	return
}

//returns our custom tls configuration
func GetTlsConfig() *tls.Config {
	logger.Debug("GetTlsConfig")
	certs := []tls.Certificate{}

	if certpath != "" {
		certs = append(certs, GenerateX509KeyPair(certpath))
	} else {
		certs = append(certs, GenerateTlsCert())
	}

	return &tls.Config{Certificates: certs, AuthenticateClient: true}
}

// replacement for ListenAndServeTLS that loads our custom configuration usage is identical to http.ListenAndServe
func ConfigListenAndServeTLS(hostname string) (err os.Error) {
	logger.Printf("ConfigListenAndServeTLS(%v)", hostname)
	listener, err := tls.Listen("tcp", hostname, GetTlsConfig())
	if err != nil {
		logger.Warn(err)
		return
	}

	if err := http.Serve(listener, nil); err != nil {
		logger.Warn(err)
	}
	return
}

func GenerateX509KeyPair(certpath string) tls.Certificate {
	certf := os.ShellExpand(certpath + "/certificate.pem")
	keyf := os.ShellExpand(certpath + "/key.pem")

	cert, err := tls.LoadX509KeyPair(certf, keyf)
	if err != nil {
		logger.Printf("GenerateX509KeyPair(%v): %v", certf, keyf)
		logger.Panic(err)
	}
	return cert
}

func GenerateTlsCert() tls.Certificate {
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	priv, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}

	now := time.Seconds()

	randomSerialNum, err := rand.Int(rand.Reader, big.NewInt(9223372036854775807))
	if err != nil {
		panic(err)
	}

	template := x509.Certificate{
		SerialNumber:       randomSerialNum,
		PublicKeyAlgorithm: x509.RSA,
		Subject: pkix.Name{
			CommonName:   hostname,
			Organization: []string{certorg},
		},
		NotBefore:    time.SecondsToUTC(now - 300),
		NotAfter:     time.SecondsToUTC(now + year),
		SubjectKeyId: []byte{1, 2, 3, 4},
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
	}

	certbyte, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		panic(err)
	}

	cert, err := tls.X509KeyPair(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certbyte}), pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)}))
	if err != nil {
		panic(err)
	}
	return cert
}

// setup master, usage is identical to http.ListenAndServe but this relies on global useTls being set
func ListenAndServeTLSorNot(hostname string) (err os.Error) {
	logger.Debug("ListenAndServeTLSorNot(%v)", hostname)
	if useTls {
		if err = ConfigListenAndServeTLS(hostname); err != nil {
			logger.Warn(err)
		}
	} else {
		if err = http.ListenAndServe(hostname, nil); err != nil {
			logger.Warn(err)
		}
	}
	return
}
