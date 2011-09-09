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
	"http"
	"url"
	"github.com/codeforsystemsbiology/rest.go"
	"github.com/codeforsystemsbiology/verboselogger.go"
)

var logger *log4go.VerboseLogger

//parse args and start as master, scribe, addama proxy or worker
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

	configFile := NewConfigurationFile(configurationFile)

    GlobalLogger(configFile)
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

// sets global logger based on verbosity level in configuration
// optional parameter:  default.verbose (defaults to true if not present or incorrectly set)
func GlobalLogger(configFile ConfigurationFile) {
    verbose, err := configFile.GetBool("default", "verbose")
    logger = log4go.NewVerboseLogger(verbose, nil, "")
    if err != nil {
        logger.Warn(err)
        verbose = true
    }
    logger.Printf("verbose set [%v]", verbose)
}

// starts master service based on the given configuration file
// required parameters:  default.hostname, default.password
// optional parameters:  master.buffersize
func StartMaster(configFile ConfigurationFile) {
	GlobalBufferSize(configFile)

	hostname := configFile.GetRequiredString("default", "hostname")
	password := configFile.GetRequiredString("default", "password")

	m := NewMaster()

	rest.Resource("jobs", MasterJobController{m, password})
	rest.Resource("nodes", MasterNodeController{m, password})
	ListenAndServeTLSorNot(hostname)
}

// starts scribe service based on the given configuration file
// required parameters:  default.hostname, default.password, scribe.target, mgodb.server, mgodb.store, mgodb.jobcollection, mgodb.taskcollection
func StartScribe(configFile ConfigurationFile) {
	hostname := configFile.GetRequiredString("default", "hostname")
	apikey := configFile.GetRequiredString("default", "password")
	target := configFile.GetRequiredString("scribe", "target")
	dbhost := configFile.GetRequiredString("mgodb", "server")
	dbstore := configFile.GetRequiredString("mgodb", "store")
	collectionJobs := configFile.GetRequiredString("mgodb", "jobcollection")
	collectionTasks := configFile.GetRequiredString("mgodb", "taskcollection")

	url, err := url.Parse(target)
	if err != nil {
		panic(err)
	}

	go LaunchScribe(&MongoJobStore{Host: dbhost, Store: dbstore, JobsCollection: collectionJobs, TasksCollection: collectionTasks}, target, apikey)

	rest.Resource("jobs", ScribeJobController{&MongoJobStore{Host: dbhost, Store: dbstore, JobsCollection: collectionJobs, TasksCollection: collectionTasks}, url, apikey})
	rest.Resource("nodes", ProxyNodeController{url, apikey})

	ListenAndServeTLSorNot(hostname)
}

// starts Addama proxy (http://addama.org) based on the given configuration file
// required parameters:  default.hostname, default.password, addama.target, addama.connectionFile, addama.host, addama.service, addama.uri, addama.label
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

	ListenAndServeTLSorNot(hostname)
}

// starts worker based on the given configuration file
// required parameters:  worker.masterhost
// optional parameters:  worker.processes
func StartWorker(configFile ConfigurationFile) {
	processes, err := configFile.GetInt("worker", "processes")
	if err != nil {
		logger.Warn(err)
		processes = 3
	}
	masterhost := configFile.GetRequiredString("worker", "masterhost")
	logger.Printf("StartWorker() [%v, %d]", masterhost, processes)
	RunNode(processes, masterhost)
}

// starts http handlers for HTML content based on the given configuration file
// optional parameters:  default.contentDirectory (location of html content to be served at https://example.com/ or https://example.com/html/index.html
func StartHtmlHandler(configFile ConfigurationFile) {
	if contentDir, _ := configFile.GetString("default", "contentDirectory"); contentDir != "" {
		logger.Printf("StartHtmlHandler(): serving HTML content from [%v]", contentDir)
		http.Handle("/html/", http.StripPrefix("/html/", http.FileServer(http.Dir(contentDir))))
		http.Handle("/", http.RedirectHandler("/html/index.html", http.StatusTemporaryRedirect))
	}
}
