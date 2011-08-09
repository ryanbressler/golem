__author__ = 'anorberg'

import golemize
import random
import json


def nyan(num, ignored):
    return "nyan" * num

def randoms(low, high, n):
    for x in range(0, n):
        yield random.randint(low, high)

if __name__ == "__main__":
    golemizer = golemize.dictToGolemizer(json.load(open("glados.json", "r")))
    kitty = golemizer.goDoIt(randoms(1, 10, 50), None, nyan)
    nyanyanyan = [meow for meow in kitty]
    print len(nyanyanyan)
    for stringinging in nyanyanyan:
        print stringinging
        for wee in stringinging.split("nyan"):
            assert wee == ""

  