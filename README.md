## tbd
### a #tag-based dependency tool for task tracking

#### [presentation](http://go-talks.appspot.com/github.com/solarsea/tbd/tbd.slide#1)

#### In order to use:

* Keep your tasks in a text file named 'tbdata', one task per line
* Tag words in your tasks with #some #tags to mark dependencies
* A task with a tag set T depends on any previous task with a tag set P when (T ∩ P) ≠ ∅

#### Invoke tbd

* without arguments to list all independent tasks
* with tag names to list all tasks that match at least one tag along with their dependencies

Live, learn and take it easy :)

