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
	"http"
	"strings"
	"url"
)

type ProxyNodeController struct {
	target *url.URL
	apikey string
}

// GET /nodes
func (this ProxyNodeController) Index(rw http.ResponseWriter) {
	preq, err := http.NewRequest("GET", "/nodes/", strings.NewReader(""))
	if err != nil {
		http.Error(rw, err.String(), http.StatusBadRequest)
		return
	}

	logger.Debug("proxying /nodes to %v", this.target)
	proxy := http.NewSingleHostReverseProxy(this.target)
	proxy.ServeHTTP(rw, preq)
}
// GET /nodes/id
func (this ProxyNodeController) Find(rw http.ResponseWriter, nodeId string) {
	preq, err := http.NewRequest("GET", "/nodes/"+nodeId, strings.NewReader(""))
	if err != nil {
		http.Error(rw, err.String(), http.StatusBadRequest)
		return
	}

	logger.Debug("proxying /nodes/%v to %v", nodeId, this.target)
	proxy := http.NewSingleHostReverseProxy(this.target)
	proxy.ServeHTTP(rw, preq)
}
// POST /nodes/restart or POST /nodes/die or POST /nodes/id/resize/new-size
func (this ProxyNodeController) Act(rw http.ResponseWriter, parts []string, r *http.Request) {
	if CheckApiKey(this.apikey, r) == false {
		http.Error(rw, "api key required in header", http.StatusForbidden)
		return
	}

	if parts[0] == "restart" {
		go RestartIn(10)
	}
	if parts[0] == "die" {
		go DieIn(10)
	}

	preq, err := http.NewRequest(r.Method, r.URL.Path, r.Body)
	if err != nil {
		http.Error(rw, err.String(), http.StatusBadRequest)
		return
	}

	preq.Header.Set("x-golem-apikey", this.apikey)

	logger.Debug("proxying %v to %v", r.URL.Path, this.target)
	proxy := http.NewSingleHostReverseProxy(this.target)
	proxy.ServeHTTP(rw, preq)
}
