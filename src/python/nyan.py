__author__ = 'anorberg'

"""
Simple example code for golemize.py. Produces a sequence of strings of "nyan" of varying length.
"""

import golemize
import random
import json


def nyan(num, ignored):
    """
    Returns the string "nyan" appended to itself num times. The ignored parameter is ignored, but required
    to match golemize.Golemizer.goDoIt's required interface.
    """
    return "nyan" * num

def randoms(low, high, n):
    """
    Generator function that yields n random integers between low and high.
    Parameters:
        low - integer representing the (inclusive) minimum value for the random results.
        high - integer representing the (exclusive) maximum value for the random results.
        n - number of results to yield before the generator stops.
    """
    for x in range(0, n):
        yield random.randint(low, high)

if __name__ == "__main__":
    golemizer = golemize.dictToGolemizer(json.load(open("glados.json", "r")))
    kitty = golemizer.goDoIt(randoms(1, 10, 50), None, nyan, quiet=True)
    nyanyanyan = [meow for meow in kitty]
    print len(nyanyanyan)
    for stringinging in nyanyanyan:
        print stringinging
        for wee in stringinging.split("nyan"):
            assert wee == ""

  