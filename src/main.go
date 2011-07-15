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
	var configFile string

	flag.BoolVar(&isMaster, "m", false, "Start as master node.")
	flag.BoolVar(&isScribe, "s", false, "Start as scribe node.")
	flag.StringVar(&certpath, "certpath", "", "The path that contains certificate.pem and key.pem to use for tls connections.")
	flag.IntVar(&iobuffersize, "iobuffer", 1000, "The size of the (per submission) buffers for standard out and standard error from client nodes.")
    flag.StringVar(&configFile, "config", "golem.config", "A configuration file for golem services")
	flag.Parse()

    configuration = NewConfiguration(configFile)
    setVerbose()
    setTls()

	if isMaster {
		m := NewMaster()
		NewRestOnJob(MasterJobController{master: m}, MasterNodeController{master: m})
	} else if isScribe {
		s := NewScribe(DoNothingJobStore{})
		NewRestOnJob(ScribeJobController{ s, NewProxyJobController() }, NewProxyNodeController())
	} else {
        atOnce := getWorkerProcesses()
		RunNode(atOnce, configuration.GetString("worker", "masterhost"))
	}
}

func setVerbose() {
    verbose, _ := configuration.ConfigFile.GetBool("default", "verbose")
    if verbose { log("running in verbose mode") }
}

func setTls() {
    useTls, err := configuration.ConfigFile.GetBool("default", "tls")
    if err != nil {
        log("useTls error, setting to 'true': %v", err)
        useTls = true
    }
    log("secure mode enabled [%v]", useTls)
}

func getWorkerProcesses() int {
    atOnce, err := configuration.ConfigFile.GetInt("worker", "processes")
    if err != nil {
        log("worker proceses error, setting to 3: %v", err)
        atOnce = 3
    }
    return atOnce
}