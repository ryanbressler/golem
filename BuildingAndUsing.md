# Introduction #
Golem is a web service architecture that allows for the distribution of work across a cluster.  Golem is written in the [Go Language](http://golang.org).  The following instructions will help users build and compile the necessary components

# Go Lang #
Visit http://golang.org/doc/install.html to get started with GO.

# Dependencies and Building #

Provided you have common source control utilities installed go makes it easy to get all of golems dependencies and build the `golem` executable:

```
hg clone https://code.google.com/p/golem/
cd golem
[sudo] go get
go build
```

(Sudo may or may not be required depending on the permission of your go package dir.)


# Golem usage #
The **golem** utility includes a number of command line options described below in the form -flag=default value. Boolean options (m, v) will be set to true if included without a value (-m, -v) or explicitly set to false using -flag=false.  Other options can be set using either "-flag value" or -flag=value.  An additional option (`-config`) is used to facilitate complex configuration through files.
```
  -h Show usage.
  -m=false: Start as master node.
  -s=false: Start as scribe node.
  -config=golem.config
```

Proceed to [Configure Golem Services](HowToConfigureServices.md)

Proceed to [Running Python Client](HowToPythonClient.md)