# 模板 A:纯 Go 项目 —— 保留完整工具链,支持 arm64+amd64
FROM golang:1.22-bookworm

WORKDIR /app

# 国内代理(单代理,Go 对 connection refused 不 fallback)
ENV GOPROXY=https://goproxy.cn,direct

COPY go.mod go.sum ./
RUN go mod download

COPY . .

CMD ["go", "test", "./..."]
