# EasyTCP

todo:

- [ ] 更多的setter 比如buffMsgChannel的大小: `server.SetBufferSize(n)`
- [x] accept 连接后和断开连接前的 hooks: `server.AfterConnected(fn)`, `server.BeforeDisconnect(fn)`
    - `s.OnDisconnect(fn)` `s.OnConnected(fn)`
- [ ] 日志: 基于logrus, 设置 defaultLogger, 而不是直接 logrus.Debug 这样调用
- [ ] 单元测试: 包括框架选择(ginkgo or testify)，代码覆盖
- [ ] 将报文进行抽象，而不是固定死结构
