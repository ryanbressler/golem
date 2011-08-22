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
	"http"
	"goconf.googlecode.com/hg"
	"github.com/hrovira/rest.go"
)

//parse args and start as master, scribe or worker
func main() {
	var configurationFile string
	var isMaster bool
	var isScribe bool
	var isAddama bool

	flag.BoolVar(&isMaster, "m", false, "Start as master node.")
	flag.BoolVar(&isScribe, "s", false, "Start as scribe node.")
	flag.BoolVar(&isAddama, "a", false, "Start as addama node.")
	flag.StringVar(&configurationFile, "config", "golem.config", "A configuration file for golem services")
	flag.Parse()

	ConfigFile = NewConfigFile(configurationFile)

	setVerbose()
	setTls()

	if isMaster {
        hostname := ConfigFile.GetRequiredString("default", "hostname")
        password := ConfigFile.GetRequiredString("default", "password")

		setBufferSize()
		m := NewMaster()

		rest.Resource("jobs", MasterJobController{m, password})
		rest.Resource("nodes", MasterNodeController{m, password})
		ListenAndServeTLSorNot(hostname, nil);
	} else if isScribe {
        hostname := ConfigFile.GetRequiredString("default", "hostname")
        password := ConfigFile.GetRequiredString("default", "password")

        url, err := http.ParseRequestURL(ConfigFile.GetRequiredString("scribe", "target"))
        if err != nil {
            panic(err.String())
        }

        proxy := http.NewSingleHostReverseProxy(url)

		mdb := NewMongoJobStore()
		go LaunchScribe(mdb)

        rest.Resource("jobs", ScribeJobController{mdb, proxy, password})
		rest.Resource("nodes", ProxyNodeController{proxy, password})
		ListenAndServeTLSorNot(hostname, nil);
	} else if isAddama {
		HandleAddamaCalls()
	} else {
		processes, masterhost := getWorkerProcesses()
		RunNode(processes, masterhost)
	}
}

func NewConfigFile(filepath string) ConfigurationFile {
	if filepath != "" {
		c, err := conf.ReadConfigFile(filepath)
		if err != nil {
			panic(err)
		}
		return ConfigurationFile{c}
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

func setBufferSize() {
	bufsize, err := ConfigFile.GetInt("master", "buffersize")
	if err != nil {
		vlog("defaulting buffer to 1000:%v", err)
		return
	}
	iobuffersize = bufsize
}

func getWorkerProcesses() (processes int, masterhost string) {
	processes, err := ConfigFile.GetInt("worker", "processes")
	if err != nil {
		log("worker proceses error, setting to 3: %v", err)
		processes = 3
	}
	masterhost = ConfigFile.GetRequiredString("worker", "masterhost")
	return
}

type ConfigurationFile struct {
	*conf.ConfigFile
}

func (this *ConfigurationFile) GetRequiredString(section string, key string) (value string) {
	value, err := this.GetString(section, key)
	if err != nil {
		panic(err)
	}
	return
}
