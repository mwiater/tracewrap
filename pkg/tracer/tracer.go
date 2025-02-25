package tracer

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/k0kubun/pp"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
)

// TraceRecord holds detailed trace information for a function call.
// Fields:
//
//	UniqueID: Unique identifier for the trace record.
//	FunctionName: Name of the function being traced.
//	CallerID: Unique identifier of the caller function, if any.
//	EntryTime: Timestamp when the function was entered.
//	ExitTime: Timestamp when the function exited.
//	Duration: Total execution duration of the function.
//	Params: Map of function parameters and their string representations.
//	ReturnValues: Slice of string representations of the function's return values.
//	MemBefore: Memory allocated (in bytes) before function execution.
//	MemAfter: Memory allocated (in bytes) after function execution.
//	MemDiff: Difference in memory allocation (in bytes).
//	PanicValue: Value captured if the function panics.
//	StackTrace: Captured stack trace in case of a panic.
//	GoroutinesDelta: Change in the number of goroutines during execution.
//	ThreadsDelta: Change in the number of OS threads (using cgo call count as proxy).
//	GCCountDelta: Change in the number of garbage collection cycles during execution.
//	HeapAllocDelta: Difference in heap allocation (in bytes).
//	HeapFreeDelta: Difference in heap free memory (in bytes).
//	NetUsageDelta: Difference in network usage (in bytes).
//	DiskUsageDelta: Difference in disk I/O usage (in bytes).
//	SystemCPULoad: System CPU load at the time of function exit.
//	SystemMemUsage: System memory usage at the time of function exit.
type TraceRecord struct {
	UniqueID        int64             `json:"uniqueId"`
	FunctionName    string            `json:"functionName"`
	CallerID        int64             `json:"callerId,omitempty"`
	EntryTime       time.Time         `json:"entryTime"`
	ExitTime        time.Time         `json:"exitTime"`
	Duration        time.Duration     `json:"duration"`
	Params          map[string]string `json:"params,omitempty"`
	ReturnValues    []string          `json:"returnValues,omitempty"`
	MemBefore       uint64            `json:"memBefore"`
	MemAfter        uint64            `json:"memAfter"`
	MemDiff         uint64            `json:"memDiff"`
	PanicValue      interface{}       `json:"panicValue,omitempty"`
	StackTrace      string            `json:"stackTrace,omitempty"`
	GoroutinesDelta int               `json:"goroutinesDelta,omitempty"`
	ThreadsDelta    int64             `json:"threadsDelta,omitempty"`
	GCCountDelta    uint32            `json:"gcCountDelta,omitempty"`
	HeapAllocDelta  int64             `json:"heapAllocDelta,omitempty"`
	HeapFreeDelta   int64             `json:"heapFreeDelta,omitempty"`
	NetUsageDelta   int64             `json:"netUsageDelta,omitempty"`
	DiskUsageDelta  int64             `json:"diskUsageDelta,omitempty"`
	SystemCPULoad   float64           `json:"systemCpuLoad,omitempty"`
	SystemMemUsage  uint64            `json:"systemMemUsage,omitempty"`
}

// Global variables used for tracing and logging.
var (
	traceRecords  []*TraceRecord         // Aggregated trace records.
	callStack     []*TraceRecord         // Stack of active trace records.
	uniqueID      int64                  // Atomic counter for generating unique IDs.
	mu            sync.Mutex             // Mutex for synchronizing access to global variables.
	logger        *log.Logger            // Logger for trace messages.
	execFrequency = make(map[string]int) // Map tracking execution frequency of functions.
)

// init initializes the tracer package by creating necessary directories and setting up the logger.
// It creates the "tracewrap" directory and opens the log file "tracewrap/tracewrap.log" for logging.
func init() {
	if err := os.MkdirAll("tracewrap", 0755); err != nil {
		log.Println("Error creating log directory:", err)
	}
	logFile, err := os.OpenFile("tracewrap/tracewrap.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_TRUNC, 0644)
	if err != nil {
		log.Println("Error opening log file:", err)
		logger = log.New(os.Stdout, "", log.LstdFlags)
	} else {
		mw := io.MultiWriter(os.Stdout, logFile)
		logger = log.New(mw, "", log.LstdFlags)
	}
}

// readMem returns the current allocated heap memory in bytes using runtime.MemStats.
func readMem() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc
}

// RecordEntry creates a new TraceRecord for a function call and pushes it onto the call stack.
// It records the function name, entry time, initial memory usage, and assigns a unique ID.
// Parameters:
//   - functionName (string): the name of the function being entered.
func RecordEntry(functionName string) {
	mu.Lock()
	defer mu.Unlock()
	id := atomic.AddInt64(&uniqueID, 1)
	record := &TraceRecord{
		UniqueID:     id,
		FunctionName: functionName,
		EntryTime:    time.Now(),
		MemBefore:    readMem(),
		Params:       make(map[string]string),
	}
	if len(callStack) > 0 {
		record.CallerID = callStack[len(callStack)-1].UniqueID
	}
	callStack = append(callStack, record)
	logger.Println("[TRACEWRAP] Entering", functionName, "ID:", id)
}

// RecordParam records a parameter value for the current function call.
// It logs the parameter and stores its string representation in the current TraceRecord.
// Parameters:
//   - paramName (string): the name of the parameter.
//   - value (interface{}): the value of the parameter.
func RecordParam(paramName string, value interface{}) {
	mu.Lock()
	defer mu.Unlock()
	if len(callStack) > 0 {
		top := callStack[len(callStack)-1]
		top.Params[paramName] = fmt.Sprintf("%+v", value)
	}
	logger.Printf("[TRACEWRAP] Parameter %s = %+v", paramName, value)
}

// RecordReturn logs and records return values for the current function call.
// It appends the string representations of the return values to the current TraceRecord.
// Parameters:
//   - functionName (string): the name of the function returning.
//   - returns (...interface{}): variadic return values.
func RecordReturn(functionName string, returns ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	if len(callStack) > 0 {
		top := callStack[len(callStack)-1]
		for _, ret := range returns {
			top.ReturnValues = append(top.ReturnValues, fmt.Sprintf("%+v", ret))
		}
	}
	logger.Printf("[TRACEWRAP] Function %s returning %+v", functionName, returns)
}

// RecordExit finalizes the current TraceRecord by capturing the exit time, computing the duration,
// measuring memory usage difference, and capturing system-level metrics.
// It then logs the function exit and aggregates the record.
// Parameters:
//   - functionName (string): the name of the function exiting.
//   - startTime (time.Time): the start time of the function call.
func RecordExit(functionName string, startTime time.Time) {
	mu.Lock()
	defer mu.Unlock()
	if len(callStack) > 0 {
		top := callStack[len(callStack)-1]
		callStack = callStack[:len(callStack)-1]
		top.ExitTime = time.Now()
		top.Duration = top.ExitTime.Sub(top.EntryTime)
		top.MemAfter = readMem()
		if top.MemAfter > top.MemBefore {
			top.MemDiff = top.MemAfter - top.MemBefore
		} else {
			top.MemDiff = 0
		}
		top.SystemCPULoad = GetSystemCPULoad()
		top.SystemMemUsage = GetSystemMemUsage()
		traceRecords = append(traceRecords, top)
		logger.Printf("[TRACEWRAP] Exiting %s, ID: %d, Duration: %v, MemDiff: %d bytes", functionName, top.UniqueID, top.Duration, top.MemDiff)
		logger.Printf("[TRACEWRAP] DEBUG: Total trace records now: %d", len(traceRecords))
		logger.Printf("[TRACEWRAP] DEBUG: System CPU Load: %f, System Mem Usage: %d bytes", top.SystemCPULoad, top.SystemMemUsage)
	}
}

// RecordPanic records panic information for the current function call.
// It updates the current TraceRecord with the panic value and the associated stack trace, and logs the panic.
// Parameters:
//   - functionName (string): the name of the function where a panic occurred.
//   - panicValue (interface{}): the value recovered from the panic.
//   - stack (string): the stack trace captured at the time of panic.
func RecordPanic(functionName string, panicValue interface{}, stack string) {
	mu.Lock()
	defer mu.Unlock()
	if len(callStack) > 0 {
		top := callStack[len(callStack)-1]
		top.PanicValue = panicValue
		top.StackTrace = stack
	}
	logger.Printf("[TRACEWRAP] Panic in %s: %+v\nStackTrace:\n%s", functionName, panicValue, stack)
}

// RecordGoroutineUsage records the change in goroutine count for the current function call.
// It updates the current TraceRecord with the delta in goroutines.
// Parameters:
//   - functionName (string): the name of the function.
//   - delta (int): the change in the number of goroutines.
func RecordGoroutineUsage(functionName string, delta int) {
	mu.Lock()
	defer mu.Unlock()
	if len(callStack) > 0 {
		top := callStack[len(callStack)-1]
		top.GoroutinesDelta = delta
	}
	logger.Printf("[TRACEWRAP] Function %s Goroutines Spawned: %d", functionName, delta)
}

// RecordThreadUsage records the change in OS thread usage (using cgo call count as a proxy) for the current function call.
// It updates the current TraceRecord with the thread usage delta.
// Parameters:
//   - functionName (string): the name of the function.
//   - delta (int64): the change in thread usage.
func RecordThreadUsage(functionName string, delta int64) {
	mu.Lock()
	defer mu.Unlock()
	if len(callStack) > 0 {
		top := callStack[len(callStack)-1]
		top.ThreadsDelta = delta
	}
	logger.Printf("[TRACEWRAP] Function %s Additional OS Threads Used: %d", functionName, delta)
}

// RecordGCActivity records the change in garbage collection cycles during the function execution.
// It updates the current TraceRecord with the delta in GC count.
// Parameters:
//   - functionName (string): the name of the function.
//   - delta (uint32): the change in the number of GC cycles.
func RecordGCActivity(functionName string, delta uint32) {
	mu.Lock()
	defer mu.Unlock()
	if len(callStack) > 0 {
		top := callStack[len(callStack)-1]
		top.GCCountDelta = delta
	}
	logger.Printf("[TRACEWRAP] Function %s GC Runs: %d", functionName, delta)
}

// RecordHeapUsage records the change in heap allocation for the current function call.
// It updates the current TraceRecord with the delta in heap allocation and freed memory.
// Parameters:
//   - functionName (string): the name of the function.
//   - heapAllocDelta (int64): the difference in heap allocation (in bytes).
//   - heapFreeDelta (int64): the difference in heap free memory (in bytes).
func RecordHeapUsage(functionName string, heapAllocDelta, heapFreeDelta int64) {
	mu.Lock()
	defer mu.Unlock()
	if len(callStack) > 0 {
		top := callStack[len(callStack)-1]
		top.HeapAllocDelta = heapAllocDelta
		top.HeapFreeDelta = heapFreeDelta
	}
	logger.Printf("[TRACEWRAP] Function %s Heap Allocated Delta: %d, Heap Freed Delta: %d", functionName, heapAllocDelta, heapFreeDelta)
}

// RecordIOUsage records the changes in network and disk I/O usage for the current function call.
// It updates the current TraceRecord with the network and disk I/O deltas.
// Parameters:
//   - functionName (string): the name of the function.
//   - netUsageDelta (int64): the change in network usage (in bytes).
//   - diskUsageDelta (int64): the change in disk I/O usage (in bytes).
func RecordIOUsage(functionName string, netUsageDelta, diskUsageDelta int64) {
	mu.Lock()
	defer mu.Unlock()
	if len(callStack) > 0 {
		top := callStack[len(callStack)-1]
		top.NetUsageDelta = netUsageDelta
		top.DiskUsageDelta = diskUsageDelta
	}
	logger.Printf("[TRACEWRAP] Function %s Network Usage Delta: %d, Disk I/O Delta: %d", functionName, netUsageDelta, diskUsageDelta)
}

// RecordExecutionFrequency increments and logs the execution counter for a function.
// Parameters:
//   - functionName (string): the name of the function.
func RecordExecutionFrequency(functionName string) {
	mu.Lock()
	execFrequency[functionName]++
	count := execFrequency[functionName]
	mu.Unlock()
	logger.Printf("[TRACEWRAP] Function %s Calls: %d", functionName, count)
}

// RecordResourceUsage logs the CPU time difference and heap allocation difference for a function execution.
// Parameters:
//   - functionName (string): the name of the function.
//   - cpuTimeDiff (time.Duration): the difference in CPU time.
//   - heapAllocDiff (int64): the difference in heap allocation (in bytes).
func RecordResourceUsage(functionName string, cpuTimeDiff time.Duration, heapAllocDiff int64) {
	logger.Printf("[TRACEWRAP] Function %s Resource Usage - CPU Time: %v, HeapAlloc Diff: %d", functionName, cpuTimeDiff, heapAllocDiff)
}

// DumpCallGraphDOT generates a DOT graph representation of the call graph using the collected trace records,
// and writes it to the specified output file.
// Parameters:
//   - outputFile (string): the path to the output DOT file.
//
// Returns:
//   - error: an error if file writing fails, or nil on success.
func DumpCallGraphDOT(outputFile string) error {
	mu.Lock()
	defer mu.Unlock()

	var sb strings.Builder
	sb.WriteString("digraph CallGraph {\n")
	sb.WriteString("  node [shape=box, style=filled, color=\"lightblue\"];\n")

	logger.Printf("[TRACEWRAP] DEBUG: Generating DOT with %d trace records", len(traceRecords))
	maxlabelLength := 40

	for _, rec := range traceRecords {
		var labelBuilder strings.Builder
		fmt.Fprintf(&labelBuilder, "%s\\nID: %d\\nDuration: %v\\nMemDiff: %d bytes", rec.FunctionName, rec.UniqueID, rec.Duration, rec.MemDiff)
		if rec.SystemCPULoad != 0 || rec.SystemMemUsage != 0 {
			fmt.Fprintf(&labelBuilder, "\\nSysLoad: %.2f, SysMem: %d bytes", rec.SystemCPULoad, rec.SystemMemUsage)
		}
		if len(rec.Params) > 0 {
			labelBuilder.WriteString("\\nParams:")
			for k, v := range rec.Params {
				escapedValue := strings.ReplaceAll(v, "\\", "\\\\")
				escapedValue = strings.ReplaceAll(escapedValue, "\"", "\\\"")
				fmt.Fprintf(&labelBuilder, "\\n  %s = %s...", k, escapedValue[:min(len(escapedValue), maxlabelLength)])
			}
		}
		if len(rec.ReturnValues) > 0 {
			labelBuilder.WriteString("\\nReturns:")
			for i, ret := range rec.ReturnValues {
				escapedRet := strings.ReplaceAll(ret, "\\", "\\\\")
				escapedRet = strings.ReplaceAll(escapedRet, "\"", "\\\"")
				fmt.Fprintf(&labelBuilder, "\\n  [%d] %s...", i, escapedRet[:min(len(escapedRet), maxlabelLength)])
			}
		}
		nodeLabel := labelBuilder.String()
		sb.WriteString(fmt.Sprintf("  %d [label=\"%s\"];\n", rec.UniqueID, nodeLabel))
	}

	for _, rec := range traceRecords {
		if rec.CallerID != 0 {
			sb.WriteString(fmt.Sprintf("  %d -> %d;\n", rec.CallerID, rec.UniqueID))
		}
	}

	sb.WriteString("}\n")
	err := os.WriteFile(outputFile, []byte(sb.String()), 0644)
	if err != nil {
		return fmt.Errorf("failed to write DOT file: %v", err)
	}
	logger.Printf("[TRACEWRAP] Call graph written to: %s\n", outputFile)
	return nil
}

// DumpTrace marshals the aggregated trace records into JSON format and logs the output.
func DumpTrace() {
	mu.Lock()
	defer mu.Unlock()
	jsonBytes, err := json.MarshalIndent(traceRecords, "", "  ")
	if err != nil {
		logger.Println("[TRACEWRAP] Error marshalling trace records:", err)
		return
	}
	logger.Println("[TRACEWRAP] Aggregated Trace Data:")
	logger.Println(string(jsonBytes))
}

// DumpTracePretty prints the aggregated trace records in a human-readable format using pretty-printing.
func DumpTracePretty() {
	pp.Println(traceRecords)
}

// min returns the smaller of two integers a and b.
// Parameters:
//   - a (int): first integer.
//   - b (int): second integer.
//
// Returns:
//   - int: the smaller integer.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetNetworkUsage computes the total network usage by summing the bytes received and sent
// across all network interfaces. It uses gopsutil's net.IOCounters.
// Returns:
//   - int64: the total network usage in bytes.
func GetNetworkUsage() int64 {
	counters, err := net.IOCounters(false)
	if err != nil {
		logger.Println("[TRACEWRAP] Error retrieving network counters:", err)
		return 0
	}
	if len(counters) == 0 {
		return 0
	}
	// When pernic is false, gopsutil returns a single aggregated counter.
	return int64(counters[0].BytesRecv + counters[0].BytesSent)
}

// GetDiskUsage computes the disk I/O usage for the current process by summing the read and write bytes.
// It uses gopsutil's process.IOCounters.
// Returns:
//   - int64: the total disk I/O usage in bytes.
func GetDiskUsage() int64 {
	proc, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		logger.Println("[TRACEWRAP] Error getting current process:", err)
		return 0
	}
	ioCounters, err := proc.IOCounters()
	if err != nil {
		logger.Println("[TRACEWRAP] Error retrieving process I/O counters:", err)
		return 0
	}
	return int64(ioCounters.ReadBytes + ioCounters.WriteBytes)
}

// GetSystemCPULoad returns the 1‑minute load average of the system.
// It uses gopsutil's load.Avg(), which may not be supported on Windows.
// Returns:
//   - float64: the 1‑minute load average, or 0.0 if an error occurs.
func GetSystemCPULoad() float64 {
	avg, err := load.Avg()
	if err != nil {
		logger.Println("[TRACEWRAP] Error retrieving system load average:", err)
		return 0.0
	}
	return avg.Load1
}

// GetSystemMemUsage returns the system memory usage.
// Here we use gopsutil's mem.VirtualMemory() to return the amount of used memory.
// Returns:
//   - uint64: the used system memory in bytes.
func GetSystemMemUsage() uint64 {
	vm, err := mem.VirtualMemory()
	if err != nil {
		logger.Println("[TRACEWRAP] Error retrieving virtual memory info:", err)
		return 0
	}
	return vm.Used
}

// GetProcessCPUTime computes the total CPU time (user + system) used by the current process.
// It uses gopsutil's process.Times().
// Returns:
//   - time.Duration: the total CPU time used, or 0 if an error occurs.
func GetProcessCPUTime() time.Duration {
	proc, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		logger.Println("[TRACEWRAP] Error getting current process:", err)
		return 0
	}
	times, err := proc.Times()
	if err != nil {
		logger.Println("[TRACEWRAP] Error retrieving process CPU times:", err)
		return 0
	}
	totalSeconds := times.User + times.System
	return time.Duration(totalSeconds * float64(time.Second))
}
