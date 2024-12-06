package routes

import (
	"farmdistribution_be/controller"
	"farmdistribution_be/controller/alamat"
	"farmdistribution_be/controller/auth"
	"farmdistribution_be/controller/image"
	"farmdistribution_be/controller/peternakan"
	"farmdistribution_be/controller/profile"
	"farmdistribution_be/controller/role"
	"net/http"

	"github.com/gorilla/mux"
)

func InitializeRoutes() *mux.Router {
	router := mux.NewRouter()

	// Middleware CORS global dari config
	// router.Use(config.CORSMiddleware)

	// Root route
	router.HandleFunc("/", controller.GetHome).Methods("GET", "OPTIONS")

	// Auth
	router.HandleFunc("/regis", handleCORS(auth.RegisterUser)).Methods("POST", "OPTIONS")
	router.HandleFunc("/login", handleCORS(auth.LoginUsers)).Methods("POST", "OPTIONS")

	// Profile
	router.HandleFunc("/profile", handleCORS(profile.GetProfile)).Methods("GET", "OPTIONS")
	router.HandleFunc("/profile/by-id", handleCORS(profile.GetProfileByID)).Methods("GET", "OPTIONS")
	router.HandleFunc("/profile/update", handleCORS(profile.UpdateProfile)).Methods("PUT", "OPTIONS")
	router.HandleFunc("/profile/delete", handleCORS(profile.DeleteProfile)).Methods("DELETE", "OPTIONS")
	router.HandleFunc("/profile/all", handleCORS(profile.GetAllProfiles)).Methods("GET", "OPTIONS")
	router.HandleFunc("/profile/add-image", handleCORS(image.AddImage)).Methods("PUT", "OPTIONS")
	router.HandleFunc("/profile/delete-image", handleCORS(image.DeleteImage)).Methods("DELETE", "OPTIONS")

	// Address
	router.HandleFunc("/add/address", handleCORS(alamat.CreateAddress)).Methods("POST", "OPTIONS")
	router.HandleFunc("/address", handleCORS(alamat.GetAddress)).Methods("GET", "OPTIONS")
	router.HandleFunc("/address/update", handleCORS(alamat.UpdateAddress)).Methods("PUT", "OPTIONS")

	// Role Management
	router.HandleFunc("/create/role-menu", handleCORS(role.CreateMenu)).Methods("POST", "OPTIONS")
	router.HandleFunc("/role-menu", handleCORS(role.GetAllMenus)).Methods("GET", "OPTIONS")
	router.HandleFunc("/role-menu", handleCORS(role.GetMenuByID)).Methods("GET", "OPTIONS")
	router.HandleFunc("/update/role-menu", handleCORS(role.UpdateMenu)).Methods("PUT", "OPTIONS")
	router.HandleFunc("/delete/role-menu", handleCORS(role.DeleteMenu)).Methods("DELETE", "OPTIONS")

	router.HandleFunc("/create/role", handleCORS(role.CreateRole)).Methods("POST", "OPTIONS")
	router.HandleFunc("/role", handleCORS(role.GetAllRoles)).Methods("GET", "OPTIONS")
	router.HandleFunc("/role-id", handleCORS(role.GetRoleByID)).Methods("GET", "OPTIONS")
	router.HandleFunc("/update/role", handleCORS(role.UpdateRole)).Methods("PUT", "OPTIONS")
	router.HandleFunc("/delete/role", handleCORS(role.DeleteRole)).Methods("DELETE", "OPTIONS")

	router.HandleFunc("/create/role/menu", handleCORS(role.CreateRoleMenu)).Methods("POST", "OPTIONS")
	router.HandleFunc("/role/menu", handleCORS(role.GetAllRoleMenus)).Methods("GET", "OPTIONS")
	router.HandleFunc("/role/menu-id", handleCORS(role.GetRoleMenuByID)).Methods("GET", "OPTIONS")
	router.HandleFunc("/update/role/menu", handleCORS(role.UpdateRoleMenu)).Methods("PUT", "OPTIONS")
	router.HandleFunc("/delete/role/menu", handleCORS(role.DeleteRoleMenu)).Methods("DELETE", "OPTIONS")

	// Peternakan
	router.HandleFunc("/peternakan", handleCORS(peternakan.CreatePeternakan)).Methods("POST", "OPTIONS")
	router.HandleFunc("/peternakan/get", handleCORS(peternakan.GetPeternakan)).Methods("GET", "OPTIONS")
	router.HandleFunc("/peternakan/update", handleCORS(peternakan.UpdatePeternakan)).Methods("PUT", "OPTIONS")
	router.HandleFunc("/peternakan/delete", handleCORS(peternakan.DeletePeternakan)).Methods("DELETE", "OPTIONS")

	// Status Product
	router.HandleFunc("/status-product", handleCORS(peternakan.CreateStatusProduct)).Methods("POST", "OPTIONS")
	router.HandleFunc("/status-product/get", handleCORS(peternakan.GetAllStatusProducts)).Methods("GET", "OPTIONS")
	router.HandleFunc("/status-product/get-by-id", handleCORS(peternakan.GetStatusProductByID)).Methods("GET", "OPTIONS")
	router.HandleFunc("/status-product/update", handleCORS(peternakan.UpdateStatusProduct)).Methods("PUT", "OPTIONS")
	router.HandleFunc("/status-product/delete", handleCORS(peternakan.DeleteStatusProduct)).Methods("DELETE", "OPTIONS")

	// Product
	router.HandleFunc("/add/product", handleCORS(peternakan.CreateProduct)).Methods("POST", "OPTIONS")
	router.HandleFunc("/product", handleCORS(peternakan.GetAllProduct)).Methods("GET", "OPTIONS")

	return router
}

// handleCORS adalah wrapper untuk menangani preflight request
func handleCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Tangani preflight request
		if r.Method == http.MethodOptions {
			if origin := r.Header.Get("Origin"); origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Tambahkan header CORS untuk semua request
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Lanjutkan ke handler berikutnya
		next(w, r)
	}
}