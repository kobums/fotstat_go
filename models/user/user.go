package user

type Column int

const (
    _ Column = iota
    
    ColumnId
    ColumnEmail
    ColumnPassword
    ColumnName
    ColumnCreateddate
    ColumnUpdateddate
)

type Params struct {
    Column Column
    Value interface{}
}




