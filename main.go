package main

import (
	"flag"
	"fmt"
	engine "github.com/mrfleap/custom-compression/engine"
	"io/ioutil"
	"os"
	"strings"
	// "github.com/pkg/profile" // Profiling package
)

// https://github.com/spf13/cobra#getting-started

// Commands represents all possible commands that can be used durinv CLI invocation
var Commands = [...]string{"compress", "decompress", "benchmark", "help"}

func main() {
	// Profiling statement here V
	// defer profile.Start().Stop()
	// ^
	compressCmd := flag.NewFlagSet("compress", flag.ExitOnError)

	decompressCmd := flag.NewFlagSet("decompress", flag.ExitOnError)

	benchmarkCmd := flag.NewFlagSet("benchmark", flag.ExitOnError)

	generateHTML := benchmarkCmd.Bool("generate", false, "Compile benchmark results as an html file")

	flag.Parse()
	command := flag.Arg(0)
	if command == "" {
		errorWithMsg(fmt.Sprintf(
			"Please provide a valid command, possible commands include: \n\t %s\n", strings.Join(Commands[:], ", ")))
	}

	// Non compression commands
	switch command {
	case "help":
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Valid commands include: \n\t %s\n", strings.Join(Commands[:], ", "))
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		return
	}

	// Get flag argument that is not a flag "-algorithm..."
	file := flag.Arg(1)
	for i := 2; len(file) > 0 && file[0] == '-'; i++ {
		file = flag.Arg(i)
	}

	if file == "" && !strings.Contains(file, ",") {
		errorWithMsg("Please provide a file to be compressed/decompressed\n")
	} else if strings.Contains(file, ",") {
		for _, filename := range strings.Split(file, ",") {
			if _, err := os.Stat(filename); os.IsNotExist(err) {
				errorWithMsg(fmt.Sprintf("Could not open file (likely does not exist): %s\n", filename))
			}
		}
	} else if _, err := os.Stat(file); os.IsNotExist(err) && file != "help" {
		errorWithMsg(fmt.Sprintf("Could not open file (likely does not exist): %s\n", file))
	}

	switch command {
	case "compress", "c":
		algorithm := compressCmd.String("algorithm", "default",
			fmt.Sprintf("Which algorithm to use, choices include: \n\t%s", strings.Join(engine.Engines[:], ", ")))

		posAfterCommand := getPosAfterCommand("compress", os.Args)
		compressCmd.Parse(os.Args[posAfterCommand:])

		if *algorithm == "default" {
			lzss := "lzss"
			algorithm = &lzss
		}
		engine.CompressFile(*algorithm, file)
	case "decompress", "d":
		algorithm := decompressCmd.String("algorithm", "default",
			fmt.Sprintf("Which algorithm to use, choices include: \n\t%s", strings.Join(engine.Engines[:], ", ")))

		posAfterCommand := getPosAfterCommand("decompress", os.Args)
		decompressCmd.Parse(os.Args[posAfterCommand:])

		if *algorithm == "default" {
			lzss := "lzss"
			algorithm = &lzss
		}
		engine.DecompressFile(*algorithm, file)
	case "benchmark":
		algorithm := benchmarkCmd.String("algorithm", "default",
			fmt.Sprintf("Which algorithm to use, choices include: \n\t%s", strings.Join(engine.Engines[:], ", ")))

		posAfterCommand := getPosAfterCommand("benchmark", os.Args)
		benchmarkCmd.Parse(os.Args[posAfterCommand:])

		if file == "help" {
			fmt.Fprintf(os.Stderr, "Flags:\n")
			flag.PrintDefaults()
			return
		}

		algorithms := parseAlgorithms(*algorithm)

		if *algorithm == "default" {
			suite := "suite"
			algorithm = &suite
		}

		files := strings.Split(file, ",")
		for i := range files {
			files[i] = strings.TrimSpace(files[i])
		}

		output := engine.BenchmarkSuite(files, algorithms, *generateHTML)
		if *generateHTML {
			err := ioutil.WriteFile("index.html", []byte(output), 0644)
			check(err)
			fmt.Println("Wrote table to index.html")
		}
	default:
		errorWithMsg(fmt.Sprintf(
			"'%s' is not a valid command, "+
				"please provide a valid command, "+
				"possible commands include: \n\t %s\n", command, strings.Join(Commands[:], ", ")))
	}
}

func parseAlgorithms(algorithmString string) (algorithms [][]string) {
	var buffer []byte
	var inLayer bool
	var layer []string
	for _, char := range []byte(algorithmString) {
		if char == ',' {
			if inLayer && len(buffer) > 0 {
				layer = append(layer, string(buffer))
			} else if len(buffer) > 0 {
				algorithms = append(algorithms, []string{string(buffer)})
			}
			buffer = make([]byte, 0)
		} else if char == '[' {
			inLayer = true
		} else if char == ']' {
			layer = append(layer, string(buffer))
			buffer = make([]byte, 0)
			inLayer = false
			algorithms = append(algorithms, layer)
			layer = make([]string, 0)
		} else {
			buffer = append(buffer, char)
		}
	}
	if len(buffer) > 0 {
		algorithms = append(algorithms, []string{string(buffer)})
	}
	return algorithms
}

func getPosAfterCommand(command string, args []string) int {
	for i, s := range args {
		if s == command {
			return i + 1
		}
	}
	return -1
}

func errorWithMsg(msg string) {
	fmt.Print(msg)
	os.Exit(1)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
