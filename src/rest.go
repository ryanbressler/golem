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
	"strings"
	"http"
	"json"
	"mime/multipart"
	"os"
)

func splitRestUrl(path string) []string {
	spliturl := strings.Split(path, "/", -1)
	pathParts := make([]string, 0, 2)
	for _, part := range spliturl {
		if part != "" {
			pathParts = append(pathParts, part)
		}
	}
	return pathParts
}

//parser for rest jobs request
func parseJobUri(path string) (jobid string, verb string) {
	pathParts := splitRestUrl(path)
	nparts := len(pathParts)
	jobid = ""
	verb = ""
	switch {
	case nparts == 2:
		jobid = pathParts[1]
	case nparts == 3:
		jobid = pathParts[1]
		verb = pathParts[2]
	}
	vlog("Parsed job request id:\"%v\" verb:\"%v\"", jobid, verb)
	return
}

func getHeader(r *http.Request, headerName string, defaultValue string) string {
	val := r.Header.Get(headerName)
	if val != "" {
		return val
	}
	return defaultValue
}

func getMultipartForm(r *http.Request) (frm *multipart.Form, err os.Error) {
	mpreader, err := r.MultipartReader()
	if err != nil {
		return
	}

	frm, err = mpreader.ReadForm(10000)
	return
}

func loadJson(r *http.Request, tasks *[]Task) (err os.Error) {
	vlog("loadJson")

	vlog("loadJson: multipart")
	frm, err := getMultipartForm(r)
	if err != nil {
		vlog("loadJson:%v", err)
		return
	}

	vlog("loadJson: open jsonfile")
	jsonfile, err := frm.File["jsonfile"][0].Open()
	if err != nil {
		vlog("loadJson:%v", err)
		return
	}
	defer jsonfile.Close()

	vlog("loadJson: decoding jsonfile")
	err = json.NewDecoder(jsonfile).Decode(&tasks)
	return
}

func CheckPassword(hashedpw string, r *http.Request) bool {
	if hashedpw != "" {
		pw := hashPw(r.Header.Get("Password"))
		log("Verifying password.")
		return hashedpw == pw
	}
	return true
}

// TODO : Deal with URI
func WriteItemAsJson(baseUri string, itemId string, r Retriever, w http.ResponseWriter) {
	item, err := r.Retrieve(itemId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	val, err := json.Marshal(item)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Write(val)
}

// TODO : Deal with URI
func WriteItemsAsJson(baseUri string, r Retriever, w http.ResponseWriter) {
	items, err := r.RetrieveAll()
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	itemsHandle := NewItemsHandle(items)

	val, err := json.Marshal(itemsHandle)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Write(val)
}

