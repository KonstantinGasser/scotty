# beam me up, ***scotty***!

<p align="center">
    <img src="resources/gopher-scotty.png" alt="scotty gopher :)" width="150px" height="150px"></img>
</p>


# Why scotty?

Often times when you develop an application on your local system it's not enough to run a single application but maybe many different ones.
The idea behind `scotty` originated from the resulting pain of having many terminal windows printing logs and stitching together the logs you
need in order to understand the bug you're searching for..tedious

With `scotty` you can multiplex your application logs into ***one consolidated*** terminal window apply ***filterss*** on specific streams and ***format*** structred (JSON format) logs. In the future the secondary goal of scotty is it to allow you to ***query/aggregate*** your logs.

![](/resources/example_v0.0.2-rc.png)
# Installation guide


## Homebrew
```
brew tap KonstantinGasser/tap
brew install scotty beam
```

## From source
```
go install github.com/KonstantinGasser/scotty@v0.0.1
go install github.com/KonstantinGasser/beam@v0.0.1
```

# How it works?

Somehow your logs need to be send or say beamed to `scotty`. This is why scotty comes with a helper command called `beam`.
Beam pushes everything it reads from stdin to scotty. Just be aware that things printed to stderr won't work..but we can
redirect `stderr` to `stdout` using `2>&1`.

# How to use it?

## ...from stdout

```
cat uss_enterprise_engine.log | beam -d engine-service
```

This above command cats the `uss_enterprise_engine.log` to stdout which is then piped to the stdin of `beam`. Note the beam's first argument
will be the name referenced in scotty.

## ...from stderr

```
go run -race cmd/my/application.go 2>&1 | beam my-application
```

Here `application.go` produces logs printed to stderr this is why we need to add `2>&1` to redirect the output to stdout. The pipe to `beam` stays unchanged.

## Format a log line

Especially when logs are structured we humans have it hard to read the unformatted JSON. Hit the `:` key and type the line number of the log you want to format.

***Hint***: once the log line is displayed use the arrow keys for up and down (or `j`, `k`) to parse the previous or next line

## Filter on streams

When you need to only look at certain logs form a subset of streams your can apply a filter. Hit `ctrl+f` followed by a comma separated list of the streams you want to focus on.
While the filter is applied you can browse through the subset of logs as usual with `: <index>`.
To remove set filters hit the `q` key

## Beam information:

Once you pipe logs to scotty a `beam` is registered and displayed in the info componenten. In there right to the beam you see the number of logs processed by this beam and left of the beam name its state.

```
┌───────────────────────────────────┐
│ ● ping-svc: 2 - ◌ pong-svc: 4   │
└───────────────────────────────────┘
  _|_ ___|___ _|_  _______|_______
   1     2     3      next beam

// 1) status: ● -> connected, ◍ -> following paused, ◌ -> disconnected
// 2) beam name as provided by the beam command
// 3) processed log count
```

# Options with beam

Currently `beam` only allows to pipe data through unix sockets..however `beam` as well as `scotty` are build such that both will support piping
data via a ***tcp:ip*** connection which enables you to beam logs from for example docker instances to `scotty` :)

# In case it fails, it fails!

Since `scotty` is still under active development panics within the program might happen. In such case you might need to delete the create ***unix socket*** before restarting scotty. The default path for the socket scotty is using is `/tmp/scotty.sock`. This does not apply if you used `tcp:ip`
