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

// Sets a global variable for verbosity in logs
// optional parameter:  default.verbose (defaults to true if not present or incorrectly set)
func GlobalVerbose(configFile ConfigurationFile) {
	v, err := configFile.GetBool("default", "verbose")
	if err != nil {
		warn("GlobalVerbose(): %v", err)
	}
	verbose=v
	log("GlobalVerbose(): verbose=[%v]", verbose)
}

// Sets global variable to enable TLS communications and other related variables (certificate path, organization)
// optional parameters:  default.certpath, default.organization, default.tls
func GlobalTls(configFile ConfigurationFile) {
	certificatepath, err := configFile.GetString("default", "certpath")
	if err != nil {
		warn("GlobalTls(): %v", err)
	} else {
		certpath = certificatepath
	}
	log("GlobalTls(): certpath=[%v]", certpath)

	certificateorg, err := configFile.GetString("default", "organization")
	if err != nil {
		warn("GlobalTls(): %v", err)
		if certificateorg, err = os.Hostname(); err != nil {
			warn("GlobalTls(): %v", err)
			certificateorg = "golem.googlecode.com"
		}
	}
	certorg = certificateorg
	log("GlobalTls(): certorg=[%v]", certorg)

	useTls, err := configFile.GetBool("default", "tls")
	if err != nil {
		warn("GlobalTls(): %v", err)
		useTls = true
	}
	log("GlobalTls(): TLS=[%v]", useTls)
}

// Sets global variable to configure buffersize for master submission channels (stdout, stderr)
// optional parameters:  master.buffersize
func GlobalBufferSize(configFile ConfigurationFile) {
	bufsize, err := configFile.GetInt("master", "buffersize")
	if err != nil {
		warn("GlobalTls(): %v", err)
	} else {
		iobuffersize = bufsize
	}

	log("GlobalBufferSize(): buffersize=[%v]", iobuffersize)
}
