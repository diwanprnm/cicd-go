# Study Case 1 — GitHub Actions + Go

**Yang akan kamu lakukan:** menjalankan pipeline yang **GAGAL**, mencari
penyebabnya, memperbaikinya, lalu melihatnya **HIJAU**.

Ini bukan tutorial "salin tempel lalu selesai". Kamu akan melihat pipeline
merah dulu — karena di situlah letak pelajarannya.

> **Baru pertama kali?** Baca [`../00-KONSEP.md`](../00-KONSEP.md) dulu.
> Sekitar 15 menit, dan itu membuat semuanya jauh lebih masuk akal.

---

## Prasyarat

Cek dulu Go sudah terpasang:

```powershell
go version
```

Kalau belum ada, unduh di <https://go.dev/dl/> lalu **tutup dan buka ulang**
terminalmu.

---

## Aplikasi: Keranjang Belanja

Aplikasi kecil yang menghitung total belanja dan diskon. Aturannya:

| Total belanja | Diskon |
|---|---|
| di bawah 100 | 0% |
| 100 sampai 499 | 10% |
| 500 ke atas | 20% |

Sengaja sederhana. Perhatianmu harus ke **pipelinenya**, bukan ke kode ini.

```
01-github-actions-go/
├── cmd/keranjang/main.go              # titik masuk aplikasi
├── internal/keranjang/keranjang.go    # logika diskon  ← ada BUG di sini
├── internal/keranjang/keranjang_test.go
├── internal/api/api.go                # HTTP API
├── internal/api/api_test.go
├── Dockerfile                          # cara aplikasi dikemas
└── .github/workflows/pipeline.yml      # ← PIPELINE-NYA DI SINI
```

---

## BAGIAN 1 — Lihat kegagalannya

### Langkah 1: Jalankan test di laptopmu

```powershell
cd 01-github-actions-go
go test ./...
```

### ✋ PREDIKSI DULU

Sebelum membaca hasilnya, tebak dulu:

> Menurutmu, apakah semua test akan lulus? Atau ada yang gagal?
> Kalau ada yang gagal, berapa banyak?

Tulis tebakanmu. **Salah tebak jauh lebih berguna daripada langsung tahu** —
karena saat tebakanmu salah, kamu akan ingat kenapa.

### Langkah 2: Jalankan dan amati

```powershell
go test ./...
```

Hasilnya seperti ini (persisnya bisa sedikit berbeda):

```
--- FAIL: TestPersenDiskonPerBatas
    PersenDiskon(100) = 0%, ingin 10%

--- FAIL: TestHitungDiskonTepatDiBatasSeratus
    persen_diskon = 0, ingin 10 (total tepat 100 harus dapat diskon)

FAIL
FAIL	github.com/diwan/keranjang/internal/keranjang
FAIL	github.com/diwan/keranjang/internal/api
```

### 💡 Yang perlu kamu perhatikan

Ada **dua** test gagal, di **dua folder berbeda**. Tapi perhatikan
pesan errornya baik-baik:

```
PersenDiskon(100) = 0%, ingin 10%          ← test di internal/keranjang
persen_diskon = 0, ingin 10                ← test di internal/api
```

**Keduanya soal angka 100 yang sama.** Keduanya bilang: "yang didapat 0,
yang seharusnya 10".

Ini pelajaran pertama yang penting:

> **Satu bug sering muncul sebagai banyak kegagalan test.**
>
> Jangan panik melihat 10 test merah. Cari **satu akar penyebabnya**.
> Sering kali 10 kegagalan itu berasal dari satu baris yang salah.

### Langkah 3: Cari akarnya

Buka `internal/keranjang/keranjang.go`, cari fungsi `persenDiskon`:

```go
func PersenDiskon(total float64) float64 {
	switch {
	case total >= 500:
		return 20
	case total > 100:      // ← BUG: seharusnya >= 100
		return 10
	default:
		return 0
	}
}
```

**Kenapa ini salah?**

Baca aturannya sekali lagi: *"total 100 sampai 499 dapat diskon 10%"*.

Tanda `>` artinya **"lebih besar dari"** — tidak termasuk 100 itu sendiri.
Jadi kalau totalnya **tepat 100**, kondisinya `100 > 100` → **salah** →
jatuh ke `default` → dapat diskon **0%**.

Padahal seharusnya dapat 10%.

Yang benar adalah `>=` — **"lebih besar atau sama dengan"**.

### 🧠 Kenapa bug ini menarik?

Perhatikan hal-hal ini:

1. **Kode ini bisa dikompilasi.** Go tidak protes sama sekali.
2. **Aplikasinya jalan normal.** Tidak crash, tidak error.
3. **Hanya bermasalah untuk satu angka:** tepat 100.

Kalau kamu menguji manual, kamu kemungkinan besar mencoba belanja 50
(dapat 0%, benar) atau 150 (dapat 10%, benar). **Kamu tidak akan mencoba
tepat 100** — karena secara naluri manusia, kita menguji angka yang
"jelas", bukan angka di titik batas.

Inilah alasan kenapa CI/CD ada:

> Robot tidak punya naluri. Robot menguji **semua** kasus yang kamu
> tuliskan, termasuk titik batas yang membosankan dan mudah terlupa.

### Langkah 4: Perbaiki

Ubah `>` menjadi `>=`:

```go
	case total >= 100:
		return 10
```

### Langkah 5: Jalankan lagi

```powershell
go test ./...
```

Sekarang:

```
ok  	github.com/diwan/keranjang/internal/api
ok  	github.com/diwan/keranjang/internal/keranjang
```

**HIJAU.** 🎉

### ✅ Yang baru saja kamu lakukan

Itu **persis** yang dilakukan pipeline, dan itu sebabnya pipeline berguna:

| | Manual (kamu) | Otomatis (pipeline) |
|---|---|---|
| Menjalankan test | Kamu ketik `go test` | Robot ketik `go test` |
| Melihat hasil | Kamu baca sendiri | Robot laporkan ke GitHub |
| Kapan | Saat kamu ingat | Setiap kali ada push |
| Kalau lupa | Bug lolos ke produksi | Tidak bisa lupa |

Pipeline CI **tidak melakukan hal yang ajaib**. Ia melakukan hal yang
sama persis dengan yang baru kamu lakukan — tapi **otomatis** dan
**tidak bisa lupa**.

---

## BAGIAN 2 — Jalankan di GitHub

Sekarang bagian yang menarik: membuat robot melakukannya untukmu.

### Langkah 1: Buat repo di GitHub

Buat repository baru di GitHub (boleh private). **Jangan** centang
"Add a README" — kita akan push dari laptop.

### Langkah 2: Push kode ini ke sana

```powershell
cd 01-github-actions-go

git init
git add .
git commit -m "Study case CI/CD: aplikasi keranjang"
git branch -M main
git remote add origin https://github.com/USERNAME/NAMA-REPO.git
git push -u origin main
```

Ganti `USERNAME` dan `NAMA-REPO` dengan punyamu.

### Langkah 3: Tonton pipeline-nya berjalan

Buka repo itu di browser → klik tab **Actions**.

Kamu akan melihat pipeline bernama **"Pipeline Keranjang"** sedang berjalan
(ikon kuning berputar). Klik untuk melihat detailnya.

### 💡 Yang akan kamu lihat

```
┌─────────────────────────────────────────────────────┐
│  1. Build — kompilasi kode          ✓  12s          │
│         ↓                                            │
│  2. Test — uji perilaku kode        ✓  18s          │
│         ↓                                            │
│  3. Image — bungkus jadi Docker     ✓  35s          │
└─────────────────────────────────────────────────────┘
```

Klik salah satu job, lalu klik salah satu step. **Kamu bisa membaca
seluruh output perintahnya** — persis seperti yang kamu lihat di terminal
laptopmu tadi.

> **Inilah CI.** Komputer milik GitHub menjalankan perintah yang sama
> dengan yang kamu jalankan di laptop — otomatis, di komputernya sendiri.

### 🧪 Percobaan: rusak lagi, sengaja

Sekarang bagian yang paling berguna. **Sengaja rusak kodenya**, lalu push:

```powershell
# Kembalikan bug-nya
# Di internal/keranjang/keranjang.go, ubah `>= 100` kembali jadi `> 100`

git add .
git commit -m "sengaja rusak, untuk melihat pipeline merah"
git push
```

Buka tab **Actions** lagi. Sekarang:

```
┌─────────────────────────────────────────────────────┐
│  1. Build — kompilasi kode          ✓  12s          │
│         ↓                                            │
│  2. Test — uji perilaku kode        ✗  18s          │
│         ↓                                            │
│  3. Image — bungkus jadi Docker     ⊘  dilewati     │
└─────────────────────────────────────────────────────┘
```

**Perhatikan dua hal:**

1. **Job `build` tetap HIJAU.** Bug ini tidak mengganggu kompilasi.
   Ini membuktikan yang saya bilang di `00-KONSEP.md`: build lulus ≠
   perilakunya benar.

2. **Job `image` DILEWATI** (ikon ⊘), tidak dijalankan. Karena job itu
   menulis `needs: test`, dan test gagal. Ini **fail fast** bekerja:
   kenapa buang waktu membuat image kalau testnya saja gagal?

Sekarang perbaiki lagi, push lagi, dan lihat kembali hijau. Kamu baru saja
mengalami sendiri siklus lengkap CI/CD.

---

## BAGIAN 3 — Membaca pipeline-nya

Sekarang buka `.github/workflows/pipeline.yml` dengan pemahaman baru.

### Struktur besarnya

```yaml
name: Pipeline Keranjang    # nama yang muncul di tab Actions

on:                         # KAPAN jalan
  push:
    branches: [main]
  pull_request:
    branches: [main]
  workflow_dispatch:        # tombol "Run workflow" manual

jobs:                       # APA saja yang dikerjakan
  build:  ...
  test:   ...
  image:  ...
```

### Job 1: `build`

```yaml
build:
  runs-on: ubuntu-latest        # komputer apa yang dipakai
  steps:
    - uses: actions/checkout@v4   # ambil kode ke komputer itu
    - uses: actions/setup-go@v5   # pasang Go
      with:
        go-version: '1.22'
    - run: go build -o bin/keranjang ./cmd/keranjang
```

**`uses:` vs `run:`** — dua hal berbeda:

| | Artinya | Contoh |
|---|---|---|
| `uses:` | Pakai aksi siap pakai dari orang lain | `actions/checkout@v4` |
| `run:` | Jalankan perintah shell biasa | `go build ./...` |

`actions/checkout@v4` itu penting dan sering dilupakan pemula:
**runner GitHub mulai dalam keadaan KOSONG.** Tidak ada kode kamu di sana.
Tanpa `checkout`, perintah `go build` akan gagal karena tidak ada file apa pun.

### Job 2: `test`

```yaml
test:
  needs: build                  # ← tunggu job build selesai dulu
  steps:
    - run: go test -race -coverprofile=coverage.out ./...
```

`needs: build` inilah yang menciptakan urutan. Tanpa baris itu,
kedua job jalan **bersamaan** — dan job `test` mungkin selesai lebih dulu
daripada `build`, yang membingungkan.

### Job 3: `image`

```yaml
image:
  needs: test
  if: github.event_name != 'pull_request'   # ← perhatikan ini
```

**Kenapa ada `if:` di sini?**

Kalau yang memicu pipeline adalah **pull request**, job ini **dilewati**.

Alasannya: pull request belum tentu akan digabung ke `main`. Kode yang
belum disetujui orang lain **tidak boleh** menghasilkan image yang bisa
dipakai deploy.

Jadi di pull request, kamu tetap dapat umpan balik *"test-nya lulus atau
tidak"* — tapi tanpa membuang waktu membuat image.

### 💡 Konsep penting: Artifact

Ini yang paling sering bikin pemula bingung, jadi perhatikan baik-baik.

**Setiap job berjalan di komputer yang BERBEDA.** Job `build` berjalan di
komputer #1, job `test` di komputer #2. Komputer #1 sudah dimatikan dan
dibuang sebelum komputer #2 menyala.

Jadi kalau job `build` menghasilkan file, **job `test` tidak akan
menemukannya.** File itu hilang bersama komputer #1.

Untuk memindahkan file antar job, pakai **artifact**:

```yaml
# Di job build — UNGGAH hasilnya
- uses: actions/upload-artifact@v4
  with:
    name: binary-keranjang
    path: bin/keranjang

# Di job lain — UNDUH hasilnya
- uses: actions/download-artifact@v4
  with:
    name: binary-keranjang
```

Anggap saja seperti menempel file di grup chat. Yang mengunggah, meletakkan;
yang mengunduh, mengambil.

**Kalau kamu lupa langkah ini**, kamu akan lihat error aneh seperti
`No such file or directory` padahal di job sebelumnya file itu jelas ada.
Kalau itu terjadi, ingat: **komputernya berbeda.**

---

## BAGIAN 4 — Latihan

Kerjakan berurutan. Yang pertama paling mudah.

### Latihan 1: Ubah aturan diskon

Ubah aturannya: belanja 200 ke atas dapat diskon 15%.

Kamu harus mengubah **tiga tempat**:
1. `internal/keranjang/keranjang.go` — logika `PersenDiskon`
2. `internal/keranjang/keranjang_test.go` — test-nya
3. `internal/api/api_test.go` — kalau ada test yang terpengaruh

**Pertanyaan:** kenapa harus tiga tempat? Apa yang terjadi kalau kamu
hanya mengubah yang pertama dan lupa yang kedua?

### Latihan 2: Tambah endpoint baru

Tambahkan endpoint `GET /versi` yang mengembalikan nomor versi aplikasi.

Petunjuk: variabel `version` di `cmd/keranjang/main.go` sudah ada, tapi
belum pernah dipakai. Kamu perlu meneruskannya ke `api.NewServer()`.

**Pertanyaan:** setelah kamu selesai, apakah cakupan test naik atau turun?
Cek dengan:

```powershell
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | tail -n 1
```

### Latihan 3: Tambah quality gate

Sekarang pipeline **tidak** memeriksa cakupan test. Tambahkan pemeriksaan
itu di job `test`, dengan batas minimal 70%.

Petunjuk: lihat versi GitLab di `../02-gitlab-ci-node/.gitlab-ci.yml`,
cari bagian "Quality gate".

**Pertanyaan:** kenapa cakupan test lebih berguna sebagai **gerbang**
yang menggagalkan pipeline, daripada sebagai **angka** di laporan?

### Latihan 4: Percepat pipeline

Lihat tab Actions, catat durasi tiap job. Totalnya berapa?

Sekarang pikirkan: apakah ada job yang bisa dijalankan **bersamaan**
alih-alih berurutan? Apa risikonya kalau kamu melakukannya?

### Latihan 5: Matriks versi

Jalankan test di Go 1.21 **dan** 1.22 sekaligus:

```yaml
test:
  strategy:
    matrix:
      go-version: ['1.21', '1.22']
  steps:
    - uses: actions/setup-go@v5
      with:
        go-version: ${{ matrix.go-version }}
```

**Pertanyaan:** kenapa ini berguna? Petunjuk: pikirkan kalau aplikasimu
dipakai 5 tim berbeda yang belum tentu memakai versi Go yang sama.

### Latihan 6: Perbaiki "quality gate" yang menipu

Ada satu masalah tersembunyi di langkah "Cek format kode" di pipeline:

```yaml
- run: |
    belum_diformat=$(gofmt -l .)
    if [ -n "$belum_diformat" ]; then exit 1; fi
```

`gofmt -l .` **selalu keluar dengan status sukses**, walaupun ada file
yang belum diformat. Jadi kalau kamu tidak memeriksa outputnya (seperti
di atas), pemeriksaan itu **tidak berguna sama sekali** — tapi tetap
terlihat hijau di pipeline.

**Pertanyaan:** coba hapus baris `if`-nya, push, dan lihat. Apakah pipeline
masih hijau padahal ada file yang belum diformat? Ini contoh paling nyata
dari: **"hijau hanya sekuat pemeriksaan yang kamu tulis."**

---

## Bereskan

Kalau kamu sudah selesai dan ingin menghapus image Docker yang dibuat
percobaan tadi:

```powershell
docker rm -f uji 2>$null
docker rmi keranjang:latest keranjang:sha-* 2>$null
```

---

## Lanjut ke mana?

Kalau kamu sudah selesai di sini, coba versi GitLab-nya:
[`../02-gitlab-ci-node/README.md`](../02-gitlab-ci-node/README.md)

Aplikasinya **sengaja dibuat sama** (keranjang belanja), tapi ditulis
dalam Node.js dan dijalankan GitLab. Kamu akan melihat sendiri bahwa
yang berubah hanya sintaksnya — tiga stage-nya tetap identik.
