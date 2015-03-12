![![](http://golem.googlecode.com/files/Slide1.jpg)](http://golem.googlecode.com/files/Slide1.jpg)

Golem (Go Launch on Every Machine) strives to be the simplest possible system for distributing computational analysis across processes and machines in a Unix/Linux based cluster. It uses web technologies, core features of the [Go language](http://golang.org) and Unix enviroment to keep internal code simple and maintainable.  It is being developed at the [Institute for Systems Biology](http://systemsbiology.org) in Seattle, WA to provide a fast, light, simple and accessible tools for parallelizing algorithms used in computational biology, cancer research and large scale data analysis.

Golem is easier to setup and use than common `MapReduce` and job queuing systems and provided a different set of features. It focuses on providing a set of tools that allow quick parallelization of existing single threaded analysis that make use of `*`nix file handle based communication, command line tools, restful web services to allow easy integration with existing infrastructure. It also includes a web interface for viewing job and cluster status built on the above.

It is well suited for problems that involve many independent calculations but small amounts of data transfer, and for which the distributed sort-and-combine included in the `MapReduce` pattern is an unnecessary expense due either to the size of the result or the lack of an interesting ordering.  In data analysis this includes algorithms that involve **repeated independent permutations**, **partitions** or **stochastic simulations** and any **embarrassingly parallelizable** analysis that _can be run independently on separate portions of the data producing results that can be easily combined or don't need to combined_.

Use cases at ISB include parallel random forest analysis (using [rf-ace](http://code.google.com/p/rf-ace)), motif  searching across entire genomes (using [Molotov](http://code.google.com/p/molotov/)), the identification of genetic structural variation at a population scale (using [FastBreak](http://code.google.com/p/fastbreak/)) and exhaustive calculation of pairwise correlations between cancer features across patients.

A typical golem job is list of tasks which can be parameterized calls to any executable or script. The golem master balances these tasks across all available processors and, as a convenience, combines standard out and standard error from all jobs line by line with no guarantees of order.

Golem is in an alpha but usable state see BuildingAndUsing for more.

Golem (from merriam-webster):

1: an artificial human being in Hebrew folklore endowed with life

2: something or someone resembling a golem: as
a : automaton
b : blockhead

For more information, please contact [codefor@systemsbiology.org](mailto:codefor@systemsbiology.org).




The project described was supported by Award Number U24CA143835 from the National Cancer Institute. The content is solely the responsibility of the authors and does not necessarily represent the official views of the National Cancer Institute or the National Institutes of Health.