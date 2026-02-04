package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Ini harusnya disimpan di .env
var jwtKey = []byte("kunci_rahasia_yang_aman")

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Struct untuk payload token
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// Database bohongan ya ges ya
var users = map[string]string{}

// Handler halaman rahasia (khusus yang ada token aja)
func welcome(w http.ResponseWriter, r *http.Request) {
	// Kita ngambil username dari context (opsional)
	w.Write([]byte("Selamat Datang! Anda berhasil masuk ke halaman rahasia"))

}

// Fungsi middleware-nya
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Ambil token dari Header Authorization
		// Format standar biasanya: "Bearer <token>"
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Token tidak ditemukan", http.StatusUnauthorized)
			return
		}

		// Pisahkan kata "Bearer" dengan token aslinya
		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

		// Parse token tersebut
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Token tidak valid", http.StatusUnauthorized)
			return
		}

		next(w, r)

	}
}

func register(w http.ResponseWriter, r *http.Request) {
	// Hanya menerima method POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds Credentials
	// Decode JSON dari body request ke struct creds
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Hash password menggunakan bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// Simpan ke database bohongan kita
	users[creds.Username] = string(hashedPassword)

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "User %s berhasil didaftarkan", creds.Username)
}

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Cek apakah user ada di map
	expectedPassword, ok := users[creds.Username]
	if !ok {
		http.Error(w, "User tidak ditemukan", http.StatusUnauthorized)
		return
	}

	// Bandingkan password input dengan hash di map
	err = bcrypt.CompareHashAndPassword([]byte(expectedPassword), []byte(creds.Password))
	if err != nil {
		http.Error(w, "Password salah", http.StatusUnauthorized)
		return
	}

	// >>>> Proses pembuatan JWT Token mulai dari sini <<<<

	// Tentukan waktu kadaluarsa tokennya (minimal 5 menit dari waktu dibuat)
	expirationTime := time.Now().Add(5 * time.Minute)

	// Buat claims atau isi data token
	claims := &Claims{
		Username: creds.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	// Buat token pake algoritma signing HS256 dan claims di atas
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Tanda tanganin token pake kunci rahasia
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Gagal generate token", http.StatusInternalServerError)
		return
	}

	// Kirim token ke user sebagai response
	// Di sini kita bisa kirim lewat cookie atau JSON Body
	// Sementara pake JSON Body aja biar mudah dibaca pake Postman
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": tokenString,
	})
}

func main() {
	http.HandleFunc("/register", register)
	http.HandleFunc("/login", login)
	http.HandleFunc("/welcome", authMiddleware(welcome))

	fmt.Println("Server berjalan di http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
