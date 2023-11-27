# Stage 1: Menggunakan image Go versi 1.19 untuk build
FROM golang:1.19 AS build

# Buat direktori kerja di dalam container
WORKDIR /app

# Instalasi Swagger
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Tambahkan kode aplikasi Go ke dalam container
COPY . .

# Inisiasi Swagger
RUN swag init

# Build aplikasi Go untuk Linux 64 bit
RUN GOOS=linux GOARCH=amd64 go build -o myapp .

# Stage 2: Menggunakan image Linux 64 bit
FROM alpine:latest

# Copy binary yang telah dibuild dari stage 1
COPY --from=build /app/myapp /app/myapp

# Port yang digunakan oleh aplikasi
EXPOSE 8080

# Perintah default saat container dijalankan
CMD ["/app/myapp"]