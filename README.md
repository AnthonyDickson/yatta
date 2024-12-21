# YATTA - Yet AnoTher Todo App

YATTA is a web app for managing your todo list. Yes... another one. What's
special about this todo app? Well first, _I_ made it. Also, it is my first
project in Go. Secondly, no React, Angular or Vue! In fact, you probably won't
see much JavaScript here, just a little bit of [HTMX](https://htmx.org/) and a
whole load of server-side rendering.

## Build and Run Server

```shell
go build
./yatta
```

You can access the web site via [localhost:8000](http://localhost:8000).
You can also use [air](https://github.com/air-verse/air) to auto-reload the
server and browser page when files are changed. Note that air is set up to
serve from [localhost:8080](http://localhost:8080) in [.air.toml](./.air.toml).

## Running tests

```shell
go test ./...
```

## Nix

There is a Nix [flake](./flake.nix) that provides a shell environment with the
commands `go` v1.23.3 and `air`. To activate the shell environment with the command:

```shell
nix develop -C $SHELL
```

## Name

YATTA comes from the Japanese phrase やった！(yatta), which means "hooray" or
literally "I did it!".
You know, the kind of thing you say when you check off an item on your todo list.
