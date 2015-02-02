## tbd
### a #tag-based dependency tool for task tracking

#### [presentation](http://go-talks.appspot.com/github.com/solarsea/tbd/tbd.slide#1)

#### In order to use:

* Keep your tasks in a text file named 'tbdata', one task per line
* Tag #some words in your tasks to #mark dependencies
* A task with a tag set T depends on any previous task with a tag set P when (T ∩ P) ≠ ∅

#### Invoke tbd

* without arguments to list all independent tasks
* with tag names to list all tasks that match at least one tag along with their dependencies

##### License

Copyright ©  2015 Stanislav Paskalev

Usage of the works is permitted provided that this instrument is retained with the works, so that any entity that uses the works is notified of this instrument.

DISCLAIMER: THE WORKS ARE WITHOUT WARRANTY.

##### Misc

* [![Build Status](https://drone.io/github.com/solarsea/tbd/status.png)](https://drone.io/github.com/solarsea/tbd/latest)
* Live, learn and take it easy :)
