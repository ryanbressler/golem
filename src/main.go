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
	"flag"
)


//////////////////////////////////////////////
//main method
//parse args and start as master, node or scribe
func main() {
	var atOnce int
	var unsecure bool
	var configFile string

	flag.BoolVar(&isMaster, "m", false, "Start as master node.")
	flag.BoolVar(&isScribe, "s", false, "Start as scribe node.")
	flag.IntVar(&atOnce, "n", 3, "For client nodes, the number of procceses to allow at once.")
	flag.StringVar(&certpath, "certpath", "", "The path that contains certificate.pem and key.pem to use for tls connections.")
	flag.BoolVar(&unsecure, "unsecure", false, "Don't use tls security.")
	flag.BoolVar(&verbose, "v", false, "Use verbose logging.")
	flag.IntVar(&iobuffersize, "iobuffer", 1000, "The size of the (per submission) buffers for standard out and standard error from client nodes.")
    flag.StringVar(&configFile, "config", "golem.config", "A configuration file for golem services")
	flag.Parse()

	if unsecure {
		useTls = false
	}

    configuration = NewConfiguration(configFile)

	if isMaster {
		m := NewMaster()
		NewRestOnJob(MasterJobController{master: m}, MasterNodeController{master: m})
	} else if isScribe {
		s := NewScribe(DoNothingJobStore{})
		NewRestOnJob(ScribeJobController{ s, NewProxyJobController() }, NewProxyNodeController())
	} else {
		RunNode(atOnce, configuration.GetString("worker", "masterhost"))
	}
}
