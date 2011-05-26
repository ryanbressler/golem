#!/usr/bin/python
import random
import sys

id = ".".join(sys.argv[1:])
#print "script entered with id %s"%(id)


for i in range(3):
	print "%s,%i,%i"%(id,i,random.randint(0, 1000))
