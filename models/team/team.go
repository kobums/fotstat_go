package team

type Column int

const (
    _ Column = iota
    
    ColumnId
    ColumnUid
    ColumnName
    ColumnCreateddate
    ColumnUpdateddate
)

type Params struct {
    Column Column
    Value interface{}
}




