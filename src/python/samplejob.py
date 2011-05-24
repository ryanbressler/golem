#!/usr/bin/python
import random
import sys

id = sys.argv[1]
print "script entered with id %s"%(id)

fo = open(id+".output.txt","w")
for i in range(1000):
	fo.write("%i\n"%(random.randint(0, 1000)))
fo.close()