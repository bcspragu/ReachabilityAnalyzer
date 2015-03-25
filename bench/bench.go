package bench

import (
	"bufio"
	"os"
)

var portMap map[string]Port

type Gate interface {
	SetOut()
	Out() bool
}

type Input struct{}
type Output struct{}

type Bench struct {
	Inputs []Input
	Ouputs []Output
	Gates  []Gate
}

func NewFromFile(filename string) *Bench {
	//portMap := make(map[string]Port)

	file, err := os.Open(filename)
	if err != nil {
		//log.Fatal(err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		//fmt.Println(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		//log.Fatal(err)
	}
	return new(Bench)
}

func (b *Bench) LoadLine(line string) {

}
