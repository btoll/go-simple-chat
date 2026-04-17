# go-simple-chat

## Use

```bash
$ ./go-simple-chat
Simple chat server started and listening on :9999
```

```bash
$ nc localhost 9999
What's your name?: Turd Ferguson
Welcome to the the simple chat server, Turd Ferguson!

```

That's it.  Simple.

## Test

If you're going to test the shutdown sequence by sending the program a `SIGINT` or a `SIGTERM`, make sure you're running a compiled binary and not starting the server using the go tooling (`go run`).  Importantly, the `PID` for `go run .` will NOT be the `PID` of the process that you want to target.

Why?  The process started by the tooling will be the parent of the actual program, which is spawned behind the scenes as a child of the `go run` process.  Building (`go build`) avoids this since no child is being spawned and the `PID` of the binary will be the one to target.

Observe:

```bash
$ go run .
Simple chat server started and listening on :9999
```

In another terminal:

```bash
$ pgrep -a go
3211 /usr/libexec/goa-daemon
3235 /usr/libexec/goa-identity-service
133165 /home/btoll/.vim/plugged/YouCompleteMe/third_party/ycmd/third_party/go/bin/gopls -logfile /tmp/gopls_stderry6gj9qz5.log
133177 /home/btoll/.vim/plugged/YouCompleteMe/third_party/ycmd/third_party/go/bin/gopls ** telemetry **
144213 /home/btoll/.vim/plugged/YouCompleteMe/third_party/ycmd/third_party/go/bin/gopls -logfile /tmp/gopls_stderr978h9ehq.log
144225 /home/btoll/.vim/plugged/YouCompleteMe/third_party/ycmd/third_party/go/bin/gopls ** telemetry **
145009 /home/btoll/.vim/plugged/YouCompleteMe/third_party/ycmd/third_party/go/bin/gopls -logfile /tmp/gopls_stderr8fsirvej.log
145021 /home/btoll/.vim/plugged/YouCompleteMe/third_party/ycmd/third_party/go/bin/gopls ** telemetry **
153052 go run .
153107 /tmp/go-build2605656490/b001/exe/go-simple-chat
```

Let's see the process tree of the `PID` of the `go run` command:

```bash
$ pstree -p 153052
go(153052)─┬─go-simple-chat(153107)─┬─{go-simple-chat}(153108)
           │                        ├─{go-simple-chat}(153109)
           │                        ├─{go-simple-chat}(153110)
           │                        ├─{go-simple-chat}(153111)
           │                        ├─{go-simple-chat}(153112)
           │                        ├─{go-simple-chat}(153113)
           │                        ├─{go-simple-chat}(153114)
           │                        └─{go-simple-chat}(153115)
           ├─{go}(153053)
           ├─{go}(153054)
           ├─{go}(153055)
           ├─{go}(153056)
           ├─{go}(153057)
           ├─{go}(153058)
           ├─{go}(153059)
           ├─{go}(153060)
           ├─{go}(153072)
           ├─{go}(153077)
           └─{go}(153106)
```

The `PID` to target is `153107`:

```bash
$ kill 153107
```

> Of course, you can also see this `PID` is the last in the list printed by the `pgrep` command above, but it's always good to know the details.

- `SIGINT`
    + `kill -INT 3200110`
    + `kill -SIGINT 3200110`
    + `kill -2 3200110`

- `SIGKILL`
    + `kill -KILL 3200110`
    + `kill -SIGKILL 3200110`
    + `kill -9 3200110`

- `SIGTERM`
    + `kill 3200110`
    + `kill -TERM 3200110`
    + `kill -SIGTERM 3200110`
    + `kill -15 3200110`

```bash
$ kill -l 11
SEGV
```

```bash
$ kill -L
 1) SIGHUP       2) SIGINT       3) SIGQUIT      4) SIGILL       5) SIGTRAP
 6) SIGABRT      7) SIGBUS       8) SIGFPE       9) SIGKILL     10) SIGUSR1
11) SIGSEGV     12) SIGUSR2     13) SIGPIPE     14) SIGALRM     15) SIGTERM
16) SIGSTKFLT   17) SIGCHLD     18) SIGCONT     19) SIGSTOP     20) SIGTSTP
21) SIGTTIN     22) SIGTTOU     23) SIGURG      24) SIGXCPU     25) SIGXFSZ
26) SIGVTALRM   27) SIGPROF     28) SIGWINCH    29) SIGIO       30) SIGPWR
31) SIGSYS      34) SIGRTMIN    35) SIGRTMIN+1  36) SIGRTMIN+2  37) SIGRTMIN+3
38) SIGRTMIN+4  39) SIGRTMIN+5  40) SIGRTMIN+6  41) SIGRTMIN+7  42) SIGRTMIN+8
43) SIGRTMIN+9  44) SIGRTMIN+10 45) SIGRTMIN+11 46) SIGRTMIN+12 47) SIGRTMIN+13
48) SIGRTMIN+14 49) SIGRTMIN+15 50) SIGRTMAX-14 51) SIGRTMAX-13 52) SIGRTMAX-12
53) SIGRTMAX-11 54) SIGRTMAX-10 55) SIGRTMAX-9  56) SIGRTMAX-8  57) SIGRTMAX-7
58) SIGRTMAX-6  59) SIGRTMAX-5  60) SIGRTMAX-4  61) SIGRTMAX-3  62) SIGRTMAX-2
63) SIGRTMAX-1  64) SIGRTMAX
```

## License

[GPLv3](COPYING)

## Author

[Benjamin Toll](https://benjamintoll.com)

