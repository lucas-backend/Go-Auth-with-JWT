# Panduan Integrasi Autentikasi JWT di Go

## Konsep Besar (The Big Picture)

Secara umum integrasi ini dibagi menjadi tiga pilar utama.

1. **Model & Setup**
Menyiapkan struktur data dan kunci rahasia.
2. **Issuing (Penerbitan Token)**
Terjadi saat Login. Server memverifikasi identitas lalu mencetak "tiket" (token).
3. **Validating (Pemeriksaan Token)**
Terjadi di Middleware. Server memeriksa keaslian "tiket" sebelum mengizinkan masuk.

---

## Fase 1 Persiapan Data (The Model)

Sebelum menulis logika kamu harus menyiapkan wadah datanya dulu.

1. **Siapkan Struct User**
Ini untuk menampung data dari database atau input JSON. Minimal berisi Username dan Password.
2. **Siapkan Struct Claims**
Ini adalah isi dari token nanti.
* Wajib menanamkan atau *embed* `jwt.RegisteredClaims` agar token memiliki fitur standar seperti waktu kadaluarsa.
* Tambahkan data spesifik aplikasi seperti `Username` atau `Role` jika perlu.


3. **Siapkan Kunci Rahasia**
Buat variabel global (idealnya dari *environment variable*) bertipe `[]byte`. Ini adalah kunci utama untuk menandatangani token.

---

## Fase 2 Pintu Masuk (Login & Issuing)

Ini adalah proses pembuatan token dan logikanya berjalan satu arah.

1. **Terima Input**
Tangkap username dan password dari *body request*.
2. **Cek Database**
Cari apakah user tersebut ada. Jika tidak ada langsung tolak request tersebut.
3. **Verifikasi Password (Bcrypt)**
Gunakan `bcrypt.CompareHashAndPassword`. Jangan pernah membandingkan string password mentah secara manual.
4. **Siapkan Expiration Time**
Tentukan kapan token basi. Gunakan `time.Now().Add(...)`.
5. **Buat Objek Claims**
Inisialisasi struct Claims yang sudah kamu buat di Fase 1. Gunakan tanda `&` (pointer) karena library memintanya demikian.
6. **Tanda Tangani Token (Signing)**
Gunakan `jwt.NewWithClaims` lalu panggil `.SignedString(kunciRahasia)`.
* Ingat bahwa kunci rahasia (`jwtKey`) harus berupa `[]byte`.
* Hasilnya adalah string panjang yang dikirim ke user sebagai respon JSON.



---

## Fase 3 Pos Penjagaan (Middleware)

Ini adalah proses pemeriksaan token dan logikanya berfungsi sebagai penyaring atau filter.

1. **Ambil Header**
Baca isi `Authorization` dari header request yang dikirim user.
2. **Bersihkan String**
Buang awalan "Bearer " menggunakan `strings.Replace` dengan limit 1. Kita hanya butuh kode tokennya saja.
3. **Parsing Token**
Gunakan fungsi `jwt.ParseWithClaims`.
* Masukkan string token yang sudah bersih.
* Masukkan `&Claims{}` kosong sebagai wadah hasil terjemahan.
* Buat fungsi *callback* yang mengembalikan `jwtKey` dan `nil` (menggunakan `interface{}`).


4. **Validasi**
Cek dua hal yaitu apakah `err` bernilai nil dan apakah `token.Valid` bernilai true.
5. **Teruskan Request**
Jika semua aman panggil `next.ServeHTTP` atau `next(w, r)` untuk mengizinkan user lanjut ke fungsi tujuan (endpoint rahasia).

---

## Ringkasan Mantra (Cheat Sheet)

Agar tidak lupa ingatlah poin-poin kunci berikut ini.

* **Hashing** -> Digunakan saat **Register** (simpan password aman).
* **Signing** -> Digunakan saat **Login** (bikin token).
* **Parsing** -> Digunakan saat **Middleware** (baca token).
* **Pointer (&)** -> Wajib dipakai saat inisialisasi **Claims**.
* **Interface{}** -> Dipakai di fungsi callback karena library menerima tipe kunci apa saja.