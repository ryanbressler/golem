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
	"io"
	"json"
	"strconv"
	"strings"
	"os"
)

type ProxyJobController struct {
	proxy *http.ReverseProxy
}

func NewProxyJobController() ProxyJobController {
	target, err := ConfigFile.GetString("scribe", "target")
	if err != nil {
		panic(err)
	}

	url, err := http.ParseRequestURL(target)
	if err != nil {
		panic(err)
	}

	return ProxyJobController{http.NewSingleHostReverseProxy(url)}
}

func (c ProxyJobController) RetrieveAll() (items []interface{}, err os.Error) {
	val, err := doProxy("GET", "/jobs/", c.proxy, nil)
	if err != nil {
		return
	}

	js := []JobSubmission{}
	if err = json.Unmarshal(val, js); err != nil {
		return
	}

	for _, s := range js {
		items = append(items, s)
	}

	return
}
func (c ProxyJobController) Retrieve(jobId string) (item interface{}, err os.Error) {
	val, err := doProxy("GET", "/jobs/"+jobId, c.proxy, nil)
	if err != nil {
		return
	}
	item = Submission{}
	err = json.Unmarshal(val, item)
	return
}
func (c ProxyJobController) NewJob(r *http.Request) (jobId string, err os.Error) {
	val, err := doProxy("POST", "/jobs/"+jobId, c.proxy, r.Body)
	if err != nil {
		return
	}

	jh := JobHandle{}
	err = json.Unmarshal(val, jh)
	if err != nil {
		return
	}

	jobId = jh.JobId
	return
}
func (c ProxyJobController) Stop(jobId string) os.Error {
	_, err := doProxy("POST", "/jobs/"+jobId+"/stop", c.proxy, nil)
	return err
}
func (c ProxyJobController) Kill(jobId string) os.Error {
	_, err := doProxy("POST", "/jobs/"+jobId+"/kill", c.proxy, nil)
	return err
}

type ProxyNodeController struct {
	proxy *http.ReverseProxy
}

func NewProxyNodeController() ProxyNodeController {
	target, err := ConfigFile.GetString("scribe", "target")
	if err != nil {
		panic(err)
	}

	url, err := http.ParseRequestURL(target)
	if err != nil {
		panic(err)
	}

	return ProxyNodeController{http.NewSingleHostReverseProxy(url)}
}

func (c ProxyNodeController) RetrieveAll() (items []interface{}, err os.Error) {
	val, err := doProxy("GET", "/nodes/", c.proxy, nil)
	if err != nil {
		return
	}
	workerNodes := []WorkerNode{}
	if err = json.Unmarshal(val, workerNodes); err != nil {
		return
	}

	for _, item := range workerNodes {
		items = append(items, item)
	}
	return
}
func (c ProxyNodeController) Retrieve(nodeId string) (item interface{}, err os.Error) {
	val, err := doProxy("GET", "/nodes/"+nodeId, c.proxy, nil)
	if err != nil {
		return
	}
	item = WorkerNode{}
	err = json.Unmarshal(val, item)
	return
}
func (c ProxyNodeController) RestartAll() os.Error {
	_, err := doProxy("POST", "/nodes/restart", c.proxy, nil)
	return err
}
func (c ProxyNodeController) Resize(nodeId string, numberOfThreads int) os.Error {
	_, err := doProxy("POST", "/nodes/"+nodeId+"/resize/"+strconv.Itoa(numberOfThreads), c.proxy, nil)
	return err
}
func (c ProxyNodeController) KillAll() os.Error {
	_, err := doProxy("POST", "/nodes/kill", c.proxy, nil)
	return err
}

// proxy support
type JsonResponseWriter struct {
	Bytes []byte
}

func (w JsonResponseWriter) Header() http.Header {
	return nil
}
func (w JsonResponseWriter) Write(b []byte) (int, os.Error) {
	w.Bytes = b
	return 0, os.EOF
}
func (w JsonResponseWriter) WriteHeader(int) {
	return
}

func doProxy(method string, uri string, proxy *http.ReverseProxy, reader io.Reader) (val []byte, err os.Error) {
	if reader == nil {
		reader = strings.NewReader("")
	}

	r, err := http.NewRequest(method, uri, reader)
	if err != nil {
		return
	}

	rw := JsonResponseWriter{}
	proxy.ServeHTTP(rw, r)
	val = rw.Bytes
	return
}
