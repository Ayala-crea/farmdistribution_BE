package radius

import (
	"database/sql"
	"encoding/json"
	"farmdistribution_be/config"
	"farmdistribution_be/helper/at"
	"farmdistribution_be/helper/watoken"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func GetAllTokoByRadius(w http.ResponseWriter, r *http.Request) {
	sqlDB, err := config.PostgresDB.DB()
	if err != nil {
		log.Fatal(err)
	}

	_, err = watoken.Decode(config.PUBLICKEY, at.GetLoginFromHeader(r))
	if err != nil {
		log.Println("Unauthorized: failed to decode token")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
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

	// Query to fetch nearby farms based on radius
	query := `
	SELECT id, nama_toko, slug, category_name, latitude, longitude, description, rating, gambar_toko,
		alamat_street, alamat_province, alamat_city, alamat_description, alamat_postal_code,
		opening_hours_opening, opening_hours_close, json_agg(json_build_object('name', u.name, 'phonenumber', u.phone_number, 'email', u.email)) AS users
	FROM toko t
	JOIN category c ON t.category_id = c.id
	LEFT JOIN users u ON t.id = u.toko_id
	WHERE earth_box(ll_to_earth($1, $2), $3) @> ll_to_earth(latitude, longitude)
	GROUP BY t.id, c.category_name
	`

	rows, err := sqlDB.Query(query, latitude, longitude, radius*1000) // Convert radius to meters
	if err != nil {
		log.Println("Error executing query:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var allMarkets []map[string]interface{}
	counter := 1
	for rows.Next() {
		var (
			tokoID         int
			namaToko       string
			slug           string
			categoryName   string
			tokoLat        float64
			tokoLon        float64
			description    string
			rating         float64
			gambarToko     string
			alamatStreet   string
			alamatProvince string
			alamatCity     string
			alamatDesc     string
			postalCode     string
			opening        string
			closing        string
			users          sql.NullString
		)

		err := rows.Scan(&tokoID, &namaToko, &slug, &categoryName, &tokoLat, &tokoLon, &description, &rating, &gambarToko,
			&alamatStreet, &alamatProvince, &alamatCity, &alamatDesc, &postalCode, &opening, &closing, &users)
		if err != nil {
			log.Println("Error scanning row:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		var filteredUsers []map[string]interface{}
		if users.Valid {
			json.Unmarshal([]byte(users.String), &filteredUsers)
		}

		location := map[string]interface{}{
			"type":       "Feature",
			"properties": map[string]interface{}{},
			"geometry": map[string]interface{}{
				"type": "Point",
				"coordinates": map[string]float64{
					"latitude":  tokoLat,
					"longitude": tokoLon,
				},
			},
		}

		allMarkets = append(allMarkets, map[string]interface{}{
			"penanda_toko":  fmt.Sprintf("Toko %d", counter),
			"id":            tokoID,
			"nama_toko":     namaToko,
			"slug":          slug,
			"category":      categoryName,
			"location":      location,
			"description":   description,
			"rating":        rating,
			"opening_hours": map[string]string{"opening": opening, "close": closing},
			"gambar_toko":   gambarToko,
			"alamat": map[string]interface{}{
				"street":      alamatStreet,
				"province":    alamatProvince,
				"city":        alamatCity,
				"description": alamatDesc,
				"postal_code": postalCode,
			},
			"user": filteredUsers,
		})

		counter++
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
