package injury

type Column int

const (
    _ Column = iota

    ColumnId
    ColumnPlayer
    ColumnType
    ColumnStartdate
    ColumnReturndate
    ColumnMemo
    ColumnCreateddate
    ColumnUpdateddate
)

type Params struct {
    Column Column
    Value interface{}
}
