FROM golang:1.23.6 AS builder

# 設定容器內的工作目錄
WORKDIR /app

# 複製 go.mod 和 go.sum 並下載依賴
COPY go.mod go.sum ./
RUN go mod download

# 複製專案內所有 Go 代碼
COPY . .

# 設定執行指令
CMD ["go", "run", "main.go"]

