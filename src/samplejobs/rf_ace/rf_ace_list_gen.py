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

def golemSubmit(pythonBin, golemPwd, commandsFile):
	#py2.6 golem.py glados.systemsbiology.net:8083 -p g0l3mm45t3r runlist 15jun_rnaseq_coadread_b.sh
	cmd = "%s golem.py glados.systemsbiology.net:8083 -p %s runlist %s" %(pythonBin, golemPwd, commandsFile)
	print "submitting to golem: " + cmd
	os.system(cmd)

def main(start, end, inputMatrixFile, associations_dir, commandsOut, gosubmit):
	config = ConfigParser.RawConfigParser()
	config.read('./rf_ace.config')
	commandwriter = open(commandsOut, 'w')
	if (not os.path.exists('./rf_ace.config')):
		print "rf_ace.config file is missing"
		sys.exit(-1)
	execpath = config.get("RF_ACE_Parameters", "execpath")
	mtry = config.getint("RF_ACE_Parameters", "mtry")
	numtrees = config.getint("RF_ACE_Parameters", "numtrees")
	permutations = config.getint("RF_ACE_Parameters", "permutations")
	pvalue_t = config.get("RF_ACE_Parameters", "pvalue_t")
	nodesize = config.get("RF_ACE_Parameters", "nodesize")
	pythonBin = config.get("PYTHON", "pythonbin")
	golemPwd = config.get("GOLEM", "golempwd")
	if (not associations_dir.endswith('/')):
		associations_dir = associations_dir + "/"
	while start < end:
			#/proj/ilyalab/TCGA/rf-ace/bin/rf_ace -I ../KruglyakGenewisePhenoProteomics.NEW.transposed.csv -i 0 -n 100 -m 1000 -p 20 -O associations_0.out
			cmd = "1 %s -I %s -i %i -n %i -m %i --nodesize %s -p %i -t %s -O %sassociations_%i.out" %(execpath, inputMatrixFile, start, numtrees, mtry, nodesize, permutations, pvalue_t, associations_dir, start)
			commandwriter.write(cmd + "\n");
			start = start + 1
	commandwriter.close()
	if (gosubmit):
		golemSubmit(pythonBin, golemPwd, commandwriter.name)

if __name__=="__main__":
	parser = optparse.OptionParser(usage="usage: %prog [options] filetureStart[start at 0] featureEnd, inputMatrixFile, associationsDir, commandsOutfile",version="%prog 1.0")
	parser.add_option('-l', '--local', help="local mode, will not check whether matrix file and output directory exists, important that you confirmed that the feature matrix and associations output path are valid before submitting jobs to grid", dest='local_mode', default=False, action='store_true')
	parser.add_option('-s', '--submit', help="Inclusion of this flag will tell the program to submit job list to GOLEM - if you are running this from local, it is important to validate that your input matrix and output directory exists", dest='go_submit', default=False, action='store_true')
	(opts, args) = parser.parse_args()
	if (len(args) == 5):
		#default number of threads is 10, sleep 40
		matrix_file = args[2]
		associations_dir = args[3]
		commandsOutfile = args[4]
		if (not os.path.exists(associations_dir) and not opts.local_mode):
                	try:
                        	os.makedirs(associations_dir)
				os.system('chmod 777 ' + associations_dir)
                	except OSError:
                        	print "Associations output path does not exist and mkdir failed %s, exiting" %associations_dir
				sys.exit(-1)
		if (not os.path.exists(matrix_file) and not opts.local_mode):
	                print "%s is not a valid file, exiting" % matrix_file
                        sys.exit(-1)
		main(int(args[0]), int(args[1]), matrix_file, associations_dir, commandsOutfile, opts.go_submit)
	else:
		print 'Try python(2.5+) rf_ace_list_gen.py --help'
		sys.exit(1)
	

