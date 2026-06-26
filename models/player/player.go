package player

type Column int

const (
    _ Column = iota
    
    ColumnId
    ColumnTeam
    ColumnName
    ColumnNumber
    ColumnBirthdate
    ColumnPosition
    ColumnCreateddate
    ColumnUpdateddate
)

type Params struct {
    Column Column
    Value interface{}
}




