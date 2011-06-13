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
import optparse

def main(start, end, inputMatrixFile, associations_dir):
	config = ConfigParser.RawConfigParser()
	config.read('./rf_ace.config')
	if (not os.path.exists('./rf_ace.config')):
		print "rf_ace.config file is missing"
		sys.exit(-1)
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

	if (not associations_dir.endswith('/')):
		associations_dir = associations_dir + "/"
	while start < end:
			#/proj/ilyalab/TCGA/rf-ace/bin/rf_ace -I ../KruglyakGenewisePhenoProteomics.NEW.transposed.csv -i 0 -n 100 -m 1000 -p 20 -O associations_0.out
			cmd = "1 %s -I %s -i %i -n %i -m %i -p %i -t %s -O %sassociations_%i.out" %(execpath, inputMatrixFile, start, numtrees, mtry, permutations, pvalue_t, associations_dir, start)
			print cmd
			start = start + 1

	
if __name__=="__main__":
	parser = optparse.OptionParser(usage="usage: %prog [options] filetureStart featureEnd, inputMatrixFile, associationsDir",version="%prog 1.0")
	parser.add_option('-l', '--local', help="local mode, will not check whether matrix file and output directory exists, important that you confirmed that the feature matrix and associations output path are valid before submitting jobs to grid", dest='local_mode', default=False, action='store_true')
	(opts, args) = parser.parse_args()
	#print "mode %s number of args %i" % (str(opts.local_mode), len(args)) 
	if (len(args) == 4):
		#default number of threads is 10, sleep 40
		matrix_file = args[2]
		associations_dir = args[3]
		if (not os.path.exists(associations_dir) and not opts.local_mode):
                	try:
                        	os.makedirs(associations_dir)
                	except OSError:
                        	print "Associations output path does not exist and mkdir failed %s, exiting" %associations_dir
				sys.exit(-1)
		if (not os.path.exists(matrix_file) and not opts.local_mode):
	                print "%s is not a valid file, exiting" % matrix_file
                        sys.exit(-1)
		main(int(args[0]), int(args[1]), matrix_file, associations_dir)
	else:
		print 'Proper usage is python(2.5+) rf_ace_scheduler.py featureStart featureEnd, inputMatrixFile, outpath'
		sys.exit(1)
	

