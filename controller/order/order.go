package order

import (
	"encoding/json"
	"farmdistribution_be/config"
	"farmdistribution_be/helper/at"
	"farmdistribution_be/helper/watoken"
	"farmdistribution_be/model"
	"fmt"
	"log"
	"net/http"
	"time"
)

func CreateOrder(w http.ResponseWriter, r *http.Request) {
	// Mendapatkan koneksi database
	sqlDB, err := config.PostgresDB.DB()
	if err != nil {
		log.Println("Database connection error:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Decode payload dari token
	payload, err := watoken.Decode(config.PUBLICKEY, at.GetLoginFromHeader(r))
	if err != nil {
		log.Println("Unauthorized: failed to decode token")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Mendapatkan ownerID dari nomor telepon
	var ownerID int
	query := `SELECT id_user FROM akun WHERE no_telp = $1`
	err = sqlDB.QueryRow(query, payload.Id).Scan(&ownerID)
	if err != nil {
		log.Println("Error retrieving user ID:", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Decode request body untuk mendapatkan data order
	var Orders model.Order
	if err := json.NewDecoder(r.Body).Decode(&Orders); err != nil {
		log.Println("Error decoding request body:", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var userID int
	if Orders.UserID != 0 {
		userID = Orders.UserID
	} else {
		userID = ownerID
	}

	// Validasi input
	if Orders.Products[0].ProductID == 0 || Orders.Products[0].Quantity == 0 || Orders.PengirimanID == 0 {
		http.Error(w, "Product ID, Quantity, and Pengiriman ID are required", http.StatusBadRequest)
		return
	}

	// Mendapatkan nama produk dan harga produk
	var productPrice float64
	var namaProduct string
	queryProduct := `SELECT name, price_per_kg FROM farm_products WHERE id = $1`
	err = sqlDB.QueryRow(queryProduct, Orders.Products[0].ProductID).Scan(&namaProduct, &productPrice)
	if err != nil {
		log.Println("Error retrieving product details:", err)
		http.Error(w, "Product not found", http.StatusBadRequest)
		return
	}

	// Mendapatkan biaya pengiriman
	var shippingCost float64
	queryShipping := `SELECT biaya_pengiriman FROM pengiriman WHERE id = $1`
	err = sqlDB.QueryRow(queryShipping, Orders.PengirimanID).Scan(&shippingCost)
	if err != nil {
		log.Println("Error retrieving shipping cost:", err)
		http.Error(w, "Shipping method not found", http.StatusBadRequest)
		return
	}

	// Hitung total harga
	Orders.TotalHarga = (productPrice * float64(Orders.Products[0].Quantity)) + shippingCost

	// Gunakan transaksi untuk memastikan konsistensi
	tx, err := sqlDB.Begin()
	if err != nil {
		log.Println("Failed to start transaction:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Generate invoice number
	invoiceNumber := fmt.Sprintf("INV-%d-%d", userID, time.Now().Unix())
	log.Println("payment method:", Orders.Invoice.PaymentMethod)
	var invoiceId int
	// Insert invoice ke dalam database
	insertInvoiceQuery := `
			INSERT INTO invoice (user_id, invoice_number, payment_status, payment_method, issued_date, due_date, total_amount, created_at, updated_at)
			VALUES ($1, $2, $3, $4, NOW(), NOW() + INTERVAL '7 days', $5, NOW(), NOW()) RETURNING id`
	err = tx.QueryRow(insertInvoiceQuery, userID, invoiceNumber, "Pending", Orders.PaymentMethod, Orders.TotalHarga).Scan(&invoiceId)
	if err != nil {
		log.Println("Error inserting invoice:", err)
		tx.Rollback()
		http.Error(w, "Failed to create invoice", http.StatusInternalServerError)
		return
	}

	// Insert order ke dalam database
	// Proses setiap produk
	var totalHarga float64
	for _, product := range Orders.Products {
		var productPrice float64
		var namaProduct string
		queryProduct := `SELECT name, price_per_kg FROM farm_products WHERE id = $1`
		err = sqlDB.QueryRow(queryProduct, product.ProductID).Scan(&namaProduct, &productPrice)
		if err != nil {
			log.Println("Error retrieving product details:", err)
			tx.Rollback()
			http.Error(w, "Product not found", http.StatusBadRequest)
			return
		}

		// Hitung harga produk
		productTotal := productPrice * float64(product.Quantity)
		totalHarga += productTotal

		// Insert order
		insertOrderQuery := `
		INSERT INTO orders (user_id, product_id, quantity, total_harga, status, pengiriman_id, invoice_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())`
		_, err = tx.Exec(insertOrderQuery, userID, product.ProductID, product.Quantity, productTotal, "Pending", Orders.PengirimanID, invoiceId)
		if err != nil {
			log.Println("Error inserting order:", err)
			tx.Rollback()
			http.Error(w, "Failed to create order", http.StatusInternalServerError)
			return
		}
	}

	// Update total harga pada invoice
	updateInvoiceQuery := `UPDATE invoice SET total_amount = $1 WHERE id = $2`
	_, err = tx.Exec(updateInvoiceQuery, totalHarga, invoiceId)
	if err != nil {
		log.Println("Error updating invoice:", err)
		tx.Rollback()
		http.Error(w, "Failed to update invoice", http.StatusInternalServerError)
		return
	}

	// Commit transaksi
	if err := tx.Commit(); err != nil {
		log.Println("Error committing transaction:", err)
		http.Error(w, "Failed to create order and invoice", http.StatusInternalServerError)
		return
	}

	// Response sukses
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":        "Order and Invoice created successfully",
		"invoice_number": invoiceNumber,
		"product_name":   namaProduct,
		"total_harga":    Orders.TotalHarga,
	})
}
