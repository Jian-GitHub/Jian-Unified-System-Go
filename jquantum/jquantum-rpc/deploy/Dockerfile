# ============================
# Build stage
# ============================
FROM golang:1.24 AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED=0

WORKDIR /build

# 先下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 拷贝源代码
COPY . .

# 编译 jquantumrpc
RUN go build -ldflags="-s -w" -o /app/jquantumrpc jquantum/jquantum-rpc/jquantumrpc.go

# ============================
# Runtime stage
# ============================
FROM ubuntu:22.04


# 设置时区（无需 tzdata）
ENV TZ=Asia/Shanghai
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

# 安装 MPI、SSH
RUN apt-get update && apt-get install -y \
    openmpi-bin openmpi-common libopenmpi-dev make cmake g++ \
    openssh-server openssh-client \
    && mkdir /var/run/sshd

# SSH 配置
RUN mkdir -p /root/.ssh
COPY ./jquantum/id_rsa /root/.ssh/id_rsa
COPY ./jquantum/id_rsa.pub /root/.ssh/id_rsa.pub
COPY ./jquantum/authorized_keys /root/.ssh/authorized_keys
RUN chmod 600 /root/.ssh/id_rsa && chmod 644 /root/.ssh/authorized_keys \
    && echo "StrictHostKeyChecking no" >> /etc/ssh/ssh_config

# 拷贝 Go 可执行文件
COPY --from=builder /app/jquantumrpc /app/jquantumrpc
COPY jquantum/jquantum-rpc/etc /app/etc

# 拷贝 QuEST
COPY jquantum/jquantum-cpp/include /harmoniacore/jquantum
COPY jquantum/jquantum-cpp/lib /harmoniacore/lib
COPY jquantum/jquantum-cpp/quest /harmoniacore/quest

# 设置工作目录
WORKDIR /app

# 暴露端口，MPI/SSH 可用
EXPOSE 22
EXPOSE 30101

# 默认启动命令
# 启动 jquantumrpc 并后台启动 sshd
RUN mkdir -p /var/run/sshd
CMD ["/app/jquantumrpc", "-f", "/app/etc/jquantumrpc.yaml"]