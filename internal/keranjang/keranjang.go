// Package keranjang menghitung total belanja dan diskon.
//
// Aplikasinya sengaja dibuat kecil. Tujuannya bukan belajar Go —
// tujuannya supaya ada kode nyata yang bisa di-build, di-test, dan di-deploy
// oleh pipeline. Perhatianmu harus ke pipelinenya, bukan ke kode ini.
package keranjang

import "errors"

// ErrJumlahTidakValid dikembalikan kalau jumlah item nol atau negatif.
var ErrJumlahTidakValid = errors.New("jumlah harus lebih dari nol")

// Item adalah satu barang di keranjang.
type Item struct {
	Nama   string  `json:"nama"`
	Harga  float64 `json:"harga"`
	Jumlah int     `json:"jumlah"`
}

// Subtotal menghitung harga dikali jumlah.
func (i Item) Subtotal() float64 {
	return i.Harga * float64(i.Jumlah)
}

// Keranjang menampung daftar item.
type Keranjang struct {
	items []Item
}

// Tambah memasukkan satu item ke keranjang.
func (k *Keranjang) Tambah(item Item) error {
	if item.Jumlah <= 0 {
		return ErrJumlahTidakValid
	}
	k.items = append(k.items, item)
	return nil
}

// Items mengembalikan salinan daftar item.
func (k *Keranjang) Items() []Item {
	out := make([]Item, len(k.items))
	copy(out, k.items)
	return out
}

// Total menghitung jumlah seluruh subtotal, sebelum diskon.
func (k *Keranjang) Total() float64 {
	var total float64
	for _, item := range k.items {
		total += item.Subtotal()
	}
	return total
}

// PersenDiskon mengembalikan persentase diskon berdasarkan total belanja.
//
// Aturannya:
//
//	total < 100        -> 0%
//	100 <= total < 500 -> 10%
//	total >= 500       -> 20%
func PersenDiskon(total float64) float64 {
	switch {
	case total >= 500:
		return 20
	case total >= 100:
		return 10
	default:
		return 0
	}
}

// Bayar menghitung yang harus dibayar setelah diskon.
func (k *Keranjang) Bayar() float64 {
	total := k.Total()
	diskon := total * PersenDiskon(total) / 100
	return total - diskon
}
