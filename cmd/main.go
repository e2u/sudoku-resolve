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
	maxRow = 8
	maxCol = 8
)

type Posit struct {
	StartX uint
	StartY uint
	EndX   uint
	EndY   uint
}

func init() {
	logrus.SetLevel(logrus.WarnLevel)
	atomic.StoreUint64(&recursionCount, 0)

	flag.StringVar(&boardFile, "b", "", "input board file")

	flag.Parse()

	if len(boardFile) == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}
}

var (
	// 3x3 分組的座標,當前座標落在哪個分組計算方法
	boardPosits = []Posit{
		NewPosit(0, 0, 2, 2), NewPosit(3, 0, 5, 2), NewPosit(6, 0, 8, 2),
		NewPosit(0, 3, 2, 5), NewPosit(3, 3, 5, 5), NewPosit(6, 3, 8, 5),
		NewPosit(0, 6, 2, 8), NewPosit(3, 6, 5, 8), NewPosit(6, 6, 8, 8),
	}

	recursionCount uint64

	boardFile string

	board [][]int = [][]int{
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
	fmt.Printf("input board: \n%v\n", boardFileString)
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
	backtrack(board, 0, 0)
	fmt.Printf("RecursionCount=%v,during=%v\n", recursionCount, time.Since(st))
	PrintBoard(board)

}

func backtrack(board [][]int, row, col uint) bool {
	atomic.AddUint64(&recursionCount, 1)
	if row > maxRow {
		return true
	}

	if col > maxCol {
		return backtrack(board, row+1, 0)
	}

	currentValue := GetPointValue(board, row, col)
	logrus.Infof(">> row=%v,col=%v,value=%v", row, col, currentValue)

	if currentValue > 0 {
		return backtrack(board, row, col+1)
	}

	mayBeArray := MayNumbers(board, row, col)
	logrus.Infof("may be=%v", mayBeArray)
	for _, mayBe := range mayBeArray {
		board[row][col] = mayBe
		if backtrack(board, row, col+1) {
			return true
		}
		board[row][col] = 0
	}

	return false
}

// GetPointValue 獲取 border 中指定座標的值
// TODO 需要檢查座標值不要越界
func GetPointValue(border [][]int, row, col uint) int {
	if row > maxRow || col > maxCol {
		return 0
	}
	return border[row][col]
}

// 返回當前位置可以填的數字
func MayNumbers(board [][]int, row, col uint) []int {
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
	posit := FindPosit(row, col)
	for sx := posit.StartX; sx <= posit.EndX; sx++ {
		for sy := posit.StartY; sy <= posit.EndY; sy++ {
			exists = append(exists, board[sx][sy])
		}
	}

	var result []int
	for i := 1; i <= 9; i++ {
		if IntArrayContains(i, exists) {
			continue
		}
		result = append(result, i)
	}

	logrus.Infof("current posint row=%v,col=%v,number=%v,may be=%v", row, col, board[row][col], result)
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

// IsZero 不是正確值
func (p Posit) IsZero() bool {
	return p.EndX == 0 && p.EndY == 0
}

func NewPosit(sx, sy, ex, ey uint) Posit {
	return Posit{StartX: sx, StartY: sy, EndX: ex, EndY: ey}
}

// 傳入 x,y 找到分組起始結束座標
// 之後可以遍歷這些座標範圍查找是否有數字存在
func FindPosit(x, y uint) Posit {
	for _, bp := range boardPosits {
		if x >= bp.StartX && y >= bp.StartY && x <= bp.EndX && y <= bp.EndY {
			return bp
		}
	}
	return Posit{}
}

// FillDone 是否全部都填充完畢
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
