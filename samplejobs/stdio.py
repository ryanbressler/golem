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
#############################################################################
""" This is a sample golem job that generates random numbers and prints them to stdio.
	
	From the comand line it can be ran specifying the number of random numbers to simulate and the
	sample, line and job ids:
	stdio.py 3 0 0 0
	
	To run it six times generating three random numbers each using a golem master listening at
	localhost:8083:
	
	golem.py localhost:8083 run 6 ~/Code/golem/src/samplejobs/stdio.py 3
	
	To run it using a list specifying provided list specifying diffrent paramaters:
	golem.py localhost:8083 runlist ~/Code/golem/src/samplejobs/list.for.stdio.py.txt


"""
import random
import sys

# the first arg is the number of times to simulate a random number
n = int(sys.argv[1])

#concactante the next three args, the submission, line and job id's into one string
id = ".".join(sys.argv[2:])



for i in range(n):
	#output the id string i and a random number. output looks like '4.0.4,2,611'
	print "%s,%i,%i"%(id,i,random.randint(0, 1000))
