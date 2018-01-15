# cmdhook

```
Put example code here
```

## Introduction

`cmdhook` is useful for running a docker `CMD` along with a few execution
hooks. The hooks are defined by these environment variables:

* `PRESTART`
* `POSTSTART`
* `POSTSTOP`
* `SIGTERM`

The program to be run is determined by the arguments passed to `cmdhook`:

```
cmdhook <mainprogram> [<arg> ...]
```

The first argument provides the location of program, and all remaining arguments
will be passed to it:

```
cmdhook echo "I am Captain Hook"
```

```
CMD ["cmdhook", "ping", "-i", "5", "8.8.8.8"]
```


The `cmdhook` program operates as `pid 1` and starts the requested program as a
child process.

## Hooks

Each hook is an environment variable containing a program path and its
arguments:

```
HOOK="<hookprogram> [<arg> ...]"
```

This facilitates easy invocation of individual programs and of the shell, as in
this `Dockerfile` snippet:

```
ENV PRESTART  /bin/sh -c echo "Performing pre-run setup..." && sleep 2s && echo "done"
ENV POSTSTART /bin/sh -c echo "Main process started"
ENV SIGTERM   /bin/sh -c echo "Termination signal received"
ENV POSTSTOP  /bin/sh -c echo "Main process stopped"
```

All hooks are optional. If a hook's environment variable is absent or blank, it
will be ignored.

The arguments supplied to a hook are parsed according to the
[posix shell parsing rules](http://pubs.opengroup.org/onlinepubs/9699919799/utilities/V3_chap02.html#tag_18_02). This means that `"`, `'`, `\` and whitespace characters are parsed in the
same manner as bash:

```
ENV POSTSTOP echo one "two two" 'three three' four\ four
```

All hooks are environment variables, and therefore can be overridden:

```
INSTANCE=hal9000
POSTSTOP="echo '$INSTANCE is done with you.' | mail -s 'Execution complete' user@example.com"

docker run -e="POSTSTOP=$POSTSTOP" example/container
```

The `stdout` and `stderr` streams of each hook will be passed
through to the `stdout` and `stderr` streams of the docker instance. This
allows their output to be visible in the docker logs. However, this may result
in the output of the `POSTSTART` and `SIGTERM` hooks being intermingled with
that of the main program.

To minimize this intermingling, the hook's output is buffered until the hook
has exited, then the output is dumped to the appropriate stream all at once.

### PRESTART

When supplied, the `PRESTART` hook will be executed before the main program.
This can be used to perform pre-launch logging, container setup and sanity
checks.

If the `PRESTART` hook returns a non-zero exit status, it will prevent the
main program from running. When this happens its exit status will be returned
by `cmdhook`. `PRESTART` is the only hook that may block execution in this way.

### POSTSTART

When supplied, the `POSTSTART` hook will be executed immediately after the main
program has been started. This can be used to perform post-launch logging,
container setup, and any actions that expect the main program to be
running.

It is important to note that `POSTSTART` executes independently of the main
program. As such, it presents some potential race conditions:

* The main program may not have completed its initialization yet
* The main program may have executed quickly and be stopping or stopped
* The main program may have encountered an error and already exited

Some use cases will not be concerned with these conditions. Others will require a
best effort at detecting and handling them.

### POSTSTOP

When supplied, the `POSTSTOP` hook will be executed immediately after the main
program has exited. This can be used to perform post-run logging and
container cleanup.

### SIGTERM

When supplied, the `SIGTERM` hook will be executed each time a `SIGINT` or
`SIGTERM` signal is received. This can be used to perform a graceful shutdown
of a program that ignores `SIGTERM` or terminates ungracefully in
its presence.

A `SIGTERM` hook can interfere with the signals that are passed to the main
program. It exerts this influence through its exit code:

* zero: The signal is blocked
* non-zero: The signal is not blocked

*TODO: Provide a way of overriding which signal is sent to the main program.*

## Security

Hooks will be executed blindly by `cmdhook`. Please be mindful of this when
deciding what goes into them. Don't allow them to be populated with untrusted
commands.

Supplying commands through environment variables may catch administrators off
guard. Provide them proper documentation.

Environment variables may receive less scrutiny than command line arguments
during security audits. Give them their due attention.

Please be careful when generating hooks via script.

It's probably a bad idea to include any sort of user-provided data in a hook.
If you absolutely must do so, please be extremely cautious and sanitize the
crap out of your inputs.
