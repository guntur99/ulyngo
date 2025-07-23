# --- STAGE 1: BUILDER ---
# Gunakan image Golang terbaru untuk membangun aplikasi Anda.
FROM golang:1.23-alpine AS builder

# Atur direktori kerja di dalam container.
WORKDIR /app

# Install git in the builder stage. This is necessary if there are Go dependencies
# that are downloaded from Git repositories.
RUN apk --no-cache add git

# Salin file go.mod dan go.sum. Ini memungkinkan Docker untuk meng-cache dependensi
# dan membangun ulang hanya jika dependensi berubah, mempercepat build.
COPY go.mod .
COPY go.sum .

# Unduh semua dependensi Golang.
RUN go mod download

# Salin semua kode sumber aplikasi Anda ke dalam container.
COPY . .

# Bangun aplikasi Golang.
# CGO_ENABLED=0: Membuat binary statis yang tidak memerlukan library C.
# GOOS=linux: Menargetkan sistem operasi Linux (standar untuk container).
# -o /app/main: Menentukan nama file output binary menjadi 'main' di direktori /app.
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main .

# --- STAGE 2: FINAL IMAGE ---
# Gunakan image Alpine Linux yang sangat ringan sebagai base image akhir.
FROM alpine:latest

# Instal 'ca-certificates' agar aplikasi bisa melakukan koneksi HTTPS dengan aman.
RUN apk --no-cache add ca-certificates

# Atur direktori kerja di dalam container akhir.
WORKDIR /app

# Salin binary 'main' yang sudah dibangun dari stage 'builder' ke dalam image akhir.
COPY --from=builder /app/main .

# Expose port yang akan didengarkan oleh aplikasi Anda.
# Cloud Run akan otomatis mencari port ini atau menggunakan variabel lingkungan PORT.
EXPOSE 8080

# Perintah default untuk menjalankan aplikasi Anda saat container dimulai.
CMD ["/app/main"]