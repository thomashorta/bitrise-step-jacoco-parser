package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

// Bitrise XML structs
type Report struct {
	XMLName  xml.Name  `xml:"report"`
	Name     string    `xml:"name,attr"`
	Counters []Counter `xml:"counter"`
}

type Counter struct {
	XMLName xml.Name `xml:"counter"`
	Type    string   `xml:"type,attr"`
	Missed  int      `xml:"missed,attr"`
	Covered int      `xml:"covered,attr"`
}

// JaCoCo counter types
const Instruction = "INSTRUCTION"
const Branch = "BRANCH"
const Line = "LINE"
const Complexity = "COMPLEXITY"
const Method = "METHOD"
const Class = "CLASS"

// Step outputs
const OutputInstructionCoverage = "JACOCO_INSTRUCTION_COVERAGE"
const OutputBranchCoverage = "JACOCO_BRANCH_COVERAGE"
const OutputLineCoverage = "JACOCO_LINE_COVERAGE"
const OutputComplexityCoverage = "JACOCO_COMPLEXITY_COVERAGE"
const OutputMethodCoverage = "JACOCO_METHOD_COVERAGE"
const OutputClassCoverage = "JACOCO_CLASS_COVERAGE"

var OutputKeys = [...]string{
	OutputInstructionCoverage,
	OutputBranchCoverage,
	OutputLineCoverage,
	OutputComplexityCoverage,
	OutputMethodCoverage,
	OutputClassCoverage,
}

func main() {
	reportPath := os.Getenv("jacoco_report_path")
	reportFormat := reportPath[strings.LastIndex(reportPath, ".")+1:]

	if reportFormat != "xml" {
		fmt.Println("Only xml is supported at the moment.")
		os.Exit(1)
	}

	// Open xml file
	xmlFile, err := os.Open(reportPath)

	if err != nil {
		fmt.Printf("Failed to open specified file %s: %s\n", reportPath, err)
		os.Exit(1)
	}

	fmt.Printf("Opened %s\n", reportPath)

	defer xmlFile.Close()

	fmt.Println("Parsing the report file")

	byteValue, _ := ioutil.ReadAll(xmlFile)

	var report Report

	xml.Unmarshal(byteValue, &report)

	coverageMap := map[string]string{}
	for _, counter := range report.Counters {
		coverage := CalcCoverage(counter.Missed, counter.Covered)
		switch counter.Type {
		case Instruction:
			coverageMap[OutputInstructionCoverage] = coverage
		case Branch:
			coverageMap[OutputBranchCoverage] = coverage
		case Line:
			coverageMap[OutputLineCoverage] = coverage
		case Complexity:
			coverageMap[OutputComplexityCoverage] = coverage
		case Method:
			coverageMap[OutputMethodCoverage] = coverage
		case Class:
			coverageMap[OutputClassCoverage] = coverage
		}
	}

	// output all coverages or N/A for the ones not available
	for _, key := range OutputKeys {
		val, ok := coverageMap[key]
		if !ok {
			val = "N/A"
		}
		SetOutput(key, val)
	}

	fmt.Println("JaCoCo report file parsed and coverage exported")
	os.Exit(0)
}

// returns the formatted coverage string, e.g.: 4.50%
func CalcCoverage(missed int, covered int) string {
	total := missed + covered
	coverage := 100.0 * float64(covered) / float64(total)
	return fmt.Sprintf("%.2f%%", coverage)
}

// outputs the desired text to the environment variable
func SetOutput(key string, value string) {
	fmt.Printf("Outputing key %s with value %s\n", key, value)
	cmdLog, err := exec.Command("bitrise", "envman", "add", "--key", key, "--value", value).CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to expose output with envman, error: %s | output: %s\n", err, cmdLog)
		os.Exit(1)
	}
}
