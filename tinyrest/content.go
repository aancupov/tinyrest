package tinyrest
import (
	"database/sql"
	"fmt"
	"net/http"
	"github.com/gorilla/mux"
	"encoding/json"
	"regexp"
)

func CreateTable(resource Resource) string {
	if(resource.Name == "") { return "Incorrect name of resurce"}
	//Открываем БД
	db, err := sql.Open(GetInstance().NameDr, GetInstance().NameDb)
	if err != nil {
		return err.Error()
	}
	defer db.Close()

	_,err = db.Exec("BEGIN TRANSACTION")
	if err != nil {
		return err.Error()
	}

	//выполняем оператор SQL, который не возвращает строк
	_,err = db.Exec("INSERT INTO resources (name,details,use) values (?,?,?)",resource.Name,resource.Details,resource.Use)
	if err != nil {
		return err.Error()
	}

	//выполняем оператор SQL, который не возвращает строк
	_, err = db.Exec("CREATE TABLE " + resource.Name + " (id integer not null primary key autoincrement,content text,note text,date text)")
	if err != nil {
		_,_ = db.Exec("ROLLBACK TRANSACTION")
		return err.Error()
	}

	_,err = db.Exec("COMMIT TRANSACTION")
	if err != nil {
		return err.Error()
	}

	return ""
}

func DeleteTable(name string) string {
	//Открываем БД
	db, err := sql.Open(GetInstance().NameDr, GetInstance().NameDb)
	if err != nil {
		return err.Error()
	}
	defer db.Close()

	_,err = db.Exec("BEGIN TRANSACTION")
	if err != nil {
		return err.Error()
	}

	//выполняем оператор SQL, который не возвращает строк
	stIns, errSQL := db.Prepare("DELETE FROM resources WHERE name=?")
	if errSQL != nil {
		return errSQL.Error()
	} else {
		defer stIns.Close()
		_, errExec := stIns.Exec(name)
		if errExec != nil {
			return errExec.Error()
		}
	}

	//выполняем оператор SQL, который не возвращает строк
	_,errSQL = db.Exec("DROP TABLE IF EXISTS "+name)
	if errSQL != nil {
		_,_ = db.Exec("ROLLBACK TRANSACTION")
		return errSQL.Error()
	}

	_,err = db.Exec("COMMIT TRANSACTION")
	if err != nil {
		return err.Error()
	}

	return "Deleted"
}

//////////////////////////////
type Record struct {
	Id int64        `json:"id"`
	Content string  `json:"content"`
	Note string  	`json:"note"`
	Date string     `json:"date"`
}

type Records struct {
	Records []Record `json:"records"`
}

func CreateContent(w http.ResponseWriter, r *http.Request) {

	//выделяем имя ресурса
	urlParams := mux.Vars(r)
	resource := urlParams["resource"]

	//читаем из тела запроса элемент ресурса
	newRecord := Record{}
	decoder := json.NewDecoder(r.Body)
	errDe := decoder.Decode(&newRecord)
	if errDe !=nil {
		fmt.Println("2:"+errDe.Error())
		msgOutput(w, errDe.Error())
		return
	}

	output, errMarshal := json.Marshal(newRecord)    //готовый "объект" превращаем в json
	if errMarshal != nil {                         //если ошибка превращения делаем простое сообщение
		msgOutput(w, errMarshal.Error())
		fmt.Println("3:"+errMarshal.Error())
		return
	}

	//открываем БД
	db, err := sql.Open(GetInstance().NameDr, GetInstance().NameDb)
	if err != nil {
		fmt.Println("1:"+err.Error())
		msgOutput(w, err.Error())
		return
	}
	defer db.Close()

	//выполняем оператор SQL, который не возвращает строк
	stIns, errSQL := db.Prepare("INSERT INTO "+resource+" (content,note,date) values (?,?,?)")
	if errSQL != nil {
		msgOutput(w, errSQL.Error())
		fmt.Println("4:"+errSQL.Error())
		return
	}
	defer stIns.Close()

	_, errExec := stIns.Exec(newRecord.Content,newRecord.Note,newRecord.Date)
	if errExec != nil {
		msgOutput(w, errExec.Error())
		fmt.Println("5:"+errExec.Error())
		return
	}

	jsonOutput(w, output)
	fmt.Println(string(output))
}

//чтение и вывод элементов ресурса
func GetContent(w http.ResponseWriter, r *http.Request) {
	urlParams := mux.Vars(r)
	resource := urlParams["resource"]

	fmt.Println("get:"+resource)

	db, err := sql.Open(GetInstance().NameDr, GetInstance().NameDb)
	if err != nil {
		fmt.Println(err.Error())
		msgOutput(w, err.Error())
		return
	}
	defer db.Close()

	mindate := ""
	maxdate := ""
	note    := ""

	fmt.Println(r.FormValue)
	//с использованием рег. выражений проверяем ключи start и next
	if m, _ := regexp.MatchString("[0-9]{4}-(0[1-9]|1[012])-(0[1-9]|1[0-9]|2[0-9]|3[01])", r.FormValue("mindate")); m {
		mindate = r.FormValue("mindate")
	}
	if m, _ := regexp.MatchString("[0-9]{4}-(0[1-9]|1[012])-(0[1-9]|1[0-9]|2[0-9]|3[01])", r.FormValue("maxdate")); m {
		maxdate = r.FormValue("maxdate")
	}
	if m, _ := regexp.MatchString("[a-z]+", r.FormValue("note")); m {
		note = r.FormValue("note")
	}

	//делаем запрос к БД, возвращающие одну и более строк
	cmd := "select * from "+resource;

	cmdnote := ""
	if note!="" {
		cmdnote = " note LIKE '%"+note+"%'";
	}

	cmddate := ""
	if (mindate!="" && maxdate=="") {
		cmddate = " date>='"+mindate+"'"
	} else if (mindate=="" && maxdate!="") {
		cmddate = " date<='"+maxdate+"'"
	} else if (mindate!="" && maxdate!="") {
		cmddate = " date>='"+mindate+"' AND date<='"+maxdate+"'"
	}

	if (cmdnote != "" && cmddate!="") {
		cmd = cmd + " WHERE "+cmdnote+" AND "+cmddate;
	} else if (cmdnote == "" && cmddate!="") {
		cmd = cmd + " WHERE "+cmddate;
	} else if (cmdnote != "" && cmddate=="") {
		cmd = cmd + " WHERE "+cmdnote;
	}
	start := ""
	next := ""
	//с использованием рег. выражений проверяем ключи start и next
	if m, _ := regexp.MatchString("^[0-9]+$", r.FormValue("start")); m {  //start д.б. числом
		start = r.FormValue("start")
	}
	if m, _ := regexp.MatchString("^[0-9]+$", r.FormValue("next")); m {   //next д.б. числом
		next = r.FormValue("next")
	}

	if start=="" && next!="" {
		cmd = cmd + " LIMIT " + next
	} else if start!="" && next!="" {
		cmd = cmd + " LIMIT " + start + ","+next
	}

	cmd = cmd + " ORDER BY date"
	fmt.Println("cmd="+cmd)

	rows,err := db.Query(cmd)


	if err != nil {
		msgOutput(w, err.Error())   //в случае ошибки, выводим ее
		fmt.Println("GetContent:"+err.Error())
		return
	}
	defer rows.Close()

	//создаем результирующий объект
	response := Records{}

	//пробегаем по строкам результата
	for rows.Next() {
		record := Record{}                                              //создаем объект
		rows.Scan(&record.Id,&record.Content,&record.Note,&record.Date) //заполняем его значениями из результатов запроса
		response.Records = append(response.Records, record)             //добавляем в результирующий объект
	}

	output,_ := json.Marshal(response)                                  //превращаем результат в json
	jsonOutput(w, output)                                               //и выводим его в w
}

func GetContentById(w http.ResponseWriter, r *http.Request) {

	urlParams := mux.Vars(r)
	resource := urlParams["resource"]
	id := urlParams["id"]

	fmt.Println("get:"+resource+" id:"+id)

	db, err := sql.Open(GetInstance().NameDr, GetInstance().NameDb)
	if err != nil {
		fmt.Println(err.Error())
		msgOutput(w, err.Error())
		return
	}
	defer db.Close()

	//делаем запрос к БД, возвращающие одну и более строк
	rows,err  := db.Query("SELECT * FROM "+resource+" WHERE id="+id)

	if err != nil {
		msgOutput(w, err.Error())   //в случае ошибки, выводим ее
		fmt.Println("GetContent:"+err.Error())
		return
	}
	defer rows.Close()

	//создаем результирующий объект
	record := Record{}                                                     //создаем объект

	//получаем результат
	if rows.Next() {
		rows.Scan(&record.Id,&record.Content,&record.Note,&record.Date)    //заполняем его значениями из результатов запроса
		output,_ := json.Marshal(record)                                   //превращаем результат в json
		jsonOutput(w, output)                                              //и выводим его в w
	} else {
		msgOutput(w, "Incorrect 'id'")   //в случае ошибки, выводим ее
	}
}

func DeleteContentById(w http.ResponseWriter, r *http.Request) {

	urlParams := mux.Vars(r)
	resource := urlParams["resource"]
	id := urlParams["id"]

	fmt.Println("delete:"+resource+" id:"+id)

	db, err := sql.Open(GetInstance().NameDr, GetInstance().NameDb)
	if err != nil {
		fmt.Println(err.Error())
		msgOutput(w, err.Error())
		return
	}
	defer db.Close()

	//делаем запрос к БД, возвращающие одну и более строк
	_,err = db.Exec("DELETE FROM "+resource+" WHERE id="+id)

	if err != nil {
		msgOutput(w, err.Error())   //в случае ошибки, выводим ее
		fmt.Println("GetContent:"+err.Error())
	} else {
		msgOutput(w,"Deleted")
	}
}

func PutContent(w http.ResponseWriter, r *http.Request) {

	//определяем имя ресурса
	urlParams := mux.Vars(r)
	resource := urlParams["resource"]

	//считываем джейсон из тела запроса в запись
	record := Record{}                       //создаем "объект"
	decoder := json.NewDecoder(r.Body)
	errDe := decoder.Decode(&record)
	if errDe != nil {
		fmt.Println("6:"+errDe.Error())
		msgOutput(w,errDe.Error())
		return
	}

	//запись превращаем в джейсон
	output, errMarshal := json.Marshal(record)    //готовый "объект" превращаем в json
	if errMarshal != nil {                         //если ошибка превращения делаем простое сообщение
		msgOutput(w, errMarshal.Error())
		fmt.Println("7:"+errMarshal.Error())
		return
	}

	if(record.Id < 0) {
		msgOutput(w, "Uncorretct 'id'")
		return
	}

	//открываем базу
	db, err := sql.Open(GetInstance().NameDr, GetInstance().NameDb)
	if err != nil {
		fmt.Println(err.Error())
		msgOutput(w, err.Error())
		return
	}
	defer db.Close()

	//пишем в базу данные из записи
	stIns, errSQL := db.Prepare("UPDATE " + resource + " SET content=?, note=?, date=? WHERE id=?")
	if errSQL != nil {
		msgOutput(w, errSQL.Error())
		fmt.Println("8:"+errSQL.Error())
		return
	}
	defer stIns.Close()
	_, errExec := stIns.Exec(record.Content, record.Note , record.Date ,record.Id)
	fmt.Println("Content:"+record.Content)
	if errExec != nil {
		msgOutput(w, errExec.Error())
		fmt.Println("9:"+errExec.Error())
		return
	}
	//выводим джейсон в качестве ответа на запрос
	jsonOutput(w,output)
	fmt.Println("Изменен:"+string(output))
}
