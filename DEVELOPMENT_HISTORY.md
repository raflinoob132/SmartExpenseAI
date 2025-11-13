# Development History - SmartExpenseAI

Catatan perkembangan dan perubahan yang telah dilakukan pada proyek SmartExpenseAI.

## 1. Inisialisasi Proyek (Awal)
**Tanggal**: 13 November 2025

### File-file yang dibuat:
- `/cmd/main.go` - Main application entry point
- `/internal/routes/telegram.go` - Telegram webhook handler
- `/internal/models/expense.go` - Expense data model
- `/internal/database/connection.go` - PostgreSQL connection with GORM
- `/internal/services/ai.go` - AI processing logic
- `/internal/services/recap.go` - Expense recap generation
- `.env` - Environment variables
- `go.mod` - Go module dependencies

### Dependencies:
- Fiber (web framework)
- GORM (ORM for PostgreSQL)
- Telegram Bot API
- OpenRouter API (AI integration)

---

## 2. Perbaikan Struktur Model dan Database (Langkah 2)
**Tanggal**: 13 November 2025

### Perubahan:
- Update `/internal/models/expense.go` dengan struktur yang lebih lengkap
- Update `/internal/database/connection.go` dengan fungsi InitDB()

### Struktur Expense:
```go
type Expense struct {
  ID          uint      `gorm:"primaryKey"`
  Description string
  Category    string
  Amount      float64
  Date        time.Time
  CreatedAt   time.Time
}
```

---

## 3. Perbaikan Fungsi AI (Langkah 3)
**Tanggal**: 13 November 2025

### Perubahan:
- Update `/internal/services/ai.go` dengan fungsi ParseExpense()
- Fungsi ini mengirim teks ke API OpenRouter model qwen2:1.5b
- Meminta output JSON sesuai format: { "description": "", "category": "", "amount": 0, "date": "YYYY-MM-DD" }
- Menggunakan OPENROUTER_API_KEY dari .env

---

## 4. Implementasi Handler Telegram (Langkah 4)
**Tanggal**: 13 November 2025

### Perubahan:
- Update `/internal/routes/telegram.go` untuk handler Telegram
- Inisialisasi bot dari token di .env
- Handle setiap pesan teks dari user
- Cek message.From.ID == TELEGRAM_USER_ID agar hanya owner yang bisa pakai
- Kirim teks user ke ParseExpense()
- Simpan hasil ke database
- Balas user dengan format: ✅ Disimpan: Kategori: Makanan, Jumlah: Rp25.000, Deskripsi: makan ayam geprek

---

## 5. Implementasi Fungsi Recap Mingguan (Langkah 5)
**Tanggal**: 13 November 2025

### Perubahan:
- Update `/internal/services/recap.go` dengan fungsi GenerateWeeklyRecap(bot *tgbotapi.BotAPI, chatID int64)
- Ambil semua data 7 hari terakhir dari DB
- Kelompokkan total per kategori
- Format rekap seperti:
  ```
  🧾 Rekap Mingguan:
  - Makanan: Rp120.000
  - Transport: Rp40.000
  Total: Rp160.000
  ```
- Kirim pesan ke chat Telegram owner
- Jadwalkan agar fungsi ini jalan seminggu sekali (gunakan gocron)

---

## 6. Update Main.go (Langkah 6)
**Tanggal**: 13 November 2025

### Perubahan:
- Update `/cmd/main.go`:
  - Jalankan InitDB()
  - Jalankan handler Telegram
  - Jalankan scheduler gocron untuk GenerateWeeklyRecap()
  - Gunakan Fiber hanya untuk kebutuhan webhook Telegram (port 8080)
  - Ambil semua env dari .env

---

## 7. Perbaikan dan Pengujian (Langkah 7)
**Tanggal**: 13 November 2025

### Perubahan:
- Fix masalah dependency dan import path Telegram Bot API
- Update go.mod untuk versi yang benar
- Fix masalah type mismatch antara int dan int64
- Fix build issues

---

## 8. Penambahan Fitur Lengkap (Langkah 8)
**Tanggal**: 13 November 2025

### Perubahan:
- Tambahkan fitur untuk user bisa on demand lihat pengeluarannya selama 30 hari di sort menjadi perbulan
- Tambahkan fitur user bisa hapus atau update pengeluaran
- Tambahkan fungsi GenerateMonthlyRecap() di services/recap.go
- Tambahkan fungsi untuk CRUD (Create, Read, Update, Delete) expense
- Tambahkan command: /lihat, /bulan, /hapus, /update, /bantuan
- Update README.md dengan fitur-fitur baru

---

## 9. Implementasi NLP untuk Perintah Natural (Langkah 9 - Dicoba lalu dikembalikan)
**Tanggal**: 13 November 2025

### Perubahan (Kemudian Dikembalikan):
- Mencoba implementasi NLP untuk mendeteksi perintah dalam bentuk natural language
- Mencoba buat sistem AI yang bisa membedakan antara pengeluaran dan perintah
- Tapi sistem ini membuat bot bingung dan kurang akurat

---

## 10. Kembali ke Pendekatan Command Tradisional (Langkah 10 - Sekarang)
**Tanggal**: 13 November 2025

### Perubahan:
- Kembali ke pendekatan command-only untuk fitur selain input pengeluaran
- AI hanya digunakan untuk ekstraksi data pengeluaran dari pesan natural
- Command seperti /lihat, /hapus tetap menggunakan command tradisional
- Output daftar pengeluaran sekarang menampilkan ID expense agar mudah dihapus
- Update respon bot agar lebih informatif saat tidak bisa mengenali pengeluaran

---

## Fitur Saat Ini:

### Input Natural (Menggunakan AI):
- "makan nasi padang 25000" → Tercatat sebagai pengeluaran

### Command Tradisional:
- `/start` - Tampilkan welcome message
- `/lihat` - Lihat 10 pengeluaran terakhir (dengan ID)
- `/bulan` - Lihat rekap pengeluaran 30 hari terakhir per bulan
- `/hapus ID` - Hapus pengeluaran dengan ID tertentu
- `/update ID deskripsi jumlah kategori` - Update data pengeluaran
- `/bantuan` - Tampilkan bantuan

### Fitur Otomatis:
- Recap mingguan otomatis dikirim ke user (dijadwalkan dengan gocron)

---

## Catatan Penting:
1. Proyek menggunakan OpenRouter API untuk AI (harus ada OPENROUTER_API_KEY di .env)
2. Proyek hanya mengizinkan satu user (via TELEGRAM_USER_ID)
3. Proyek membutuhkan database PostgreSQL (di DATABASE_URL)
4. Untuk testing lokal, perlu ngrok karena menggunakan webhook Telegram
5. Model AI saat ini menggunakan openai/gpt-3.5-turbo melalui OpenRouter API