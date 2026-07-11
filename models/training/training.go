package training

type Column int

const (
    _ Column = iota

    ColumnId
    ColumnTeam
    ColumnTrainingdate
    ColumnCreateddate
    ColumnUpdateddate
)

type Params struct {
    Column Column
    Value interface{}
}
