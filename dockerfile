# 使用 Go 官方鏡像作為構建環境
FROM golang:1.17 as builder

# 設置工作目錄，之後的指令都會在這個目錄下執行
WORKDIR /app

# 將 go.mod 和 go.sum 文件複製到容器中
COPY go.mod go.sum ./

# 下載項目依賴。利用 Docker 緩存層，除非 go.mod 和 go.sum 改變，否則不需要重複下載依賴
RUN go mod download

# 將項目中的所有文件複製到容器中
COPY . .

# 構建應用程序
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o myapp .

# 使用 scratch 作為最終運行時鏡像，它是一個空白的鏡像，非常小巧
FROM scratch

# 將工作目錄設置為 /root/ （這是 scratch 鏡像的根目錄）
WORKDIR /root/

# 從構建階段的 /app 目錄中複製構建好的二進制文件到當前目錄
COPY --from=builder /app/myapp .

# 從構建階段複製 .env 文件
COPY --from=builder /app/.env .

# 配置容器啟動時運行的命令
ENTRYPOINT ["./myapp"]

# 暴露端口（假設你的應用使用 8080 端口）
EXPOSE 8080
