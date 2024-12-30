package order

import (
	"database/sql"
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

func GetAllOrder(w http.ResponseWriter, r *http.Request) {
	sqlDB, err := config.PostgresDB.DB()
	if err != nil {
		log.Println("Database connection error:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	payload, err := watoken.Decode(config.PUBLICKEY, at.GetLoginFromHeader(r))
	if err != nil {
		log.Println("Unauthorized: failed to decode token")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var ownerID int
	query := `SELECT id_user FROM akun WHERE no_telp = $1`
	err = sqlDB.QueryRow(query, payload.Id).Scan(&ownerID)
	if err != nil {
		log.Println("Error retrieving user ID:", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Query untuk mendapatkan semua produk milik user
	productIDs := []int{}
	queryProduct := `SELECT id FROM farm_products WHERE farm_id = (SELECT farm_id FROM farms WHERE owner_id = $1)`
	rows, err := sqlDB.Query(queryProduct, ownerID)
	if err != nil {
		log.Println("Error retrieving products for user:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var productID int
		if err := rows.Scan(&productID); err != nil {
			log.Println("Error scanning product ID:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		productIDs = append(productIDs, productID)
	}

	// Query untuk mendapatkan semua order terkait produk tersebut
	orders := []model.Order{}
	queryOrders := `
	SELECT o.id, o.user_id, o.product_id, o.quantity, o.total_harga, o.status, o.pengiriman_id, i.invoice_number, i.payment_status, i.payment_method, i.total_amount, i.issued_date, i.due_date
	FROM orders o
	JOIN invoice i ON o.invoice_id = i.id
	WHERE o.product_id = ANY($1)`
	rows, err = sqlDB.Query(queryOrders, productIDs)
	if err != nil {
		log.Println("Error retrieving orders:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var order model.Order
		if err := rows.Scan(&order.ID, &order.UserID, &order.ProductID, &order.Quantity, &order.TotalHarga, &order.Status, &order.PengirimanID, &order.Invoice.InvoiceNumber, &order.Invoice.PaymentStatus, &order.PaymentMethod, &order.Invoice.TotalAmount, &order.Invoice.IssuedDate, &order.Invoice.DueDate); err != nil {
			log.Println("Error scanning order:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		orders = append(orders, order)
	}

	// Response sukses
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(orders)
}

func GetOrderById(w http.ResponseWriter, r *http.Request) {
	// Mendapatkan koneksi database
	sqlDB, err := config.PostgresDB.DB()
	if err != nil {
		log.Println("Database connection error:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Validasi token
	_, err = watoken.Decode(config.PUBLICKEY, at.GetLoginFromHeader(r))
	if err != nil {
		log.Println("Unauthorized: failed to decode token")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Mendapatkan farm_id dari query parameter
	farmId := r.URL.Query().Get("farm_id")
	if farmId == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Farm ID is required",
		})
		return
	}

	// Query untuk mendapatkan data farm
	queryGet := `
		SELECT 
			f.name AS farm_name, f.farm_type, f.phonenumber_farm, f.email, f.description, 
			ST_Y(location::geometry) AS lat, ST_X(location::geometry) AS lon, f.image_farm,
			o.name AS owner_name, 
			a.street, a.city, a.province, a.postal_code, a.country
		FROM farms f
		LEFT JOIN owners o ON f.owner_id = o.id
		LEFT JOIN addresses a ON f.address_id = a.id
		WHERE f.id = $1`

	// Eksekusi query
	row := sqlDB.QueryRow(queryGet, farmId)

	// Variabel untuk menyimpan data farm
	var farm model.Farms
	var lat, lon float64
	err = row.Scan(
		&farm.Name, &farm.Farm_Type, &farm.PhonenumberFam, &farm.Email, &farm.Description,
		&lat, &lon, &farm.FamrsImageURL,
		&farm.Name, &farm.Street, &farm.City, &farm.State, &farm.PostalCode, &farm.Country,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("No farm found with the provided ID")
			http.Error(w, "Farm not found", http.StatusNotFound)
			return
		}
		log.Println("Error scanning farm data:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Set lokasi (latitude dan longitude)
	farm.Lat = lat
	farm.Lon = lon

	// Membentuk respons
	response := map[string]interface{}{
		"message": "Farm data retrieved successfully",
		"farm":    farm,
	}

	// Kirimkan respons sukses
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
