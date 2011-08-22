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
)

type ProxyNodeController struct {
	proxy  *http.ReverseProxy
	apikey string
}

// GET /nodes
func (this ProxyNodeController) Index(rw http.ResponseWriter) {
	preq, _ := http.NewRequest("GET", "/nodes", nil)
	go this.proxy.ServeHTTP(rw, preq)
}
// GET /nodes/id
func (this ProxyNodeController) Find(rw http.ResponseWriter, nodeId string) {
	preq, _ := http.NewRequest("GET", "/nodes/"+nodeId, nil)
	go this.proxy.ServeHTTP(rw, preq)
}
// POST /nodes/restart or POST /nodes/die or POST /nodes/id/resize/new-size
func (this ProxyNodeController) Act(rw http.ResponseWriter, parts []string, r *http.Request) {
	if CheckApiKey(this.apikey, r) == false {
		http.Error(rw, "api key required in header", http.StatusForbidden)
		return
	}

	preq, _ := http.NewRequest(r.Method, r.URL.Path, r.Body)
	preq.Header.Set("x-golem-apikey", this.apikey)
	go this.proxy.ServeHTTP(rw, preq)
}
