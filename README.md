# EasyTCP

[![Run Actions](https://github.com/DarthPestilane/easytcp/actions/workflows/actions.yml/badge.svg?branch=master&event=push)](https://github.com/DarthPestilane/easytcp/actions/workflows/actions.yml)
[![codecov](https://codecov.io/gh/DarthPestilane/easytcp/branch/master/graph/badge.svg?token=002KJ5IV4Z)](https://codecov.io/gh/DarthPestilane/easytcp)

## Introduction

`EasyTCP` is a light-weight TCP framework written in Go (Golang), featured with:

- Non-invasive design
- Pipelined middlewares for route handler
- Customizable message packer and codec
- Handy functions to handle request data and send response

`EasyTCP` helps you build a TCP server easily and fast.

## Install

This package, so far, has been tested with

- go1.14.x
- go1.15.x
- go1.16.x

on the latest Linux, Macos and Windows.

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
	s.AddRoute(uint(1001), func(ctx *router.Context) (*packet.MessageEntry, error) {
		fmt.Printf("[server] request received | id: %d; size: %d; data: %s\n", ctx.MsgID(), ctx.MsgSize(), ctx.MsgData())
		return ctx.Response(uint(1002), []byte("copy that"))
	})

	// listen and serve
	if err := s.Serve(":5896"); err != nil && err != server.ErrServerStopped {
		fmt.Println("serve error: ", err.Error())
	}
}
```

Above is the server side example. There are client and more detailed examples in [examples/tcp](./examples/tcp)

## Architecture

```
accepting connection:

+------------+    +-------------------+    +----------------+
|            |    |                   |    |                |
|            |    |                   |    |                |
| tcp server |--->| accept connection |--->| create session |
|            |    |                   |    |                |
|            |    |                   |    |                |
+------------+    +-------------------+    +----------------+

in session:

+------------------+    +-----------------------+    +----------------------------------+
| read connection  |--->| unpack packet payload |--->|                                  |
+------------------+    +-----------------------+    |                                  |
                                                     | router (middlewares and handler) |
+------------------+    +-----------------------+    |                                  |
| write connection |<---| pack packet payload   |<---|                                  |
+------------------+    +-----------------------+    +----------------------------------+
```

## API

### Routing

EasyTCP considers every message has a `ID` segment.
A message will be routed, according to it's id, to the handler through middelwares.

```
request flow:

+----------+    +--------------+    +--------------+    +---------+
| request  |--->|              |--->|              |--->|         |
+----------+    |              |    |              |    |         |
                | middleware 1 |    | middleware 2 |    | handler |
+----------+    |              |    |              |    |         |
| response |<---|              |<---|              |<---|         |
+----------+    +--------------+    +--------------+    +---------+
```

#### Register a route

```go
s.AddRoute(reqID, func(ctx *router.Context) (*packet.MessageEntry, error) {
	// handle the request via ctx
	fmt.Printf("[server] request received | id: %d; size: %d; data: %s\n", ctx.MsgID(), ctx.MsgSize(), ctx.MsgData())

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
	return func(ctx *router.Context) (resp *packet.MessageEntry, err error) {
		// do things before...
		resp, err := next(ctx)
		// do things after...
		return resp, err
	}
}
```

### Packer

A packer is to pack and unpack packets' payload. We can set the Packer when creating the server.

```go
s := easytcp.NewTCPServer(&server.TCPOption{
	MsgPacker: new(MyPacker), // this is optional, the default one is packet.DefaultPacker
})
```

We can set our own Packer or EasyTCP uses [`DefaultPacker`](./packet/packer.go).

The `DefaultPacker` considers packet's payload as a `Size(4)|ID(4)|Data(n)` format.

This may not covery some particular cases, fortunately, we can create our own Packer.

```go
// Packer16bit is a custom packer, implements packet.Packer interafce.
// THe Packet format is `size[2]id[2]data`
type Packer16bit struct{}

func (p *Packer16bit) bytesOrder() binary.ByteOrder {
	return binary.BigEndian
}

func (p *Packer16bit) Pack(msg *packet.MessageEntry) ([]byte, error) {
	size := len(msg.Data) // without id
	buff := bytes.NewBuffer(make([]byte, 0, size+2+2))
	if err := binary.Write(buff, p.bytesOrder(), uint16(size)); err != nil {
		return nil, fmt.Errorf("write size err: %s", err)
	}
	if err := binary.Write(buff, p.bytesOrder(), uint16(msg.ID)); err != nil {
		return nil, fmt.Errorf("write id err: %s", err)
	}
	if err := binary.Write(buff, p.bytesOrder(), msg.Data); err != nil {
		return nil, fmt.Errorf("write data err: %s", err)
	}
	return buff.Bytes(), nil
}

func (p *Packer16bit) Unpack(reader io.Reader) (*packet.MessageEntry, error) {
	sizeBuff := make([]byte, 2)
	if _, err := io.ReadFull(reader, sizeBuff); err != nil {
		return nil, fmt.Errorf("read size err: %s", err)
	}
	size := p.bytesOrder().Uint16(sizeBuff)

	idBuff := make([]byte, 2)
	if _, err := io.ReadFull(reader, idBuff); err != nil {
		return nil, fmt.Errorf("read id err: %s", err)
	}
	id := p.bytesOrder().Uint16(idBuff)

	data := make([]byte, size)
	if _, err := io.ReadFull(reader, data); err != nil {
		return nil, fmt.Errorf("read data err: %s", err)
	}

	msg := &packet.MessageEntry{ID: uint(id), Data: data}
	return msg, nil
}
```

### Codec

A Codec is to encode and decode message data. The Codec is optional, EasyTCP won't encode/decode message data if Codec is not set.
We can set Codec when creating the server.

```go
s := easytcp.NewTCPServer(&server.TCPOption{
	MsgCodec: &packet.JsonCodec{}, // this is optional. The JsonCodec is a built-in codec
})
```

Since we set the codec, we may want to decode the request data in route handler.

```go
s.AddRoute(reqID, func(ctx *router.Context) (*packet.MessageEntry, error) {
	var reqData map[string]interface{}
	if err := ctx.Bind(&reqData); err != nil { // here we decode message data and bind to reqData
		// handle error...
	}
	fmt.Printf("[server] request received | id: %d; size: %d; data-decoded: %+v\n", ctx.MsgID(), ctx.MsgSize(), reqData)
	respData := map[string]string{"key": "value"}
	return ctx.Response(respID, respData)
})
```

Codec's encoding will be invoked before message packed,
and decoding should be invoked in the route handler which is after message unpacked.

## Contribute

Check out a new branch for the job, and make sure github action passed.
