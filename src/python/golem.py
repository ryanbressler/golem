#!/usr/bin/python
#    Copyright (C) 2003-2010 Institute for Systems Biology
#                            Seattle, Washington, USA.
# 
#    This library is free software; you can redistribute it and/or
#    modify it under the terms of the GNU Lesser General Public
#    License as published by the Free Software Foundation; either
#    version 2.1 of the License, or (at your option) any later version.
# 
#    This library is distributed in the hope that it will be useful,
#    but WITHOUT ANY WARRANTY; without even the implied warranty of
#    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
#    Lesser General Public License for more details.
# 
#    You should have received a copy of the GNU Lesser General Public
#    License along with this library; if not, write to the Free Software
#    Foundation, Inc., 59 Temple Place, Suite 330, Boston, MA 02111-1307  USA
# 
# 

import sys
import json
import urllib
import urllib2


def main():
	master = sys.argv[1]
	cmd = sys.argv[2]
	
	url = "http://"+master+"/jobs/"
	if cmd == "run":
		print "Submiting run request to %s"%(url)
		jobs = [{"Count":int(sys.argv[3]),"Args":sys.argv[4:]}]
		data = urllib.urlencode([('data', json.dumps(jobs))])
		req = urllib2.Request(url)
		fd = urllib2.urlopen(req, data)
		print fd.read()
	if cmd == "runlist":
		fo = open(sys.argv[3])
		jobs=[]
		for line in fo:
			vals = line.split()
			jobs.append({"Count":int(vals[0]),"Args":vals[1:]})
		data = urllib.urlencode([('data', json.dumps(jobs))])
		req = urllib2.Request(url)
		fd = urllib2.urlopen(req, data)
		print fd.read()
	
	

if __name__ == "__main__":
	main()