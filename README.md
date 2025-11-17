# 在线聊天室 (onlineChatRoom)
一个基于 Go 语言实现的在线聊天室系统，支持用户注册、登录、群聊、私聊、查看在线用户及活跃度排名等功能，使用 MySQL 存储用户信息，Redis 存储聊天记录和用户活跃度数据。

## 功能特点
用户注册与登录  

群聊消息发送与接收    

一对一私聊功能    

查看在线用户列表    

用户活跃度排名展示      

心跳检测机制，自动处理离线用户  

聊天记录持久化存储

## 技术栈
编程语言: Go 1.24  

网络通信: TCP Socket  

Mysql:8.0+  

Redis:5.0+  

## 数据存储:
MySQL: 存储用户账号信息  

Redis: 存储聊天记录 (使用 Streams) 和用户活跃度  

## 第三方库:
github.com/go-redis/redis: Redis 客户端  

github.com/go-sql-driver/mysql: MySQL 驱动  

github.com/jmoiron/sqlx: 增强型 SQL 操作库  

## 项目结构

onlineChatRoom/  

├── client/               # 客户端代码  

│   ├── tool/             # 客户端工具函数  

│   └── client.go         # 客户端入口  

├── server/               # 服务端代码 

│   ├── tool/             # 服务端工具函数 

│   └── server.go         # 服务端入口 

├── db/                   # 数据库操作 

│   ├── mysql.go          # MySQL相关操作 

│   └── redis.go          # Redis相关操作 

├── msg/                  # 消息处理 

│   ├── msg.go            # 消息结构定义 

│   └── tool.go           # 消息处理工具 

├── utils/                # 通用工具函数 

├── go.mod                # 项目依赖
 
└── go.sum                # 依赖校验 

## 快速开始
### 前置条件
Go 1.24+ 环境  

MySQL 数据库  

Redis 服务  

### 环境配置
**MySQL 配置:**

创建数据库 onlinechatroom  
```sql
CREATE database onlinechatroom;
```
创建用户表:
```sql
CREATE TABLE user (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    password VARCHAR(50) NOT NULL
);
```
修改 db/mysql.go 中的数据库连接信息  

**Redis 配置:**  

确保 Redis 服务正常运行  

根据需要修改 db/redis.go 中的 Redis 连接信息  

### 运行步骤
克隆项目代码 

**安装依赖:**
```bash
go mod tidy
```
**启动服务端:**
```bash
cd server
go run server.go
```
**启动客户端 (可多个):**
```bash
cd client
go run client.go
```
### 使用说明
客户端连接后可选择 "注册" 或 "登录" 

成功进入聊天室后，支持以下命令: 

list: 查看在线用户列表 

quit: 退出聊天室 

To:用户名-->内容: 发送私聊消息 

rank: 查看用户活跃度排名 

直接输入内容：发送群聊消息 

### 实现细节
消息格式: 使用 JSON 格式进行消息序列化与反序列化 

网络通信: 基于 TCP 协议，采用自定义的消息长度前缀 + 消息内容的格式 

并发处理: 服务端使用 goroutine 为每个客户端连接提供独立处理 

心跳机制: 客户端每 10 秒发送心跳包，服务端检测超时连接 (20 秒) 并强制下线 

数据持久化: 聊天记录通过 Redis Streams 存储，用户信息存储在 MySQL 

### 注意事项
确保 MySQL 和 Redis 服务在运行前已正确配置并启动 

多客户端测试时，使用不同的用户名登录以体验完整功能 

如需修改服务端端口，可修改server/server.go中的监听端口设置 

