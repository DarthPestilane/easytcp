# EasyTCP

[![Run Actions](https://github.com/DarthPestilane/easytcp/actions/workflows/actions.yml/badge.svg?branch=master&event=push)](https://github.com/DarthPestilane/easytcp/actions/workflows/actions.yml)
[![Go Report](https://goreportcard.com/badge/github.com/darthPestilane/easytcp)](https://goreportcard.com/report/github.com/darthPestilane/easytcp)
[![codecov](https://codecov.io/gh/DarthPestilane/easytcp/branch/master/graph/badge.svg?token=002KJ5IV4Z)](https://codecov.io/gh/DarthPestilane/easytcp)

```
$ ./start

[EASYTCP ROUTE TABLE]:
+------------+-----------------------+
| Message ID |     Route Handler     |
+------------+-----------------------+
|       1000 | path/to/handler.Func1 |
+------------+-----------------------+
|       1002 | path/to/handler.Func2 |
+------------+-----------------------+
[EASYTCP] Serving at: tcp://[::]:10001
```

## Introduction

`EasyTCP` is a light-weight TCP server framework written in Go (Golang), featured with:

- Non-invasive design
- Pipelined middlewares for route handler
- Customizable message packer and codec
- Handy functions to handle request data and send response
- Common hooks
- Customizable logger

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

Note: EasyTCP uses **Go Modules** to manage dependencies.

## Quick start

```go
package main

import (
	"fmt"
	"github.com/DarthPestilane/easytcp"
	"github.com/DarthPestilane/easytcp/message"
)

func main() {
	// create a new server
	s := easytcp.NewServer(&easytcp.ServerOption{})

	// add a route to message id
	s.AddRoute(uint(1001), func(c *easytcp.Context) (*message.Entry, error) {
		fmt.Printf("[server] request received | id: %d; size: %d; data: %s\n", c.Message().ID, len(c.Message().Data), c.Message().Data)
		return c.Response(uint(1002), []byte("copy that"))
	})

	// set custom logger (optional)
	easytcp.SetLogger(lg)

	// add global middlewares (optional)
	s.Use(recoverMiddleware)

	// set hooks (optional)
	s.OnSessionCreate(fn)
	s.OnSessionClose(fn)

	// set not-found route handler (optional)
	s.NotFoundHandler(handler)

	// listen and serve
	if err := s.Serve(":5896"); err != nil && err != server.ErrServerStopped {
		fmt.Println("serve error: ", err.Error())
	}
}
```

Above is the server side example. There are client and more detailed examples in [examples/tcp](./examples/tcp)

## Benchmark

```
goos: darwin
goarch: amd64
pkg: github.com/DarthPestilane/easytcp
BenchmarkTCPServer_NoRoute-8             	  196898	      7487 ns/op	     103 B/op	       4 allocs/op
BenchmarkTCPServer_NotFoundHandler-8     	  268335	      8524 ns/op	     740 B/op	      13 allocs/op
BenchmarkTCPServer_OneHandler-8          	  197395	      7231 ns/op	     425 B/op	      12 allocs/op
BenchmarkTCPServer_ManyHandlers-8        	  210126	      9533 ns/op	     571 B/op	      17 allocs/op
BenchmarkTCPServer_OneRouteSet-8         	  230790	      7615 ns/op	     626 B/op	      20 allocs/op
BenchmarkTCPServer_OneRouteJsonCodec-8   	  250588	      8773 ns/op	    1587 B/op	      29 allocs/op
PASS
ok  	github.com/DarthPestilane/easytcp	13.048s
```

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
s.AddRoute(reqID, func(c *easytcp.Context) (*message.Entry, error) {
	// handle the request via ctx
	fmt.Printf("[server] request received | id: %d; size: %d; data: %s\n", c.Message().ID, len(c.Message().Data), c.Message().Data)

	// do things...

	// return response
	return c.Response(respID, []byte("copy that"))
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
var exampleMiddleware easytcp.MiddlewareFunc = func(next easytcp.HandlerFunc) easytcp.HandlerFunc {
	return func(c *easytcp.Context) (*message.Entry, error) {
		// do things before...
		resp, err := next(c)
		// do things after...
		return resp, err
	}
}
```

### Packer

A packer is to pack and unpack packets' payload. We can set the Packer when creating the server.

```go
s := easytcp.NewServer(&easytcp.ServerOption{
	Packer: new(MyPacker), // this is optional, the default one is DefaultPacker
})
```

We can set our own Packer or EasyTCP uses [`DefaultPacker`](./packer.go).

The `DefaultPacker` considers packet's payload as a `Size(4)|ID(4)|Data(n)` format.

This may not covery some particular cases, fortunately, we can create our own Packer.

```go
// Packer16bit is a custom packer, implements Packer interafce.
// THe Packet format is `size[2]id[2]data`
type Packer16bit struct{}

func (p *Packer16bit) bytesOrder() binary.ByteOrder {
	return binary.BigEndian
}

func (p *Packer16bit) Pack(msg *message.Entry) ([]byte, error) {
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

func (p *Packer16bit) Unpack(reader io.Reader) (*message.Entry, error) {
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

	msg := &message.Entry{ID: uint(id), Data: data}
	return msg, nil
}
```

### Codec

A Codec is to encode and decode message data. The Codec is optional, EasyTCP won't encode/decode message data if Codec is not set.
We can set Codec when creating the server.

```go
s := easytcp.NewServer(&easytcp.ServerOption{
	Codec: &easytcp.JsonCodec{}, // this is optional. The JsonCodec is a built-in codec
})
```

Since we set the codec, we may want to decode the request data in route handler.

```go
s.AddRoute(reqID, func(c *easytcp.Context) (*message.Entry, error) {
	var reqData map[string]interface{}
	if err := c.Bind(&reqData); err != nil { // here we decode message data and bind to reqData
		// handle error...
	}
	fmt.Printf("[server] request received | id: %d; size: %d; data-decoded: %+v\n", c.Message().ID, len(c.Message().Data), reqData)
	respData := map[string]string{"key": "value"}
	return c.Response(respID, respData)
})
```

Codec's encoding will be invoked before message packed,
and decoding should be invoked in the route handler which is after message unpacked.

## Contribute

Check out a new branch for the job, and make sure github action passed.

Use issues for everything

- For a small change, just send a PR.
- For bigger changes open an issue for discussion before sending a PR.
- PR should have:
  - Test case
  - Documentation
  - Example (If it makes sense)
- You can also contribute by:
  - Reporting issues
  - Suggesting new features or enhancements
  - Improve/fix documentation
