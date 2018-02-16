package test

import (
	"fmt"
	"strings"
	"testing"
)

var cases = []TestCase{
	{
		Name: "1",
		Got:  `1234567890`,
		Want: `qwertyuiop`,
	},
	{
		Name: "2",
		Got: `1234567890
		1234567890`,
		Want: `qwertyuiop
		qwertyuiop`,
	},
	{
		Name: "3",
		Got: `1234567890
		1234567890
		1234567890
		1234567890
		1234567890
		1234567890
		1234567890`,
		Want: `1234567890
		1234567890
		1234567890
		1234567890
		1234567890
		qwertyuiop
		qwertyuiop
		1234567890`,
	},
}

func TestGenerationCases(t *testing.T) {
	for _, tt := range cases {
		t.Run(tt.Name, func(t *testing.T) {
			line, _ := findStringDifference(tt.Want, tt.Got)
			fmt.Println(cutWithLinesAround(tt.Want, line))
			fmt.Println(cutWithLinesAround(tt.Got, line))
		})
	}
}

func findStringDifference(str1, str2 string) (lineNum, symbolNum int) {
	splited1 := strings.Split(str1, "\n")
	splited2 := strings.Split(str2, "\n")
	for ; lineNum < len(splited1) && lineNum < len(splited2); lineNum++ {
		if splited1[lineNum] == splited2[lineNum] {
			continue
		}
		line1, line2 := splited1[lineNum], splited2[lineNum]
		for i := 0; i < len(line1) && i < len(line2); i++ {
			if line1[i] != line2[i] {
				symbolNum = i
				break
			}
		}
	}
	return
}

const around = 2

func cutWithLinesAround(str string, lineNum int) string {
	splited := strings.Split(str, "\n")
	i := 0
	var ans []string
	for lineNum-around < i && i < lineNum+around && i < len(splited) {
		if lineNum-around < i && i < lineNum+around {
			ans = append(ans, fmt.Sprintf("%d\t%s", i+1, splited[i]))
		}
		i++
	}
	return strings.Join(ans, "\n")
}
