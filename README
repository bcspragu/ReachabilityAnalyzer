# ReachabilityAnalyzer

Given a bench file, this program can compute whether or not a goal state is
reachable from an initial state.

## Installation

To run this program, the picosat command line utility needs to be installed. To
compile this program, Golang and all of its build tools need to be installed,
then running 'make' will generate the binary with the name analyzer.

## Usage

The application is run as a command line utility, with flags specifying the
majority of the behavior.

--input
  Specifies the relative path and name of the bench file to be read
  in, defaults to bench/ex1. Note: The file extension should be left off, and
  the state file needs to have the same name. Check the examples at the end for
  more information.

--runner
  Specifies the number of runner threads to use for explicit search, defaults
  to 10, but can be safely set from 1 to several thousand.

--unroll
  Specifies the number of unrollings to do for symbolic search, defaults to 2.

--log
  Specifies how much logging information to display. 0 is none and the default,
  1 is debug information, and 2 displays all information.

-c
  Run explicit search, ignoring the goal state and returning a count of all reachable states.

-e
  Run explicit search.

-s 
  Run symbolic search.

### Examples

./analyzer --input=bench/ex3 --runners=250 -e
  Runs explicit search on bench/ex3 with 250 runners

./analyzer --input=bench/ex4 --unroll=17 -s
  Runs symbolic search on bench/ex4 with 17 unrollings

./analyzer --input=bench/ex2 --log=1 -c
  Runs explicit search on bench/ex2 with debugging output and count all reachable states

## Test Suite

The program also has a benchmarking/test suite, which can be run by entering
the bench/ directory and running 'go test' or 'go test --bench=.'
