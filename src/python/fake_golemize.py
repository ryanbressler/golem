import json

__author__ = 'anorberg'

# A mock version of the golemize.py library.
# This has the exact same function signatures as golemize.py, but very different behavior.
# While golemize.py distributes work over a Golem computational cluster, fake_golemize simply runs all calculations
# on the local machine. This allows for a more rapid testing loop that permits conventional debugging, and a "real"
# Golemizer can be thrown in later.

# This is explicitly designed for "import fake_golemize as golemize" to work perfectly as long as golemize is
# being used properly.

def doItLater(taskInputSeq, commonInput, function):
    for input in taskInputSeq:
        yield function(input, commonInput)

class Golemizer:
    def __init__(self, serverUrl = None, serverPass = None,
                 golemOutputPath = None, pickleScratch = None, thisLibraryPath = None,
                 pyPath = None, pickleOut = None, taskSize = 1):
        #none of this is used! This is all to mock the real Golemizer interface.
        pass

    def setTaskSize(self, value):
        pass

    def _spill(self,nextList,pickleCount):
        raise RuntimeError("_spill is an internal function to Golemizer. Fakes don't have one. Don't call it.")

    def goDoIt(self, inputSeq, commonData, targetFunction, binplace=None, alternateSource=None, recursive=None):
        #we're not using the last three parameters- they are to mock the interface
        return doItLater(inputSeq, commonData, targetFunction)

def dictToGolemizer(config):
    #all fake golemizers are the same
    return Golemizer()

def jsonToGolemizer(jsonfile):
    #all fake golemizers are still the same, even if you'd rather use a file than a dict
    #still, it seems polite to make sure your file at least is legible
    asFile = open(jsonfile, "r")
    json.load(asFile) #...and throw it away
    asFile.close()
    return Golemizer()

def jumpToTask():
    raise RuntimeError("jumpToTask is a highly volatile internal function that shouldn't be getting called outside of Golemizer, and the fake doesn't call it. Whatever is trying to call it is probably doing something wrong.")