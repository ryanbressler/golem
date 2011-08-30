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
	"os"
)

const (
	second = 1e9 // one second is 1e9 nanoseconds
	year   = 60 * 60 * 24 * 365
)

var verbose = false
var iobuffersize = 1000
var useTls bool = true
var certpath string = ""
var certorg string = "golem.googlecode.com"

func GlobalVerbose(configFile ConfigurationFile) {
	verbose, err := configFile.GetBool("default", "verbose")
	if err != nil {
		warn("%v", err)
	}
	info("verbose=[%v]", verbose)
}

func GlobalTls(configFile ConfigurationFile) {
	certificatepath, err := configFile.GetString("default", "certpath")
	if err != nil {
		warn("%v", err)
	} else {
		certpath = certificatepath
	}
	info("certpath=[%v]", certpath)

	certificateorg, err := configFile.GetString("default", "organization")
	if err != nil {
		warn("%v", err)
		if certificateorg, err = os.Hostname(); err != nil {
			warn("%v", err)
			certificateorg = "golem.googlecode.com"
		}
	}
	certorg = certificateorg
	info("certorg=[%v]", certorg)

	useTls, err := configFile.GetBool("default", "tls")
	if err != nil {
		warn("%v", err)
		useTls = true
	}
	info("TLS=[%v]", useTls)
}

func GlobalBufferSize(configFile ConfigurationFile) {
	bufsize, err := configFile.GetInt("master", "buffersize")
	if err != nil {
		warn("%v", err)
	} else {
		iobuffersize = bufsize
	}

	info("buffersize=[%v]", iobuffersize)
}
