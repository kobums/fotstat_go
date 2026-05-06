package team

type Column int

const (
    _ Column = iota
    
    ColumnId
    ColumnUser
    ColumnName
    ColumnCreateddate
    ColumnUpdateddate
)

type Params struct {
    Column Column
    Value interface{}
}




