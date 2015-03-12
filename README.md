# golem
Automatically exported from code.google.com/p/golem

This is an unmaintaned archive fork since googel code is shutting down. The wiki is duplicated bellow:

Golem (Go Launch on Every Machine) strives to be the simplest possible system for distributing computational analysis across processes and machines in a Unix/Linux based cluster. It uses web technologies, core features of the Go language and Unix enviroment to keep internal code simple and maintainable. It is being developed at the Institute for Systems Biology in Seattle, WA to provide a fast, light, simple and accessible tools for parallelizing algorithms used in computational biology, cancer research and large scale data analysis.

Golem is easier to setup and use than common MapReduce and job queuing systems and provided a different set of features. It focuses on providing a set of tools that allow quick parallelization of existing single threaded analysis that make use of *nix file handle based communication, command line tools, restful web services to allow easy integration with existing infrastructure. It also includes a web interface for viewing job and cluster status built on the above.

It is well suited for problems that involve many independent calculations but small amounts of data transfer, and for which the distributed sort-and-combine included in the MapReduce pattern is an unnecessary expense due either to the size of the result or the lack of an interesting ordering. In data analysis this includes algorithms that involve repeated independent permutations, partitions or stochastic simulations and any embarrassingly parallelizable analysis that can be run independently on separate portions of the data producing results that can be easily combined or don't need to combined.

Use cases at ISB include parallel random forest analysis (using rf-ace), motif searching across entire genomes (using Molotov), the identification of genetic structural variation at a population scale (using FastBreak) and exhaustive calculation of pairwise correlations between cancer features across patients.

A typical golem job is list of tasks which can be parameterized calls to any executable or script. The golem master balances these tasks across all available processors and, as a convenience, combines standard out and standard error from all jobs line by line with no guarantees of order.

Golem is in an alpha but usable state see BuildingAndUsing for more.

Golem (from merriam-webster):

1: an artificial human being in Hebrew folklore endowed with life

2: something or someone resembling a golem: as a : automaton b : blockhead

For more information, please contact codefor@systemsbiology.org.



The project described was supported by Award Number U24CA143835 from the National Cancer Institute. The content is solely the responsibility of the authors and does not necessarily represent the official views of the National Cancer Institute or the National Institutes of Health.

Introduction
Golem is intended as a light weight solution that allows researchers to run analysis on a lot of machines. It is intended to be simpler then mapreduce and more flexible then qsub and will be well suited for creating ad hock clusters to tackle problems such as Stochastic simulation and paramater space searching where a large number of long running processes need to be run but little data needs to be transfered.

Goals
Simple for a researcher to use to get their analysis script running in parallel.

Input is a list of commands to run and how many times you want to them run.

Light enough to run in background when not in use.

Client can limit processes to allow use in spare cycles.

Built using only the python and go standard libraries for portability.

Built using configurable web services for communication to allow use where other protocols might be blocked.

Possibly basic catting of stdio from distributed processes into a single file, by line but analysis script responsible for disk and data stuff.

For fault tolerance client must accept a pid token and use it to make filenames so names don't colide if command is run twice because it fails/slows on one node.

Planned Components
Server/Master - Go Receives jobs from user and sends it out to clients.

Keep track of what has finished and does basic load balancing, fault tolerance and straggler prevention.

Takes stdio sent from clients and puts it all into a file (without regard for order).

Exposes restful service for submitting, listing and deleting jobs.

Exposes web sockets for client nodes.

Client - Go Connects to server's web sockets and spawns local processes.

Might need to be configurable to allow different paths to scripts/data?

Might need to be able to download scripts.

Send stdio back to server via websocket.

Interface Scripts - Python Start, List and Delete scripts.

HIt servers restful services.

Possibly post script and small amounts of data to some sort of content repository.

Configuring Core Golem Services
These are the configuration options for golem services that are set in a configuration file passed at the command line. A more complete example can be found in the source tree.

#global configuration
#the location of the master
hostname = localhost:8083
#password to require for job submission
password = test
#use verbose logging
verbose = true
#run the cluster over https using ssl
tls = true
#use ssl certificate in the specific location (if not specified golem will generate a self signed certificate)
#certpath = $HOME/.golem
#the organization to generate a cert with
organization = example.org
#the size of the chanel of strings on either side of a connection
conbuffersize=10

[master]
#the number of cpu's to allow the master to use 
gomaxproc = 8
#the number of go routines to have parsing io for each job
iomonitors = 2
#the size of the channel used
subiobuffersize = 1000
#overrides conbuffersize above for master
conbuffersize=1000



[worker]
#the master to connect to
masterhost = localhost:8083
#the number of cpu's to allow the worker process itself to use
gomaxproc = 1
#the number of tasks to run at once (the number of cpus - gomaxproc is recomended)
processes = 3
#overrides conbuffersize above for workers
conbuffersize=10000
For this configuration to work you will need to perchase or generate an ssl certificate manually and out it in the specified location. You can also use tls = false to run the cluster over unsecure channels or leave out the certpath line in which case golem will randomly generate a self signed certificate on startup.

Configuring The Scribe
If you wish to store persistant information about jobs in a database you will need to setup MongoDB and add the following sections to a config file.

#Sections below are used only for the scribe and are not needed if the scribe is not used.
[scribe]
# the master to keep track of
target = localhost:8083
#the number of cpu's to allow the scribe to use 
gomaxproc = 2

[mgodb]
server = example.systemsbiology.net
store = golemstore
port = 1000
user = user
password = password
Starting a Master
cd to the output directory (where concatenated output will be placed) and:

golem -m -config=master.config
Connecting Client Nodes
golem -config=worker.config
Starting a Scribe

golem -s -config=scribe.config

GenerateCertificate  
Commands to generate necessary key/certificate pair for secure (tls) authorized communication Updated Dec 27, 2011 by rkreisberg@systemsbiology.org
Introduction
This assumes that openssl has been installed locally. This step is also not necessary if you use either tls = false or remove the certpath setting line in your config file (in which case a random cert will be generated on startup).

Details
Command sequence:

make a user-specific golem directory:

mkdir ~/.golem
make a PEM encoded key

openssl genrsa -out ~/.golem/key.pem 1024
generate the certificate request
openssl req -new -key ~/.golem/key.pem -out ~/.golem/certificate.csr
generate the certificate file from the request using the key

openssl x509 -req -days 365 -in ~/.golem/certificate.csr -signkey ~/.golem/key.pem -out ~/.golem/certificate.crt
translate the CRT certificate to DER format (1st step to generate a PEM)

openssl x509 -in ~/.golem/certificate.crt -out ~/.golem/certificate.der -outform DER
translate the DER formate to a PEM file

openssl x509 -in ~/.golem/certificate.der -inform DER -out ~/.golem/certificate.pem -outform PEM

HowToPythonClient  
How to submit jobs and manage golems through Python client Updated May 5, 2014 by ryanbressler
Preparing a Job
A golem job can be any unix executable accessible from every worker node.

Jobs are executed with any specified parameters plus three additional ID parameters indicating the submission, line (within the submission) and job (to allow lines to be run multiple time) submission id. Any files output by the executable may use these IDs to avoid filename conflicts. A script to pass through a command striping these parameters is included at "python/ignoreThree.py".

Executables can print to standard I/O and golem workers will collect the results line by line (with no guarantees about order) and send them to the master node where they are results are collated into single files.

Using Python Client golem.py
golem .py is used to submit either single jobs (to be run a specified number times) or lists of jobs. Its usage (which can be seen by running it with no parameters) is:

golem.py hostname [-p password] [-L label] [-u email] command and args
where command and arguments can be:

run n job_executable exeutable args	run job_executable n times with the supplied args
runlist listofjobs.txt	run each line (n n job_executable exeutable args) of the file
runerrors listofjobs.txt oldjobid	rerun the tasks that errored during the old job
rundnf listofjobs.txt oldjobid	rerun the tasks that did not finish during the old job
get jobid	Download the out, err, and log files for the specified job
list	list statuses of all submissions on cluster
jobs	same as list
status subid	get status of a single submission
stop subid	stop a submission from submitting more jobs but let running jobs finish
kill subid	stop a submission from submitting more jobs and kill running jobs
nodes	list the nodes connected to the clus ter
resize nodeid newmax	change the number of tasks a node takes at once
resizeall newmax	change the number of taska of all nodes that aren't set to take 0 tasks
resizehost cname newmax	change max tasks of a worker by cname
restart	cycle all golem proccess on the cluster...use only for udating core components
die	kill everything ... rarelly used
Running Jobs
You can submit jobs in several ways:

   python golem.py localhost:8083 run 6 samplejobs/stdio.py

   python golem.py localhost:8083 run 6 samplejobs/fileio.py

   python golem.py localhost:8083 runlist joblist.txt

   python golem.py localhost:8083 -L "Test Run" -u "user@example.com" runlist joblist.txt

   python golem.py localhost:8083 -u "user@example.com" runlist joblist.txt

   python golem.py localhost:8083 -L "Test Run" runlist joblist.txt
Joblist is specified as a white space delimited file with the form:

1 executable param1 param2
1 executable param1 param2
Where the first field in each row is the number of times to run each command and the remaining fields are the command and any arguments. All executables (excluding those in the path of the user running the client nodes) and files should be referred to by their full path on the client nodes.

See the sample jobs for more information.

Scribe  
Persistence Service for Golem Jobs 
Phase-Implementation Updated Jun 15, 2011 by hrovira.isb
Introduction
The Scribe is a REST+JSON service that provides persistence for jobs submitted to the Golem master. This optional service integrates with the Golem stack using a SOA design.

Details
The service stack for the Scribe/Master/Worker receives authenticated client requests through the Scribe, which persists jobs to the database. An asynchronous thread will submit the pending jobs from the database to the Master. An asynchronous thread polls the Master for running jobs and update the database. A REST API and user interface will be provided to monitor, stop and delete jobs.

HowToMongoDB  
Installation and setup for MongoDB. Updated Sep 12, 2011 by hrovira@systemsbiology.org
Introduction
The following steps were taken to install and setup MongoDB on MacOS X.

Download binaries from http://www.mongodb.org/downloads
Unix Instructions http://www.mongodb.org/display/DOCS/Quickstart+Unix
Sql to MongoDB Guide: http://www.mongodb.org/display/DOCS/SQL+to+Mongo+Mapping+Chart
Interactive Tutorial http://try.mongodb.org/
The following section was added to the scribe configuration file:

[mgodb]
server = localhost
store = golem_store
Links
MongoDB Quickstart
Downloads Page

HowToUserInterface  
How to setup user interface for Golem Updated Nov 2, 2011 by ryanbressler
Introduction
The golem project provides hooks for user interfaces. The RestOnJob interface provides an /html URL mapping that serves HTML content from the filesystem. In addition, we are including example HTML content under http://golem.googlecode.com/hg/src/html to demonstrate how to construct a web user interface.

The example code is based on the ExtJS Javascript framework, but this is not a required dependency. The javascript is loaded from ISB servers and the Addama project.

Details
To serve HTML content from golem services (e.g. Master, Scribe), there must be an html directory in the path of the service. For instance:

    cd /local/projects/golem/src
    gomake
    cp golem /local/webapps/mygolem/
    cd /local/apps/webapps/mygolem/
    mkdir html
    cp $CONTENT_HOME/* html/
    ./golem -m -hostname myhost:8083
The cluster ui will be available at https://myhost:8083/html/. Visualizations of cluster performance will be available at https://myhost:8083//html/stream.html.

This mechanism allows any content contained in the html directory to be served by the golem services.
