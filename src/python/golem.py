#!/usr/bin/python
#	 Copyright (C) 2003-2010 Institute for Systems Biology
#							 Seattle, Washington, USA.
# 
#	 This library is free software; you can redistribute it and/or
#	 modify it under the terms of the GNU Lesser General Public
#	 License as published by the Free Software Foundation; either
#	 version 2.1 of the License, or (at your option) any later version.
# 
#	 This library is distributed in the hope that it will be useful,
#	 but WITHOUT ANY WARRANTY; without even the implied warranty of
#	 MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
#	 Lesser General Public License for more details.
# 
#	 You should have received a copy of the GNU Lesser General Public
#	 License along with this library; if not, write to the Free Software
#	 Foundation, Inc., 59 Temple Place, Suite 330, Boston, MA 02111-1307  USA
# 
# 

import sys
try: import json #python 2.6 included simplejson as json
except ImportError: import simplejson as json
import urllib
import httplib
import urlparse
import os
supporttls=True
try: 
	from tlslite.integration.HTTPTLSConnection import HTTPTLSConnection
	from tlslite.X509 import X509
	from tlslite.X509CertChain import X509CertChain
	from tlslite.utils.keyfactory import parsePEMKey
except ImportError:
	supporttls=False
	print "Error importing tlslite."

usage = """Usage: golem.py hostname [-p password] command and args
where command and args can be:
run n job_executable exeutable args : run job_executable n times with the supplied args
runlist listofjobs.txt              : run each line (n n job_executable exeutable args) of the file 
list                                : list statuses of all submissions on cluster
status subid                        : get status of a single submission
stop subid                          : stop a submission
nodes                               : list the nodes connected to the cluster
resize nodeid newmax                : change the number of jobs a node takes at once
restart                             : cycle all golem proccess on the cluster...use only for udating core components
die                                 : kill everything ... rarelly used
"""

"""multipart encodes a form. data should be a dictionary of the the form fields and filebody 
should be a string of the body of the file"""
def encode_multipart_formdata(data, filebody):
    BOUNDARY = '----------ThIs_Is_tHe_bouNdaRY_$'
    CRLF = '\r\n'
    L = []
    for key, value in data.iteritems():
        L.append('--' + BOUNDARY)
        L.append('Content-Disposition: form-data; name="%s"' % key)
        L.append('')
        L.append(value)
	if filebody != "":
		L.append('--' + BOUNDARY)
		L.append('Content-Disposition: form-data; name="jsonfile"; filename="data.json"')
		L.append('Content-Type: text/plain')
		L.append('')
		L.append(filebody)
    L.append('--' + BOUNDARY + '--')
    L.append('')
    body = CRLF.join(L)
    content_type = 'multipart/form-data; boundary=%s' % BOUNDARY
    return content_type, body

"""
posts a multipart form to url, paramMap should be a dictionary of the form fields, json data 
should be a string of the body of the file (json in our case) and password should be the password 
to icnlude in the header
"""
def doGet(url):
	u = urlparse.urlparse(url)
	conn=0
	if u.scheme == "http":
		conn = httplib.HTTPConnection(u.hostname,u.port)
	else:

		conn = HTTPTLSConnection(u.hostname,u.port)#,privateKey=key,certChain=X509CertChain([cert]))
	
		
	conn.request("GET", u.path)

	resp = conn.getresponse()
	if resp.status == 200:
		output = resp.read()
		try:
			print json.dumps(json.JSONDecoder().decode(output), sort_keys=True, indent=4)
		except:
			print output
	else:
		print resp.status, resp.reason

	#conn.close()

"""
posts a multipart form to url, paramMap should be a dictionary of the form fields, json data 
should be a string of the body of the file (json in our case) and password should be the password 
to icnlude in the header
"""
def doPost(url, paramMap, jsondata,password):

	u = urlparse.urlparse(url)
	
	content_type, body =encode_multipart_formdata(paramMap,jsondata)
	headers = { "Content-type": content_type,
		'content-length':str(len(body)),
		"Accept": "text/plain",
		"Password":password }


	print "scheme: %s host: %s port: %s"%(u.scheme, u.hostname, u.port)
	

	
	conn=0
	if u.scheme == "http":
		conn = httplib.HTTPConnection(u.hostname,u.port)
	else:

		conn = HTTPTLSConnection(u.hostname,u.port)#,privateKey=key,certChain=X509CertChain([cert]))
	
		
	conn.request("POST", u.path, body, headers)

	resp = conn.getresponse()
	if resp.status == 200:
		output = resp.read()
		try:
			print json.dumps(json.JSONDecoder().decode(output), sort_keys=True, indent=4)
		except:
			print output
	else:
		print resp.status, resp.reason

	#conn.close()

def main():
	if len(sys.argv)==1:
		print usage
		return
		
	master = sys.argv[1]
	cmdi = 2
	pwd = ""
	if sys.argv[2] == "-p":
		pwd = sys.argv[3]	
	cmdi = 4
	
	
	cmd = sys.argv[cmdi]

	#Todo: default to http when not tls.
	
	if master[0:4] != "http":
		if supporttls:
			print "Using https."
			master = "https://"+master
		else:
			print "Using http (unsecure)."
			master = "http://"+master
	if master[0:5] == "https" and supporttls == False:
		return "To use https please instal the python package tlslite."
	
	url = master+"/jobs/"
	if cmd == "run":
		
		jobs = [{"Count":int(sys.argv[cmdi+1]),"Args":sys.argv[cmdi+2:]}]
		jobs = json.dumps(jobs)
		data = {'command':cmd}
		print "Submiting run request to %s."%(url)
		doPost(url,data,jobs,pwd)
	
	if cmd == "runlist":
		fo = open(sys.argv[cmdi+1])
		jobs=[]
		for line in fo:
			vals = line.split()
			jobs.append({"Count":int(vals[0]),"Args":vals[1:]})
		jobs = json.dumps(jobs)
		data = {'command':cmd}
		print "Submiting run request to %s."%(url)
		doPost(url,data,jobs,pwd)
		
	if cmd == "runoneach":
		
		jobs = [{"Args":sys.argv[cmdi+1:]}]
		jobs = json.dumps(jobs)
		data = {'command':cmd}
		print "Submiting run request to %s."%(url)
		doPost(url,data,jobs,pwd)
	if cmd == "list":
		doGet(url)
		
	if cmd == "stop":
		jobid = sys.argv[cmdi+1]
		doPost(url+jobid+"/stop",{},"",pwd)
	if cmd == "status":
		jobid = sys.argv[cmdi+1]
		doGet(url+jobid)
		
	if cmd == "nodes":
		doGet(master+"/admin/")
	if cmd == "resize":
		doPost(master+"/admin/"+sys.argv[cmdi+1]+"/resize/"+sys.argv[cmdi+2],{},"",pwd)
	if cmd == "restart":
		input = raw_input("This will kill all jobs on the cluster and is only used for updating golem version. Enter \"Y\" to continue.>")
		if input == "Y":
			doPost(master+"/admin/restart",{},"",pwd)
		else:
			print "Canceled"
	if cmd == "die":
		input = raw_input("This kill the entire cluster down and is almost never used. Enter \"Y\" to continue.>")
		if input == "Y":
			doPost(master+"/admin/die",{},"",pwd)
		else:
			print "Canceled"
	
	if cmd == "ls":
		print "not yet implemented"
	
		
	
	
	

if __name__ == "__main__":
	main()