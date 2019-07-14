package tinyrest
import (
	"net/http"
	"encoding/json"
	"fmt"
)

type CreateResponse struct {
	Msg string "json:msg"
}

// вывод строки в виде json в ответ на http-запрос
func msgOutput(w http.ResponseWriter, msg string) {
	Response := CreateResponse{}
	Response.Msg = msg
	createOutput,_ := json.Marshal(Response)
	jsonOutput(w,createOutput)
}

// вывод последовательности байтов как json в ответ на http-запрос
func jsonOutput(w http.ResponseWriter, recs []byte) {
	//w.Header().Set("Pragma","no-cache") //устанавливаем значение Pragma http-заголовка в "no-cache", т.е. гарантируем получение "свежих" данных
	w.Header().Set("Access-Control-Allow-Origin","*");
	w.Header().Set("Access-Control-Allow-Methods","POST, GET, OPTIONS, DELETE");
	//w.Header().Set("Access-Control-Max-Age","3600");
	//w.Header().Set("Access-Control-Allow-Headers","Content-Type, x-requested-with");
	w.Header().Set("Content-type","application/json; charset=utf-8");
	//w.Header().Set("Link","<http://localhost:8080/api/users?start="+string(next)+"; rel=\"next\"")
	fmt.Fprintf(w, string(recs))
}
