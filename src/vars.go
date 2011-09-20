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
	"runtime"
	"goconf.googlecode.com/hg"
)

const (
	second = 1e9 // one second is 1e9 nanoseconds
	year   = 60 * 60 * 24 * 365
)

var iobuffersize = 1000
var conbuffersize = 10
var iomonitors = 2
var useTls bool = true
var certpath string = ""
var certorg string = "golem.googlecode.com"

// Sets global variable to enable TLS communications and other related variables (certificate path, organization)
// optional parameters:  default.certpath, default.organization, default.tls
func GlobalTls(configFile *conf.ConfigFile) {
	certificatepath, err := configFile.GetString("default", "certpath")
	if err != nil {
		logger.Warn(err)
	} else {
		certpath = certificatepath
	}
	logger.Printf("certpath=[%v]", certpath)

	certificateorg, err := configFile.GetString("default", "organization")
	if err != nil {
		logger.Warn(err)
		if certificateorg, err = os.Hostname(); err != nil {
			logger.Warn(err)
			certificateorg = "golem.googlecode.com"
		}
	}
	certorg = certificateorg
	logger.Printf("certorg=[%v]", certorg)

	useTlsl, err := configFile.GetBool("default", "tls")
	if err != nil {
		logger.Warn(err)
		useTls = true
	} else {
		useTls = useTlsl
	}
	logger.Printf("TLS=[%v]", useTls)
}

// Sets global variable to configure buffersize for master submission channels (stdout, stderr)
// optional parameters:  master.buffersize
func SubIOBufferSize(section string, configFile *conf.ConfigFile) {
	bufsize, err := configFile.GetInt(section, "subiobuffersize")
	if err != nil {
		logger.Warn(err)
	} else {
		iobuffersize = bufsize
	}

	logger.Printf("buffersize=[%v]", iobuffersize)
}

// Sets global variable to configure buffersize for channels wrapping connections between worker and master (stdout, stderr)
// optional parameters:  master.buffersize
func ConBufferSize(section string, config *conf.ConfigFile) {
	bufsize, err := config.GetInt(section, "conbuffersize")
	if err != nil {
		logger.Printf("conbuffersize not fount in %v", section)
	} else {
		conbuffersize = bufsize
		logger.Printf("conbuffersize=[%v]", conbuffersize)
	}
}

//get the number of IO monitors to run per node
func IOMOnitors(config *conf.ConfigFile) {
	iomons, err := config.GetInt("master", "iomonitors")
	if err != nil {
		logger.Printf("iomonitors not fount in master")
	} else {
		if iomons > 0 {
			iomonitors = iomons
		}

	}
	logger.Printf("iomonitors=[%v]", iomonitors)
}

//get the number of proccessors to use for golem itself
func GoMaxProc(section string, config *conf.ConfigFile) {
	gomaxproc, err := config.GetInt(section, "gomaxproc")
	if err != nil {
		logger.Printf("gomaxproc not fount in %v", section)
	} else {
		runtime.GOMAXPROCS(gomaxproc)
		logger.Printf("gomaxproc=[%v]", gomaxproc)
	}
}

//conveniance function for requiring a string in ghte config file
func GetRequiredString(config *conf.ConfigFile, section string, key string) (value string) {
	value, err := config.GetString(section, key)
	if err != nil {

		logger.Fatalf("[CONFIG] %v is required: [section=%v]", key, section)

	}
	return
}
