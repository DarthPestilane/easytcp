# EasyTCP

[![Run Actions](https://github.com/DarthPestilane/easytcp/actions/workflows/actions.yml/badge.svg?branch=master&event=push)](https://github.com/DarthPestilane/easytcp/actions/workflows/actions.yml)
[![codecov](https://codecov.io/gh/DarthPestilane/easytcp/branch/master/graph/badge.svg?token=002KJ5IV4Z)](https://codecov.io/gh/DarthPestilane/easytcp)

## Introduction

`EasyTCP` is a light-weight TCP framework written in Go (Golang), features with:

- Non-invasive design
- Pipelined middlewares for route handler
- Customizable message packer and codec
- Handy functions to handle request data and send response

`EasyTCP` helps you build a TCP server easily and fast.

## Install

This package, so far, has been tested in

- go1.14.x
- go1.15.x
- go1.16.x

Use the below Go command to install EasyTCP.

```sh
$ go get -u github.com/DarthPestilane/easytcp
```

## Quick start

```go
package main

import (
	"fmt"
	"github.com/DarthPestilane/easytcp"
	"github.com/DarthPestilane/easytcp/packet"
	"github.com/DarthPestilane/easytcp/router"
	"github.com/DarthPestilane/easytcp/server"
)

func main() {
	// create a new server
	s := easytcp.NewTCPServer(&server.TCPOption{})

	// add a route to message id
	s.AddRoute(uint(1001), func(ctx *router.Context) (packet.Message, error) {
		fmt.Printf("[server] request received | id: %d; size: %d; data: %s\n", ctx.MsgID(), ctx.MsgSize(), ctx.MsgRawData())
		return ctx.Response(uint(1002), []byte("copy that"))
	})

	// listen and serve
	if err := s.Serve(":5896"); err != nil && err != server.ErrServerStopped {
		fmt.Println("serve error: ", err.Error())
	}
}
```

There are more detailed examples in [examples/tcp](./examples/tcp)

## API

### Architecture

### Routing

EasyTCP considers every message has a `ID` segment.
A message will be routed, according to it's id, to the handler through middelwares.

#### Register a route

```go
s.AddRoute(reqID, func(ctx *router.Context) (packet.Message, error) {
	// handle the request via ctx
	fmt.Printf("[server] request received | id: %d; size: %d; data: %s\n", ctx.MsgID(), ctx.MsgSize(), ctx.MsgRawData())

	// do things...

	// return response
	return ctx.Response(respID, []byte("copy that"))
})
```

#### Using middleware

```go
// register global middlewares.
// global middlewares are priorer than per-route middlewares, they will be invoked first
s.Use(recoverMiddleware, logMiddleware, ...)

// register middlewares for one route
s.AddRoute(reqID, handler, middleware1, middleware2)

// a middleware looks like:
var exampleMiddleware router.MiddlewareFunc = func(next router.HandlerFunc) router.HandlerFunc {
	return func(ctx *router.Context) (resp packet.Message, err error) {
		// do things before...
		resp, err := next(ctx)
		// do things after...
		return resp, err
	}
}
```

### Packer

### Codec

## Contribute

Check out a new branch for the job, and make sure git action passed.
