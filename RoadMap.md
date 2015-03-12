This is a space to list features that may or may not be going into golem in an order that may or may not make sense.

# Road Map #

  1. Tasks status including something like pending, errored, retry\_1, retry\_1\_errored, retry\_1\_failed, retry\_all\_nodes, retry\_all\_nodes\_errored, retry\_all\_nodes\_failed.
  1. HIgh and low task priority.
  1. Tracking Task In node Handles and extending node reconnecting to retry taks when a node disappears or when the task errors in accordance with above codes.
  1. Rest interface to doanload task list with status. Also allow restart (from webapp?).
  1. GolemFD ([The Golem File Distribution System](FileSystemSpecandProposal.md)) as seperate stand alone utility  centered around a master Rest store for tracking file locations.
  1. golemfdin a utility for streaming files onto local disc tracked by golemfd. Usage something like:
```
task.py | golemfdin jobid fname"
```
  1. golemfdout a utility that accepts a task id and a number of files and cats files from golemfd or returns filenames. Also accepts a flag to instruct it to error if files aren't availible locally in such a way that golem will redistribute it until it finds a node that has local files. Usage something like:
```
golemfdout oldjobid -l -n 4 | combinetask.py | golemfdin %newjobid%
```
> > This will give us a set of independent simple tools that can be easily wired together using basic bash to process large amounts of data in a map reduce like paradigm.
  1. Nodes optionally monitor available resources and error a job in a way that it will be retried if there are not enough available resources to prevent tasks dyeing due to lack of resources and allow us to take advantage of spare cycles.



