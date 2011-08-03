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
import re
import sys

import golem
import signal
import optparse
import time

try:
    import json #python 2.6 included simplejson as json
except ImportError:
    import simplejson as json

QUERY_INTERVAL = 3.0 #in seconds

def stall(jobid, composedUrl):
    """
    Waits until the specified job is no longer Running.
    If it can't communicate with the server, it will throw an IOError.
    If there is no "Running" field in the response at all, Stall will terminate normally.
    If a sleep is interrupted via keyboard, stall will throw a KeyboardInterrupt.
    """
    decoder = json.JSONDecoder()
    while True:
        response, content = golem.getJobStatus(jobid, composedUrl)
        if response.status != 200:
            raise IOError("Unsuccessful status when communicating with server: " + response)
        contentDict = decoder.decode(content)
        if not contentDict["Running"]:
            return contentDict
        time.sleep(QUERY_INTERVAL)

#TODO: invent function printUsage()

def main(argv):
    parser = optparse.OptionParser()
    parser.add_option("-p", "--password", dest="password", help="Specify the password for connecting to the server.",
                      default="")
    parser.add_option("-e", "--echo", dest="echo", action="store_true", default=False)
    flags, args = parser.parse_args(argv[1:4]) #because "late params" are actually arguments to the target script

    password = flags.password
    args = args + argv[4:]

    if len(args) < 3:
        print "Not enough arguments."
        printUsage()
        sys.exit(status=1)

    master = args[0]

    master = golem.canonizeMaster(master)
    url = master+"/jobs/"

    command = args[1]
    cmdArgs = args[2:]

    if command=="run":
        response, content = golem.runOneLine(int(cmdArgs[0]), cmdArgs[1:], password, url)
    elif command=="runlist":
        response, content = golem.runList(open(cmdArgs[0]), password, url)
    elif command=="runoneach":
        response, content = golem.runOnEach([{"Args": cmdArgs}],password,url)
    else:
        raise ValueError("golemBlocking can only handle the commands 'run', 'runlist', and 'runoneach'.")

    try:
        contentDict = json.JSONDecoder().decode(content)
        id = contentDict["id"]
    except ValueError:
        try:
            id = re.search(r'[\s\{]"?id:"(\w*)"',content).group(1)
        except AttributeError:
            id = re.search(r"[\s\{]'?id:'(\w*)'", content).group(1)

    try:
        stall(id, url)
    except KeyboardInterrupt:
        golem.stopJob(id, password, url)
        print "Job halted."


if __name__ == "__main__":
    main(sys.argv)
    