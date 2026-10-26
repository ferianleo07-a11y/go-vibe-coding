# 🏨 Aplikasi Akuntansi & Keuangan Bisnis Hotel (Standar USALI)

Aplikasi web akuntansi dan manajemen operasional perhotelan berbasis **Go (Gin Framework)**, **GORM**, **PostgreSQL**, dan **Vanilla HTML/CSS/JavaScript**. Sistem ini dirancang untuk hotel bintang 3–5 dengan kepatuhan penuh terhadap standar akuntansi perhotelan internasional **USALI** (*Uniform System of Accounts for the Lodging Industry*).

---

## 🌟 Fitur Utama

1. **🔐 Autentikasi & Isolasi Data Pengguna (JWT & Bcrypt)**
   - Pendaftaran akun admin/keuangan hotel dan login aman berbasis token JWT.
   - Kata sandi dienkripsi menggunakan algoritma `bcrypt`.
   - Token disimpan di `localStorage` peramban dan diverifikasi via header `Authorization: Bearer <token>`.
   - **Isolasi Data Penuh:** Setiap kamar, transaksi reservasi, dan beban pengeluaran terikat ke `user_id`. Pengguna tidak dapat melihat atau memanipulasi data milik pengguna lain.

2. **🛌 Manajemen Kamar & Ketersediaan Real-Time (CRUD Kamar)**
   - Status kamar real-time: `🟢 Tersedia`, `🔵 Terisi`, `🟡 Sedang Dibersihkan`, dan `🔴 Dalam Perbaikan`.
   - Pengaturan tipe kamar, tarif sewa per malam, dan fasilitas.
   - Filter cepat berdasarkan status ketersediaan.

3. **🧳 Check-in, Check-out & Transaksi Reservasi Tamu**
   - Pencatatan identitas tamu (KTP / Passport, No. Telepon).
   - Kalkulasi otomatis durasi malam dan total biaya sewa kamar secara *real-time*.
   - Manajemen uang muka (DP), sisa tagihan, dan status pelunasan.
   - Status otomatis: Check-in otomatis mengubah kamar menjadi `Terisi`, sedangkan Check-out mengubah status kamar menjadi `Sedang Dibersihkan`.
   - **🧾 Cetak Nota / Invoice:** Lembar nota resmi siap cetak langsung via peramban (`window.print()`).

4. **💸 Manajemen Biaya Operasional (8 Klasifikasi Beban USALI)**
   - Pencatatan beban pengeluaran operasional hotel berdasarkan klasifikasi standar USALI:
     1. `Utilities` (Listrik, Air, Gas, Internet)
     2. `Operating Supplies` (Linen, Amenities kamar mandi, Bahan Kebersihan)
     3. `Repairs & Maintenance` (Servis AC, Plumbing, Pemeliharaan Gedung)
     4. `Salaries & Wages` (Gaji Staf, Tunjangan, Lembur)
     5. `Marketing & Sales` (Iklan, Promosi, Komisi Online Travel Agent)
     6. `Administrative & General` (ATK, Software, Asuransi, Legalitas)
     7. `Food & Beverage` (Bahan Baku Restoran & Sarapan Hotel)
     8. `Other Expenses` (Beban Operasional Lain-lain)

5. **📊 Dashboard Analisis & Metrik KPI Hotel**
   - **Occupancy Rate (%)** $= (\text{Kamar Terisi} / \text{Total Kamar Tersedia}) \times 100$
   - **ADR (Average Daily Rate)** $= \text{Total Pendapatan Kamar} / \text{Kamar Terjual}$
   - **RevPAR (Revenue per Available Room)** $= \text{Total Pendapatan Kamar} / \text{Total Kamar Tersedia}$
   - **Laba / Rugi Bersih (Net Income)** $= \text{Pendapatan Kamar} - \text{Total Beban Operasional}$
   - Distribusi persentase pengeluaran per kategori USALI.

---

## 📁 Struktur Direktori Proyek

```text
hotel-app/
├── cmd/
│   └── api/
│       └── main.go                 # Titik masuk utama aplikasi (entry point)
├── config/
│   └── database.go                 # Setup GORM PostgreSQL via DATABASE_URL & migrasi aman
├── controllers/                    # HTTP Controller & validasi input
│   ├── auth_controller.go          # Registrasi, login, profil pengguna
│   ├── room_controller.go          # CRUD kamar & validasi status
│   ├── reservation_controller.go   # Check-in, check-out, kalkulasi, invoice
│   ├── expense_controller.go       # CRUD pengeluaran operasional USALI
│   └── report_controller.go        # Kalkulasi KPI & Rekap Laba/Rugi
├── middlewares/
│   └── auth_middleware.go          # Validasi Bearer JWT & inject user_id ke context
├── models/                         # Definisi struct & skema database GORM
│   ├── user.go                     # Struct User & hashing bcrypt
│   ├── room.go                     # Struct Room
│   ├── reservation.go              # Struct Reservation
│   └── expense.go                  # Struct Expense & kategori USALI
├── routes/
│   └── routes.go                   # Pendaftaran routing REST API & file statis web
├── tests/                          # Pengujian otomatis (Unit & Integration Test)
│   ├── api_health_test.go          # Test endpoint kesehatan API
│   ├── auth_test.go                # Test hash bcrypt & validasi JWT
│   └── kpi_test.go                 # Test formula Occupancy, ADR, RevPAR, Laba/Rugi
├── web/                            # Frontend statis (Vanilla HTML/CSS/JS)
│   ├── css/
│   │   └── style.css               # Desain UI modern, responsif, dan layout cetak nota
│   ├── js/
│   │   ├── auth.js                 # Logika autentikasi & token di localStorage
│   │   ├── rooms.js                # Manajemen kamar & filter
│   │   ├── reservations.js         # Check-in/out, estimasi, dan cetak invoice
│   │   ├── expenses.js             # Biaya operasional USALI
│   │   └── dashboard.js            # Visualisasi metrik KPI & grafik beban
│   ├── index.html                  # Main Page (Dashboard & Workspace Lengkap)
│   ├── login.html                  # Halaman masuk
│   └── register.html               # Halaman pendaftaran akun
├── .env.example                    # Template konfigurasi environment (aman di Git)
├── .gitignore                      # Mengabaikan .env dan binary hasil build
├── go.mod
├── go.sum
└── README.md
```

---

## ⚙️ Cara Menjalankan Aplikasi

### 1. Prasyarat
- Go versi 1.21+ (diuji pada Go 1.27)
- PostgreSQL Server aktif

### 2. Konfigurasi Lingkungan (.env)
Salin template konfigurasi `.env.example` menjadi `.env`:
```bash
cp .env.example .env
```
Sesuaikan nilai `DATABASE_URL` pada `.env` dengan kredensial PostgreSQL lokal Anda:
```dotenv
APP_PORT=8080
DATABASE_URL=postgres://username:password@localhost:5432/hotel_finance?sslmode=disable
JWT_SECRET=rahasia-kunci-jwt-hotel-app-yang-sangat-aman
```

> [!IMPORTANT]
> File `.env` bersifat rahasia dan telah didaftarkan ke `.gitignore` sehingga tidak akan terunggah ke Git.

### 3. Menjalankan Server
```bash
go run ./cmd/api
```
Server akan aktif di: **`http://localhost:8080`**

Buka peramban di:
- Masuk / Login: `http://localhost:8080/login`
- Daftar Akun: `http://localhost:8080/register`
- Main Page: `http://localhost:8080/`

---

## 🧪 Menjalankan Pengujian Otomatis

Jalankan seluruh rangkaian unit test dan integration test dasar:
```bash
go test ./tests/... -v
```

Hasil pengujian mencakup:
- Keabsahan enkripsi satu arah *password hashing* (`bcrypt`).
- Pembuatan dan verifikasi token *JSON Web Token* (`jwt`).
- Akurasi kalkulasi formula perhotelan: *Occupancy Rate*, *ADR*, *RevPAR*, dan *Net Income*.
- Validasi 8 kategori beban standar *USALI*.
- Kesiapan *health-check* router API.

---

## 🔒 Kepatuhan Keamanan & Aturan Pengembangan

1. **DATABASE_URL Tanpa Hardcode:** Koneksi PostgreSQL dibaca secara dinamis dari `os.Getenv("DATABASE_URL")`. Jika variabel belum didefinisikan, aplikasi mengembalikan pesan error informatif tanpa mencetak kredensial ke log.
2. **Perlindungan Rahasia:** File `.env` terabaikan oleh Git, sementara `.env.example` disediakan tanpa menyertakan kata sandi asli.
3. **Penyembunyian Data Sensitif:** Field `Password` pada model pengguna menggunakan tag `json:"-"` sehingga tidak pernah diekspos dalam respons API atau log.
4. **Isolasi Data Antar Pengguna:** Seluruh operasi query (Read, Update, Delete) pada Kamar, Reservasi, dan Biaya selalu menyertakan filter `WHERE user_id = ?` yang diekstrak langsung dari token JWT pengguna yang terverifikasi.
