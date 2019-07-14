package tinyrest

import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
    "encoding/json"
    "fmt"
    "net/http"
    "regexp"
    //"io/ioutil"
	"strconv"
	"github.com/gorilla/mux"
)


type Resource struct {
    Name string     `json:"name"`
    Details string  `json:"details"`
    Id int64        `json:"id"`
    Use string      `json:"use"`
}

type Resources struct {
    Resources []Resource `json:"resources"`
}

func (Resources) CreateResource(w http.ResponseWriter, r *http.Request) {

    newResource := Resource{}                   //создаем "объект" Resource
    decoder := json.NewDecoder(r.Body)
    errDe := decoder.Decode(&newResource)
    if errDe !=nil {
        fmt.Println("2:"+errDe.Error())
        fmt.Println(r.Body)
        msgOutput(w, errDe.Error())
    } else if(newResource.Name != "") {
        msg := CreateTable(newResource)
        if (msg == "") {
            fmt.Println(newResource)
            createOutput,_ := json.Marshal(newResource)
            jsonOutput(w, createOutput)
        } else {
            fmt.Println(msg)
            msgOutput(w, msg)
        }
    }
}

func (Resources) DeleteResource(w http.ResponseWriter, r *http.Request) {

    //читаем джейсон из запроса
    newResource := Resource{}                       //создаем "объект" User
    decoder := json.NewDecoder(r.Body)
    errDe := decoder.Decode(&newResource)
    if errDe != nil {
        fmt.Println("6:"+errDe.Error())
        msgOutput(w, errDe.Error())
        return
    }
    //если id корректен, то ищем по id имя ресурса и пытаемся выполнить двойное действие
    if(newResource.Id >= 0) {
        nameOfResource := nameById(newResource.Id)
        msg := DeleteTable(nameOfResource) //удалить из таблицы ресурсов и удалить таблицу соотв этому ресурсу
        fmt.Println(msg)
        msgOutput(w, msg)
    } else {
        msgOutput(w, "Uncorrect 'id'")
    }
}

func (Resources) PutResource(w http.ResponseWriter, r *http.Request) {

    db, err := sql.Open(GetInstance().NameDr, GetInstance().NameDb)
    if err != nil {
        fmt.Println(err.Error())
        msgOutput(w, err.Error())
        return
    }
    defer db.Close()

    newResource := Resource{}                       //создаем "объект"
    decoder := json.NewDecoder(r.Body)
    errDe := decoder.Decode(&newResource)
    if errDe != nil {
        fmt.Println("6:"+errDe.Error())
        msgOutput(w,errDe.Error())
        return
    }

    output, errMarshal := json.Marshal(newResource)    //готовый "объект" превращаем в json
    if errMarshal != nil {                         //если ошибка превращения делаем простое сообщение
        msgOutput(w, errMarshal.Error())
        fmt.Println("7:"+errMarshal.Error())
        return
    }

    if(newResource.Id < 0) {
        msgOutput(w, "Uncorretct 'id'")
        return
    }

    //выполняем оператор SQL, который не возвращает строк
    stIns, errSQL := db.Prepare("UPDATE resources SET details=?, use=? WHERE id=?")
    if errSQL != nil {
        msgOutput(w, errSQL.Error())
        fmt.Println("8:"+errSQL.Error())
        return
    }
    defer stIns.Close()
    _, errExec := stIns.Exec(newResource.Details,newResource.Use,newResource.Id)
    if errExec != nil {
        msgOutput(w, errExec.Error())
        fmt.Println("9:"+errExec.Error())
        return
    }
    jsonOutput(w,output)
    fmt.Println("Изменен:"+string(output))
}

func (Resources) GetResources(w http.ResponseWriter, r *http.Request) {
    db, err := sql.Open(GetInstance().NameDr, GetInstance().NameDb)
    if err != nil {
        fmt.Println(err.Error())
    }
    defer db.Close()

    fmt.Println(r.FormValue("start"))
    start := ""
    next := ""
    //с использованием рег. выражений проверяем ключи start и next
    if m, _ := regexp.MatchString("^[0-9]+$", r.FormValue("start")); m {  //start д.б. числом
        start = r.FormValue("start")
    }
    if m, _ := regexp.MatchString("^[0-9]+$", r.FormValue("next")); m {   //next д.б. числом
        next = r.FormValue("next")
    }

    //делаем запрос к БД, возвращающие одну и более строк
    rows,err := db.Query("select 1")
    if start=="" && next!="" {
        rows,err  = db.Query("select * from resources limit ?",next)
    } else if start!="" && next!="" {
        rows,err  = db.Query("select * from resources limit ?,?",start,next)
    } else {
        rows,err  = db.Query("select * from resources")
    }

    if err != nil {
        msgOutput(w, err.Error())   //в случае ошибки, выводим ее
    } else {
        defer rows.Close()

        //создаем результирующий объект
        response := Resources{}

        //пробегаем по строкам результата
        for rows.Next() {
            resource := Resource{}                                                          //создаем объект
            rows.Scan(&resource.Id,&resource.Name,&resource.Details,&resource.Use)                            //заполняем его значениями из результатов запроса
            response.Resources = append(response.Resources, resource)                           //добавляем в результирующий объект
        }

        output,_ := json.Marshal(response)                                          //превращаем результат в json
        //fmt.Fprintf(w, string(output))                                              //и выводим его в w
        jsonOutput(w,output)
    }
}

func (Resources) GetResourceById(w http.ResponseWriter, r *http.Request) {
    //выделяем имя ресурса
    urlParams := mux.Vars(r)
    param,_ := strconv.Atoi(urlParams["id"])
    db, err := sql.Open(GetInstance().NameDr, GetInstance().NameDb)
    if err != nil {
        fmt.Println(err.Error())
    }
    defer db.Close()

    //делаем запрос к БД, возвращающие одну и более строк
    rows,err := db.Query("select 1")
    rows,err  = db.Query("select id,name,details,use from resources where id=?",param)

    if err != nil {
        msgOutput(w, err.Error())   //в случае ошибки, выводим ее
    } else {
        defer rows.Close()

        //создаем результирующий объект
        response := Resource{}

        //пробегаем по строкам результата
        for rows.Next() {
            rows.Scan(&response.Id,&response.Name,&response.Details,&response.Use)    //заполняем его значениями из результатов запроса
        }

        output,_ := json.Marshal(response)                                          //превращаем результат в json
        //fmt.Println(string(output))                                              //и выводим его в w
        jsonOutput(w,output)
    }
}

func nameById(id int64) string {
    db, err := sql.Open(GetInstance().NameDr, GetInstance().NameDb)
    if err != nil {
        fmt.Println(err.Error())
        return ""
    }
    defer db.Close()

    //делаем запрос к БД, возвращающие одну и более строк
    rows,err  := db.Query("select * from resources where id=?",id)

    if err == nil {
        defer rows.Close()

        //пробегаем по строкам результата
        for rows.Next() {
            resource := Resource{}                                                          //создаем объект
            rows.Scan(&resource.Id,&resource.Name,&resource.Details,&resource.Use)          //заполняем его значениями из результатов запроса
            fmt.Println("id="+string(id)+" name="+resource.Name)
            return resource.Name;
        }

    }
    return ""
}

func (Resources) Init() {

    //Открываем БД
    db, err := sql.Open(GetInstance().NameDr, GetInstance().NameDb)
    if err != nil {
        fmt.Println(err.Error())
    }
    defer db.Close()

    _, errCreate := db.Exec("create table if not exists resources (id integer not null primary key autoincrement,name text not null, details text, use text)")
    if errCreate != nil {
        fmt.Println(errCreate.Error())
    }

}
