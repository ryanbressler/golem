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

def unpickleSequence(pickleFiles):
    """Generator function that unpickles sequences of (Boolean, result?exception?) tuples from the files
    in the parametric list. If an exception is expressed, it is rethrown here.
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

class Golemizer:
    def __init__(self, serverUrl, serverPass, golemOutputPath, golemIdSeq, pickleScratch, thisLibraryPath, pyPath = "/hpc/bin/python", pickleOut = None, taskSize = 10):
        self.masterPath = golem.canonizeMaster(serverUrl) + "/jobs/"
        self.serverPass = serverPass
        self.golemOutPath = golemOutputPath
        self.golemIds = ["{0:02d}".format(id) for id in golemIdSeq]
        self.pickleInputShare = pickleScratch
        self.pyPath = pyPath
        self.thisLibraryPath = thisLibraryPath
        if pickleOut:
            self.jobOutputPath = pickleOut
        else:
            self.jobOutputPath = "./"
        self.taskSize = taskSize

    def setTaskSize(self, value):
        """
        Setters aren't necessary in Python, but their presence makes it much clearer that
        this is an intended portion of the API for this object.
        """
        self.taskSize = value

    def _spill(self, nextList, pickleCount):
        nextPickle = open(str(pickleCount) + ".pkl", "wb", -1)
        cPickle.dump(nextList, nextPickle, 2)
        nextPickle.flush()
        nextPickle.close()

    def goDoIt(self, inputSeq, commonData, targetFunction, binplace = True, alternateSource = None):
        """
        Executes a function on the Golem cluster indicated by the settings for this object.
        inputSeq: inputs to the function to run. This will be desequenced and run in a batch of method calls, one
        call per item.
        """
        restoreThisCwdOrPeopleWillHateMePassionately = os.getcwd()
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
                    n = 0;
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

            response, content = golem.runBatch(runlist, self.serverPass, self.masterPath)
            jobId = golemBlocking.jobIdFromResponse(content)
            golemBlocking.stall(jobId, self.masterPath)

            #Note: We're choosing to ignore stdout/stderr. We can revisit this design decision later and decide to
            #do something instead, if we really desperately want to

            resultPathGenerator = (os.path.abspath(
                os.path.join(
                    self.golemOutPath, "golem_" + x + os.sep, self.jobOutputPath,
                )
            ) for x in self.golemIds)

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

            resultFilesNumbered.sort()

            return unpickleSequence((pair[1] for pair in resultFilesNumbered))
        finally:
            os.chdir(restoreThisCwdOrPeopleWillHateMePassionately)

def dictToGolemizer(config):
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
        range(
            int(config["lowGolemID"]),
            int(config["highGolemID"]),
            1
        ),
        config["golemStagingRoot"],
        config["golemizeScriptPath"],
        pythonBinPath,
        pickleOut,
        taskSize
    )

def jsonToGolemizer(jsonfile):
    return dictToGolemizer(json.load(jsonfile))

def jumpToTask():
    #The traditionally "Right" thing to do is to use a ConfigParser or equivalent. However,
    #the case-sensitive position-sensitive spot equality comparison for --golemtask is fine when we've forcibly
    #constructed the relevant args ourselves. It minimizes delay, and minimizes chance of interfering
    #with some legit command line that for some reason uses --golemtask (hopefully not in position 1).
    if len(sys.argv) < 9:
        raise ValueError("Not a valid command line (wrong count)") #not one of our command lines

    if sys.argv[1] != "--golemtask":
        raise ValueError("Not a valid command line (not a --golemtask)")

    #NORETURN beyond this point

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
    targetModule = __import__(modname)
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
        print jumpToTask()
    except ValueError:
        print "This Python module is a library."
        print "It is invoked as a script in its own right as part of its operation, but this is not such an invocation."
        print "Please read the documentation for more details."