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
	"http"
	"strings"
	"goconf.googlecode.com/hg"
	"github.com/codeforsystemsbiology/httplib.go"
	"url"
)

type AddamaConnection struct {
	target         string
	connectionFile string
	serviceHost    string
	serviceName    string
	uri            string
	label          string
	apikey         string
}

func NewAddamaProxy(addamaConn AddamaConnection) *AddamaProxy {
	target := addamaConn.target
	connectionFile := addamaConn.connectionFile
	serviceHost := addamaConn.serviceHost
	serviceName := addamaConn.serviceName
	uri := addamaConn.uri
	label := addamaConn.label
	apikey := addamaConn.apikey

	registrar := NewRegistrar(connectionFile)

	serviceUri := "/addama/services/golem/" + serviceName
	service := fmt.Sprintf(" { uri: '%v', url: '%v', label: '%v' } ", serviceUri, serviceHost, label)
	header := registrar.Register("/addama/registry/services/"+serviceName, "service", service)
	registrykey := header.Get("x-addama-registry-key")
	logger.Printf("registrykey:%v", registrykey)

	mapping := fmt.Sprintf(" { uri: '%v', label: '%v', service: '%v' } ", uri, label, serviceUri)
	registrar.Register("/addama/registry/mappings"+uri, "mapping", mapping)

	targetUrl, _ := url.Parse(target)
	return &AddamaProxy{target: targetUrl, registrykey: registrykey, apikey: apikey, baseuri: uri}
}

func NewRegistrar(connectionFilePath string) *Registrar {
	connectionFile, _ := conf.ReadConfigFile(connectionFilePath)
	host, _ := connectionFile.GetString("Connection", "host")
	apikey, _ := connectionFile.GetString("Connection", "apikey")
	logger.Debug("NewRegistrar(%v):%v,%v", connectionFilePath, host, apikey)
	return &Registrar{host: "https://" + host, apikey: apikey}
}

type Registrar struct {
	host   string
	apikey string
}

func (this *Registrar) Register(uri string, registrationType string, registration string) http.Header {
	logger.Debug("Register(%v%v, %v, %v)", this.host, uri, registrationType, registration)

	requestBuilder := httplib.Post(this.host + uri)
	requestBuilder.Header("x-addama-apikey", this.apikey)
	requestBuilder.Param(registrationType, registration)
	resp, err := requestBuilder.AsResponse()
	if err != nil {
		panic(err)
	}

	if resp.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("unable to properly register:%v", resp))
	}

	logger.Debug("Register(%v%v): %d", this.host, uri, resp.StatusCode)

	return resp.Header
}

type AddamaProxy struct {
	target      *url.URL
	registrykey string
	apikey      string
	baseuri     string
}

func (this *AddamaProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger.Printf("AddamaProxy.ServeHTTP(%v %v)", r.Method, r.URL.Path)

	if this.registrykey != r.Header.Get("x-addama-registry-key") {
		http.Error(w, "Registry key does not match", http.StatusForbidden)
		return
	}

	uri := strings.Replace(r.URL.Path, this.baseuri, "", -1)
	preq, _ := http.NewRequest(r.Method, uri, r.Body)
	preq.Header.Set("x-golem-apikey", this.apikey)

	proxy := http.NewSingleHostReverseProxy(this.target)
	go proxy.ServeHTTP(w, preq)
}
