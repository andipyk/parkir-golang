package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"
)

type Kendaraan struct {
	ID         string `json:"id"`
	WaktuMasuk [4]int `json:"waktu_masuk"`
}

type Tagihan struct {
	Bayar int    `json:"bayar"`
	Plat  string `json:"plat"`
	Tipe  string `json:"tipe"`
	Waktu int    `json:"waktu"`
}

func (k Kendaraan) Masuk() {
	fmt.Println("id : ", k.ID)
	fmt.Println("Waktu Masuk :", k.WaktuMasuk)
}

func (k Kendaraan) Keluar(WaktuKeluar [4]int, Tipe string, platKendaraan string) {
	mobilDetPertama := 5_000
	motorDetPertama := 3_000
	mobilPerSecond := 3_000
	motorPerSecond := 2_000

	selesihArray, selisihDay := rangeTime(k.WaktuMasuk, WaktuKeluar)
	if selisihDay >= 1 {
		bayar := 300_000_000 //tarif flat bayar jika > 1 hari
		fmt.Printf("%s parkir sudah lewat %d hari.\nAnda harus membayar %d ", platKendaraan, selesihArray[1], bayar)
		return
	}

	selisihSecond := convertToSecond(selesihArray)
	if Tipe == "mobil" {
		bayar := mobilDetPertama + (mobilPerSecond * (selisihSecond - 1))
		fmt.Println("Yang harus dibayar : ", bayar)
		fmt.Printf("%s telah parkir MOBIL selama %d detik\n", platKendaraan, selisihSecond)
	} else if Tipe == "motor" {
		bayar := motorDetPertama + (motorPerSecond * (selisihSecond - 1))
		fmt.Println("Yang harus dibayar : ", bayar)
		fmt.Printf("%s telah parkir MOTOR selama %d detik\n", platKendaraan, selisihSecond)
	} else {
		fmt.Println("ada masalah bray")
	}

}

//var kendaraan map[string]Kendaraan // dari HTTP GET & JSON

type Waktu struct {
	WaktuMasuk time.Time `json:"waktu_masuk"`
}

type Keluar struct {
	Tipe string `json:"tipe"`
	Plat string `json:"plat"`
}

func main() {

	menu := 0 // inisiasi menu
	for menu != 4 {
		fmt.Println("==== ANDI PARKING ☕ ====")
		fmt.Println("1. Kendaraan Masuk")
		fmt.Println("2. Kendaraan Keluar")
		fmt.Println("3. Listen API")
		fmt.Println("4. Exit")
		menu = ScanPilihan()
		switch menu {
		case 1:
			// =============== MASUK ===============
			waktu := Waktu{time.Now()}

			jsonReq, err := json.Marshal(waktu)
			resp, err := http.Post("http://127.0.0.1:8080/parkir/parkir_server/masuk", "application/json; charset=utf-8", bytes.NewBuffer(jsonReq))
			if err != nil {
				log.Fatalln(err)
			} else {
				var kendaraan Kendaraan
				_ = json.NewDecoder(resp.Body).Decode(&kendaraan)

				// =============== PRINT ALL KENDARAAN ===============

				fmt.Println("Key:", kendaraan.ID, "Value:", kendaraan.WaktuMasuk)

			}

		case 2:
			// =============== PRINT ALL KENDARAAN ===============
			response, err := http.Get("http://127.0.0.1:8080/parkir/parkir_server")
			var kendaraan map[string]Kendaraan
			if err != nil {
				fmt.Printf("The HTTP request failed with error %s\n", err)
			} else {
				_ = json.NewDecoder(response.Body).Decode(&kendaraan)

				for key, value := range kendaraan {
					fmt.Println("Key:", key, "Value:", value)
				}
			}

			idKendaraan := ScanString()
			if value, found := kendaraan[idKendaraan]; found {
				fmt.Println(value)
				fmt.Println("Tipe Kendaraan mobil/motor")
				tipeKendaraan := ScanString()
				fmt.Println("plat kendaraan")
				platKendaraan := ScanString()

				keluar := Keluar{tipeKendaraan, platKendaraan}
				jsonReqKeluar, err := json.Marshal(keluar)
				response, err = http.Post("http://127.0.0.1:8080/parkir/parkir_server/keluar/"+idKendaraan, "application/json; charset=utf-8", bytes.NewBuffer(jsonReqKeluar))
				if err != nil {
					log.Fatalln(err)
				} else {
					var tagihan Tagihan
					_ = json.NewDecoder(response.Body).Decode(&tagihan)
					fmt.Println("waktu", tagihan.Waktu, "detik")
					fmt.Println("anda membayar", tagihan.Bayar)
					fmt.Println("success di hapus")
				}
			}

		case 3:
			router := mux.NewRouter()
			router.HandleFunc("/parkir/parkir_client", getKendaraanAll).Methods("GET")
			router.HandleFunc("/parkir/parkir_client/masuk", kendaraanMasuk).Methods("GET")
			router.HandleFunc("/parkir/parkir_client/keluar/{id}", kendaraanKeluar).Methods("POST")

			fmt.Println("Starting the application...\n ")

			// untuk tau kapan fail koneksi
			log.Fatal(http.ListenAndServe(":8088", router))

		case 4:
			fmt.Println("Terimakasih Bapak !")

		default:
			fmt.Println("Maaf Inputan salah")

		}
	}

}

func kendaraanKeluar(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")

	params := mux.Vars(request)
	var keluar Keluar
	_ = json.NewDecoder(request.Body).Decode(&keluar)

	var urlKeluar = "http://127.0.0.1:8080/parkir/parkir_server/keluar/" + params["id"]

	jsonReqKeluar, err := json.Marshal(keluar)
	response, err := http.Post(urlKeluar, "application/json; charset=utf-8", bytes.NewBuffer(jsonReqKeluar))
	if err != nil {
		log.Fatalln(err)
	} else {

		var tagihan Tagihan
		_ = json.NewDecoder(response.Body).Decode(&tagihan)

		_ = json.NewEncoder(writer).Encode(tagihan)
	}
}

func kendaraanMasuk(writer http.ResponseWriter, request *http.Request) {
	// =============== MASUK ===============
	waktu := Waktu{time.Now()}

	jsonReq, err := json.Marshal(waktu)
	_, err = http.Post("http://127.0.0.1:8080/parkir/parkir_server/masuk", "application/json; charset=utf-8", bytes.NewBuffer(jsonReq))
	if err != nil {
		log.Fatalln(err)
	} else {
		_, _ = fmt.Fprintf(writer, "masuk pak eko !")
	}
}

func getKendaraanAll(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	var kendaraan map[string]Kendaraan
	response, err := http.Get("http://127.0.0.1:8080/parkir/parkir_server")
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	} else {
		_ = json.NewDecoder(response.Body).Decode(&kendaraan)
		_ = json.NewEncoder(writer).Encode(kendaraan)
	}
}

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

func ScanPilihan() int {
	var pilihan int
	fmt.Print("Pilihan Anda : ")
	_, err := fmt.Scanf("%d", &pilihan)

	if err != nil {
		fmt.Println(err)
		return 0
	}

	return pilihan
}

func ScanString() string {
	var pilihan string
	fmt.Print("Pilihan Anda : ")
	_, err := fmt.Scanf("%s", &pilihan)

	if err != nil {
		fmt.Println(err)
		return "inputan kosong" // debug mode
	}

	return pilihan
}
