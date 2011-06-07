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
import urllib2
import httplib
import urlparse


def doPost(url, paramMap):
	u = urlparse.urlparse(url)
	
	headers = { "Content-type": "application/x-www-form-urlencoded","Accept": "text/plain" }

   

	#params = "empty"
	#if paramMap:
	
	params = urllib.urlencode(paramMap)

	print "scheme: %s host: %s port: %s"%(u.scheme, u.hostname, u.port)
	
	print "parameters: [" + params + "]"
	
	conn=0
	if u.scheme == "http":
		conn = httplib.HTTPConnection(u.hostname,u.port,None,2,("localhost","8080"))
	else:
		conn = httplib.HTTPSConnection(u.hostname,u.port,"/Users/rbressle/.golem/key.pem","/Users/rbressle/.golem/certificate.pem")#,None,2,("localhost","8080"))
	
	conn.set_debuglevel(100)
		
	conn.request("POST", u.path, params, headers)

	resp = conn.getresponse()
	if resp.status == 200:
		output = resp.read()
		try:
			print json.dumps(json.JSONDecoder().decode(output), sort_keys=True, indent=4)
		except:
			print output
	else:
		print resp.status, resp.reason

	conn.close()

def main():
	master = sys.argv[1]
	cmd = sys.argv[2]
	
	
	if master[0:4] != "http":
		print "Using http (unsecure)."
		master = "http://"+master
	url = master+"/jobs/"
	if cmd == "run":
		
		jobs = [{"Count":int(sys.argv[3]),"Args":sys.argv[4:]}]
		data = {'data': json.dumps(jobs)}
		print "Submiting run request to %s."%(url)
		doPost(url,data)
	
	if cmd == "runlist":
		fo = open(sys.argv[3])
		jobs=[]
		for line in fo:
			vals = line.split()
			jobs.append({"Count":int(vals[0]),"Args":vals[1:]})
		data = {'data': json.dumps(jobs)}
		print "Submiting run request to %s."%(url)
		doPost(url,data)
		
	if cmd == "ls":
		print "not yet implemented"
	
	if cmd == "status":
		print "not yet implemented"
		
	if cmd == "kill":
		print "not yet implemented"
	
	

if __name__ == "__main__":
	main()