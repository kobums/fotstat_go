package record

type Column int

const (
    _ Column = iota
    
    ColumnId
    ColumnQuarter
    ColumnPlayer
    ColumnMin
    ColumnGoal
    ColumnAssist
    ColumnYellowcard
    ColumnRedcard
    ColumnCreateddate
    ColumnUpdateddate
)

type Params struct {
    Column Column
    Value interface{}
}




