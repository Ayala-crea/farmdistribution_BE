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
		sendErrorResponse(w, http.StatusInternalServerError, "Database connection error", err.Error())
		return
	}

	latStr := r.URL.Query().Get("lat")
	lonStr := r.URL.Query().Get("lon")
	radiusStr := r.URL.Query().Get("radius")

	latitude, err := strconv.ParseFloat(latStr, 64)
	if err != nil || latitude < -90 || latitude > 90 {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid latitude", err.Error())
		return
	}

	longitude, err := strconv.ParseFloat(lonStr, 64)
	if err != nil || longitude < -180 || longitude > 180 {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid longitude", err.Error())
		return
	}

	radiusInKm, err := strconv.ParseFloat(radiusStr, 64)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid radius", err.Error())
		return
	}

	radius := radiusInKm * 1000

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
		sendErrorResponse(w, http.StatusInternalServerError, "Internal Server Error", err.Error())
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
			sendErrorResponse(w, http.StatusInternalServerError, "Error scanning row", err.Error())
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
		sendErrorResponse(w, http.StatusNotFound, "No stores found within the given radius", "")
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

// Helper function to send error responses in a structured format
func sendErrorResponse(w http.ResponseWriter, statusCode int, message string, errDetail string) {
	response := map[string]interface{}{
		"status":  "error",
		"message": message,
		"error":   errDetail,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
