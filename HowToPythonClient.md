# Preparing a Job #
A golem job can be any unix executable accessible from every worker node.

Jobs are executed with any specified parameters plus three additional ID parameters indicating the submission, line (within the submission) and job (to allow lines to be run multiple time) submission id.  Any files output by the executable may use these IDs to avoid filename conflicts. A script to pass through a command striping these parameters is included at "python/ignoreThree.py".

Executables can print to standard I/O and golem workers will collect the results line by line (with no guarantees about order) and send them to the master node where they are results are collated into single files.

# Using Python Client golem.py #
golem .py is used to submit either single jobs (to be run a specified number times) or lists of jobs.  Its usage (which can be seen by running it with no parameters) is:

```
golem.py hostname [-p password] [-L label] [-u email] command and args

```
where command and arguments can be:

|run n job\_executable exeutable args | run job\_executable n times with the supplied args|
|:------------------------------------|:--------------------------------------------------|
|runlist listofjobs.txt              | run each line (n n job\_executable exeutable args) of the file|
|runerrors listofjobs.txt oldjobid   | rerun the tasks that errored during the old job|
|rundnf listofjobs.txt oldjobid      | rerun the tasks that did not finish during the old job|
|get jobid                           | Download the out, err, and log files for the specified job|
|list                                | list statuses of all submissions on cluster|
|jobs                                | same as list|
|status subid                        | get status of a single submission|
|stop subid                          | stop a submission from submitting more jobs but let running jobs finish|
|kill subid                          | stop a submission from submitting more jobs and kill running jobs|
|nodes                               | list the nodes connected to the clus  ter|
|resize nodeid newmax                | change the number of tasks a node takes at once|
|resizeall newmax                    | change the number of taska of all nodes that aren't set to take 0 tasks|
|resizehost cname newmax             | change max tasks of a worker by cname|
|restart                             | cycle all golem proccess on the cluster...use only for udating core components|
|die                                 | kill everything ... rarelly used|



# Running Jobs #
You can submit jobs in several ways:
```
   python golem.py localhost:8083 run 6 samplejobs/stdio.py

   python golem.py localhost:8083 run 6 samplejobs/fileio.py

   python golem.py localhost:8083 runlist joblist.txt

   python golem.py localhost:8083 -L "Test Run" -u "user@example.com" runlist joblist.txt

   python golem.py localhost:8083 -u "user@example.com" runlist joblist.txt

   python golem.py localhost:8083 -L "Test Run" runlist joblist.txt
```

Joblist is specified as a white space delimited file  with the form:
```
1 executable param1 param2
1 executable param1 param2
```

Where the first field in each row is the number of times to run each command and the remaining fields are the command and any arguments. All executables (excluding those in the path of the user running the client nodes) and files should be referred to by their full path on the client nodes.

See the [sample jobs](http://code.google.com/p/golem/source/browse/#hg%2Fsrc%2Fsamplejobs) for more information.