#!/usr/bin/env python
import os
import sys

if __name__ == '__main__':
	cmd = " ".join(sys.argv[1:-3])
	os.system(cmd)