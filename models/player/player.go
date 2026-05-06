package player

type Column int

const (
    _ Column = iota
    
    ColumnId
    ColumnTeam
    ColumnName
    ColumnNumber
    ColumnCreateddate
    ColumnUpdateddate
)

type Params struct {
    Column Column
    Value interface{}
}




