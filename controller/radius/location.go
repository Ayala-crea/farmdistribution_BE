package radius

import (
	"encoding/json"
	"farmdistribution_be/config"
	"log"
	"net/http"
	"strconv"
)

func GetAllTokoByRadius(w http.ResponseWriter, r *http.Request) {
	sqlDB, err := config.PostgresDB.DB()
	if err != nil {
		log.Fatal(err)
	}

	latStr := r.URL.Query().Get("lat")
	lonStr := r.URL.Query().Get("lon")
	radiusStr := r.URL.Query().Get("radius")

	latitude, err := strconv.ParseFloat(latStr, 64)
	if err != nil || latitude < -90 || latitude > 90 {
		http.Error(w, "Invalid latitude", http.StatusBadRequest)
		return
	}

	longitude, err := strconv.ParseFloat(lonStr, 64)
	if err != nil || longitude < -180 || longitude > 180 {
		http.Error(w, "Invalid longitude", http.StatusBadRequest)
		return
	}

	radius, err := strconv.ParseFloat(radiusStr, 64)
	if err != nil {
		http.Error(w, "Invalid radius", http.StatusBadRequest)
		return
	}

	// Query untuk toko berdasarkan radius
	query := `
	SELECT 
    id, 
    name AS nama_toko, 
    farm_type AS kategori, 
    phonenumber_farm, 
    email, 
    description, 
    image_farm AS gambar_toko,
    ST_AsGeoJSON(location) AS location,
    created_at
FROM farms
WHERE ST_DWithin(
    location, 
    ST_SetSRID(ST_MakePoint($1, $2), 4326), 
    $3
);
	`

	rows, err := sqlDB.Query(query, longitude, latitude, radius)
	if err != nil {
		log.Println("Error executing query:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var allMarkets []map[string]interface{}
	for rows.Next() {
		var (
			tokoID      int
			namaToko    string
			kategori    string
			phonenumber string
			email       string
			description string
			gambarToko  string
			location    string
			createdAt   string
		)

		err := rows.Scan(&tokoID, &namaToko, &kategori, &phonenumber, &email, &description, &gambarToko, &location, &createdAt)
		if err != nil {
			log.Println("Error scanning row:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		allMarkets = append(allMarkets, map[string]interface{}{
			"id":          tokoID,
			"nama_toko":   namaToko,
			"kategori":    kategori,
			"phonenumber": phonenumber,
			"email":       email,
			"description": description,
			"gambar_toko": gambarToko,
			"location":    location,
			"created_at":  createdAt,
		})
	}

	if len(allMarkets) == 0 {
		http.Error(w, "No stores found within the given radius", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"status":  "success",
		"message": "Stores found within radius",
		"data":    allMarkets,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
