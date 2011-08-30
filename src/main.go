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
	"github.com/codeforsystemsbiology/rest.go"
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

	configFile := NewConfigFile(configurationFile)

	GlobalVerbose(configFile)
	GlobalTls(configFile)
	StartHtmlHandler(configFile)

	if isMaster {
		StartMaster(configFile)
	} else if isScribe {
		StartScribe(configFile)
	} else if isAddama {
		StartAddama(configFile)
	} else {
		StartWorker(configFile)
	}
}

func StartMaster(configFile ConfigurationFile) {
	GlobalBufferSize(configFile)

	hostname := configFile.GetRequiredString("default", "hostname")
	password := configFile.GetRequiredString("default", "password")

	m := NewMaster()

	rest.Resource("jobs", MasterJobController{m, password})
	rest.Resource("nodes", MasterNodeController{m, password})
	ListenAndServeTLSorNot(hostname, nil)
}

func StartScribe(configFile ConfigurationFile) {
	hostname := configFile.GetRequiredString("default", "hostname")
	apikey := configFile.GetRequiredString("default", "password")
	target := configFile.GetRequiredString("scribe", "target")
	dbhost := configFile.GetRequiredString("mgodb", "server")
	dbstore := configFile.GetRequiredString("mgodb", "store")
	collectionJobs := configFile.GetRequiredString("mgodb", "jobcollection")
	collectionTasks := configFile.GetRequiredString("mgodb", "taskcollection")

	url, err := http.ParseRequestURL(target)
	if err != nil {
		panic(err)
	}

	go LaunchScribe(NewMongoJobStore(dbhost, dbstore, collectionJobs, collectionTasks), target, apikey)

	rest.Resource("jobs", ScribeJobController{NewMongoJobStore(dbhost, dbstore, collectionJobs, collectionTasks), url, apikey})
	rest.Resource("nodes", ProxyNodeController{url, apikey})

	ListenAndServeTLSorNot(hostname, nil)
}

func StartAddama(configFile ConfigurationFile) {
	hostname := configFile.GetRequiredString("default", "hostname")

	addamaConn := AddamaConnection{
		target:         configFile.GetRequiredString("addama", "target"),
		connectionFile: configFile.GetRequiredString("addama", "connectionFile"),
		serviceHost:    configFile.GetRequiredString("addama", "host"),
		serviceName:    configFile.GetRequiredString("addama", "service"),
		uri:            configFile.GetRequiredString("addama", "uri"),
		label:          configFile.GetRequiredString("addama", "label"),
		apikey:         configFile.GetRequiredString("default", "password")}

	http.Handle("/", NewAddamaProxy(addamaConn))

	ListenAndServeTLSorNot(hostname, nil)
}
func StartWorker(configFile ConfigurationFile) {
	processes, err := configFile.GetInt("worker", "processes")
	if err != nil {
		log("worker proceses error, setting to 3: %v", err)
		processes = 3
	}

	masterhost := configFile.GetRequiredString("worker", "masterhost")

	RunNode(processes, masterhost)
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

func StartHtmlHandler(configFile ConfigurationFile) {
	if contentDir, _ := configFile.GetString("default", "contentDirectory"); contentDir != "" {
		info("serving HTML content from [%v]", contentDir)
		http.Handle("/html/", http.StripPrefix("/html/", http.FileServer(http.Dir(contentDir))))
		http.Handle("/", http.RedirectHandler("/html/index.html", http.StatusTemporaryRedirect))
	}
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
