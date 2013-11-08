# tasktogo

`tasktogo` is a command-line todo list manager designed for both
interactive and script-based use. It has support for due dates and
varying priorities of tasks, and is capable of ranking tasks by
urgency, (which is calculated considering both the nearness of due
date and priority). For those who like rainbows, it can also colorize
output.

Primarily, it is designed for single users, rather than teams. Tasks
are represented in a simple [JSON][] format, and can be accessed
either directly (as a text file) or via `tasktogo`'s interfaces, which
are suitable for scripting.

It is written in pure Go.

![Demo screenshot][]

## Install

`tasktogo` is not yet packaged for any operating system or
distribution, but should be capable of functioning on any UNIX-alike,
and maybe even Windows.

If you have Go installed already, then you can do either of the
following. If not, you will need to [install it][Install Go] first.

For a full installation with manpage on UNIX-alikes:

```
git clone https://github.com/SashaCrofter/tasktogo.git
cd tasktogo
go get
make
sudo make install
```

For a source and binary-only install on any system with Go:

```
go install github.com/SashaCrofter/tasktogo.git
```

### Options

#### Changing install location

By default, the included Makefile installs `tasktogo` and its
resources (such as the manpage) to `/usr/local`, but this can be
changed by prepending `prefix=/usr` (for example) to the `make
install` command. In this case, then, the executable will go to
`/usr/bin` and so on.

It is worth nothing that one should not invoke with `prefix=` with no
argument, because then resources will not go to the correct directory.

#### Using a different compiler

The default compile command used is `go build`, with some flags used
to set the version automatically. If one wants to use a different
compiler or set of flags, one can prepend `GOCOMPILER=gccgo` (for
example) to the `make` command.

Note that in the compiled binary, the version number will only reflect
the lated tagged commit, rather than a verbose descriptor including
the current commit ID.

## Instructions

See the man page (if installed) via `man tasktogo`, or otherwise,
`tasktogo -h` will provide some help.

## Reporting bugs

`tasktogo` is hopelessly early-stage software, made primarily for its
author's use. Bug reports and feature requests are *greatly*
appreciated, and should both be made to the [issues page][], or to the
author, Alexander Bauer <sasha@crofter.org>.

When filing a bug report, please include the first line of output of
`tasktogo help`, containing the version number.

## License

`tasktogo` copyright © 2013 Alexander Bauer

License GPLv3+: GNU GPL version 3 or later <http://gnu.org/licenses/gpl.html>.  
This is free software: you are free to change and redistribute
it. There is *NO WARRANTY*, to the extent permitted by law.

[Go]: http://golang.org
[Install Go]: http://golang.org/install
[JSON]: http://json.org

[Issues page]: https://github.com/SashaCrofter/tasktogo/issues

[Demo Screenshot]: http://i.imgur.com/nSzJhyC.png "Any resemblance this bears to the todo list of any persons living or dead is purely coincidental."
