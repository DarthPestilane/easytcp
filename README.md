# EasyTCP

[![Run Actions](https://github.com/DarthPestilane/easytcp/actions/workflows/actions.yml/badge.svg?branch=master&event=push)](https://github.com/DarthPestilane/easytcp/actions/workflows/actions.yml)
[![codecov](https://codecov.io/gh/DarthPestilane/easytcp/branch/master/graph/badge.svg?token=002KJ5IV4Z)](https://codecov.io/gh/DarthPestilane/easytcp)

## Introduction

`EasyTCP` is a light-weight TCP framework written in Go (Golang), helps you build a TCP server easily and fast.

## Install

To install `EasyTCP` package, you need to install Go and set your Go workspace first.

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
		fmt.Printf("message received <<< id: %d; size: %d; data: %s\n", ctx.MsgID(), ctx.MsgSize(), ctx.MsgRawData())
		return ctx.Response(uint(1002), []byte("copy that"))
	})

	// listen and serve
	if err := s.Serve(":5896"); err != nil && err != server.ErrServerStopped {
		fmt.Println("serve error: ", err.Error())
	}
}
```

## API

### Architecture

### Sever

### Packer

### Codec

### Routing

## Contribute

Check out a new branch for the job, and make sure git action passed.
