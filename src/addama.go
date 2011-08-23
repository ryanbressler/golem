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
)

func NewAddamaProxy(apikey string) *AddamaProxy {
	target := ConfigFile.GetRequiredString("addama", "target")
	connectionFile := ConfigFile.GetRequiredString("addama", "connectionFile")
	serviceHost := ConfigFile.GetRequiredString("addama", "host")
	serviceName := ConfigFile.GetRequiredString("addama", "service")
	uri := ConfigFile.GetRequiredString("addama", "uri")
	label := ConfigFile.GetRequiredString("addama", "label")

	registrar := NewRegistrar(connectionFile)

	serviceUri := "/addama/services/golem/" + serviceName
	service := fmt.Sprintf(" { uri: '%v', url: '%v', label: '%v' } ", serviceUri, serviceHost, label)
	header := registrar.Register("/addama/registry/services/"+serviceName, "service", service)
	registrykey := header.Get("x-addama-registry-key")
	log("registrykey:%v", registrykey)

	mapping := fmt.Sprintf(" { uri: '%v', label: '%v', service: '%v' } ", uri, label, serviceUri)
	registrar.Register("/addama/registry/mappings"+uri, "mapping", mapping)

	url, _ := http.ParseRequestURL(target)
	proxy := http.NewSingleHostReverseProxy(url)
	return &AddamaProxy{proxy: proxy, registrykey: registrykey, apikey: apikey, baseuri: uri}
}

func NewRegistrar(connectionFilePath string) *Registrar {
	c, _ := conf.ReadConfigFile(connectionFilePath)
	connectionFile := ConfigurationFile{c}
	host, _ := connectionFile.GetString("Connection", "host")
	apikey, _ := connectionFile.GetString("Connection", "apikey")
	vlog("NewRegistrar(%v):%v,%v", connectionFilePath, host, apikey)
	return &Registrar{host: "https://" + host, apikey: apikey}
}

type Registrar struct {
	host   string
	apikey string
}

func (this *Registrar) Register(uri string, registrationType string, registration string) http.Header {
	vlog("Register(%v%v, %v, %v)", this.host, uri, registrationType, registration)

	requestBuilder := Post(this.host + uri)
	requestBuilder.Header("x-addama-apikey", this.apikey)
	requestBuilder.Param(registrationType, registration)
	resp, err := requestBuilder.getResponse()
	if err != nil {
		panic(err)
	}

	if resp.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("unable to properly register:%v", resp))
	}

	vlog("Register(%v%v): %d", this.host, uri, resp.StatusCode)

	return resp.Header
}

type AddamaProxy struct {
	proxy       *http.ReverseProxy
	registrykey string
	apikey      string
	baseuri     string
}

func (this *AddamaProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log("AddamaProxy.ServeHTTP(%v %v)", r.Method, r.URL.Path)

	if this.registrykey != r.Header.Get("x-addama-registry-key") {
		http.Error(w, "Registry key does not match", http.StatusForbidden)
		return
	}

	uri := strings.Replace(r.URL.Path, this.baseuri, "", -1)
	preq, _ := http.NewRequest(r.Method, uri, r.Body)
	preq.Header.Set("x-golem-apikey", this.apikey)

	go this.proxy.ServeHTTP(w, preq)
}
