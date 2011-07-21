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
	"fmt"
	"goconf.googlecode.com/hg"
)

//parse args and start as master, scribe or worker
func main() {
	var configurationFile string
	var isMaster bool
	var isScribe bool

	flag.BoolVar(&isMaster, "m", false, "Start as master node.")
	flag.BoolVar(&isScribe, "s", false, "Start as scribe node.")
	flag.StringVar(&configurationFile, "config", "golem.config", "A configuration file for golem services")
	flag.Parse()

	ConfigFile = NewConfigFile(configurationFile)

	setVerbose()
	setTls()

	if isMaster {
		m := NewMaster()
		HandleRestJson(MasterJobController{master: m}, MasterNodeController{master: m})
	} else if isScribe {
		mdb := NewMongoJobStore()
		s := NewScribe(mdb)
		HandleRestJson(ScribeJobController{s, mdb, NewProxyJobController()}, NewProxyNodeController())
	} else {
		processes, masterhost := getWorkerProcesses()
		RunNode(processes, masterhost)
	}
}

func NewConfigFile(filepath string) *conf.ConfigFile {
	if filepath != "" {
		c, err := conf.ReadConfigFile(filepath)
		if err != nil {
			panic(err)
		}
		return c
	}
	panic(fmt.Sprintf("configuration file not found [%v]", filepath))
}

func setVerbose() {
	verbose, _ = ConfigFile.GetBool("default", "verbose")
	if verbose {
		log("running in verbose mode")
	}
}

func setTls() {
	useTls, err := ConfigFile.GetBool("default", "tls")
	if err != nil {
		log("useTls error, setting to 'true': %v", err)
		useTls = true
	}
	log("secure mode enabled [%v]", useTls)
}

func getWorkerProcesses() (processes int, masterhost string) {
	processes, err := ConfigFile.GetInt("worker", "processes")
	if err != nil {
		log("worker proceses error, setting to 3: %v", err)
		processes = 3
	}

	masterhost, err = ConfigFile.GetString("worker", "masterhost")
	if err != nil {
		panic(err)
	}

	return
}
