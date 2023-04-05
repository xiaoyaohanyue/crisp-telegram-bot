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
  identifier: 049sk12f-8349-8274-9d91-f21jv91kafa7
  key: 078f2106a5d89179gkqn38e5e82e3c7j30ajfkelqnvd874fb2378573499ff505
telegram:
  key: 
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
