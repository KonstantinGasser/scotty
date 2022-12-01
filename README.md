# scotty

<div style="text-align: center">
    <img src="resources/gopher-scotty.png" alt="scotty gopher :)" width="250px" height="250px"></img>
</div>

> Multiplex log streams for aggregation and analysis while developing local services

## Draft 

Often enough in order to develop applications in your local development you are required to run more than one service. Let's say one service dealing with authentication and another service taking care about user data. If everything is working - great why do we need logs in the first place?😆...but if the login fails what happens, you have to trace through the logs to find the issue. However, since you have two services you need to stich together the logs from the terminal windows; yak 😒.

Scotty's goal is to make this process of tracing logs in local environments easier and similar to what developers are used to in production environments. At least to funnel all logs into one output and allow for filtering yay even tracing of an id. As a side goal Scotty aims to provide better log output for logs which can be represented as JSON making it again easier and quicker to find the root cause of the issue.