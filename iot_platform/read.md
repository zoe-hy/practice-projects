**项目简介**

一个简单的物联网数据处理示例工程，演示生产者（API） -> Kafka -> 消费者 -> MySQL 的数据流：接收大量设备上报的数据，经过消息队列缓冲与异步处理后持久化到数据库。

**技术栈**

- **语言**: Go
- **消息队列**: Kafka
- **数据库**: MySQL
- **容器/编排**: Docker / docker-compose
- **开发/调试**: REST 客户端（用于发送请求）、Database 客户端（用于查看数据）

**整体架构**

- **API / 生产者**: `main.go` 提供 HTTP 接口，接收设备或测试请求，并将消息写入 Kafka。
- **消息总线**: Kafka 作为高并发消息缓冲与分发层（通常搭配 Zookeeper）。
- **消费者**: `Consumer.go` 从 Kafka 消费消息，进行业务处理后写入 MySQL。
- **持久层**: MySQL（建表与初始化见 `sql.sql`）。
- **压力测试**: `stress.go` 用于模拟大量设备并发上报以验证系统吞吐与稳定性。

**核心流程（高层）**

1. 设备或测试脚本通过 HTTP 请求将数据发送到 `main.go`。
2. `main.go` 将数据序列化并发送到 Kafka 的指定 Topic。
3. `Consumer.go` 订阅该 Topic，消费并处理消息（解析、校验、必要的聚合/清洗）。
4. 处理后的数据写入 MySQL，供查询与分析使用。

**主要文件说明**

- `main.go`: 生产者与 HTTP 接口实现，负责将请求转为 Kafka 消息。
- `Consumer.go`: Kafka 消费者实现，包含消息处理与写入 MySQL 的逻辑。
- `stress.go`: 压力测试脚本，批量/并发发送模拟设备数据到 API。
- `docker-compose.yml`: 本地启动依赖（如 Kafka、Zookeeper、MySQL）的编排文件。
- `sql.sql`: 数据库建表/初始化 SQL。
- `iot-api.http`: REST 请求示例，便于手动测试 API。
- `go.mod`: Go 模块依赖描述文件。

**快速运行（本地）**

1. 启动依赖服务（需已安装 Docker）：

```bash
docker-compose up -d
```

2. 在项目目录启动生产者或直接运行：

```bash
go run main.go
```

3. 启动消费者：

```bash
go run Consumer.go
```

4. 使用 `iot-api.http` 或 `go run stress.go` 发送测试数据。

**监控与日志**

- 本项目在生产者和消费者中集成了结构化日志（`zerolog`）以及 Prometheus 指标。默认会在生产者监听端口的下一个端口（`HTTP_PORT+1`，例如 `8081`）暴露 `/metrics`，消费者在 `:9100` 暴露 `/metrics`。你可以把这些地址接入 Prometheus。

**构建镜像**

```bash
docker build -f Dockerfile.producer -t iot-producer:local .
docker build -f Dockerfile.consumer -t iot-consumer:local .
```

**部署到 k8s（快速）**

```bash
kubectl apply -k k8s/
```



