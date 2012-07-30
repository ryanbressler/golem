__author__ = 'anorberg'

"""
Simple test code to validate new enhanced exception processing.
"""

import golemize

def recip(num, ignored):
	return 1.0/num

if __name__ == "__main__":
	golemizer = golemize.jsonToGolemizer(open("enhancedExceptionGlados.json", "r"))
	bustedResult = golemizer.goDoIt(range(-100, 100), None, recip, quiet=True)
	n = -100
	for value in bustedResult:
		print value
