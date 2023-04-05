# Crisp Telegram Bot

功能：将crisp消息转发到telegram，通过telegram回复crisp消息。

### 编译为linux程序

`CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build`

## config.yaml

```
debug: true
redis:
  host: localhost:6379
  db: 0
  password: ''
crisp:
  identifier: 049sk12f-8349-8274-9d91-f21jv91kafa7 #的 Crisp Marketplace 插件 ID
  key: 078f2106a5d89179gkqn38e5e82e3c7j30ajfkelqnvd874fb2378573499ff505 # 你的 Crisp Marketplace 插件秘钥
telegram:
  key: #你的Bot Token
admins:
  - 93847124
```

# 上传到linux服务器

上传执行程序和配置文件

⚠️要上传到能访问telegram的服务器，否则会不能获取到消息

创建守护：

```
vi /etc/systemd/system/crisp.service
```

添加如下内容:

```shell
[Unit]
Description = crisp
After = network.target
Wants = network.target

[Service]
WorkingDirectory = /root/crispbot/
ExecStart = /root/crispbot/crisp_tg_bot
Restart = on-abnormal
RestartSec = 5s
KillMode = mixed

StandardOutput = null
StandardError = syslog

[Install]
WantedBy = multi-user.target

```

保存、启动

```
systemctl daemon-reload
systemctl start crisp
#检查是否启动成功
systemctl status crisp
#设置开机自启
systemctl enable crisp
```

## 附：申请 Crisp 以及 MarketPlace 插件

1. 注册 [https://app.crisp.chat/initiate/signup](https://app.crisp.chat/initiate/signup)
2. 前往 MarketPlace， 需要重新注册账号 [https://marketplace.crisp.chat/](https://marketplace.crisp.chat/)
3. 点击 New Plugin，选择 Private，输入名字以及描述。会获得开发者ID和Key。
4. 需要Production Key，点击 Ask a production token，再点击Add a Scope。
5. 需要 2 条**read**和**write**权限：`website:conversation:sessions` 和 `website:conversation:messages`
6. 保存后即可获得ID和Key，此时点击右上角 Install Plugin on Website 即可。
