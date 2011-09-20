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
	"goconf.googlecode.com/hg"
	"launchpad.net/mgo"
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

	configFile, err := conf.ReadConfigFile(configurationFile)
	if err != nil {
		panic(err)
	}

	GlobalLogger(configFile)
	GlobalTls(configFile)
	SubIOBufferSize("default", configFile)
	GoMaxProc("default", configFile)
	ConBufferSize("default", configFile)
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
func GlobalLogger(configFile *conf.ConfigFile) {
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
func StartMaster(configFile *conf.ConfigFile) {
	SubIOBufferSize("master", configFile)
	GoMaxProc("master", configFile)
	ConBufferSize("master", configFile)
	IOMOnitors(configFile)

	hostname := GetRequiredString(configFile, "default", "hostname")
	password := GetRequiredString(configFile, "default", "password")

	m := NewMaster()

	rest.Resource("jobs", MasterJobController{m, password})
	rest.Resource("nodes", MasterNodeController{m, password})

	rest.ResourceContentType("jobs", "application/json")
	rest.ResourceContentType("nodes", "application/json")

	ListenAndServeTLSorNot(hostname)
}

// starts scribe service based on the given configuration file
// required parameters:  default.hostname, default.password, scribe.target, mgodb.server, mgodb.store, mgodb.jobcollection, mgodb.taskcollection
func StartScribe(configFile *conf.ConfigFile) {
	MongoLogger(configFile)

	GoMaxProc("scribe", configFile)

	hostname := GetRequiredString(configFile, "default", "hostname")
	apikey := GetRequiredString(configFile, "default", "password")
	target := GetRequiredString(configFile, "scribe", "target")
	dbhost := GetRequiredString(configFile, "mgodb", "server")
	dbstore := GetRequiredString(configFile, "mgodb", "store")

	url, err := url.Parse(target)
	if err != nil {
		panic(err)
	}

	go LaunchScribe(&MongoJobStore{Host: dbhost, Store: dbstore}, target, apikey)

	rest.Resource("jobs", ScribeJobController{&MongoJobStore{Host: dbhost, Store: dbstore}, url, apikey})
	rest.ResourceContentType("jobs", "application/json")

	rest.Resource("nodes", ProxyNodeController{url, apikey})
	rest.ResourceContentType("nodes", "application/json")

	rest.Resource("cluster", ScribeClusterController{&MongoJobStore{Host: dbhost, Store: dbstore}, url})
	rest.ResourceContentType("cluster", "application/json")

	var numberOfSeconds int = 10
	if pollingSecs, oops := configFile.GetInt("scribe", "clusterStatsPollingInSeconds"); oops == nil {
		numberOfSeconds = pollingSecs
	}

	logger.Printf("polling for cluster stats every %d secs", numberOfSeconds)
	go MonitorClusterStats(&MongoJobStore{Host: dbhost, Store: dbstore}, target, int64(numberOfSeconds))

	ListenAndServeTLSorNot(hostname)
}

func MongoLogger(configFile *conf.ConfigFile) {
	verbose, err := configFile.GetBool("mgodb", "verbose")
	if err != nil {
		logger.Warn(err)
		verbose = false
	}

	logger.Printf("mongo logger verbose set [%v]", verbose)
	if verbose {
		mgo.SetLogger(logger)
		mgo.SetDebug(verbose)
	}
}

// starts Addama proxy (http://addama.org) based on the given configuration file
// required parameters:  default.hostname, default.password, addama.target, addama.connectionFile, addama.host, addama.service, addama.uri, addama.label
func StartAddama(configFile *conf.ConfigFile) {
	hostname := GetRequiredString(configFile, "default", "hostname")

	addamaConn := AddamaConnection{
		target:         GetRequiredString(configFile, "addama", "target"),
		connectionFile: GetRequiredString(configFile, "addama", "connectionFile"),
		serviceHost:    GetRequiredString(configFile, "addama", "host"),
		serviceName:    GetRequiredString(configFile, "addama", "service"),
		uri:            GetRequiredString(configFile, "addama", "uri"),
		label:          GetRequiredString(configFile, "addama", "label"),
		apikey:         GetRequiredString(configFile, "default", "password")}

	http.Handle("/", NewAddamaProxy(addamaConn))

	ListenAndServeTLSorNot(hostname)
}

// starts worker based on the given configuration file
// required parameters:  worker.masterhost
// optional parameters:  worker.processes
func StartWorker(configFile *conf.ConfigFile) {

	GoMaxProc("worker", configFile)
	ConBufferSize("worker", configFile)
	processes, err := configFile.GetInt("worker", "processes")
	if err != nil {
		logger.Warn(err)
		processes = 3
	}
	masterhost := GetRequiredString(configFile, "worker", "masterhost")
	logger.Printf("StartWorker() [%v, %d]", masterhost, processes)
	RunNode(processes, masterhost)
}

// starts http handlers for HTML content based on the given configuration file
// optional parameters:  default.contentDirectory (location of html content to be served at https://example.com/ or https://example.com/html/index.html
func StartHtmlHandler(configFile *conf.ConfigFile) {
	if contentDir, _ := configFile.GetString("default", "contentDirectory"); contentDir != "" {
		logger.Printf("StartHtmlHandler(): serving HTML content from [%v]", contentDir)
		http.Handle("/html/", http.StripPrefix("/html/", http.FileServer(http.Dir(contentDir))))
		http.Handle("/", http.RedirectHandler("/html/index.html", http.StatusTemporaryRedirect))
	}
}
