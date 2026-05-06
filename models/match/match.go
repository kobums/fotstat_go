package match

type Column int

const (
    _ Column = iota
    
    ColumnId
    ColumnTeam
    ColumnAwayname
    ColumnMatchdate
    ColumnCreateddate
    ColumnUpdateddate
)

type Params struct {
    Column Column
    Value interface{}
}




