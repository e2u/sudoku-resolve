package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	maxRow       = 8
	maxCol       = 8
	modeClassic  = "c"
	modeDiagonal = "d"
)

// AreaCoordinate 區域座標
type AreaCoordinate struct {
	Start Coordinate
	End   Coordinate
}

// Coordinate
type Coordinate struct {
	X uint
	Y uint
}

func init() {
	logrus.SetLevel(logrus.WarnLevel)
	atomic.StoreUint64(&recursionCount, 0)

	flag.StringVar(&boardFile, "b", "", "input board file")
	flag.StringVar(&puzzleMode, "m", "c", "sudoku puzzleMode: c (classic) or d (diagonal)")

	flag.Parse()

	if len(boardFile) == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}
}

var (
	// 模式,
	puzzleMode = modeClassic
	// 預設 3x3 分組的座標,當前座標落在哪個分組計算方法
	coordinates = []AreaCoordinate{
		NewAreaCoordinate(Coordinate{X: 0, Y: 0}, Coordinate{X: 2, Y: 2}), NewAreaCoordinate(Coordinate{X: 3, Y: 0}, Coordinate{X: 5, Y: 2}), NewAreaCoordinate(Coordinate{X: 6, Y: 0}, Coordinate{X: 8, Y: 2}),
		NewAreaCoordinate(Coordinate{X: 0, Y: 3}, Coordinate{X: 2, Y: 5}), NewAreaCoordinate(Coordinate{X: 3, Y: 3}, Coordinate{X: 5, Y: 5}), NewAreaCoordinate(Coordinate{X: 6, Y: 3}, Coordinate{X: 8, Y: 5}),
		NewAreaCoordinate(Coordinate{X: 0, Y: 6}, Coordinate{X: 2, Y: 8}), NewAreaCoordinate(Coordinate{X: 3, Y: 6}, Coordinate{X: 5, Y: 8}), NewAreaCoordinate(Coordinate{X: 6, Y: 6}, Coordinate{X: 8, Y: 8}),
	}

	// 預設對角線座標,對角線模式裡，LT -> RB 以及 RT -> LT 兩條對角線都不能有重複數字
	// LT to RB
	diagonalLTCoordinates = []Coordinate{
		Coordinate{X: 0, Y: 0}, Coordinate{X: 1, Y: 1}, Coordinate{X: 2, Y: 2}, Coordinate{X: 3, Y: 3}, Coordinate{X: 4, Y: 4}, Coordinate{X: 5, Y: 5}, Coordinate{X: 6, Y: 6}, Coordinate{X: 7, Y: 7}, Coordinate{X: 8, Y: 8},
	}
	// RT to LT
	diagonalRTCoordinates = []Coordinate{
		Coordinate{X: 8, Y: 0}, Coordinate{X: 7, Y: 1}, Coordinate{X: 6, Y: 2}, Coordinate{X: 5, Y: 3}, Coordinate{X: 4, Y: 4}, Coordinate{X: 3, Y: 5}, Coordinate{X: 2, Y: 6}, Coordinate{X: 1, Y: 7}, Coordinate{X: 0, Y: 8},
	}

	// 遞歸計數器
	recursionCount uint64

	// 輸入文件
	boardFile string

	board = [][]int{
		//------------------------------
		{0, 0, 0 /**/, 0, 0, 0 /**/, 0, 0, 0},
		{0, 0, 0 /**/, 0, 0, 0 /**/, 0, 0, 0},
		{0, 0, 0 /**/, 0, 0, 0 /**/, 0, 0, 0},
		//------------------------------
		{0, 0, 0 /**/, 0, 0, 0 /**/, 0, 0, 0},
		{0, 0, 0 /**/, 0, 0, 0 /**/, 0, 0, 0},
		{0, 0, 0 /**/, 0, 0, 0 /**/, 0, 0, 0},
		//------------------------------
		{0, 0, 0 /**/, 0, 0, 0 /**/, 0, 0, 0},
		{0, 0, 0 /**/, 0, 0, 0 /**/, 0, 0, 0},
		{0, 0, 0 /**/, 0, 0, 0 /**/, 0, 0, 0},
		//------------------------------
	}
)

func main() {

	boardFileBytes, err := ioutil.ReadFile(boardFile)
	if err != nil {
		logrus.Errorf(err.Error())
		os.Exit(1)
	}

	boardFileString := strings.TrimSpace(string(boardFileBytes))
	var modStr string
	switch puzzleMode {
	case modeClassic:
		modStr = "classic"
	case modeDiagonal:
		modStr = "diagonal"
	}
	fmt.Printf("puzzle mode %v, input board: \n%v\n", modStr, boardFileString)
	fmt.Println("------------------------------")
	row := 0
	for _, col := range strings.Split(boardFileString, "\n") {
		cs := strings.TrimSpace(strings.ReplaceAll(col, " ", ""))
		if cs == "" {
			continue
		}
		for idx, c := range cs {
			v, err := strconv.ParseInt(string(c), 10, 64)
			if err != nil {
				logrus.Errorf(err.Error())
				continue
			}
			board[row][idx] = int(v)
		}
		row++
	}
	st := time.Now()

	// 優先處理中心點,用經典模式處理
	if board[4][4] == 0 {
		MayNumbers(modeClassic, board, 4, 4)
	}

	backtrack(puzzleMode, board, 0, 0)
	fmt.Printf("recursion=%v,during=%v\n", recursionCount, time.Since(st))
	PrintBoard(board)
}

func backtrack(mode string, board [][]int, row, col uint) bool {
	atomic.AddUint64(&recursionCount, 1)
	if row > maxRow {
		return true
	}

	if col > maxCol {
		return backtrack(mode, board, row+1, 0)
	}

	currentValue := GetPointValue(board, row, col)
	logrus.Infof(">> row=%v,col=%v,value=%v", row, col, currentValue)

	if currentValue > 0 {
		return backtrack(mode, board, row, col+1)
	}

	mayBeArray := MayNumbers(mode, board, row, col)
	logrus.Infof("may be=%v", mayBeArray)
	for _, mayBe := range mayBeArray {
		board[row][col] = mayBe
		if backtrack(mode, board, row, col+1) {
			return true
		}
		board[row][col] = 0
	}

	return false
}

// GetPointValue 獲取 border 中指定座標的值
func GetPointValue(border [][]int, row, col uint) int {
	if row > maxRow || col > maxCol {
		return 0
	}
	return border[row][col]
}

// 返回當前位置可以填的數字
func MayNumbers(mode string, board [][]int, row, col uint) []int {
	if row > maxRow || col > maxRow {
		return []int{}
	}
	// 如果當前位置的數字大於 0，則說明已經有數字佔用，直接返回空數組
	if board[row][col] > 0 {
		return []int{}
	}
	var exists []int
	// 找當前行已經存在的數字
	for _, v := range board[row] {
		if v == 0 {
			continue
		}
		exists = append(exists, v)
	}
	// 查找當前列已經存在的數字
	for i := 0; i <= maxRow; i++ {
		v := board[i][col]
		if v == 0 {
			continue
		}
		exists = append(exists, v)
	}
	// 尋找當前座標所在位置所在的九宮格中的數字
	areaCoordinate := GetAreaCoordinate(row, col)
	for sx := areaCoordinate.Start.X; sx <= areaCoordinate.End.X; sx++ {
		for sy := areaCoordinate.Start.Y; sy <= areaCoordinate.End.Y; sy++ {
			exists = append(exists, board[sx][sy])
		}
	}
	// 如果運行模式是 對角線模式,則需要檢查再檢查對角線上的值
	// 如果當前座標落在兩條對角線上，則排除對角線上已存在的數字
	// 檢查是否落在指定的對角線座標上
	inDiagonalCoordinates := func(dcs []Coordinate, x, y uint) bool {
		for _, c := range dcs {
			if c.X == x && c.Y == y {
				return true
			}
		}
		return false
	}
	// 獲取指定對角線上已經存在的值
	getDiagonalValues := func(dcs []Coordinate) {
		for _, c := range dcs {
			exists = append(exists, board[c.X][c.Y])
			logrus.Infof("diagonal puzzleMode append exists value=%v", board[c.X][c.Y])
		}
	}
	// 如果是對角線模式，而且中心點 4,4 是需要填充的話，優先處理中心點
	if mode == modeDiagonal {
		logrus.Infof("diagonal puzzleMode")
		switch {
		case inDiagonalCoordinates(diagonalLTCoordinates, row, col): // 對角線 LT to RB
			getDiagonalValues(diagonalLTCoordinates)
		case inDiagonalCoordinates(diagonalRTCoordinates, row, col): // 對角線 RT to LB
			getDiagonalValues(diagonalRTCoordinates)
		}
	}

	var result []int
	for i := 1; i <= 9; i++ {
		if IntArrayContains(i, exists) {
			continue
		}
		result = append(result, i)
	}

	logrus.Infof("current coordinate row=%v,col=%v,number=%v,may be=%v", row, col, board[row][col], result)
	return result
}

func IntArrayContains(i int, a []int) bool {
	for idx := range a {
		if i == a[idx] {
			return true
		}
	}
	return false
}

func NewAreaCoordinate(start, end Coordinate) AreaCoordinate {
	return AreaCoordinate{Start: start, End: end}
}

// IsZero 是否是正確的座標值
func (p AreaCoordinate) IsZero() bool {
	return p.End.X == 0 && p.End.Y == 0
}

// 傳入 x,y 找到分組起始結束座標
// 之後可以遍歷這些座標範圍查找是否有數字存在
func GetAreaCoordinate(x, y uint) AreaCoordinate {
	for _, bp := range coordinates {
		if x >= bp.Start.X && y >= bp.Start.Y && x <= bp.End.X && y <= bp.End.Y {
			return bp
		}
	}
	return AreaCoordinate{}
}

// FillDone 檢查是否全部都填充完畢
func FillDone(board [][]int) bool {
	for row := 0; row <= maxRow; row++ {
		for col := 0; col <= maxCol; col++ {
			if board[row][col] == 0 {
				return false
			}
		}
	}
	return true
}

// PrintBoard 打印結果
func PrintBoard(board [][]int) {
	for rc, row := range board {
		if rc > 0 && rc%3 == 0 {
			fmt.Println()
		}
		for idx, v := range row {
			c := ""
			if idx > 0 && idx%3 == 0 {
				c = "   "
			}
			fmt.Printf("%s%v", c, v)
		}
		fmt.Println()
	}
}
