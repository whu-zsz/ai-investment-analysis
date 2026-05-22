# 构建阶段
FROM golang:1.23-alpine AS builder

WORKDIR /app

# 安装依赖
RUN apk add --no-cache git make

ENV GOPROXY=https://goproxy.cn,https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.org
ENV GODEBUG=http2client=0

# 复制go.mod和go.sum
COPY go.mod go.sum ./
RUN for i in 1 2 3; do go mod download && break || if [ "$i" -eq 3 ]; then exit 1; fi; done

# 复制源代码
COPY . .

# 编译
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/server/main.go

# 运行阶段
FROM alpine:latest

WORKDIR /app

# 安装ca证书（用于HTTPS请求）
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .
COPY --from=builder /app/docs ./docs

# 创建上传目录
RUN mkdir -p uploads

# 暴露端口
EXPOSE 8080

# 运行
CMD ["./main"]
