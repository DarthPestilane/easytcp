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

`EasyTCP` is a light-weight and less painful TCP server framework written in Go (Golang) based on the standard `net` package.

✨ Features:

- Non-invasive design
- Pipelined middlewares for route handler
- Customizable message packer and codec, and logger
- Handy functions to handle request data and send response
- Common hooks

`EasyTCP` helps you build a TCP server easily and fast.

This package, so far, has been tested with

- go1.14.x
- go1.15.x
- go1.16.x
- go1.17.x

on the latest Linux, Macos and Windows.

## Install

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
    // Create a new server with default options.
    s := easytcp.NewServer(&easytcp.ServerOption{})

    // Register a route with message's ID.
    // The `DefaultPacker` treats id as uint32,
    // so when we add routes or return response, we should use uint32 or *uint32.
    s.AddRoute(uint32(1001), func(c *easytcp.Context) (*message.Entry, error) {
        fmt.Printf("[server] request received | id: %d; size: %d; data: %s\n", c.Message().ID, len(c.Message().Data), c.Message().Data)
        return c.Response(uint32(1002), []byte("copy that"))
    })

    // Set custom logger (optional).
    easytcp.SetLogger(lg)

    // Add global middlewares (optional).
    s.Use(recoverMiddleware)

    // Set hooks (optional).
    s.OnSessionCreate = func(sess *easytcp.Session) {}
    s.OnSessionClose = func(sess *easytcp.Session) {}

    // Set not-found route handler (optional).
    s.NotFoundHandler(handler)

    // Listen and serve.
    if err := s.Serve(":5896"); err != nil && err != server.ErrServerStopped {
        fmt.Println("serve error: ", err.Error())
    }
}
```

Above is the server side example. There are client and more detailed examples including:

- [broadcasting](./examples/tcp/broadcast)
- [custom packet](./examples/tcp/custom_packet)
- [communicating with protobuf](./examples/tcp/proto_packet)

in [examples/tcp](./examples/tcp).

## Benchmark

- goversion: 1.15.15
- goos: darwin
- goarch: amd64

| Benchmark name                    | (1)    | (2)        | (3)       | (4)          | remark                        |
| --------------------------------- | ------ | ---------- | --------- | ------------ | ----------------------------- |
| Benchmark_NoRoute-8               | 250000 | 8583 ns/op | 159 B/op  | 3 allocs/op  |                               |
| Benchmark_NotFoundHandler-8       | 250000 | 7546 ns/op | 810 B/op  | 11 allocs/op |                               |
| Benchmark_OneHandler-8            | 250000 | 7829 ns/op | 781 B/op  | 12 allocs/op |                               |
| Benchmark_ManyHandlers-8          | 250000 | 7369 ns/op | 822 B/op  | 14 allocs/op |                               |
| Benchmark_OneRouteSet-8           | 250000 | 7818 ns/op | 1031 B/op | 17 allocs/op |                               |
| Benchmark_OneRouteJsonCodec-8     | 250000 | 9152 ns/op | 1956 B/op | 28 allocs/op | build with `encoding/json`    |
| Benchmark_OneRouteJsonCodec-8     | 250000 | 9127 ns/op | 1685 B/op | 23 allocs/op | build with `json-jsoniter/go` |
| Benchmark_OneRouteProtobufCodec-8 | 250000 | 8010 ns/op | 1016 B/op | 15 allocs/op |                               |
| Benchmark_OneRouteMsgpackCodec-8  | 250000 | 9121 ns/op | 1281 B/op | 19 allocs/op |                               |

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

in route handler:

+----------------------------+    +------------+
| codec decode request data  |--->|            |
+----------------------------+    |            |
                                  | user logic |
+----------------------------+    |            |
| codec encode response data |<---|            |
+----------------------------+    +------------+
```

## Conception

### Routing

EasyTCP considers every message has a `ID` segment to distinguish one another.
A message will be routed, according to it's id, to the handler through middlewares.

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
// global middlewares are prior than per-route middlewares, they will be invoked first
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

The `DefaultPacker` considers packet's payload as a `Size(4)|ID(4)|Data(n)` format. (`Size` only represents the length of `Data` instead of the whole payload length)

This may not covery some particular cases, but fortunately, we can create our own Packer.

```go
// Packer16bit is a custom packer, implements Packer interafce.
// THe Packet format is `size[2]id[2]data`
type Packer16bit struct{}

func (p *Packer16bit) bytesOrder() binary.ByteOrder {
    return binary.BigEndian
}

func (p *Packer16bit) Pack(entry *message.Entry) ([]byte, error) {
    size := len(entry.Data) // without id
    buff := bytes.NewBuffer(make([]byte, 0, size+2+2))
    if err := binary.Write(buff, p.bytesOrder(), uint16(size)); err != nil {
        return nil, fmt.Errorf("write size err: %s", err)
    }
    if err := binary.Write(buff, p.bytesOrder(), entry.ID.(uint16)); err != nil {
        return nil, fmt.Errorf("write id err: %s", err)
    }
    if err := binary.Write(buff, p.bytesOrder(), entry.Data); err != nil {
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
    // since id here is the type of uint16, we need to use a uint16 when adding routes.
    // eg: server.AddRoute(uint16(123), ...)

    data := make([]byte, size)
    if _, err := io.ReadFull(reader, data); err != nil {
        return nil, fmt.Errorf("read data err: %s", err)
    }

    entry := &message.Entry{ID: id, Data: data}
    entry.Set("theWholeLength", 2+2+size) // we can set our custom kv data here.
    // c.Message().Get("theWholeLength")  // and get them in route handler.
    return entry, nil
}
```

And see the custom packer [here](./examples/fixture/packer.go).

### Codec

A Codec is to encode and decode message data. The Codec is optional, EasyTCP won't encode or decode message data if the Codec is not set.

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

> NOTE:
>
> If the Codec is not set (or is `nil`), EasyTCP will try to convert the `respData` (the second parameter of `c.Response`) into a `[]byte`.
> So the type of `respData` should be one of `string`, `[]byte` or `fmt.Stringer`.

#### JSON Codec

`JsonCodec` is an EasyTCP's built-in codec, which uses `encoding/json` as the default implementation.
Can be changed by build from other tags.

[jsoniter](https://github.com/json-iterator/go) :

```sh
go build -tags=jsoniter .
```

#### Protobuf Codec

`ProtobufCodec` is an EasyTCP's built-in codec, which uses `google.golang.org/protobuf` as the implementation.

#### Msgpack Codec

`MsgpackCodec` is an EasyTCP's built-in codec, which uses `github.com/vmihailenco/msgpack` as the implementation.

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
