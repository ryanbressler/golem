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
	"io"
	"json"
	"strconv"
	"os"
)

type ProxyJobController struct {
	proxy  *http.ReverseProxy
	apikey string
}

func NewProxyJobController() ProxyJobController {
	target := ConfigFile.GetRequiredString("scribe", "target")
	apikey := ConfigFile.GetRequiredString("default", "password")
	url, err := http.ParseRequestURL(target)
	if err != nil {
		panic(err)
	}

	return ProxyJobController{http.NewSingleHostReverseProxy(url), apikey}
}

func (c ProxyJobController) RetrieveAll() (items []interface{}, err os.Error) {
	decoder, err := Proxy("GET", "/jobs/", c.apikey, c.proxy, nil)
	if err != nil {
		return
	}

	js := JobDetailsList{}
	if err = decoder.Decode(&js); err != nil {
		return
	}

	for _, s := range js.Items {
		items = append(items, s)
	}

	return
}
func (c ProxyJobController) Retrieve(jobId string) (item interface{}, err os.Error) {
	decoder, err := Proxy("GET", "/jobs/"+jobId, c.apikey, c.proxy, nil)
	if err != nil {
		return
	}
	item = JobDetails{}
	err = decoder.Decode(&item)
	return
}
func (c ProxyJobController) NewJob(r *http.Request) (jobId string, err os.Error) {
	decoder, err := Proxy("POST", "/jobs/"+jobId, c.apikey, c.proxy, r.Body)
	if err != nil {
		return
	}

	jd := JobDetails{}
	if err = decoder.Decode(&jd); err != nil {
		return
	}

	jobId = jd.JobId
	return
}
func (c ProxyJobController) Stop(jobId string) os.Error {
	_, err := Proxy("POST", "/jobs/"+jobId+"/stop", c.apikey, c.proxy, nil)
	return err
}
func (c ProxyJobController) Kill(jobId string) os.Error {
	_, err := Proxy("POST", "/jobs/"+jobId+"/kill", c.apikey, c.proxy, nil)
	return err
}

type ProxyNodeController struct {
	proxy  *http.ReverseProxy
	apikey string
}

func NewProxyNodeController() ProxyNodeController {
	target := ConfigFile.GetRequiredString("scribe", "target")
	apikey := ConfigFile.GetRequiredString("default", "password")
	url, err := http.ParseRequestURL(target)
	if err != nil {
		panic(err)
	}

	return ProxyNodeController{http.NewSingleHostReverseProxy(url), apikey}
}

func (c ProxyNodeController) RetrieveAll() (items []interface{}, err os.Error) {
	decoder, err := Proxy("GET", "/nodes/", c.apikey, c.proxy, nil)
	if err != nil {
		return
	}

	lst := WorkerNodeList{}
	if err = decoder.Decode(&lst); err != nil {
		return
	}

	for _, item := range lst.Items {
		items = append(items, item)
	}
	return
}
func (c ProxyNodeController) Retrieve(nodeId string) (item interface{}, err os.Error) {
	decoder, err := Proxy("GET", "/nodes/"+nodeId, c.apikey, c.proxy, nil)
	if err != nil {
		return
	}
	item = WorkerNode{}
	err = decoder.Decode(&item)
	return
}
func (c ProxyNodeController) RestartAll() os.Error {
	_, err := Proxy("POST", "/nodes/restart", c.apikey, c.proxy, nil)
	return err
}
func (c ProxyNodeController) Resize(nodeId string, numberOfThreads int) (err os.Error) {
	_, err = Proxy("POST", "/nodes/"+nodeId+"/resize/"+strconv.Itoa(numberOfThreads), c.apikey, c.proxy, nil)
	return
}
func (c ProxyNodeController) KillAll() os.Error {
	_, err := Proxy("POST", "/nodes/kill", c.apikey, c.proxy, nil)
	return err
}

// proxy support
type PipeResponseWriter struct {
	Reader      io.Reader
	Writer      io.Writer
	StatusCode chan int
}
func NewPipeResponseWriter() PipeResponseWriter {
	preader, pwriter := io.Pipe()
	return PipeResponseWriter{ preader, pwriter, make(chan int, 1) }
}
func (w PipeResponseWriter) Header() http.Header {
	return http.Header{}
}
func (w PipeResponseWriter) Write(b []byte) (int, os.Error) {
    return w.Writer.Write(b)
}
func (w PipeResponseWriter) WriteHeader(i int) {
	w.StatusCode <- i
	return
}

func Proxy(method string, uri string, apikey string, proxy *http.ReverseProxy, body io.ReadCloser) (decoder *json.Decoder, err os.Error) {
	vlog("Proxy(%v %v)", method, uri)

	r, _ := http.NewRequest(method, uri, body)
	r.Header.Set("x-golem-apikey", apikey)

	responseWriter := NewPipeResponseWriter()
    go proxy.ServeHTTP(responseWriter, r)

    statusCode := <- responseWriter.StatusCode
	if http.StatusOK != statusCode {
		err = os.NewError(fmt.Sprintf("proxy [%v %v]: %d", method, uri, statusCode))
		return
	}

    decoder = json.NewDecoder(responseWriter.Reader)
	return
}
