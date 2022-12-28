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

## Loading logs
### With `cat logs.log | beam` output from the `stdout`

### With output from `stderr`


### Multiplexing logs streams from web-services

## Querying your logs

### View filtered log streams

### Aggregate log streams

