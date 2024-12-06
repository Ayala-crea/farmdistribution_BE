package controller

import (
	"app_farm_be/helper/at"
	"app_farm_be/model"
	"net/http"
)

func GetHome(respw http.ResponseWriter, req *http.Request) {
	var resp model.Response
	resp.Response = at.GetIPaddress()
	at.WriteJSON(respw, http.StatusOK, resp)
}
