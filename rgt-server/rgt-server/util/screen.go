package util

const (
	SCREEN_SIMPLE   = 0x00
	SCREEN_EXTENDED = 0x01
)

type Screen struct {
	data        []byte
	rows        int16
	columns     int16
	cursorStyle int16
	cursorRow   int16
	cursorCol   int16
	screenType  int8
}

func NewScreen(rows, columns int16, screenType int8) *Screen {
	cellSize := cellSize(screenType)
	return &Screen{rows: rows,
		columns:    columns,
		screenType: screenType,
		data:       make([]byte, int16(cellSize)*rows*columns)}
}

func cellSize(screenType int8) int16 {
	if screenType == SCREEN_SIMPLE {
		return 2
	} else {
		return 4
	}
}

func (s *Screen) GetRows() int16 {
	return s.rows
}

func (s *Screen) GetCols() int16 {
	return s.columns
}

func (s *Screen) GetCursorStyle() int16 {
	return s.cursorStyle
}

func (s *Screen) GetCursorRow() int16 {
	return s.cursorRow
}

func (s *Screen) GetCursorCol() int16 {
	return s.cursorCol
}

func (s *Screen) SetCursorStyle(style int16) {
	s.cursorStyle = style
}

func (s *Screen) SetCursorRow(row int16) {
	s.cursorRow = row
}

func (s *Screen) SetCursorCol(col int16) {
	s.cursorCol = col
}

func (s *Screen) GetData() []byte {
	return s.data
}

func (s *Screen) GetScreenType() int8 {
	return s.screenType
}

func (s *Screen) SetScreenType(newScreenType int8) {
	s.screenType = newScreenType
}

func (s *Screen) PutData(data []byte) {
	s.data = make([]byte, cellSize(s.screenType)*s.rows*s.columns)
	copy(s.data, data)
}

func (s *Screen) DataLen() int {
	return len(s.data)
}
