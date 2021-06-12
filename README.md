# EasyTCP

[![Run Actions](https://github.com/DarthPestilane/easytcp/actions/workflows/actions.yml/badge.svg?branch=master&event=push)](https://github.com/DarthPestilane/easytcp/actions/workflows/actions.yml)
[![codecov](https://codecov.io/gh/DarthPestilane/easytcp/branch/master/graph/badge.svg?token=002KJ5IV4Z)](https://codecov.io/gh/DarthPestilane/easytcp)

road map:

- [x] TCP/UDP server
- [x] Routing incoming message to handler through middlewares
- [x] Customize `Packer` to pack and unpack message packet, and `Codec` to encode and decode message data
- [x] Customize logger

todo:

- [ ] Refactor `Session`, `Request` and `Response` into a `context` thing in router's `HandlerFunc` ?
    - [x] Introduce `context` to contain session and message
    - [x] refactor session's SendResp and RecvReq methods, do the pack and unpack there
