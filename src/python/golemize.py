#!/tools/bin/python

"""
A module to distribute a Python function across a Golem cluster.

This module contains the class Golemizer, which represents a set of connection information to a Golem server and the
ability to distribute a job from a sequence of input. It also contains a few helper functions to create a Golemizer
from a configuration file.
"""

import os
import re
import shutil

__author__ = 'anorberg'

try:
    import json
except ImportError:
    import simplejson as json
import golem
import golemBlocking
import sys
import inspect
import cPickle
import uuid

def _unpickleSequence(pickleFiles):
    """Generator function that unpickles results from job runs.

    Parameter:
        pickleFiles - A sequence of strings containing paths to valid pickled files containing sequences of results.

    Yields:
        The results of the function that was passed to the goDoIt call that generated the pickle, sequentially. If
        that function threw an exception, the generator throws that exception at the corresponding point.

    A result is a tuple of (boolean, anything), where the boolean is interpreted as an error flag. If the flag is
    not set, the other data is a standard result, which is yielded. If the flag is set, the result was a caught
    exception, which will be reraised (killing the generator, but allowing user debugging of their code).
    """
    for filePath in pickleFiles:
        picklefile = open(filePath, "rb")
        try:
            seq = cPickle.load(picklefile)
            for errorflag, data in seq:
                if errorflag:
                    raise data
                else:
                    yield data
        finally:
            picklefile.close()

class ExecutionFailure(Exception):
    """
    Exception class that represents any unknown failure preventing results from being generated on a node.
    """
    def __init__(self, message):
        """
        Constructs an ExecutionFailure exception with the given message.
        """
        self.message = message
    def __str__(self):
        """
        Returns the name of the exception class, and the message assigned to this exception.
        """
        return "ExecutionFailure: " + str(self.message)
    def __repr__(self):
        """
        Returns the constructor expression for this class.
        """
        return "ExecutionFailure(message=" + repr(self.message) + ")"

class InfiniteRecursionError(Exception):
    """
    Exception class that represents the detection of goDoIt being called from a job that is already a Golem task,
    which almost guarantees an infinite hang, and usually represents a lack of __name__ == '__main__' guards.
    """
    def __init__(self, message):
        """
        Constructs an InfiniteRecursionError with the given message.
        """
        self.message = message
    def __str__(self):
        """
        Returns the name of the exception class, and the message assigned to this exception.
        """
        return "InfiniteRecursionError: " + str(self.message)
    def __repr__(self):
        """
        Returns the constructor expression for this class.
        """
        return "InfiniteRecursionError(message=" + repr(self.message) + ")"

class Golemizer:
    """
    Class that represents connection/configuration information for a Golem server, and the logic to distribute a job.
    """
    def __init__(self, serverUrl, serverPass, golemOutputPath, pickleScratch, thisLibraryPath, pyPath = "/hpc/bin/python", pickleOut = None, taskSize = 10):
        """
        Constructor for a Golemizer.
        Parameters:
            serverUrl - String representing the URL of the server (including port number).
            serverPass - String representing the server password.
            golemOutputPath - String representing the file system path containing the golem_n folders used as
                              working directories for the worker nodes.
            pickleScratch - String representing the file system path where Golemizer should write the pickled input,
                            and possibly other files. Needs to be reachable, via the same path, on all workers.
            thisLibraryPath - String representing the file system path, accessible from all machines, to golemize.py.
            pyPath - String representing the file system path to the installation of Python on the workers.
                     Default: "/hpc/bin/python"
            pickleOut - Alternate output directory for pickle files. The working directory for the Golem nodes
                        will be used if this is not provided. Default: "None", which is equivalent to "./" or
                        anything that evaluates to false
            taskSize - Number of tasks per execution of the software on a worker. Larger values (possibly significantly)
                       reduce job distribution overhead. Smaller values can be distributed more evenly.
        """
        self.masterPath = golem.canonizeMaster(serverUrl, False) + "/jobs/"
        self.serverPass = serverPass
        self.golemOutPath = golemOutputPath
        #self.golemIds = ["{0:02d}".format(id) for id in golemIdSeq]
        self.pickleInputShare = pickleScratch
        self.pyPath = pyPath
        self.thisLibraryPath = thisLibraryPath
        if pickleOut:
            self.jobOutputPath = pickleOut
        else:
            self.jobOutputPath = "./"
        self.taskSize = taskSize

    def __repr__(self):
        """
        Returns the constructor expression for this class.
        """
        return "Golemizer(serverUrl={0},serverPass={1},golemOutputPath={2},pickleScratch={3},thisLibraryPath={4},pyPath={5},pickleOut={6},taskSize={7})".format(
            self.masterPath[:-6],
            self.serverPass,
            self.golemOutPath,
            self.pickleInputShare,
            self.pyPath,
            self.thisLibraryPath,
            self.jobOutputPath,
            self.taskSize
        )

    def setTaskSize(self, value):
        """
        Setters aren't necessary in Python, but their presence makes it much clearer that
        this is an intended portion of the API for this object.
        """
        self.taskSize = value

    def _spill(self, nextList, pickleCount):
        """
        Internal function to dump a list of items into a numbered pickle file. Returns None.
        Parameters:
            nextList - List of items to pickle. (The list itself is pickled as a unit, not just the elements. This
                       results in a file that is trivially larger, but is significantly better in performance
                       because it keeps execution in cpickle rather than gating back to Python.)
            pickleCount - ID number for this pickle.
        """
        nextPickle = open(str(pickleCount) + ".pkl", "wb", -1)
        cPickle.dump(nextList, nextPickle, 2)
        nextPickle.flush()
        nextPickle.close()

    def goDoIt(self, inputSeq, commonData, targetFunction, binplace = True, alternateSource = None, recursive = False, quiet = False, label="", email=""):
        """
        Executes a function on the Golem cluster indicated by the settings for this object.
        Parameters:
            inputSeq - inputs to the function to run. This will be desequenced and run in a batch of method calls, one
                       call per item. Objects in the input must be possible and meaningful to pickle, and
                       unpickle on a different machine.
            commonData - input that should be provided to every invocation of the function. Must be possible and
                         meaningful to pickle, then unpickle on a different machine.
            targetFunction - Function to execute. Must have prototype func(item, item), by any name, where the
                             first item is something off the inputSeq sequence and the second is the commonData.
                             It must return its result, as only these return values will be pickled and sent back;
                             changes to global variables are not captured or returned.
            binPlace - Should the script containing the target function be copied to the path that the input
                       data is being copied to? If so, that one file must be the ONLY non-library file required
                       for the function to work properly, since files around it won't be known about. This
                       permits development out of a directory that isn't world-readable. If this is set to something
                       that evaluates to false, the file will be used in place, and must be visible to the
                       workers by the same path it is visible to the client. Default: True.
            alternateSource - Path to the file to use as the source for the script containing the targetFunction.
                              This should be blank in almost all instances, since this function will determine this
                              data reflectively in that case and that is much more likely to be correct. Use this
                              if, for some reason, the file detected by inspect.getAbsFilePath is wrong, or if
                              a different path is simply required to access the data on the workers.
                              If the binPlace flag is set, this is the file that is copied, if it is not None (if it is,
                              then the detected file is copied). Default: None.
            recursive - Deprecated.
            quiet - Suppress server responses from being printed to stdout while waiting for results. Default: False,
                    for backwards compatibility. "True" is more likely to be desirable.
            label - Alternate identifier for locating the job in the log later. Optional.
            email - Informational field to identify the person running the job in case they need to be contacted. Optional.
        """
        if len(sys.argv) > 1 and sys.argv[1] == "--golemtask":
            #uh-oh
            raise InfiniteRecursionError("goDoIt called from something that was already a Golem task, " +
                                        "without the 'recursive' flag indicating that this is intentional." +
                                        "Make sure to test for __name__ == '__main__' in your main program," +
                                        "or it will try to execute in its entirety when Golemizer tries to import.")

        restoreThisCwdOrPeopleWillHateMePassionately = os.getcwd()
        loud = not quiet
        try:
            outName = str(uuid.uuid1())
            os.chdir(self.pickleInputShare)
            os.mkdir(outName) #insecure: mode 0777
            os.chdir(outName)
            picklePath = os.getcwd()
            pickleCount = 0
            nextList = []
            n = 0
            localLimit = self.taskSize

            for parameter in inputSeq:
                nextList.append(parameter)
                n += 1
                if n >= localLimit:
                    self._spill(nextList, pickleCount)
                    nextList = []
                    n = 0
                    pickleCount += 1
            if nextList:
                self._spill(nextList, pickleCount)
                pickleCount += 1

            if not alternateSource:
                #restore original path or getabsfile doesn't work correctly as of 2.7
                os.chdir(restoreThisCwdOrPeopleWillHateMePassionately)
                target = inspect.getabsfile(targetFunction)
                os.chdir(self.pickleInputShare)
                os.chdir(outName)
                #print "===> Original file:", target
            else:
                target = alternateSource

            if binplace:
                #print "===> Original file:", target
                newTarget = os.path.join(picklePath, os.path.basename(target))
                #print "===> New file:", newTarget
                shutil.copy2(target, newTarget)
                target = newTarget
                
            commonFile = open("common.pkl", "wb")
            commonObjectPickler = cPickle.Pickler(commonFile, 2)
            commonObjectPickler.dump(commonData)
            commonObjectPickler.dump(targetFunction)
            commonFile.flush()
            commonFile.close()

            runlist = [
                    {"Count":1, "Args":[self.pyPath,
                                        self.thisLibraryPath,
                                        "--golemtask",
                                        os.path.join(picklePath, "common.pkl"),
                                        os.path.join(picklePath, str(n)+".pkl"), #we are making certain filename assumptions on the client side
                                        self.jobOutputPath,
                                        target]
                    }
                for n in range(0, pickleCount)]

            response, content = golem.runBatch(runlist, self.serverPass, self.masterPath, loud, label, email)
            jobId = golemBlocking.jobIdFromResponse(content)
            finalStatus = golemBlocking.stall(jobId, self.masterPath, loud)
            if loud and (finalStatus["Status"] != "SUCCESS"):
                print "Uh-oh- job status is", finalStatus["Status"], "and we're probably going to crash soon"

            #Note: We're choosing to ignore stdout/stderr. We can revisit this design decision later and decide to
            #do something instead, if we really desperately want to

            #resultPathGenerator = (os.path.abspath(
            #    os.path.join(
            #        self.golemOutPath, "golem_" + x + os.sep, self.jobOutputPath,
            #    )
            #) for x in self.golemIds)

            golemDirPattern =re.compile("golem_\\d+")

            resultPathGenerator = (os.path.abspath(
                    os.path.join(self.golemOutPath, foo))\
                    for foo in os.listdir(self.golemOutPath)\
                    if golemDirPattern.match(foo))

            resultFilesNumbered = []

            filenamePattern = re.compile("^{0}_(\\d+)\\.out\\.pkl$".format(jobId))

            #because we're already performing the match, decorate-sort-undecorate is the best sort strategy here
            for resultPath in resultPathGenerator:
                #print "==>", resultPath
                for file in os.listdir(resultPath):
                    match = filenamePattern.match(file)
                    if match:
                        #print "====>", file
                        resultFilesNumbered.append((int(match.group(1)), os.path.join(resultPath, file)))

            if len(resultFilesNumbered) != pickleCount:
                raise ExecutionFailure(
                    "Unknown error prevented {0} of {1} task bundles from completing.".format(
                        pickleCount - len(resultFilesNumbered), pickleCount))
            resultFilesNumbered.sort()

            return _unpickleSequence((pair[1] for pair in resultFilesNumbered))
        finally:
            os.chdir(restoreThisCwdOrPeopleWillHateMePassionately)

def dictToGolemizer(config):
    """Constructs a Golemizer from a group of settings stored in  a dictionary, keyed by string.
    Parameters:
        config - A dict with the following key-value pairs (all keys are strings):
            "pickleOut" - optional. A string representing the full path to where worker nodes should put results,
                          if somewhere other than their working directory.
            "taskSize" - optional. An integer number of tasks per job bundle.
            "pythonBin" - optional. The path to the Python interpreter on each worker. If this field is
                          not present, the default is "/hpc/bin/python".
            "serverURL" - The string representing the URL to reach the Golem server with, including port.
            "serverPassword" - The string representing the login password for the Golem server.
            "golemResultRoot" - The string representing the root path to where Golem aggregates worker node
                                working directories.
            "golemStagingRoot" - The string representing a file path, identical on all machines, where the input
                                 can be serialized and then retrieved later.
            "golemizeScriptPath" - The string representing the path to golemize.py, visible to all workers.
    Returns:
        A Golemizer constructed from the fields in the provided dict.
    Throws:
        KeyError if a required field is missing.
    """
    pickleOut = None
    if "pickleOut" in config:
        pickleOut = str(config["pickleOut"])
    taskSize = 10
    if "taskSize" in config:
        taskSize = int(config["taskSize"])
    pythonBinPath = "/hpc/bin/python"
    if "pythonBin" in config:
        pythonBinPath = str(config["pythonBin"])

    return Golemizer(
        config["serverURL"],
        config["serverPassword"],
        config["golemResultRoot"],
        #range(
        #    int(config["lowGolemID"]),
        #    int(config["highGolemID"]),
        #    1
        #),
        config["golemStagingRoot"],
        config["golemizeScriptPath"],
        pythonBinPath,
        pickleOut,
        taskSize
    )

def jsonToGolemizer(jsonfile):
    """
    Constructs a Golemizer from a file open for read that contains a correctly-formatted JSON object.

    The file is parsed as a JSON object into a dict, then fed into dictToGolemizer. The object must fit
    the same schema as imposed by dictToGolemizer.
    """
    return dictToGolemizer(json.load(jsonfile))

def _jumpToTask():
    """
    Internal method used when golemize.py is invoked as a script to execute a single job.

    Takes no input. Returns the number of individual tasks that failed.

    argv:
        1   --golemtask
        2   path to common.pkl
        3   path to pickle file for this specific task
        4   output path (usually ./)
        5   path to script that contains calculation function- must be safe for import
        6   job ID (automatically provided by golem, used during computation)
        7   row ID (automatically provided by golem, ignored)
        8   task ID (automatically provided by golem, ignored)
    """
    #The traditionally "Right" thing to do is to use a ConfigParser or equivalent. However,
    #the case-sensitive position-sensitive spot equality comparison for --golemtask is fine when we've forcibly
    #constructed the relevant args ourselves. It minimizes delay, and minimizes chance of interfering
    #with some legit command line that for some reason uses --golemtask (hopefully not in position 1).
    if len(sys.argv) < 9:
        raise ValueError("Not a valid command line (wrong count)") #not one of our command lines

    if sys.argv[1] != "--golemtask":
        raise ValueError("Not a valid command line (not a --golemtask)")

    #argv standard:
    # 1:    --golemtask
    # 2:    common data path (contains common data and function pointer)
    # 3:    task data path (contains sequence of Stuff that should be given to calculation function)
    # 4:    output path (usually "./", but available in case we want to centralize output)
    # 5:    host script path
    # 6:    job ID (automatically added by golem, we use it)
    # 7:    row ID (automatically added by golem, we ignore it)
    # 8:    task ID (automatically added, we ignore it, better be equal to 6 since we're only firing tasks once)

    inScript = sys.argv[5]

    sys.path.append(os.path.dirname(inScript))#puts the original script on the module search path for depickle
    modname = os.path.basename(inScript).split(".")[0]
    try:
        targetModule = __import__(modname)
    except InfiniteRecursionError as emergencyBrake:
        #provide a useful explanation for the kill
        trunc = (os.path.basename(sys.argv[3]))[:-4] #truncates ".pkl"
        outFileName = sys.argv[6]+"_"+trunc+".out.pkl" #this name is sacred to finding the results, including the jobID
        outFileName = os.path.join(sys.argv[4], outFileName)
        outFile = open(outFileName, "wb")
        cPickle.dump([(True, emergencyBrake)], outFile, 2)
        outFile.flush()
        outFile.close()
        raise emergencyBrake

    # this is the nasty part
    # The next four statements recreate the environment out of which the computation method was pickled
    # Thus, the computation method can be unpickled. Unpickling will fail otherwise.
    # Lookup-by-name from the module fails when the calculation function is an instance method of a constructed
    #   object instance, which cPickle is smart enough to carry along with the function reference, but a by-name
    #   lookup clearly wouldn't be.
    globalRef = globals()
    for thingie in dir(targetModule):
        if thingie not in globalRef:
            globalRef[thingie] = targetModule.__dict__[thingie]
    
    commonFile = open(sys.argv[2], "rb")
    commonUnpickle = cPickle.Unpickler(commonFile)
    commonData = commonUnpickle.load()
    doIt = commonUnpickle.load()
    commonUnpickle = None #intentional dead store for safety reasons
    commonFile.close()

    taskFile = open(sys.argv[3], "rb")
    taskList = cPickle.load(taskFile)
    taskFile.close()

    ret = []
    failureCount = 0

    for task in taskList:
        errored = False
        try:
            result = doIt(task, commonData)
        except Exception as miserableFailure:
            errored = True
            result = miserableFailure
            failureCount += 1
        ret.append((errored, result))

    trunc = (os.path.basename(sys.argv[3]))[:-4] #truncates ".pkl"

    outFileName = sys.argv[6]+"_"+trunc+".out.pkl" #this name is sacred to finding the results, including the jobID
    outFileName = os.path.join(sys.argv[4], outFileName)
    outFile = open(outFileName, "wb")

    cPickle.dump(ret, outFile, 2)
    outFile.flush()
    outFile.close()

    return failureCount

if __name__ == "__main__":
    try:
        _jumpToTask()
    except ValueError:
        print "This Python module is a library."
        print "It is invoked as a script in its own right as part of its operation, but this is not such an invocation."
        print "Please read the documentation for more details."