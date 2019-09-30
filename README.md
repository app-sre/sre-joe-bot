# sre-joe-bot

*Work-in-progress*

`sre-joe-bot` is a Slack bot used by Red Hat Application SREs

# Instructions

## Installing

    go get github.com/jfchevrette/sre-joe-bot

## Running

    sre-joe-bot


## Configuration

`sre-joe-bot` expects some environment variables to be defined. A sample of these variables can be found in the file `.env.example`

The app-interface integration variables are used by some bot commands to retrieve information from app-interface (https://github.com/app-sre/qontract-reconcile)


# TODO
- More commands to view data from app-interface
- Commands to launch actions against resources
