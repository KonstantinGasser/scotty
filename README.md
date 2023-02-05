# beam me up, ***scotty***!

<p align="center">
    <img src="resources/gopher-scotty.png" alt="scotty gopher :)" width="150px" height="150px"></img>
</p>


# Why scotty?

Often times when you develop an application on your local system it's not enough to run a single application but maybe many different once.
The idea behind `scotty` originated from the resulting pain of having many terminal windows printing logs and stitching together the logs you
need in order to understand the bug you're searching for..tedious

With `scotty` you can multiplex your application logs into one consolidated terminal window apply filers on specific streams and query your logs (well once its implemented..working on it)

![](/resources/example_v0.0.1.png)
# Installation guide


## Homebrew
```
brew tab KonstantinGasser/scotty
brew install scotty
```

## From source
```
go install github.com/KonstantinGasser/scotty@latest
```

# How it works?

Somehow your logs need to be send or say beamed to `scotty`. This is why scotty comes with a helper command called `beam`.
Beam pushes everything it reads from stdin to scotty. Just be aware that things printed to stderr won't work..but we can
redirect `stderr` to `stdout` using `2>&1`. 

# Examples

## ...from stdout

```
cat uss_enterprise_engine.log | beam engine-service -d
```

This above command cats the `uss_enterprise_engine.log` to stdout which is then piped to the stdin of `beam`. Note the beam's first argument
will be the name referenced in scotty.

## ...from stderr

```
go run -race cmd/my/application.go 2>&1 | beam my-application
```

Here `application.go` produces logs printed to stderr this is why we need to add `2>&1` to redirect the output to stdout. The pipe to `beam` stays unchanged.

# Format a log line

Especially when logs are structured we humans have it hard to read the unformatted JSON. Hit the `:` key and type the line number of the log you want to format.

***Hint***: once the log line is displayed use the arrow keys for up and down (or `j`, `k`) to parse the previous or next line

# Options with beam

Currently `beam` only allows to pipe data through unix sockets..however `beam` as well as `scotty` are build such that both will support piping
data via a ***tcp:ip*** connection which enables you to beam logs from for example docker instances to `scotty` :)

# In case it fails, it fails!

Since `scotty` is still under active development panics within the program might happen. In such case you might need to delete the create ***unix socket* before restarting scotty. The default path for the socket scotty is using is `/tmp/scotty.unix`. This does not apply it you used `tcp:ip`