# scotty

<p align="center">
    <img src="resources/gopher-scotty.png" alt="scotty gopher :)" width="150px" height="150px"></img>
</p>

> Multiplex and query your logs during development

## Idea behind ***scotty*** 

`scotty` aims to improve the process of understanding logs and tracing request or span ids while developing services locally. With `scotty` and its sub command `beam` you can multiplex many log streams into a consolidated view on which you can apply aggregations and filters. Thereby, you no longer need to manually stich together logs from multiple terminal windows.

As a secondary goal `scotty` tries parsing your logs (if they are structured) into JSON helping the readability while browsing the logs.

By using *unix pipes* to pipe your program output into `beam` you are free to append commands prior to calling `beam`. By default `beam` will try to connect to a *unix socket* however, since applications nowadays usually are shipped within replicated environments (such as ***docker***) `scotty` will also allow you to beam **sorry stream** your logs via a `tcp:ip` connection. 


## Installation

### Install with Homebrew (not available yet :( will be done in v0.0.1)

### Go installed? Install from source

`go install github.com/KonstantinGasser/scotty@latest`

## Beaming logs to scotty

`scotty` can read any logs/messages piped to the `os.Stdin` of `beam`. Just be aware that application log
are not necessarily printed to `os.Stdout` but usually to `os.Stderr`. Therefore you might need to redirect `os.Stderr` to `os.Stdout` in order for `beam` to receive and read the data from `os.Stdin`.<br>
Redirecting the output can be done as such: `2>&1`

***Example reading from os.Stdout:***<br>
When using `cat` data is printed to `os.Stdout` as such no redirect is required and in order to receive the data in `scotty` you can use a standard pipe like so: `cat engine.log | beam` which will pipe the engine logs line by line to scotty. 

***Example redirect to os.Stdout:***<br>
Beaming application logs might requires to redirect the log output from `os.Stderr` to `os.Stdout` first before they can be piped to `beam`. Say your application is a python application your command could look like so: `python3 my/application.py -some flag -second flag 2>&1 | beam`. 
<br><br>
***Note:*** By default `beam` will print send logs also to `os.Stdout` similar to the `tee` command. If you want to suppress the output you can run `beam` with the `-d` flag (`beam -d`)

## Streams
In scotty a stream is any command/application which pipes data via `beam` to `scotty`. In order to differentiate between logs from different streams you can provide a label per stream. The label is a flag on the `beam` command (`beam -label=engine-service`). It is recommended to reuse the same label after re-connecting a beam to scotty as scotty maintains a state of color codes per stream to help with readability. Changing the stream label has no downside besides the fact that scotty assigns a new color to the stream. The state is discarded whenever `scotty` is stopped.

## Querying your logs
yet to be implemented
### View filtered log streams
yet to be implemented
### Aggregate log streams
yet to be implemented
