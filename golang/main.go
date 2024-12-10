package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"gorm.io/driver/sqlite"
	"strconv"
)

type Product struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
	StockQuantity int   `json:"stock_quantity"`
	IsDeleted   bool    `json:"is_deleted"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type ApiResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
	Errors  []string    `json:"errors"`
}

var db *gorm.DB
var err error

// Initialize the database
func InitDb() {
	db, err = gorm.Open(sqlite.Open("./product.db"), &gorm.Config{})
    if err != nil {
        log.Fatal("Error connecting to database: ", err)
    }
    db.AutoMigrate(&Product{})
}

// Generic repository for CRUD operations
type GenericRepository struct {
	DB *gorm.DB
}

func (repo *GenericRepository) GetAll() ([]Product, error) {
	var products []Product
	err := repo.DB.Where("is_deleted = ?", false).Find(&products).Error
	if err != nil {
		return nil, err
	}
	return products, nil
}

func (repo *GenericRepository) GetById(id uint) (*Product, error) {
	var product Product
	err := repo.DB.Where("id = ? AND is_deleted = ?", id, false).First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (repo *GenericRepository) Create(product *Product) (*Product, error) {
	err := repo.DB.Create(product).Error
	if err != nil {
		return nil, err
	}
	return product, nil
}

func (repo *GenericRepository) Update(product *Product) (*Product, error) {
	err := repo.DB.Save(product).Error
	if err != nil {
		return nil, err
	}
	return product, nil
}

func (repo *GenericRepository) Delete(id uint) (bool, error) {
	var product Product
	err := repo.DB.Where("id = ? AND is_deleted = ?", id, false).First(&product).Error
	if err != nil {
		return false, err
	}
	product.IsDeleted = true
	repo.DB.Save(&product)
	return true, nil
}

// Initialize repository
var productRepo *GenericRepository

// Handlers
func GetAllProducts(w http.ResponseWriter, r *http.Request) {
	products, err := productRepo.GetAll()
	if err != nil {
		http.Error(w, "Error fetching products", http.StatusInternalServerError)
		return
	}
	response := ApiResponse{Success: true, Data: products, Message: "Products retrieved successfully"}
	respondWithJSON(w, response)
}

func GetProductById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	// Convert string id to uint
	productID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}
	product, err := productRepo.GetById(uint(productID))
	if err != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}
	response := ApiResponse{Success: true, Data: product, Message: "Product retrieved successfully"}
	respondWithJSON(w, response)
}

func CreateProduct(w http.ResponseWriter, r *http.Request) {
	var product Product
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&product)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	createdProduct, err := productRepo.Create(&product)
	if err != nil {
		http.Error(w, "Error creating product", http.StatusInternalServerError)
		return
	}
	response := ApiResponse{Success: true, Data: createdProduct, Message: "Product created successfully"}
	respondWithJSON(w, response)
}

func UpdateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	var product Product
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&product)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	// Convert string id to uint
	productID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}
	product.ID = uint(productID)
	updatedProduct, err := productRepo.Update(&product)
	if err != nil {
		http.Error(w, "Error updating product", http.StatusInternalServerError)
		return
	}
	response := ApiResponse{Success: true, Data: updatedProduct, Message: "Product updated successfully"}
	respondWithJSON(w, response)
}

func DeleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	// Convert string id to uint
	productID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}
	success, err := productRepo.Delete(uint(productID))
	if err != nil {
		http.Error(w, "Error deleting product", http.StatusInternalServerError)
		return
	}
	if !success {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}
	response := ApiResponse{Success: true, Message: "Product deleted successfully"}
	respondWithJSON(w, response)
}

func respondWithJSON(w http.ResponseWriter, response ApiResponse) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Setup routes
func InitializeRoutes() {
	r := mux.NewRouter()
	r.HandleFunc("/products", GetAllProducts).Methods("GET")
	r.HandleFunc("/products/{id}", GetProductById).Methods("GET")
	r.HandleFunc("/products", CreateProduct).Methods("POST")
	r.HandleFunc("/products/{id}", UpdateProduct).Methods("PUT")
	r.HandleFunc("/products/{id}", DeleteProduct).Methods("DELETE")
	http.Handle("/", r)
}

func main() {
	// Initialize DB
	InitDb()

	// Initialize repository
	productRepo = &GenericRepository{DB: db}

	// Initialize routes
	InitializeRoutes()

	// Start server
	fmt.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
