package main

import (
	"bufio"
	"fmt"
	"hack/assembler"
	"log"
	"os"
)

// read xxx.asm and output xxx.hack
func main() {
	if len(os.Args) < 2 {
		log.Fatal("Please specify the asm file")
	}
	inputFilePath := os.Args[1]
	lines, err := ReadInputFile(inputFilePath)
	if err != nil {
		log.Fatal(err)
	}

	commands, err := assembler.Parse(lines)
	if err != nil {
		log.Fatal(err)
	}

	instructions, err := assembler.Translate(commands)

	if err != nil {
		log.Fatal(err)
	}
	code, err := assembler.OutputBinaryCode(instructions)
	if err != nil {
		log.Fatal(err)
	}
	for _, c := range code {
		fmt.Println(c)
	}

}

func ReadInputFile(inputFilePath string) ([]string, error) {
	inputFile, err := os.Open(inputFilePath)
	lines := make([]string, 0)
	if err != nil {
		return lines, err
	}
	scanner := bufio.NewScanner(bufio.NewReader(inputFile))
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
	}
	return lines, nil
}

// parser
