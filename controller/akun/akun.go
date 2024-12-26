package akun

import (
	"encoding/json"
	"farmdistribution_be/config"
	"farmdistribution_be/model"
	"log"
	"net/http"
)

func GetAllAkun(w http.ResponseWriter, r *http.Request) {
	sqlDB, err := config.PostgresDB.DB()
	if err != nil {
		log.Fatal(err)
	}

	var akun []model.Akun
	w.Header().Set("Content-Type", "application/json")

	query := `SELECT id_user, nama, no_telp, email, id_role, password FROM akun`
	rows, err := sqlDB.Query(query)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "Database error",
			"message": "Failed to fetch users.",
		})
		return
	}

	defer rows.Close()

	for rows.Next() {
		var a model.Akun
		err := rows.Scan(&a.ID, &a.Nama, &a.NoTelp, &a.Email, &a.RoleID, &a.Password)
		if err != nil {
			log.Fatal(err)
		}
		akun = append(akun, a)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Users fetched successfully",
		"users":   akun,
	})
}

func EditDataAkun(w http.ResponseWriter, r *http.Request) {
	sqlDB, err := config.PostgresDB.DB()
	if err != nil {
		log.Fatal(err)
	}

	var akun model.Akun
	w.Header().Set("Content-Type", "application/json")

	id := r.URL.Query().Get("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "Missing ID",
			"message": "Please provide a valid user ID.",
		})
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&akun); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Invalid request payload",
			"message": "The JSON request body could not be decoded. Please check the structure of your request.",
		})
		return
	}

	if akun.Nama == "" || akun.NoTelp == "" || akun.Email == "" || akun.RoleID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "Missing required fields",
			"message": "Please provide valid data for the user.",
		})
		return
	}

	query := `UPDATE akun SET nama = $1, no_telp = $2, email = $3, id_role = $4 WHERE id_user = $5`
	_, err = sqlDB.Exec(query, akun.Nama, akun.NoTelp, akun.Email, akun.RoleID, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "Database error",
			"message": "Failed to update user.",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User updated successfully",
	})
}

func GetById(w http.ResponseWriter, r *http.Request) {
	sqlDB, err := config.PostgresDB.DB()
	if err != nil {
		log.Fatal(err)
	}

	var akun model.Akun
	w.Header().Set("Content-Type", "application/json")

	// Ambil ID dari query parameter
	id := r.URL.Query().Get("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "Missing ID",
			"message": "Please provide a valid user ID.",
		})
		return
	}

	// Query untuk mengambil data berdasarkan ID
	query := `SELECT id_user, nama, no_telp, email, id_role, password FROM akun WHERE id_user = $1`
	err = sqlDB.QueryRow(query, id).Scan(&akun.ID, &akun.Nama, &akun.NoTelp, &akun.Email, &akun.RoleID, &akun.Password)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "User not found",
				"message": "No user found with the provided ID.",
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "Database error",
			"message": "Failed to fetch user.",
		})
		return
	}

	// Kirim data akun sebagai response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User fetched successfully",
		"user":    akun,
	})
}
