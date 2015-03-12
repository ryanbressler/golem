This page discusses what we need from a system designed to handle large amounts of disk IO. It is intended to be used either to evaluate existing solutions (HDFS, GridFS, Fuse) or design our own.

# Rationale #
Jobs like [Molotov](http://molotov.googlecode.com) generate enough output to block on writing to our current system (which pipes all standard out back to a single file on the master node) or even direct writes to disks on one of our large file servers. We don't necessarily need to keep all of this output in backed up long term storage (like our file servers) but we may want to keep it around for short periods of time so that we can repeat downstream analysis as we develop it.

Because the each of the nodes in our cluster have 4 large local disks originally intended for use with _HDFS_ it makes sense to seek a solution that allows us to leverage these disks.

# Requirements #

  1. Modularity...golem must continue to work on clusters that don't wish to use this feature.
  1. Low footprint when not in use.
  1. Ability to load balance writes from tasks across the 4 local disks on each worker node.
  1. Ability to send downstream analysis to nodes that have the appropriate file minimizing network traffic.
  1. Ability to organize output by (unique) job id.
  1. Ability to find, list and cat job output from remote nodes as needed.
  1. Since all of our science aspires to be repeatable and our cluster is relatively small we don't need much in the way of fault tolerance provided faults do not pass silently... i.e. it is acceptable for us to have to repeat an analysis if a disk fails.

# Alternatives We Have Looked at #

  1. HDFS has not proved to play nice with others on our cluster and has a lot of overhad due to features we don't need.
  1. GridFS doesn't offer the level control needed to allow us to send jobs close to files.

# Straw Man Proposal #
I propose that we develop a File Distribution System much lighter existing Distributed File System solutions.

  1. A service that runs on each node that tracks disk IO and and free space. When someone wants to write a file it makes a request of this service (via rest) including its unique job and task id and receives a path to write too. This service provides basic load balancing to ensure all disks on the node are being used as appropriate.
  1. A service that runs on the master node that receives posts from the above and tracks where file outputs ended up and exposes a restful interface to locate files.
  1. Utilities for loading filesets onto the cluster and getting them off of the nodes (we can just thinly wrap scp, fuse or http servers running on the nodes as appropriate).

In addition to allowing us to handle large amounts of io, we can add support for golem jobs that take multiple files as input... i.e. reducers allowing users to perform MapReduce-like tasks.