# scotty

<p align="center">
    <img src="resources/gopher-scotty.png" alt="scotty gopher :)" width="250px" height="250px"></img>
</p>

> Multiplex log streams for aggregation and analysis while developing local services

## Draft 

`scotty` aims to improve the process of understanding logs and tracing request or span ids while developing services locally. With `scotty` and its sub command `beam` you can multiplex many log streams into a consolidated view on which you can apply aggregations and filters. Thereby, you no longer need to manually stich together logs from multiple terminal windows.

As a secondary goal `scotty` tries parsing your logs (if they are structured) into JSON helping the readability while browsing the logs.

By using *unix pipes* to pipe your program output into `beam` you are free to append commands prior to calling `beam`. By default `beam` will try to connect to a *unix socket* however, since applications nowadays usually are shipped within replicated environments (such as ***docker***) `scotty` will also allow you to beam **sorry stream** your logs via a `tcp:ip` connection. 

