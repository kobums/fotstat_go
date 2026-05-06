package quarter

type Column int

const (
    _ Column = iota
    
    ColumnId
    ColumnMatch
    ColumnNumber
    ColumnCreateddate
    ColumnUpdateddate
)

type Params struct {
    Column Column
    Value interface{}
}




