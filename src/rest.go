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
