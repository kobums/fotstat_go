package attendance

type Column int

const (
    _ Column = iota

    ColumnId
    ColumnTraining
    ColumnPlayer
    ColumnMin
    ColumnCreateddate
    ColumnUpdateddate
)

type Params struct {
    Column Column
    Value interface{}
}
