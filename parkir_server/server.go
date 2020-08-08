package main

import (
	"encoding/json"
	"github.com/chilts/sid"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"
)

type Kendaraan struct {
	ID         string `json:"id"`
	WaktuMasuk [4]int `json:"waktu_masuk"`
}

type Detail struct {
	Plat string `json:"plat"`
	Tipe string `json:"tipe"`
}

type Tagihan struct {
	Bayar int    `json:"bayar"`
	Plat  string `json:"plat"`
	Tipe  string `json:"tipe"`
	Waktu int    `json:"waktu"`
}

func (k Kendaraan) Keluar(WaktuKeluar [4]int, Tipe string, platKendaraan string) Tagihan {
	mobilDetPertama := 5_000
	motorDetPertama := 3_000
	mobilPerSecond := 3_000
	motorPerSecond := 2_000
	bayar := 0

	selesihArray, selisihDay := rangeTime(k.WaktuMasuk, WaktuKeluar)
	if selisihDay >= 1 {
		//tarif flat bayar jika > 1 hari
		bayar := 300_000_000

		return Tagihan{
			Bayar: bayar,
			Plat:  platKendaraan,
			Tipe:  Tipe,
			// satu hari = 86400 detik
			Waktu: 86400,
		}
	}

	selisihSecond := convertToSecond(selesihArray)
	if Tipe == "mobil" {
		bayar = mobilDetPertama + (mobilPerSecond * (selisihSecond - 1))
	} else if Tipe == "motor" {
		bayar = motorDetPertama + (motorPerSecond * (selisihSecond - 1))
	}

	return Tagihan{
		Bayar: bayar,
		Plat:  platKendaraan,
		Tipe:  Tipe,
		Waktu: selisihSecond,
	}

}

// initating kendaraan ( MAP )
var kendaraan = make(map[string]Kendaraan)

func main() {

	// ========= Mock Data Test ============
	for i := 0; i < 2; i++ {
		time.Sleep(2 * time.Second)
		idParkir2 := sid.Id()
		kendaraan[idParkir2] = Kendaraan{
			ID:         idParkir2,
			WaktuMasuk: waktuTerkini(time.Now()),
		}
	}

	// init Router
	router := mux.NewRouter()

	// Handle EndPoints/Routing
	router.HandleFunc("/parkir/parkir_server", getKendaraanAll).Methods("GET")
	router.HandleFunc("/parkir/parkir_server/{id}", getKendaraan).Methods("GET")
	router.HandleFunc("/parkir/parkir_server/masuk", createKendaraanMasuk).Methods("POST")
	router.HandleFunc("/parkir/parkir_server/keluar/{id}", updateKendaraanKeluar).Methods("POST")

	// untuk tau kapan fail koneksi
	log.Fatal(http.ListenAndServe(":8080", router))
}

// ========================= PARKING API ==============================
// Show All from Kendaraan
func getKendaraanAll(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "application/json")

	jsonData, err := json.Marshal(kendaraan)
	if err != nil {
		log.Println(err)
	}
	_, _ = writer.Write(jsonData)
}

// Show Kendaraan by ID
func getKendaraan(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")

	params := mux.Vars(request)
	if key, found := kendaraan[params["id"]]; found {
		jsonData, err := json.Marshal(key)
		if err != nil {
			log.Println(err)
		}
		_, _ = writer.Write(jsonData)
	}
}

// Crate new kendaraan masuk
func createKendaraanMasuk(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "application/json")

	idParkir := sid.Id()
	kendaraan[idParkir] = Kendaraan{
		ID:         idParkir,
		WaktuMasuk: waktuTerkini(time.Now()),
	}

	jsonData, err := json.Marshal(kendaraan[idParkir])
	if err != nil {
		log.Println(err)
	}
	_, _ = writer.Write(jsonData)
}

// Pembayaran kendaraan masuk
func updateKendaraanKeluar(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")

	params := mux.Vars(request)

	id := params["id"]
	if _, found := kendaraan[id]; found {
		var detail Detail
		_ = json.NewDecoder(request.Body).Decode(&detail)

		waktuKeluar := waktuTerkini(time.Now())

		tagihan := kendaraan[id].Keluar(waktuKeluar, detail.Tipe, detail.Plat)
		delete(kendaraan, id)
		jsonData, err := json.Marshal(tagihan)

		if err != nil {
			log.Println(err)
		}
		_, _ = writer.Write(jsonData)
	}
}

// ======================= MY FUNC ===================================
func convertToSecond(array [3]int) int {
	array[0] = array[0] * 3600
	array[1] = array[1] * 60

	result := 0
	for _, v := range array {
		result += v
	}

	return result
}

func rangeTime(masuk [4]int, keluar [4]int) ([3]int, int) {
	var newArr [4]int
	for i := 0; i < 4; i++ {
		newArr[i] = keluar[i] - masuk[i]
	}
	return [3]int{newArr[1], newArr[2], newArr[3]}, newArr[0]
}

func waktuTerkini(now time.Time) [4]int {
	day := now.Day()
	hour := now.Hour()
	minute := now.Minute()
	second := now.Second()

	return [4]int{day, hour, minute, second}
}
