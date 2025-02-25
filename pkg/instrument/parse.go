// pkg/instrument/parse.go

package instrument

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// TraceRecord holds parsed information from the tracewrap log.
type TraceRecord struct {
	ID       string
	FuncName string
	Duration string
	MemDiff  string
	SysLoad  string
	SysMem   string
	Params   []string
	Returns  []string
}

var (
	reEntering  = regexp.MustCompile(`Entering (\S+) ID: (\d+)`)
	reExiting   = regexp.MustCompile(`Exiting (\S+), ID: (\d+), Duration: ([^,]+), MemDiff: (\d+) bytes`)
	reSysDebug  = regexp.MustCompile(`System CPU Load: ([\d\.]+), System Mem Usage: (\d+) bytes`)
	reParameter = regexp.MustCompile(`Parameter (\S+) = (.+)`)
	reReturning = regexp.MustCompile(`Function (\S+) returning \[(.*?)\](.*)`)
)

// ParseLogAndGenerateCallGraph parses the provided tracewrap log file and generates a callgraph.dot file.
func ParseLogAndGenerateCallGraph(logPath string) error {
	file, err := os.Open(logPath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var records []*TraceRecord
	var currentRecord *TraceRecord

	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "Entering") {
			matches := reEntering.FindStringSubmatch(line)
			if len(matches) == 3 {
				rec := &TraceRecord{
					FuncName: matches[1],
					ID:       matches[2],
				}
				records = append(records, rec)
				currentRecord = rec
			}
		} else if strings.Contains(line, "Parameter") {
			matches := reParameter.FindStringSubmatch(line)
			if len(matches) == 3 && currentRecord != nil {
				paramStr := fmt.Sprintf("%s = %s", matches[1], matches[2])
				currentRecord.Params = append(currentRecord.Params, paramStr)
			}
		} else if strings.Contains(line, "returning") {
			matches := reReturning.FindStringSubmatch(line)
			if len(matches) >= 3 && currentRecord != nil {
				retStr := fmt.Sprintf("[%s] %s", matches[2], strings.TrimSpace(matches[3]))
				currentRecord.Returns = append(currentRecord.Returns, retStr)
			}
		} else if strings.Contains(line, "Exiting") {
			matches := reExiting.FindStringSubmatch(line)
			if len(matches) == 5 {
				// Locate the record with the matching ID.
				for _, rec := range records {
					if rec.ID == matches[2] {
						rec.Duration = matches[3]
						rec.MemDiff = matches[4]
						break
					}
				}
			}
		} else if strings.Contains(line, "System CPU Load:") {
			matches := reSysDebug.FindStringSubmatch(line)
			if len(matches) == 3 && currentRecord != nil {
				currentRecord.SysLoad = matches[1]
				currentRecord.SysMem = matches[2]
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// Determine output file path (same directory as the log file).
	outPath := filepath.Join(filepath.Dir(logPath), "callgraph.dot")
	outFile, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Write DOT file header.
	fmt.Fprintln(outFile, "digraph CallGraph {")
	fmt.Fprintln(outFile, `  node [shape=box, style=filled, color="lightblue"];`)

	// Assume that the record for "main" is the parent. Do not create a node for main.
	var mainID string
	for _, rec := range records {
		if rec.FuncName == "main" {
			mainID = rec.ID
			continue
		}
		label := fmt.Sprintf("%s\\nID: %s\\nDuration: %s\\nMemDiff: %s bytes\\nSysLoad: %s, SysMem: %s bytes",
			rec.FuncName, rec.ID, rec.Duration, rec.MemDiff, rec.SysLoad, rec.SysMem)
		if len(rec.Params) > 0 {
			label += "\\nParams:"
			//for _, p := range rec.Params {
			//	label += fmt.Sprintf("\\n  %s", p)
			//}

			paramsString := concatenateAndTruncateString(rec.Params, 20)
			label += paramsString + "\\n"
		}
		if len(rec.Returns) > 0 {
			label += "\\nReturns:"
			for _, r := range rec.Returns {
				label += fmt.Sprintf("\\n  %s", r)
			}
		}
		fmt.Fprintf(outFile, "  %s [label=\"%s\"];\n", rec.ID, label)
	}

	// Write edges from main to every other function (if main exists).
	if mainID != "" {
		for _, rec := range records {
			if rec.FuncName != "main" {
				fmt.Fprintf(outFile, "  %s -> %s;\n", mainID, rec.ID)
			}
		}
	}

	fmt.Fprintln(outFile, "}")
	return nil
}

func concatenateAndTruncateString(stringSlice []string, length int) string {
	concatenatedString := strings.Join(stringSlice, "")

	if len(concatenatedString) > length {
		return concatenatedString[:length]
	}
	return concatenatedString
}
