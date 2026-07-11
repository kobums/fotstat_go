package inbody

type Column int

const (
    _ Column = iota

    ColumnId
    ColumnPlayer
    ColumnTestdate
    ColumnHeight
    ColumnWeight
    ColumnMuscle
    ColumnFat
    ColumnRightleg
    ColumnLeftleg
    ColumnScore
    ColumnCreateddate
    ColumnUpdateddate
)

type Params struct {
    Column Column
    Value interface{}
}
