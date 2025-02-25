package instrument

import (
	"fmt"
)

// DynamicTracerImport holds the dynamic tracer import string set by SetDynamicTracerImport.
// It is used to dynamically specify the tracer package import.
var DynamicTracerImport string

// SetDynamicTracerImport sets the DynamicTracerImport variable to the tracer package import path.
// The workspace parameter is accepted for interface consistency though it is not used in this implementation.
//
// Parameters:
//   - workspace (string): the path to the workspace directory.
//
// Returns:
//   - error: an error object if setting the tracer import fails (currently always nil).
func SetDynamicTracerImport(workspace string) error {
	DynamicTracerImport = "\"github.com/mwiater/tracewrap/pkg/tracer\""
	fmt.Println("DEBUG: SetDynamicTracerImport set to", DynamicTracerImport)
	return nil
}
