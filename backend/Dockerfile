FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git make

ENV GOPROXY=https://goproxy.cn,https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.org
ENV GODEBUG=http2client=0

COPY go.mod go.sum ./
RUN for i in 1 2 3; do go mod download && break || if [ "$i" -eq 3 ]; then exit 1; fi; done

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/server/main.go

FROM python:3.11-slim

WORKDIR /app

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates tzdata \
    && rm -rf /var/lib/apt/lists/*

RUN pip install --no-cache-dir akshare==1.18.63

ENV TZ=Asia/Shanghai

COPY --from=builder /app/main .
COPY --from=builder /app/docs ./docs
COPY --from=builder /app/scripts ./scripts

RUN mkdir -p uploads

EXPOSE 8080

CMD ["./main"]
