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


import time
import os
import sys
import ConfigParser



def main(start, end, inputMatrixFile, outpath):
	config = ConfigParser.RawConfigParser()
	config.read('./rf_ace.config')
	"""
	[RF_ACE_Parameters]
	execpath=/proj/ilyalab/TCGA/rf-ace/bin/rf_ace
	mtry=175
	numtrees=300
	permutations=20
	pvalue_t=.1
	nodesize=5
	alpha=.95
	"""
	execpath = config.get("RF_ACE_Parameters", "execpath")
	mtry = config.getint("RF_ACE_Parameters", "mtry")
	numtrees = config.getint("RF_ACE_Parameters", "numtrees")
	permutations = config.getint("RF_ACE_Parameters", "permutations")
	pvalue_t = config.get("RF_ACE_Parameters", "pvalue_t")
	nodesize = config.get("RF_ACE_Parameters", "nodesize")
	alpha = config.get("RF_ACE_Parameters", "alpha")

	while start < end:
			#/proj/ilyalab/TCGA/rf-ace/bin/rf_ace -I ../KruglyakGenewisePhenoProteomics.NEW.transposed.csv -i 0 -n 100 -m 1000 -p 20 -O associations_0.out
			cmd = "1 %s -I %s -i %i -n %i -m %i -p %i -t %s -O %sassociations_%i.out" %(execpath, inputMatrixFile, start, numtrees, mtry, permutations, pvalue_t, outpath, start)
			print cmd
			start = start + 1

	

if __name__=="__main__":
	if (len(sys.argv) == 5):
		#default number of threads is 10, sleep 40 
		main(int(sys.argv[1]), int(sys.argv[2]), sys.argv[3], sys.argv[4])
	else:
		print 'Proper usage is python(2.5+) rf_ace_scheduler.py featureStart featureEnd, inputMatrixFile, outpath'
		sys.exit(1)
	

