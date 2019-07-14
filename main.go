package main

import (
	"net/http"
	tcms "./tinyrest"
	"github.com/gorilla/mux"
)

func main() {

	tcms.GetInstance().NameDb = "/users/andrew/db/tinyrest.db"
	tcms.GetInstance().NameDr = "sqlite3"

	resources := tcms.Resources{} //tcms - это имя package Tags - тип определенный в package tags
	resources.Init()

	//используем фреймворк Gorilla
	gorillaRoute := mux.NewRouter()

	// А Д М И Н И С Т Р И Р О В А Н И Е
	//Создание ресурса
	gorillaRoute.HandleFunc("/admin/resources", resources.CreateResource).Methods("POST")
	//Изменение ресурса
	gorillaRoute.HandleFunc("/admin/resources/put", resources.PutResource).Methods("POST")
	//Удаление ресурса
	gorillaRoute.HandleFunc("/admin/resources/delete", resources.DeleteResource).Methods("POST")
	//Запрос ресурса по id
	gorillaRoute.HandleFunc("/admin/resources/{id:[0-9]+}", resources.GetResourceById).Methods("GET")
	//Запрос списка ресурсов
	gorillaRoute.HandleFunc("/admin/resources", resources.GetResources).Methods("GET")

	// Р А Б О Т А   С   С О Д Е Р Ж А Н И Е М
	//Запрос строк(и) ресурса
	gorillaRoute.HandleFunc("/api/{resource:[a-z]+}/{id:[0-9]+}", tcms.GetContentById).Methods("GET") //строка по id
	gorillaRoute.HandleFunc("/api/{resource:[a-z]+}", tcms.GetContent).Methods("GET") //все строки
	//Удаление строки ресурса
	gorillaRoute.HandleFunc("/api/{resource:[a-z]+}/{id:[0-9]+}/delete", tcms.DeleteContentById).Methods("POST")
	gorillaRoute.HandleFunc("/api/{resource:[a-z]+}/{id:[0-9]+}", tcms.DeleteContentById).Methods("DELETE")
	//Изменение строки ресурса
	gorillaRoute.HandleFunc("/api/{resource:[a-z]+}/put", tcms.PutContent).Methods("POST")
	gorillaRoute.HandleFunc("/api/{resource:[a-z]+}", tcms.PutContent).Methods("PUT")
	//Создание строки ресурса
	gorillaRoute.HandleFunc("/api/{resource:[a-z]+}", tcms.CreateContent).Methods("POST")

	http.Handle("/", gorillaRoute) //подключаем обработку URL через mux Gorilla

	http.ListenAndServe(":1967", nil) //запускаем сервер на 1967-порту
}
