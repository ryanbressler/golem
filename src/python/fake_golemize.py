import json

__author__ = 'anorberg'

"""
A mock version of the golemize.py library.

This has the exact same function signatures as golemize.py, but very different behavior.
While golemize.py distributes work over a Golem computational cluster, fake_golemize simply runs all calculations
on the local machine. This allows for a more rapid testing loop that permits conventional debugging, and a "real"
Golemizer can be thrown in later.

This is explicitly designed for "import fake_golemize as golemize" to work perfectly as long as golemize is
being used properly.
"""

def _doItLater(taskInputSeq, commonInput, function):
    """
    A generator function that performs late closure over a function given its sequence of inputs.
    Parameters:
        taskInputSeq - Iterable object (sequence type, generator) that yields the first parameter to the function.
        commonInput - object to provide on every invocation of the function.
        function - a function that takes two parameters and yields something.
    Yields:
        The value of repeated calls to the provided function with parameters (taskInputSeq[?], commonInput), for
        increasing indexes (more accurately, iteration) over taskInputSeq.
    Throws:
        Anything that the function or the input sequence throws- this makes no attempt to catch or handle errors.
    """
    for input in taskInputSeq:
        yield function(input, commonInput)

class Golemizer:
    """
    A fake version of the Golemizer class.

    Pretends to be a Golemizer, but it ignores all of its settings and 'goDoIt' returns a generator that performs
    the specified function on-the-spot on the local machine.
    """
    def __init__(self, serverUrl = None, serverPass = None,
                 golemOutputPath = None, pickleScratch = None, thisLibraryPath = None,
                 pyPath = None, pickleOut = None, taskSize = 1):
        """
        Creates a fake Golemizer. This constructor is empty and contains only a 'pass'. However, it supports
        the full syntax of a real Golemizer constructor, all of which it ignores and are optional.
        Parameters:
            serverUrl - ignored, default None
            serverPass - ignored, default None
            golemOutputPath - ignored, default None
            pickleScratch - ignored, default None
            thisLibraryPath - ignored, default None
            pyPath - ignored, default None
            pickleOut - ignored, default None
            taskSize  - ignored, default 1
        """
        #none of this is used! This is all to mock the real Golemizer interface.
        pass

    def setTaskSize(self, value):
        """
        Does nothing. On a real Golemizer, this would change the number of items per job bundle.
        Parameter:
            value - ignored
        """
        pass

    def _spill(self,nextList,pickleCount):
        """
        Complains that nothing should be calling _spill here. On a real Golemizer, this would write a pickle file.

        Parameters:
            nextList - ignored
            pickleCount - ignored
        Throws:
            RuntimeError - always. This function should not be called.
        """
        raise RuntimeError("_spill is an internal function to Golemizer. Fakes don't have one. Don't call it.")
    def __repr__ (self):
        """
        Returns an eval()uable string representation of the object. All fake Golemizer objects are the same.
        """
        return "Golemizer()"
    def goDoIt(self, inputSeq, commonData, targetFunction, binplace=None, alternateSource=None, recursive=None, quiet=None):
        """
        Mocks the operation of distributed computation by returning a generator that performs the calculation locally.

        Parameters:
            inputSeq - Iterable object (sequence, generator) that provides inputs to the calculation function.
            commonData - Object that is provided to all calls to the calculation function.
            targetFunction - Function to pretend to distribute across a cluster. Of course, no such thing actually
                             happens; instead, this will be invoked during iteration over the generator that
                             this function provides.
            binplace - ignored, default None
            alternateSource - ignored, default None
            recursive - ignored, deprecated, default None
            quiet - ignored, default None
        """
        # not using the last three parameters- they are to mock the interface
        return _doItLater(inputSeq, commonData, targetFunction)

def dictToGolemizer(config):
    """
    Ignores a dictionary and creates a fake Golemizer that has nothing to do with the parameters in it.
    Parameters:
        config - ignored
    Returns:
        A (fake) Golemizer created by passing no parameters to its constructor. The fake Golemizer ignores all
        of its parameters anyway, so this is identical to an object the has parameters.
    """
    return Golemizer()

def jsonToGolemizer(jsonfile):
    """
    Loads a JSON-formatted file, then throws it out and creates an unrelated fake Golemizer.
    """
    asFile = open(jsonfile, "r")
    json.load(asFile) #...and throw it away
    asFile.close()
    return Golemizer()

def _jumpToTask():
    raise RuntimeError("jumpToTask is a highly volatile internal function that shouldn't be getting called outside of Golemizer, and the fake doesn't call it. Whatever is trying to call it is probably doing something wrong.")