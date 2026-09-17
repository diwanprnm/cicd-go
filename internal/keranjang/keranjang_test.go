package keranjang_test

import (
	"errors"
	"testing"

	"github.com/diwan/keranjang/internal/keranjang"
)

func TestSubtotalMengalikanHargaDenganJumlah(t *testing.T) {
	item := keranjang.Item{Nama: "kopi", Harga: 25000, Jumlah: 3}

	if got := item.Subtotal(); got != 75000 {
		t.Errorf("Subtotal() = %.0f, ingin 75000", got)
	}
}

func TestTambahMenolakJumlahNolAtauNegatif(t *testing.T) {
	for _, jumlah := range []int{0, -1, -100} {
		k := &keranjang.Keranjang{}

		err := k.Tambah(keranjang.Item{Nama: "teh", Harga: 5000, Jumlah: jumlah})
		if !errors.Is(err, keranjang.ErrJumlahTidakValid) {
			t.Errorf("Tambah(jumlah=%d) error = %v, ingin %v",
				jumlah, err, keranjang.ErrJumlahTidakValid)
		}
	}
}

func TestTotalMenjumlahkanSemuaItem(t *testing.T) {
	k := &keranjang.Keranjang{}

	// 25000 x 2 = 50000
	if err := k.Tambah(keranjang.Item{Nama: "kopi", Harga: 25000, Jumlah: 2}); err != nil {
		t.Fatalf("Tambah() error: %v", err)
	}
	// 15000 x 1 = 15000
	if err := k.Tambah(keranjang.Item{Nama: "roti", Harga: 15000, Jumlah: 1}); err != nil {
		t.Fatalf("Tambah() error: %v", err)
	}

	if got := k.Total(); got != 65000 {
		t.Errorf("Total() = %.0f, ingin 65000", got)
	}
}

// TestPersenDiskonPerBatas menguji SETIAP titik batas dengan tepat.
//
// Inilah gunanya unit test yang tidak bisa digantikan manual testing:
// menguji nilai TEPAT di titik batas, yaitu 100 dan 500.
//
// Manusia cenderung menguji 99 dan 501 — angka yang "jelas" — lalu
// melewatkan justru titik yang paling rawan salah. Robot tidak punya
// naluri seperti itu, dan itulah keunggulannya.
func TestPersenDiskonPerBatas(t *testing.T) {
	kasus := []struct {
		total float64
		ingin float64
	}{
		// Di bawah batas pertama
		{0, 0},
		{99, 0},

		// Tepat di batas pertama — paling rawan salah
		{100, 10},

		// Di antara dua batas
		{499, 10},

		// Tepat di batas kedua — paling rawan salah
		{500, 20},

		// Jauh di atas batas kedua
		{1000, 20},
	}

	for _, k := range kasus {
		if got := keranjang.PersenDiskon(k.total); got != k.ingin {
			t.Errorf("PersenDiskon(%.0f) = %.0f%%, ingin %.0f%%",
				k.total, got, k.ingin)
		}
	}
}

func TestBayarMengurangiTotalDenganDiskon(t *testing.T) {
	k := &keranjang.Keranjang{}

	// 3 x 20000 = 60000 -> diskon 20% -> bayar 48000
	if err := k.Tambah(keranjang.Item{Nama: "sepatu", Harga: 20000, Jumlah: 3}); err != nil {
		t.Fatalf("Tambah() error: %v", err)
	}

	if got := k.Bayar(); got != 48000 {
		t.Errorf("Bayar() = %.0f, ingin 48000", got)
	}
}

func TestBayarTanpaDiskonDiBawahSeratus(t *testing.T) {
	k := &keranjang.Keranjang{}

	// 5 x 10 = 50. Totalnya di bawah 100, jadi TIDAK dapat diskon.
	// Hati-hati memilih angka di sini: kalau kamu pakai 5 x 1000 = 5000,
	// totalnya justru masuk kategori diskon 20%, dan test-nya akan gagal
	// karena alasan yang tidak kamu duga.
	if err := k.Tambah(keranjang.Item{Nama: "permen", Harga: 10, Jumlah: 5}); err != nil {
		t.Fatalf("Tambah() error: %v", err)
	}

	if got := k.Bayar(); got != 50 {
		t.Errorf("Bayar() = %.0f, ingin 50 (tanpa diskon)", got)
	}
}

func TestKeranjangKosongBayarNol(t *testing.T) {
	k := &keranjang.Keranjang{}

	if got := k.Bayar(); got != 0 {
		t.Errorf("Bayar() = %.0f, ingin 0", got)
	}
}

func TestItemsMengembalikanSalinan(t *testing.T) {
	k := &keranjang.Keranjang{}
	if err := k.Tambah(keranjang.Item{Nama: "kopi", Harga: 10000, Jumlah: 1}); err != nil {
		t.Fatalf("Tambah() error: %v", err)
	}

	// Mengubah hasil Items() tidak boleh mengubah isi keranjang.
	salinan := k.Items()
	salinan[0].Nama = "DIUBAH"

	if k.Items()[0].Nama != "kopi" {
		t.Error("Items() membocorkan isi internal keranjang")
	}
}
