package main

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"kasir-api/database"
	"kasir-api/handlers"
	"kasir-api/repositories"
	"kasir-api/services"
)

type Category struct {
	ID int `json:"id"`
	Name string `json:"name"`
	Description string `json:"description"`
}

var category = []Category {
	{ID: 1, Name: "Makanan", Description: "Ini kategori makanan"},
	{ID: 2, Name: "Minuman", Description: "Ini kategori minuman"},
	{ID: 3, Name: "Alat Tulis", Description: "Ini kategori alat tulis"},
}

type Config struct {
	Port string `mapstructure:"PORT"`
	DBConn string `mapstructure:"DB_CONN"`
}

func main() {

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if _, err := os.Stat(".env"); err == nil {
		viper.SetConfigFile(".env")
		_ = viper.ReadInConfig()
	}

	config := Config{
		Port: viper.GetString("PORT"),
		DBConn: viper.GetString("DB_CONN"),
	}

	fmt.Println("DB Connection String:", config.DBConn)

	db, err := database.InitDB(config.DBConn)
	if err != nil {
		log.Println("Warning: Failed to initialize database:", err)
		log.Println("Running without database connection")
		// Continue without database
	} else {
		defer db.Close()
		productRepo := repositories.NewProductRepository(db)
		productService := services.NewProductService(productRepo)
		productHandler := handlers.NewProductHandler(productService)
		
		http.HandleFunc("/api/produk", productHandler.HandleProducts)
		http.HandleFunc("/api/produk/", func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				productHandler.GetByID(w, r)
			case http.MethodPut:
				productHandler.Update(w, r)
			case http.MethodDelete:
				productHandler.Delete(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})
	}

	http.HandleFunc("/api/categories/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			getCategoryByID(w, r)
		} else if r.Method == "PUT" {
			updateCategory(w, r)
		} else if r.Method == "DELETE" {
			deleteCategory(w, r)
		}
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request){
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "OK",
			"message": "API Running",
		})
	})

/*
	http.HandleFunc("/api/categories", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(category)
		
		} else if r.Method == "POST" {
			var categoryBaru Category
			err := json.NewDecoder(r.Body).Decode(&categoryBaru)
			if err != nil {
				http.Error(w, "Invalid Request", http.StatusBadRequest)
				return
			}

			categoryBaru.ID = len(category) + 1
			category = append(category, categoryBaru)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(categoryBaru)
		}
	})
*/

	addr := "0.0.0.0:" + config.Port
	fmt.Println("Server running di", addr)

	err = http.ListenAndServe(addr, nil)
	if (err != nil) {
		fmt.Println("gagal running server", err)
	}
}

// func getProdukByID(w http.ResponseWriter, r *http.Request) {
// 	idStr := strings.TrimPrefix(r.URL.Path, "/api/produk/")
// 	id, err := strconv.Atoi(idStr)
// 	if err != nil {
// 		http.Error(w, "Invalid Produk ID", http.StatusBadRequest)
// 		return
// 	}

// 	for _, p := range produk {
// 		if p.ID == id {
// 			w.Header().Set("Content-Type", "application/json")
// 			json.NewEncoder(w).Encode(p)
// 			return
// 		}
// 	}
// }

// func updateProduk(w http.ResponseWriter, r *http.Request) {
// 	idStr := strings.TrimPrefix(r.URL.Path, "/api/produk/")

// 	id, err := strconv.Atoi(idStr)
// 	if err != nil {
// 		http.Error(w, "Invalid Produk ID", http.StatusBadRequest)
// 		return
// 	}

// 	var updateProduk Produk
// 	err = json.NewDecoder(r.Body).Decode(&updateProduk)
// 	if err != nil {
// 		http.Error(w, "Invalid request", http.StatusBadRequest)
// 		return
// 	}

// 	for i := range produk {
// 		if produk[i].ID == id {
// 			updateProduk.ID = id
// 			produk[i] = updateProduk

// 			w.Header().Set("Content-Type", "application/json")
// 			json.NewEncoder(w).Encode(updateProduk)
// 			return
// 		}
// 	}

// 	http.Error(w, "Produk belum ada", http.StatusNotFound)
// }

// func deleteProduk(w http.ResponseWriter, r *http.Request) {

// 	idStr := strings.TrimPrefix(r.URL.Path, "/api/produk/")

// 	id, err := strconv.Atoi(idStr)
// 	if err != nil {
// 		http.Error(w, "Invalid Produk ID", http.StatusBadRequest)
// 		return
// 	}

// 	for i, p := range produk {
// 		if p.ID == id {
// 			produk = append(produk[:i], produk[i+1:]...)

// 			w.Header().Set("Content-Type", "application/json")
// 			json.NewEncoder(w).Encode(map[string]string{
// 				"message": "sukses delete",
// 			})
// 			return
// 		}
// 	}

// 	http.Error(w, "Produk belum ada", http.StatusNotFound)
// }

func getCategoryByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/categories/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Category ID", http.StatusBadRequest)
		return
	}

	for _, c := range category {
		if c.ID == id {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(c)
			return
		}
	}
}

func updateCategory(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/categories/")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Category ID", http.StatusBadRequest)
		return
	}

	var updateCategory Category
	err = json.NewDecoder(r.Body).Decode(&updateCategory)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	for i := range category {
		if category[i].ID == id {
			updateCategory.ID = id
			category[i] = updateCategory

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(updateCategory)
			return
		}
	}

	http.Error(w, "Kategori belum ada", http.StatusNotFound)
}

func deleteCategory(w http.ResponseWriter, r *http.Request) {

	idStr := strings.TrimPrefix(r.URL.Path, "/api/categories/")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Category ID", http.StatusBadRequest)
		return
	}

	for i, c := range category {
		if c.ID == id {
			category = append(category[:i], category[i+1:]...)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"message": "sukses delete",
			})
			return
		}
	}

	http.Error(w, "Kategori belum ada", http.StatusNotFound)
}