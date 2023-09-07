# mow.cli
[![Build Status](https://travis-ci.org/jawher/mow.cli.svg?branch=master)](https://travis-ci.org/jawher/mow.cli)
[![GoDoc](https://godoc.org/github.com/github.com/jawher/mow.cli?status.svg)](https://godoc.org/github.com/jawher/mow.cli)
[![Coverage Status](https://coveralls.io/repos/github/jawher/mow.cli/badge.svg?branch=master)](https://coveralls.io/github/jawher/mow.cli?branch=master)

Package cli provides a framework to build command line applications in Go with most of the burden of arguments parsing and validation placed on the framework instead of the user.


## Getting Started
The following examples demonstrate basic usage the package.


### Simple Application
In this simple application, we mimic the argument parsing of the standard UNIX
cp command. Our application requires the user to specify one or more source
files followed by a destination.  An optional recursive flag may be provided.

```go
package main

import (
	"fmt"
	"os"

	"github.com/jawher/mow.cli"
)

func main() {
	// create an app
	app := cli.App("cp", "Copy files around")

	// Here's what differentiates mow.cli from other CLI libraries:
	// This line is not just for help message generation.
	// It also validates the call to reject calls with less than 2 arguments
	// and split the arguments between SRC or DST
	app.Spec = "[-r] SRC... DST"

	var (
		// declare the -r flag as a boolean flag
		recursive = app.BoolOpt("r recursive", false, "Copy files recursively")
		// declare the SRC argument as a multi-string argument
		src = app.StringsArg("SRC", nil, "Source files to copy")
		// declare the DST argument as a single string (string slice) arguments
		dst = app.StringArg("DST", "", "Destination where to copy files to")
	)

	// Specify the action to execute when the app is invoked correctly
	app.Action = func() {
		fmt.Printf("Copying %v to %s [recursively: %v]\n", *src, *dst, *recursive)
	}

	// Invoke the app passing in os.Args
	app.Run(os.Args)
}
```

### Pointers to existing variables

This variant of the cp command uses the Ptr variants, where you can pass pointers to existing variables
instead of declaring new ones for the options/arguments:

```go
package main

import (
	"fmt"
	"os"

	cli "github.com/jawher/mow.cli"
)

type Config struct {
	Recursive bool
	Src       []string
	Dst       string
}

func main() {
	var (
		app = cli.App("cp", "Copy files around")
		cfg Config
	)
	// Here's what differentiates mow.cli from other CLI libraries:
	// This line is not just for help message generation.
	// It also validates the call to reject calls with less than 2 arguments
	// and split the arguments between SRC or DST
	app.Spec = "[-r] SRC... DST"

	// declare the -r flag as a boolean flag
	app.BoolOptPtr(&cfg.Recursive, "r recursive", false, "Copy files recursively")
	// declare the SRC argument as a multi-string argument
	app.StringsArgPtr(&cfg.Src, "SRC", nil, "Source files to copy")
	// declare the DST argument as a single string (string slice) arguments
	app.StringArgPtr(&cfg.Dst, "DST", "", "Destination where to copy files to")

	// Specify the action to execute when the app is invoked correctly
	app.Action = func() {
		fmt.Printf("Copying using config: %+v\n", cfg)
	}
	// Invoke the app passing in os.Args
	app.Run(os.Args)
}
```

### Multi-Command Application
In the next example, we create a multi-command application in the same style as
familiar commands such as git and docker.  We build a fictional utility called
uman to manage users in a system.  It provides two commands that can be invoked:
list and get. The list command takes an optional flag to specify all users
including disabled ones.  The get command requries one argument, the user ID,
and takes an optional flag to specify a detailed listing.

```go
package main

import (
	"fmt"
	"os"

	"github.com/jawher/mow.cli"
)

func main() {
	app := cli.App("uman", "User Manager")

	app.Spec = "[-v]"

	var (
		verbose = app.BoolOpt("v verbose", false, "Verbose debug mode")
	)

	app.Before = func() {
		if *verbose {
			// Here we can enable debug output in our logger for example
			fmt.Println("Verbose mode enabled")
		}
	}

	// Declare our first command, which is invocable with "uman list"
	app.Command("list", "list the users", func(cmd *cli.Cmd) {
		// These are the command-specific options and args, nicely scoped
		// inside a func so they don't pollute the namespace
		var (
			all = cmd.BoolOpt("all", false, "List all users, including disabled")
		)

		// Run this function when the command is invoked
		cmd.Action = func() {
			// Inside the action, and only inside, we can safely access the
			// values of the options and arguments
			fmt.Printf("user list (including disabled ones: %v)\n", *all)
		}
	})

	// Declare our second command, which is invocable with "uman get"
	app.Command("get", "get a user details", func(cmd *cli.Cmd) {
		var (
			detailed = cmd.BoolOpt("detailed", false, "Disaply detailed info")
			id       = cmd.StringArg("ID", "", "The user id to display")
		)

		cmd.Action = func() {
			fmt.Printf("user %q details (detailed mode: %v)\n", *id, *detailed)
		}
	})

	// With the app configured, execute it, passing in the os.Args array
	app.Run(os.Args)
}

```


### A Larger Multi-Command Example
This example shows an alternate way of organizing our code when dealing with a
larger number of commands and subcommands. This layout emphasizes the command
structure and defers the details of each command to subsequent functions. Like
the prior examples, options and arguments are still scoped to their respective
functions and don't pollute the global namespace.

```go
package main

import (
	"fmt"
	"os"

	"github.com/jawher/mow.cli"
)

// Global options available to any of the commands
var filename *string

func main() {
	app := cli.App("vault", "Password Keeper")

	// Define our top-level global options
	filename = app.StringOpt("f file", os.Getenv("HOME")+"/.safe", "Path to safe")

	// Define our command structure for usage like this:
	app.Command("list", "list accounts", cmdList)
	app.Command("creds", "display account credentials", cmdCreds)
	app.Command("config", "manage accounts", func(config *cli.Cmd) {
		config.Command("list", "list accounts", cmdList)
		config.Command("add", "add an account", cmdAdd)
		config.Command("remove", "remove an account(s)", cmdRemove)
	})

	app.Run(os.Args)
}

// Sample use: vault list OR vault config list
func cmdList(cmd *cli.Cmd) {
	cmd.Action = func() {
		fmt.Printf("list the contents of the safe here")
	}
}

// Sample use: vault creds reddit.com
func cmdCreds(cmd *cli.Cmd) {
	cmd.Spec = "ACCOUNT"
	account := cmd.StringArg("ACCOUNT", "", "Name of account")
	cmd.Action = func() {
		fmt.Printf("display account info for %s\n", *account)
	}
}

// Sample use: vault config add reddit.com -u username -p password
func cmdAdd(cmd *cli.Cmd) {
	cmd.Spec = "ACCOUNT [ -u=<username> ] [ -p=<password> ]"
	var (
		account  = cmd.StringArg("ACCOUNT", "", "Account name")
		username = cmd.StringOpt("u username", "admin", "Account username")
		password = cmd.StringOpt("p password", "admin", "Account password")
	)
	cmd.Action = func() {
		fmt.Printf("Adding account %s:%s@%s", *username, *password, *account)
	}
}

// Sample use: vault config remove reddit.com twitter.com
func cmdRemove(cmd *cli.Cmd) {
	cmd.Spec = "ACCOUNT..."
	var (
		accounts = cmd.StringsArg("ACCOUNT", nil, "Account names to remove")
	)
	cmd.Action = func() {
		fmt.Printf("Deleting accounts: %v", *accounts)
	}
}
```


## Comparison to Other Tools
There are several tools in the Go ecosystem to facilitate the creation of
command line tools.  The following is a comparison to the built-in flag package
as well as the popular urfave/cli (formerly known as codegangsta/cli):

|                                                                     | mow.cli | urfave/cli | flag |
|---------------------------------------------------------------------|---------|------------|------|
| Contextual help                                                     | ✓       | ✓          |      |
| Commands                                                            | ✓       | ✓          |      |
| Option folding `-xyz`                                               | ✓       |            |      |
| Option value folding `-fValue`                                      | ✓       |            |      |
| Option exclusion `--start ❘ --stop`                                 | ✓       |            |      |
| Option dependency `[-a -b]` or `[-a [-b]]`                          | ✓       |            |      |
| Arguments validation `SRC DST`                                      | ✓       |            |      |
| Argument optionality `SRC [DST]`                                    | ✓       |            |      |
| Argument repetition `SRC... DST`                                    | ✓       |            |      |
| Option/argument dependency `SRC [-f DST]`                           | ✓       |            |      |
| Any combination of the above `[-d ❘ --rm] IMAGE [COMMAND [ARG...]]` | ✓       |            |      |

Unlike the simple packages above, docopt is another library that supports rich
set of flag and argument validation. It does, however, fall short for many use
cases including:

|                            | mow.cli | docopt |
|----------------------------|---------|--------|
| Contextual help            | ✓       |        |
| Backtracking `SRC... DST`  | ✓       |        |
| Backtracking `[SRC] DST`   | ✓       |        |
| Branching `(SRC ❘ -f DST)` | ✓       |        |


## Installation
To install this package, run the following:

```shell
go get github.com/jawher/mow.cli
```

# Package Documentation

<!-- Do NOT edit past here. This is replaced by the contents of the package documentation -->
Package cli provides a framework to build command line applications in Go with
most of the burden of arguments parsing and validation placed on the framework
instead of the user.

## Basics
To create a new application, initialize an app with cli.App. Specify a name and
a brief description for the application:

```
cp := cli.App("cp", "Copy files around")
```

To attach code to execute when the app is launched, assign a function to the
Action field:

```
cp.Action = func() {
    fmt.Printf("Hello world\n")
}
```

To assign a version to the application, use Version method and specify the flags
that will be used to invoke the version command:

```
cp.Version("v version", "cp 1.2.3")
```

Finally, in the main func, call Run passing in the arguments for parsing:

```
cp.Run(os.Args)
```

## Options
To add one or more command line options (also known as flags), use one of the
short-form StringOpt, StringsOpt, IntOpt, IntsOpt, Float64Opt, Floats64Opt, or BoolOpt methods on App (or
Cmd if adding flags to a command or a subcommand). For example, to add a boolean
flag to the cp command that specifies recursive mode, use the following:

```
recursive := cp.BoolOpt("R recursive", false, "recursively copy the src to dst")
```

or:

```
cp.BoolOptPtr(&cfg.recursive, "R recursive", false, "recursively copy the src to dst")
```

The first version returns a new pointer to a bool value which will be populated when the app is run,
whereas the second version will populate a pointer to an existing variable you specify.

The option name(s) is a space separated list of names (without the
dashes). The one letter names can then be called with a single dash (short option, -R), the others with two dashes (long options, --recursive).

You also specify the default value for the option if it is not supplied by the user.

The last parameter is the description to be shown in help messages.

There is also a second set of methods on App called String, Strings, Int, Ints,
and Bool, which accept a long-form struct of the type: cli.StringOpt,
cli.StringsOpt, cli.IntOpt, cli.IntsOpt, cli.Float64Opt, cli.Floats64Opt, cli.BoolOpt. The struct describes the
option and allows the use of additional features not available in the short-form
methods described above:

```
recursive = cp.Bool(cli.BoolOpt{
    Name:       "R recursive",
    Value:      false,
    Desc:       "copy src files recursively",
    EnvVar:     "VAR_RECURSIVE",
    SetByUser:  &recursiveSetByUser,
})
```

Or:

```
recursive = cp.BoolPtr(&recursive, cli.BoolOpt{
    Name:       "R recursive",
    Value:      false,
    Desc:       "copy src files recursively",
    EnvVar:     "VAR_RECURSIVE",
    SetByUser:  &recursiveSetByUser,
})
```

The first version returns a new pointer to a value which will be populated when the app is run,
whereas the second version will populate a pointer to an existing variable you specify.

Two features, EnvVar and SetByUser, can be defined in the long-form struct
method. EnvVar is a space separated list of environment variables used to
initialize the option if a value is not provided by the user. When help messages
are shown, the value of any environment variables will be displayed. SetByUser
is a pointer to a boolean variable that is set to true if the user specified the
value on the command line. This can be useful to determine if the value of the
option was explicitly set by the user or set via the default value.

You can only access the values stored in the pointers in the Action func, which is invoked after
argument parsing has been completed. This precludes using the value of one
option as the default value of another.

On the command line, the following syntaxes are supported when specifying
options.

Boolean options:

```
-f         single dash one letter name
-f=false   single dash one letter name, equal sign followed by true or false
--force    double dash for longer option names
-it        single dash for multiple one letter names (option folding), this is equivalent to: -i -t
```

String, int and float options:

```
-e=value       single dash one letter name, equal sign, followed by the value
-e value       single dash one letter name, space followed by the value
-Ivalue        single dash one letter name, immediately followed by the value
--extra=value  double dash for longer option names, equal sign followed by the value
--extra value  double dash for longer option names, space followed by the value
```

Slice options (StringsOpt, IntsOpt, Floats64Opt) where option is repeated to accumulate
values in a slice:

```
-e PATH:/bin    -e PATH:/usr/bin     resulting slice contains ["/bin", "/usr/bin"]
-ePATH:/bin     -ePATH:/usr/bin      resulting slice contains ["/bin", "/usr/bin"]
-e=PATH:/bin    -e=PATH:/usr/bin     resulting slice contains ["/bin", "/usr/bin"]
--env PATH:/bin --env PATH:/usr/bin  resulting slice contains ["/bin", "/usr/bin"]
--env=PATH:/bin --env=PATH:/usr/bin  resulting slice contains ["/bin", "/usr/bin"]
```

## Arguments
To add one or more command line arguments (not prefixed by dashes), use one of
the short-form StringArg, StringsArg, IntArg, IntsArg, Float64Arg, Floats64Arg, or BoolArg methods on App
(or Cmd if adding arguments to a command or subcommand). For example, to add two
string arguments to our cp command, use the following calls:

```
src := cp.StringArg("SRC", "", "the file to copy")
dst := cp.StringArg("DST", "", "the destination")
```

Or:

```
cp.StringArgPtr(&src, "SRC", "", "the file to copy")
cp.StringArgPtr(&dst, "DST", "", "the destination")
```

The first version returns a new pointer to a value which will be populated when the app is run,
whereas the second version will populate a pointer to an existing variable you specify.

You then specify the argument as will be displayed in help messages.
Argument names must be specified as all uppercase.  The next parameter is the
default value for the argument if it is not supplied. And the last is
the description to be shown in help messages.

There is also a second set of methods on App called String, Strings, Int, Ints,
Float64, Floats64 and Bool, which accept a long-form struct of the type: cli.StringArg,
cli.StringsArg, cli.IntArg, cli.IntsArg, cli.BoolArg. The struct describes the
arguments and allows the use of additional features not available in the
short-form methods described above:

```
src = cp.Strings(StringsArg{
    Name:      "SRC",
    Desc:      "The source files to copy",
    Value:     "default value",
    EnvVar:    "VAR1 VAR2",
    SetByUser: &srcSetByUser,
})
```

Or:

```
src = cp.StringsPtr(&src, StringsArg{
    Name:      "SRC",
    Desc:      "The source files to copy",
    Value:     "default value",
    EnvVar:    "VAR1 VAR2",
    SetByUser: &srcSetByUser,
})
```

The first version returns a new pointer to a value which will be populated when the app is run,
whereas the second version will populate a pointer to an existing variable you specify.

Two features, EnvVar and SetByUser, can be defined in the long-form struct
method. EnvVar is a space separated list of environment variables used to
initialize the argument if a value is not provided by the user. When help
messages are shown, the value of any environment variables will be displayed.
SetByUser is a pointer to a boolean variable that is set to true if the user
specified the value on the command line. This can be useful to determine if the
value of the argument was explicitly set by the user or set via the default
value.

You can only access the values stored in the pointers in the Action func, which is invoked after
argument parsing has been completed. This precludes using the value of one
argument as the default value of another.

## Operators
The -- operator marks the end of command line options. Everything that follows
will be treated as an argument, even if starts with a dash.  For example, the
standard POSIX touch command, which takes a filename as an argument (and
possibly other options that we'll ignore here), could be defined as:

```
file := cp.StringArg("FILE", "", "the file to create")
```

If we try to create a file named "-f" via our touch command:

```
$ touch -f
```

It will fail because the -f will be parsed as an option, not as an argument. The
fix is to insert -- after all flags have been specified, so the remaining
arguments are parsed as arguments instead of options as follows:

```
$ touch -- -f
```

This ensures the -f is parsed as an argument instead of a flag named f.

## Commands
This package supports nesting of commands and subcommands. Declare a top-level
command by calling the Command func on the top-level App struct. For example,
the following creates an application called docker that will have one command
called run:

```
docker := cli.App("docker", "A self-sufficient runtime for linux containers")

docker.Command("run", "Run a command in a new container", func(cmd *cli.Cmd) {
    // initialize the run command here
})
```

The first argument is the name of the command the user will specify on the
command line to invoke this command.  The second argument is the description of
the command shown in help messages.  And, the last argument is a CmdInitializer,
which is a function that receives a pointer to a Cmd struct representing the
command.

Within this function, define the options and arguments for the command by
calling the same methods as you would with top-level App struct (BoolOpt,
StringArg, ...).  To execute code when the command is invoked, assign a function
to the Action field of the Cmd struct. Within that function, you can safely
refer to the options and arguments as command line parsing will be completed at
the time the function is invoked:

```
docker.Command("run", "Run a command in a new container", func(cmd *cli.Cmd) {
    var (
        detached = cmd.BoolOpt("d detach", false, "Run container in background")
        memory   = cmd.StringOpt("m memory", "", "Set memory limit")
        image    = cmd.StringArg("IMAGE", "", "The image to run")
    )

    cmd.Action = func() {
        if *detached {
            // do something
        }
        runContainer(*image, *detached, *memory)
    }
})
```

Optionally, to provide a more extensive description of the command, assign a
string to LongDesc, which is displayed when a user invokes --help. A LongDesc
can be provided for Cmds as well as the top-level App:

```
cmd.LongDesc = `Run a command in a new container

With the docker run command, an operator can add to or override the
image defaults set by a developer. And, additionally, operators can
override nearly all the defaults set by the Docker runtime itself.
The operator’s ability to override image and Docker runtime defaults
is why run has more options than any other docker command.`
```

Subcommands can be added by calling Command on the Cmd struct. They can by
defined to any depth if needed:

```
docker.Command("job", "actions on jobs", func(job *cli.Cmd) {
    job.Command("list", "list jobs", listJobs)
    job.Command("start", "start a new job", startJob)
    job.Command("log", "log commands", func(log *cli.Cmd) {
        log.Command("show", "show logs", showLog)
        log.Command("clear", "clear logs", clearLog)
    })
})
```

Command and subcommand aliases are also supported. To define one or more
aliases, specify a space-separated list of strings to the first argument of
Command:

```
job.Command("start run r", "start a new job", startJob)
```

With the command structure defined above, users can invoke the app in a variety
of ways:

```
$ docker job list
$ docker job start
$ docker job run   # using the alias we defined
$ docker job r     # using the alias we defined
$ docker job log show
$ docker job log clear
```

As a convenience, to assign an Action to a func with no arguments, use
ActionCommand when defining the Command. For example, the following two
statements are equivalent:

```
app.Command("list", "list all configs", cli.ActionCommand(list))

// Exactly the same as above, just more verbose
app.Command("list", "list all configs", func(cmd *cli.Cmd)) {
    cmd.Action = func() {
        list()
    }
}
```

Please note that options, arguments, specs, and long descriptions cannot be
provided when using ActionCommand. This is intended for very simple command
invocations that take no arguments.

Finally, as a side-note, it may seem a bit weird that this package uses a
function to initialize a command instead of simply returning a command struct.
The motivation behind this API decision is scoping: as with the standard flag
package, adding an option or an argument returns a pointer to a value which will
be populated when the app is run.  Since you'll want to store these pointers in
variables, and to avoid having dozens of them in the same scope (the main func
for example or as global variables), this API was specifically tailored to take
a func parameter (called CmdInitializer), which accepts the command struct. With
this design, the command's specific variables are limited in scope to this
function.

## Interceptors
Interceptors, or hooks, can be defined to be executed before and after a command
or when any of its subcommands are executed.  For example, the following app
defines multiple commands as well as a global flag which toggles verbosity:

```
app := cli.App("app", "bla bla")
verbose := app.BoolOpt("verbose v", false, "Enable debug logs")

app.Command("command1", "...", func(cmd *cli.Cmd) {
    if (*verbose) {
        logrus.SetLevel(logrus.DebugLevel)
    }
})

app.Command("command2", "...", func(cmd *cli.Cmd) {
    if (*verbose) {
        logrus.SetLevel(logrus.DebugLevel)
    }
})
```

Instead of duplicating the check for the verbose flag and setting the debug
level in every command (and its sub-commands), a Before interceptor can be set
on the top-level App instead:

```
app.Before = func() {
    if (*verbose) {
        logrus.SetLevel(logrus.DebugLevel)
    }
}
```

Whenever a valid command is called by the user, all the Before interceptors
defined on the app and the intermediate commands will be called, in order from
the root to the leaf.

Similarly, to execute a hook after a command has been called, e.g. to cleanup
resources allocated in Before interceptors, simply set the After field of the
App struct or any other Command. After interceptors will be called, in order,
from the leaf up to the root (the opposite order of the Before interceptors).

The following diagram shows when and in which order multiple Before and After
interceptors are executed:

```
+------------+    success    +------------+   success   +----------------+     success
| app.Before +---------------> cmd.Before +-------------> sub_cmd.Before +---------+
+------------+               +-+----------+             +--+-------------+         |
                               |                           |                     +-v-------+
                 error         |           error           |                     | sub_cmd |
       +-----------------------+   +-----------------------+                     | Action  |
       |                           |                                             +-+-------+
+------v-----+               +-----v------+             +----------------+         |
| app.After  <---------------+ cmd.After  <-------------+  sub_cmd.After <---------+
+------------+    always     +------------+    always   +----------------+      always
```

## Exiting
To exit the application, use cli.Exit function, which accepts an exit code and
exits the app with the provided code.  It is important to use cli.Exit instead
of os.Exit as the former ensures that all of the After interceptors are executed
before exiting.

```
cli.Exit(1)
```

## Spec Strings
An App or Command's invocation syntax can be customized using spec strings. This
can be useful to indicate that an argument is optional or that two options are
mutually exclusive.  The spec string is one of the key differentiators between
this package and other CLI packages as it allows the developer to express usage
in a simple, familiar, yet concise grammar.

To define option and argument usage for the top-level App, assign a spec string
to the App's Spec field:

```
cp := cli.App("cp", "Copy files around")
cp.Spec = "[-R [-H | -L | -P]]"
```

Likewise, to define option and argument usage for a command or subcommand,
assign a spec string to the Command's Spec field:

```
docker := cli.App("docker", "A self-sufficient runtime for linux containers")
docker.Command("run", "Run a command in a new container", func(cmd *cli.Cmd) {
    cmd.Spec = "[-d|--rm] IMAGE [COMMAND [ARG...]]"
    :
    :
}
```

The spec syntax is mostly based on the conventions used in POSIX command line
applications (help messages and man pages). This syntax is described in full
below. If a user invokes the app or command with the incorrect syntax, the app
terminates with a help message showing the proper invocation. The remainder of
this section describes the many features and capabilities of the spec string
grammar.

Options can use both short and long option names in spec strings.  In the
example below, the option is mandatory and must be provided.  Any options
referenced in a spec string MUST be explicitly declared, otherwise this package
will panic. I.e. for each item in the spec string, a corresponding *Opt or *Arg
is required:

```
x.Spec = "-f"  // or x.Spec = "--force"
forceFlag := x.BoolOpt("f force", ...)
```

Arguments are specified with all-uppercased words.  In the example below, both
SRC and DST must be provided by the user (two arguments).  Like options, any
argument referenced in a spec string MUST be explicitly declared, otherwise this
package will panic:

```
x.Spec="SRC DST"
src := x.StringArg("SRC", ...)
dst := x.StringArg("DST", ...)
```

With the exception of options, the order of the elements in a spec string is
respected and enforced when command line arguments are parsed.  In the example
below, consecutive options (-f and -g) are parsed regardless of the order they
are specified (both "-f=5 -g=6" and "-g=6 -f=5" are valid).  Order between
options and arguments is significant (-f and -g must appear before the SRC
argument). The same holds true for arguments, where SRC must appear before DST:

```
x.Spec = "-f -g SRC -h DST"
var (
    factor = x.IntOpt("f", 1, "Fun factor (1-5)")
    games  = x.IntOpt("g", 1, "# of games")
    health = x.IntOpt("h", 1, "# of hosts")
    src    = x.StringArg("SRC", ...)
    dst    = x.StringArg("DST", ...)
)
```

Optionality of options and arguments is specified in a spec string by enclosing
the item in square brackets []. If the user does not provide an optional value,
the app will use the default value specified when the argument was defined. In
the example below, if -x is not provided, heapSize will default to 1024:

```
x.Spec = "[-x]"
heapSize := x.IntOpt("x", 1024, "Heap size in MB")
```

Choice between two or more items is specified in a spec string by separating
each choice with the | operator. Choices are mutually exclusive. In the examples
below, only a single choice can be provided by the user otherwise the app will
terminate displaying a help message on proper usage:

```
x.Spec = "--rm | --daemon"
x.Spec = "-H | -L | -P"
x.Spec = "-t | DST"
```

Repetition of options and arguments is specified in a spec string with the ...
postfix operator to mark an item as repeatable. Both options and arguments
support repitition. In the example below, users may invoke the command with
multiple -e options and multiple SRC arguments:

```
x.Spec = "-e... SRC..."

// Allows parsing of the following shell command:
//   $ app -eeeee file1 file2
//   $ app -e -e -e -e file1 file2
```

Grouping of options and arguments is specified in a spec string with
parenthesis.  When combined with the choice | and repetition ... operators,
complex syntaxes can be created. The parenthesis in the example below indicate a
repeatable sequence of a -e option followed by an argument, and that is mutually
exclusive to a choice between -x and -y options.

```
x.Spec = "(-e COMMAND)... | (-x|-y)"

// Allows parsing of the following shell command:
//   $ app -e show -e add
//   $ app -y
// But not the following:
//   $ app -e show -x
```

Option groups, or option folding, are a shorthand method to declaring a choice
between multiple options.  I.e. any combination of the listed options in any
order with at least one option selected. The following two statements are
equivalent:

```
x.Spec = "-abcd"
x.Spec = "(-a | -b | -c | -d)..."
```

Option groups are typically used in conjunction with optionality [] operators.
I.e. any combination of the listed options in any order or none at all. The
following two statements are equivalent:

```
x.Spec = "[-abcd]"
x.Spec = "[-a | -b | -c | -d]..."
```

All of the options can be specified using a special syntax: [OPTIONS]. This is a
special token in the spec string (not optionality and not an argument called
OPTIONS). It is equivalent to an optional repeatable choice between all the
available options. For example, if an app or a command declares 4 options a, b,
c and d, then the following two statements are equivalent:

```
x.Spec = "[OPTIONS]"
x.Spec = "[-a | -b | -c | -d]..."
```

Inline option values are specified in the spec string with the =<some-text>
notation immediately following an option (long or short form) to provide users
with an inline description or value. The actual inline values are ignored by the
spec parser as they exist only to provide a contextual hint to the user. In the
example below, "absolute-path" and "in seconds" are ignored by the parser:

```
x.Spec = "[ -a=<absolute-path> | --timeout=<in seconds> ] ARG"
```

The -- operator can be used to automatically treat everything following it as
arguments.  In other words, placing a -- in the spec string automatically
inserts a -- in the same position in the program call arguments. This lets you
write programs such as the POSIX time utility for example:

```
x.Spec = "-lp [-- CMD [ARG...]]"

// Allows parsing of the following shell command:
//   $ app -p ps -aux
```

## Spec Grammar
Below is the full EBNF grammar for the Specs language:

```
spec         -> sequence
sequence     -> choice*
req_sequence -> choice+
choice       -> atom ('|' atom)*
atom         -> (shortOpt | longOpt | optSeq | allOpts | group | optional) rep?
shortOp      -> '-' [A-Za-z]
longOpt      -> '--' [A-Za-z][A-Za-z0-9]*
optSeq       -> '-' [A-Za-z]+
allOpts      -> '[OPTIONS]'
group        -> '(' req_sequence ')'
optional     -> '[' req_sequence ']'
rep          -> '...'
```

By combining a few of these building blocks together (while respecting the
grammar above), powerful and sophisticated validation constraints can be created
in a simple and concise manner without having to define in code. This is one of
the key differentiators between this package and other CLI packages. Validation
of usage is handled entirely by the package through the spec string.

Behind the scenes, this package parses the spec string and constructs a finite
state machine used to parse the command line arguments. It also handles
backtracking, which allows it to handle tricky cases, or what I like to call
"the cp test":

```
cp SRC... DST
```

Without backtracking, this deceptively simple spec string cannot be parsed
correctly. For instance, docopt can't handle this case, whereas this package
does.

## Default Spec
By default an auto-generated spec string is created for the app and every
command unless a spec string has been set by the user.  This can simplify use of
the package even further for simple syntaxes.

The following logic is used to create an auto-generated spec string: 1) start
with an empty spec string, 2) if at least one option was declared, append
"[OPTIONS]" to the spec string, and 3) for each declared argument, append it, in
the order of declaration, to the spec string. For example, given this command
declaration:

```
docker.Command("run", "Run a command in a new container", func(cmd *cli.Cmd) {
    var (
        detached = cmd.BoolOpt("d detach", false, "Run container in background")
        memory   = cmd.StringOpt("m memory", "", "Set memory limit")
        image    = cmd.StringArg("IMAGE", "", "The image to run")
        args     = cmd.StringsArg("ARG", nil, "Arguments")
    )
})
```

The auto-generated spec string, which should suffice for simple cases, would be:

```
[OPTIONS] IMAGE ARG
```

If additional constraints are required, the spec string must be set explicitly
using the grammar documented above.

## Custom Types
By default, the following types are supported for options and arguments: bool,
string, int, float64, strings (slice of strings), ints (slice of ints) and floats64 (slice of float64).
You can, however, extend this package to handle other types, e.g. time.Duration, float64,
or even your own struct types.

To define your own custom type, you must implement the flag.Value interface for
your custom type, and then declare the option or argument using VarOpt or VarArg
respectively if using the short-form methods. If using the long-form struct,
then use Var instead.

The following example defines a custom type for a duration. It defines a
duration argument that users will be able to invoke with strings in the form of
"1h31m42s":

```
// Declare your type
type Duration time.Duration

// Make it implement flag.Value
func (d *Duration) Set(v string) error {
    parsed, err := time.ParseDuration(v)
    if err != nil {
        return err
    }
    *d = Duration(parsed)
    return nil
}

func (d *Duration) String() string {
    duration := time.Duration(*d)
    return duration.String()
}

func main() {
    duration := Duration(0)
    app := App("var", "")
    app.VarArg("DURATION", &duration, "")
    app.Run([]string{"cp", "1h31m42s"})
}
```

To make a custom type to behave as a boolean option, i.e. doesn't take a value,
it must implement the IsBoolFlag method that returns true:

```
type BoolLike int

func (d *BoolLike) IsBoolFlag() bool {
    return true
}
```

To make a custom type behave as a multi-valued option or argument, i.e. takes
multiple values, it must implement the Clear method, which is called whenever
the values list needs to be cleared, e.g. when the value was initially populated
from an environment variable, and then explicitly set from the CLI:

```
type Durations []time.Duration

// Make it implement flag.Value
func (d *Durations) Set(v string) error {
    parsed, err := time.ParseDuration(v)
    if err != nil {
        return err
    }
    *d = append(*d, Duration(parsed))
    return nil
}

func (d *Durations) String() string {
    return fmt.Sprintf("%v", *d)
}

// Make it multi-valued
func (d *Durations) Clear() {
    *d = []Duration{}
}
```

To hide the default value of a custom type, it must implement the IsDefault
method that returns a boolean. The help message generator will use the return
value to decide whether or not to display the default value to users:

```
type Action string

func (a *Action) IsDefault() bool {
    return (*a) == "nop"
}
```





## License
This work is published under the MIT license.

Please see the `LICENSE` file for details.

* * *
Automatically generated by [autoreadme](https://github.com/jimmyfrasche/autoreadme) on 2019.02.24
