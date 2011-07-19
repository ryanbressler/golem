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
	"crypto/rsa"
	"crypto/x509"
	"crypto/rand"
	"encoding/pem"
	"time"
	"http"
)

//connect a websocket to the master as a worker
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

//returns our custom tls configuration
func getTlsConfig() *tls.Config {
	certs := []tls.Certificate{}

    certpath, _ := ConfigFile.GetString("default", "certpath")
    if certpath != "" {
		certs = append(certs, GenerateX509KeyPair(certpath))
    } else {
		certs = append(certs, GenerateTlsCert())
    }

	return &tls.Config{Certificates: certs, AuthenticateClient: true}
}

//a replacment for ListenAndServeTLS that loads our custom confiuration usage is identical to http.ListenAndServe
func ConfigListenAndServeTLS(hostname string, handler http.Handler) (err os.Error) {
	listener, err := tls.Listen("tcp", hostname, getTlsConfig())
	if err != nil {
		vlog("Tls Listen Error : %v", err)
		return
	}

	if err := http.Serve(listener, handler); err != nil {
		vlog("Tls Serve Error : %v", err)
	}
	return
}

func GenerateX509KeyPair(certpath string) tls.Certificate {
	certf := os.ShellExpand(certpath + "/certificate.pem")
	keyf := os.ShellExpand(certpath + "/key.pem")

    cert, err := tls.LoadX509KeyPair(certf, keyf)
    if err != nil {
        vlog("Err loading tls keys from %v and %v: %v", certf, keyf, err)
        panic(err)
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
	organization, err := ConfigFile.GetString("default", "organization")
	if err != nil {
	    organization = "Golem"
	}

	template := x509.Certificate{
		SerialNumber: []byte{0},
		PublicKeyAlgorithm: x509.RSA,
		Subject: x509.Name{
			CommonName:   hostname,
			Organization: []string{organization},
		},
		NotBefore: time.SecondsToUTC(now - 300),
		NotAfter:  time.SecondsToUTC(now + year),
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
