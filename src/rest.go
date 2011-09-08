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
	"json"
	"os"
)

func GetHeader(r *http.Request, headerName string, defaultValue string) string {
	logger.Debug("GetHeader(%v,%v,%v)", r.URL.Path, headerName, defaultValue)
	val := r.Header.Get(headerName)
	if val != "" {
		return val
	}
	return defaultValue
}

func LoadTasksFromJson(r *http.Request, tasks *[]Task) (err os.Error) {
	logger.Debug("LoadTasksFromJson(%v)", r.URL.Path)

	mpreader, err := r.MultipartReader()
	if err != nil {
		logger.Warn(err)
		return
	}

	frm, err := mpreader.ReadForm(10000)
	if err != nil {
		logger.Warn(err)
		return
	}

	jsonfile, err := frm.File["jsonfile"][0].Open()
	if err != nil {
		logger.Warn(err)
		return
	}
	defer jsonfile.Close()

	err = json.NewDecoder(jsonfile).Decode(&tasks)
	if err != nil {
		logger.Warn(err)
	}
	return
}

func CheckApiKey(apikey string, r *http.Request) bool {
	logger.Debug("CheckApiKey(%v %v)", r.Method, r.URL.Path)
	if apikey != "" {
		headerkey := r.Header.Get("x-golem-apikey")
		if headerkey == "" {
			return false
		}
		return headerkey == apikey
	}
	return true
}
