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

//////////////////////
//tls
//connect a websocket to the master as a client
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

//get the cert file paths for running as the master
func getCertFilePaths() (string, string) {
	certf := os.ShellExpand("$HOME/.golem/certificate.pem")
	keyf := os.ShellExpand("$HOME/.golem/key.pem")
	return certf, keyf
}

//returns our custom tls configuration
func getTlsConfig() (*tls.Config, os.Error) {
	if certpath!=""{
		
	}
	certs := []tls.Certificate{}
	if isMaster {
		var cert tls.Certificate
		var err os.Error
		switch {
		case certpath!="":
			certf, keyf := getCertFilePaths()
			cert, err = tls.LoadX509KeyPair(certf, keyf)
			if err != nil {
				vlog("Err loading tls keys from %v and %v: %v\n", certf, keyf, err)
				return nil, err
			}
		default:
			cert, err =GenerateTlsCert()
			if err != nil {
				log("Error generating tls cert: %v", err)
			}
		}
		
		certs = append(certs, cert)
	}

	return &tls.Config{Certificates: certs, AuthenticateClient: true}, nil

}

//a replacment for ListenAndServeTLS that loads our custom confiuration usage is identical to http.ListenAndServe
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


//Create tls certificate
func GenerateTlsCert()(cert tls.Certificate, err os.Error){
	hostname, err := os.Hostname()
	if err != nil {
		log("Error getting hostname")
		return
	}
	
	priv, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		log("failed to generate private key: %s", err)
		return
	}
	
	now := time.Seconds()
	
	template := x509.Certificate{
		SerialNumber: []byte{0},
		
		//SignatureAlgorithm: x509.MD5WithRSA,
		PublicKeyAlgorithm: x509.RSA,
		Subject: x509.Name{
			CommonName:   hostname,
			Organization: []string{"Golem"},
		},
		NotBefore: time.SecondsToUTC(now - 300),
		NotAfter:  time.SecondsToUTC(now + 60*60*24*365), // valid for 1 year.
	
		SubjectKeyId: []byte{1, 2, 3, 4},
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
	}
	
	
	certbyte, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		log("Failed to create certificate: %s", err)
		return 
	}
	
	cert, err =  tls.X509KeyPair(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certbyte}), pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)}))
	if err != nil {
		log("Failed to X509KeyPair certificate: %s", err)
		return 
	}
	return
	
}

//this is the main function to setup a server as used by the master, usage is identical to 
//http.ListenAndServe but this relys on global useTls being set
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
