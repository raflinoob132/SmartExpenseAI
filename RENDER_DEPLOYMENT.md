# Panduan Deploy ke Render

## File render.yaml yang dibutuhkan

```yaml
services:
- type: web
  name: smartexpenseai
  env: go
  region: singapore  # atau ganti ke region terdekatmu
  buildCommand: |
    go mod download &&
    go build -o bin/server ./cmd/main.go
  startCommand: ./bin/server
  envVars:
    - key: PORT
      value: 8080
    - key: TELEGRAM_BOT_TOKEN
      sync: false  # Harus diisi manual di dashboard Render
    - key: TELEGRAM_USER_ID
      sync: false  # Harus diisi manual di dashboard Render
    - key: DATABASE_URL
      sync: false  # Harus diisi manual di dashboard Render
    - key: OPENROUTER_API_KEY
      sync: false  # Harus diisi manual di dashboard Render
  healthCheckPath: /setup-webhook  # Ganti dengan path yang sesuai jika ada endpoint health check
  disk:  # Jika dibutuhkan simpan data lokal
    name: smartexpenseai-data
    sizeGB: 1
```

## Cara Deploy

1. Pastikan repository kamu sudah di-push ke GitHub
2. Login ke [Render Dashboard](https://dashboard.render.com/)
3. Klik "New +" dan pilih "Web Service"
4. Pilih repository GitHub kamu yang berisi proyek SmartExpenseAI
5. Render akan otomatis mendeteksi file `render.yaml` dan menggunakan konfigurasi tersebut
6. Tambahkan environment variables di dashboard Render:
   - `TELEGRAM_BOT_TOKEN` = token bot Telegram kamu
   - `TELEGRAM_USER_ID` = user ID Telegram kamu (agar hanya kamu yang bisa pakai bot)
   - `DATABASE_URL` = connection string PostgreSQL (bisa buat dari Render PostgreSQL atau database lain)
   - `OPENROUTER_API_KEY` = API key dari OpenRouter untuk AI
7. Klik "Create Web Service"

Catatan: Aplikasi ini sudah dirancang untuk membaca PORT dari environment variable, jadi Render akan otomatis menyediakan port yang tersedia.

## Catatan Penting

- Pastikan repository kamu bersifat public jika kamu tidak menghubungkan akun GitHub premium
- Jika kamu ingin private repository, kamu perlu menghubungkan akun GitHub premium
- Pastikan kamu juga membuat PostgreSQL instance di Render atau menggunakan database eksternal
- Setelah deploy selesai, kamu harus setup webhook Telegram dengan mengakses: `https://[nama-service-kamu].onrender.com/setup-webhook`

## Setup Webhook

Setelah deployment selesai:
1. Akses URL: `https://[nama-service-kamu].onrender.com/setup-webhook`
2. Ini akan mengatur webhook Telegram ke URL public kamu
3. Bot siap digunakan!

## Troubleshooting

- Jika service crash, cek logs di dashboard Render
- Pastikan semua environment variables telah diisi dengan benar
- Pastikan database kamu bisa diakses dari Render
- Jika menggunakan Render PostgreSQL, pastikan security group mengizinkan koneksi dari service kamu