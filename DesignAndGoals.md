# Introduction #

Golem is intended as a light weight solution that allows researchers to run analysis on a lot of machines. It is intended to be simpler then mapreduce and more flexible then qsub and will be well suited for creating ad hock clusters to tackle problems such as Stochastic simulation and paramater space searching where a large number of long running processes need to be run but little data needs to be transfered.

# Goals #
Simple for a researcher to use to get their analysis script running in parallel.

Input is a list of commands to run and how many times you want to them run.

Light enough to run in background when not in use.

Client can limit processes to allow use in spare cycles.

Built using only the python and go standard libraries for portability.

Built using configurable web services for communication to allow use where other protocols might be blocked.

Possibly basic catting of stdio from distributed processes into a single file, by line but analysis script responsible for disk and data stuff.

For fault tolerance client must accept a pid token and use it to make filenames so names don't colide if command is run twice because it fails/slows on one node.


# Planned Components #

Server/Master - Go
Receives jobs from user and sends it out to clients.

Keep track of what has finished and does basic load balancing, fault tolerance and straggler prevention.

Takes stdio sent from clients and puts it all into a file (without regard for order).

Exposes restful service for submitting, listing and deleting jobs.

Exposes web sockets for client nodes.

Client - Go
Connects to server's web sockets and spawns local processes.

Might need to be configurable to allow different paths to scripts/data?

Might need to be able to download scripts.

Send stdio back to server via websocket.


Interface Scripts - Python
Start, List and Delete scripts.

HIt servers restful services.

Possibly post script and small amounts of data to some sort of content repository.