package http

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type httpResponse struct {
	Status string      `json:"status,omitempty"`
	Data   interface{} `json:"data,omitempty"`
	Error  string      `json:"error,omitempty"`
}

func WriteHTTPResponse(w http.ResponseWriter, data interface{}, err error, httpStatus ...int) {
	var response httpResponse
	code := http.StatusOK
	if err != nil {
		response.Error = err.Error()
		code = http.StatusInternalServerError
	} else {
		response.Data = data
		response.Status = "Success"
		code = http.StatusOK
	}

	if len(httpStatus) > 0 {
		if httpStatus[0] != http.StatusInternalServerError {
			response.Data = data
		}
		code = httpStatus[0]
	}

	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Status-code", strconv.Itoa(code))

	res, err := json.Marshal(response)
	if err != nil {
		resError := httpResponse{
			Error: err.Error(),
		}
		res, _ = json.Marshal(resError)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(res)
		return
	}

	w.WriteHeader(code)
	w.Write(res)
}

func WriteHTTPAjax(w http.ResponseWriter, data interface{}, code int) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)

	js, _ := json.Marshal(data)

	w.Write(js)
}
