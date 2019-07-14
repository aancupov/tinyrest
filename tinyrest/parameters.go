package tinyrest

type Parameters struct{
    NameDb string
    NameDr string
}

// именование с маленькой буквы позволяет защитить от экспорта
var instance *Parameters

func GetInstance() *Parameters {
    if instance == nil {

        instance = &Parameters{}
    }

    return instance
}
