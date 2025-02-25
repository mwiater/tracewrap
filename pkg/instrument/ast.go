package instrument

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/mwiater/tracewrap/config"
)

// InstrumentWorkspace traverses all files within the workspace directory and instruments
// each Go source file according to the provided configuration. Files matching any exclude
// patterns from cfg.Instrumentation.Exclude or located within the "tracer" directory are skipped.
//
// Parameters:
//   - workspace (string): the path to the workspace directory.
//   - cfg (config.Config): the configuration settings used for instrumentation.
//
// Returns:
//   - error: an error object if any file fails to be instrumented.
func InstrumentWorkspace(workspace string, cfg config.Config) error {
	return filepath.Walk(workspace, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(workspace, path)
		if err != nil {
			return err
		}
		if strings.HasPrefix(rel, "tracer") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !info.IsDir() && filepath.Ext(path) == ".go" {
			for _, pattern := range cfg.Instrumentation.Exclude {
				matched, err := filepath.Match(pattern, rel)
				if err != nil {
					return fmt.Errorf("error matching pattern %s: %v", pattern, err)
				}
				if matched {
					fmt.Printf("Skipping file (matches exclude pattern '%s'): %s\n", pattern, rel)
					return nil
				}
			}
			fmt.Printf("Instrumenting file: %s\n", path)
			if err := instrumentFile(path); err != nil {
				return fmt.Errorf("failed to instrument file %s: %v", path, err)
			}
		}
		return nil
	})
}

// instrumentFile parses and instruments a single Go source file located at filePath.
// It modifies the AST of the file to inject instrumentation code and then writes
// the modified AST back to the file.
//
// Parameters:
//   - filePath (string): the path to the Go source file to instrument.
//
// Returns:
//   - error: an error object if parsing, instrumentation, or file writing fails.
func instrumentFile(filePath string) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parsing error: %v", err)
	}

	ensureImport := func(pkg string) {
		found := false
		for _, imp := range f.Imports {
			if imp.Path != nil && imp.Path.Value == "\""+pkg+"\"" {
				found = true
				break
			}
		}
		if !found {
			newImport := &ast.ImportSpec{
				Path: &ast.BasicLit{
					Kind:  token.STRING,
					Value: "\"" + pkg + "\"",
				},
			}
			for _, decl := range f.Decls {
				if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
					genDecl.Specs = append(genDecl.Specs, newImport)
					return
				}
			}
			importDecl := &ast.GenDecl{
				Tok:   token.IMPORT,
				Specs: []ast.Spec{newImport},
			}
			f.Decls = append([]ast.Decl{importDecl}, f.Decls...)
		}
	}
	ensureImport("time")
	ensureImport("fmt")
	ensureImport("runtime/debug")
	ensureImport("runtime")
	tracerPkg := strings.Trim(DynamicTracerImport, "\"")
	ensureImport(tracerPkg)

	for _, imp := range f.Imports {
		if imp.Path != nil && strings.Contains(imp.Path.Value, "ghost/tracer") {
			fmt.Printf("DEBUG: Replacing import %s with %s in file %s\n", imp.Path.Value, DynamicTracerImport, filePath)
			imp.Path.Value = DynamicTracerImport
		}
	}

	for _, decl := range f.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Body != nil {
			if fn.Name.Name == "init" {
				continue
			}

			fnNameLit := "\"" + fn.Name.Name + "\""

			recoverStmt := &ast.DeferStmt{
				Call: &ast.CallExpr{
					Fun: &ast.FuncLit{
						Type: &ast.FuncType{
							Params: &ast.FieldList{},
						},
						Body: &ast.BlockStmt{
							List: []ast.Stmt{
								&ast.AssignStmt{
									Lhs: []ast.Expr{&ast.Ident{Name: "r"}},
									Tok: token.DEFINE,
									Rhs: []ast.Expr{
										&ast.CallExpr{
											Fun:  &ast.Ident{Name: "recover"},
											Args: []ast.Expr{},
										},
									},
								},
								&ast.IfStmt{
									Cond: &ast.BinaryExpr{
										X:  &ast.Ident{Name: "r"},
										Op: token.NEQ,
										Y:  &ast.Ident{Name: "nil"},
									},
									Body: &ast.BlockStmt{
										List: []ast.Stmt{
											&ast.ExprStmt{
												X: &ast.CallExpr{
													Fun: &ast.SelectorExpr{
														X:   &ast.Ident{Name: "tracer"},
														Sel: &ast.Ident{Name: "RecordPanic"},
													},
													Args: []ast.Expr{
														&ast.BasicLit{Kind: token.STRING, Value: fnNameLit},
														&ast.Ident{Name: "r"},
														&ast.CallExpr{
															Fun: ast.NewIdent("string"),
															Args: []ast.Expr{
																&ast.CallExpr{
																	Fun: &ast.SelectorExpr{
																		X:   &ast.Ident{Name: "debug"},
																		Sel: &ast.Ident{Name: "Stack"},
																	},
																	Args: []ast.Expr{},
																},
															},
														},
													},
												},
											},
											&ast.ExprStmt{
												X: &ast.CallExpr{
													Fun:  &ast.Ident{Name: "panic"},
													Args: []ast.Expr{&ast.Ident{Name: "r"}},
												},
											},
										},
									},
								},
							},
						},
					},
					Args: []ast.Expr{},
				},
			}

			startTimeDecl := &ast.AssignStmt{
				Lhs: []ast.Expr{&ast.Ident{Name: "__tracewrap_startTime"}},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "time"},
							Sel: &ast.Ident{Name: "Now"},
						},
						Args: []ast.Expr{},
					},
				},
			}

			startCPUTimeDecl := &ast.AssignStmt{
				Lhs: []ast.Expr{&ast.Ident{Name: "__tracewrap_startCPUTime"}},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "tracer"},
							Sel: &ast.Ident{Name: "GetProcessCPUTime"},
						},
						Args: []ast.Expr{},
					},
				},
			}

			deferExit := &ast.DeferStmt{
				Call: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "tracer"},
						Sel: &ast.Ident{Name: "RecordExit"},
					},
					Args: []ast.Expr{
						&ast.BasicLit{Kind: token.STRING, Value: fnNameLit},
						&ast.Ident{Name: "__tracewrap_startTime"},
					},
				},
			}

			recordEntryCall := &ast.ExprStmt{
				X: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "tracer"},
						Sel: &ast.Ident{Name: "RecordEntry"},
					},
					Args: []ast.Expr{
						&ast.BasicLit{Kind: token.STRING, Value: fnNameLit},
					},
				},
			}

			var paramLogs []ast.Stmt
			if fn.Type.Params != nil {
				for _, field := range fn.Type.Params.List {
					for _, name := range field.Names {
						logCall := &ast.ExprStmt{
							X: &ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X:   &ast.Ident{Name: "tracer"},
									Sel: &ast.Ident{Name: "RecordParam"},
								},
								Args: []ast.Expr{
									&ast.BasicLit{Kind: token.STRING, Value: "\"" + name.Name + "\""},
									&ast.CallExpr{
										Fun: &ast.Ident{Name: "fmt.Sprintf"},
										Args: []ast.Expr{
											&ast.BasicLit{Kind: token.STRING, Value: "\"%+v\""},
											&ast.Ident{Name: name.Name},
										},
									},
								},
							},
						}
						paramLogs = append(paramLogs, logCall)
					}
				}
			}

			if fn.Name.Name == "main" && fn.Recv == nil {
				dumpCallGraphStmt := &ast.ExprStmt{
					X: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   ast.NewIdent("tracer"),
							Sel: ast.NewIdent("DumpCallGraphDOT"),
						},
						Args: []ast.Expr{
							&ast.BasicLit{
								Kind:  token.STRING,
								Value: "\"tracewrap/callgraph.dot\"",
							},
						},
					},
				}
				fn.Body.List = append(fn.Body.List, dumpCallGraphStmt)
			}

			startGoroutinesDecl := &ast.AssignStmt{
				Lhs: []ast.Expr{&ast.Ident{Name: "__tracewrap_startGoroutines"}},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   ast.NewIdent("runtime"),
							Sel: ast.NewIdent("NumGoroutine"),
						},
						Args: []ast.Expr{},
					},
				},
			}

			startThreadsDecl := &ast.AssignStmt{
				Lhs: []ast.Expr{&ast.Ident{Name: "__tracewrap_startThreads"}},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   ast.NewIdent("runtime"),
							Sel: ast.NewIdent("NumCgoCall"),
						},
						Args: []ast.Expr{},
					},
				},
			}

			memStatsBeforeDecl := &ast.DeclStmt{
				Decl: &ast.GenDecl{
					Tok: token.VAR,
					Specs: []ast.Spec{
						&ast.ValueSpec{
							Names: []*ast.Ident{ast.NewIdent("__tracewrap_memStatsBefore")},
							Type: &ast.SelectorExpr{
								X:   ast.NewIdent("runtime"),
								Sel: ast.NewIdent("MemStats"),
							},
						},
					},
				},
			}

			readMemStatsBefore := &ast.ExprStmt{
				X: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   ast.NewIdent("runtime"),
						Sel: ast.NewIdent("ReadMemStats"),
					},
					Args: []ast.Expr{
						&ast.UnaryExpr{
							Op: token.AND,
							X:  ast.NewIdent("__tracewrap_memStatsBefore"),
						},
					},
				},
			}

			startNetUsageDecl := &ast.AssignStmt{
				Lhs: []ast.Expr{&ast.Ident{Name: "__tracewrap_startNetUsage"}},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   ast.NewIdent("tracer"),
							Sel: ast.NewIdent("GetNetworkUsage"),
						},
						Args: []ast.Expr{},
					},
				},
			}

			startDiskUsageDecl := &ast.AssignStmt{
				Lhs: []ast.Expr{&ast.Ident{Name: "__tracewrap_startDiskUsage"}},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   ast.NewIdent("tracer"),
							Sel: ast.NewIdent("GetDiskUsage"),
						},
						Args: []ast.Expr{},
					},
				},
			}

			newDefer := &ast.DeferStmt{
				Call: &ast.CallExpr{
					Fun: &ast.FuncLit{
						Type: &ast.FuncType{
							Params: &ast.FieldList{},
						},
						Body: &ast.BlockStmt{
							List: []ast.Stmt{
								&ast.DeclStmt{
									Decl: &ast.GenDecl{
										Tok: token.VAR,
										Specs: []ast.Spec{
											&ast.ValueSpec{
												Names: []*ast.Ident{ast.NewIdent("__tracewrap_endCPUTime")},
												Type: &ast.SelectorExpr{
													X:   ast.NewIdent("time"),
													Sel: ast.NewIdent("Duration"),
												},
												Values: []ast.Expr{
													&ast.BasicLit{Kind: token.INT, Value: "0"},
												},
											},
											&ast.ValueSpec{
												Names: []*ast.Ident{ast.NewIdent("__tracewrap_cpuTimeDiff")},
												Type: &ast.SelectorExpr{
													X:   ast.NewIdent("time"),
													Sel: ast.NewIdent("Duration"),
												},
												Values: []ast.Expr{
													&ast.BasicLit{Kind: token.INT, Value: "0"},
												},
											},
											&ast.ValueSpec{
												Names: []*ast.Ident{ast.NewIdent("__tracewrap_memStatsAfter")},
												Type: &ast.SelectorExpr{
													X:   ast.NewIdent("runtime"),
													Sel: ast.NewIdent("MemStats"),
												},
												Values: []ast.Expr{
													&ast.CompositeLit{
														Type: &ast.SelectorExpr{
															X:   ast.NewIdent("runtime"),
															Sel: ast.NewIdent("MemStats"),
														},
													},
												},
											},
											&ast.ValueSpec{
												Names: []*ast.Ident{ast.NewIdent("__tracewrap_endGoroutines")},
												Type:  ast.NewIdent("int"),
												Values: []ast.Expr{
													&ast.BasicLit{Kind: token.INT, Value: "0"},
												},
											},
											&ast.ValueSpec{
												Names: []*ast.Ident{ast.NewIdent("__tracewrap_endThreads")},
												Type:  ast.NewIdent("int64"),
												Values: []ast.Expr{
													&ast.BasicLit{Kind: token.INT, Value: "0"},
												},
											},
											&ast.ValueSpec{
												Names: []*ast.Ident{ast.NewIdent("__tracewrap_endNetUsage")},
												Type:  ast.NewIdent("int64"),
												Values: []ast.Expr{
													&ast.BasicLit{Kind: token.INT, Value: "0"},
												},
											},
											&ast.ValueSpec{
												Names: []*ast.Ident{ast.NewIdent("__tracewrap_endDiskUsage")},
												Type:  ast.NewIdent("int64"),
												Values: []ast.Expr{
													&ast.BasicLit{Kind: token.INT, Value: "0"},
												},
											},
										},
									},
								},
								&ast.AssignStmt{
									Lhs: []ast.Expr{&ast.Ident{Name: "__tracewrap_endCPUTime"}},
									Tok: token.ASSIGN,
									Rhs: []ast.Expr{
										&ast.CallExpr{
											Fun: &ast.SelectorExpr{
												X:   ast.NewIdent("tracer"),
												Sel: ast.NewIdent("GetProcessCPUTime"),
											},
											Args: []ast.Expr{},
										},
									},
								},
								&ast.AssignStmt{
									Lhs: []ast.Expr{&ast.Ident{Name: "__tracewrap_cpuTimeDiff"}},
									Tok: token.ASSIGN,
									Rhs: []ast.Expr{
										&ast.BinaryExpr{
											X:  &ast.Ident{Name: "__tracewrap_endCPUTime"},
											Op: token.SUB,
											Y:  &ast.Ident{Name: "__tracewrap_startCPUTime"},
										},
									},
								},
								&ast.ExprStmt{
									X: &ast.CallExpr{
										Fun: &ast.SelectorExpr{
											X:   ast.NewIdent("runtime"),
											Sel: ast.NewIdent("ReadMemStats"),
										},
										Args: []ast.Expr{
											&ast.UnaryExpr{
												Op: token.AND,
												X:  ast.NewIdent("__tracewrap_memStatsAfter"),
											},
										},
									},
								},
								&ast.ExprStmt{
									X: &ast.CallExpr{
										Fun: &ast.SelectorExpr{
											X:   ast.NewIdent("tracer"),
											Sel: ast.NewIdent("RecordResourceUsage"),
										},
										Args: []ast.Expr{
											&ast.BasicLit{
												Kind:  token.STRING,
												Value: fnNameLit,
											},
											&ast.Ident{Name: "__tracewrap_cpuTimeDiff"},
											&ast.BinaryExpr{
												X: &ast.CallExpr{
													Fun: ast.NewIdent("int64"),
													Args: []ast.Expr{
														&ast.SelectorExpr{
															X:   ast.NewIdent("__tracewrap_memStatsAfter"),
															Sel: ast.NewIdent("HeapAlloc"),
														},
													},
												},
												Op: token.SUB,
												Y: &ast.CallExpr{
													Fun: ast.NewIdent("int64"),
													Args: []ast.Expr{
														&ast.SelectorExpr{
															X:   ast.NewIdent("__tracewrap_memStatsBefore"),
															Sel: ast.NewIdent("HeapAlloc"),
														},
													},
												},
											},
										},
									},
								},
								&ast.AssignStmt{
									Lhs: []ast.Expr{&ast.Ident{Name: "__tracewrap_endGoroutines"}},
									Tok: token.ASSIGN,
									Rhs: []ast.Expr{
										&ast.CallExpr{
											Fun: &ast.SelectorExpr{
												X:   ast.NewIdent("runtime"),
												Sel: ast.NewIdent("NumGoroutine"),
											},
											Args: []ast.Expr{},
										},
									},
								},
								&ast.ExprStmt{
									X: &ast.CallExpr{
										Fun: &ast.SelectorExpr{
											X:   ast.NewIdent("tracer"),
											Sel: ast.NewIdent("RecordGoroutineUsage"),
										},
										Args: []ast.Expr{
											&ast.BasicLit{
												Kind:  token.STRING,
												Value: fnNameLit,
											},
											&ast.BinaryExpr{
												X:  ast.NewIdent("__tracewrap_endGoroutines"),
												Op: token.SUB,
												Y:  ast.NewIdent("__tracewrap_startGoroutines"),
											},
										},
									},
								},
								&ast.AssignStmt{
									Lhs: []ast.Expr{&ast.Ident{Name: "__tracewrap_endThreads"}},
									Tok: token.ASSIGN,
									Rhs: []ast.Expr{
										&ast.CallExpr{
											Fun: &ast.SelectorExpr{
												X:   ast.NewIdent("runtime"),
												Sel: ast.NewIdent("NumCgoCall"),
											},
											Args: []ast.Expr{},
										},
									},
								},
								&ast.ExprStmt{
									X: &ast.CallExpr{
										Fun: &ast.SelectorExpr{
											X:   ast.NewIdent("tracer"),
											Sel: ast.NewIdent("RecordThreadUsage"),
										},
										Args: []ast.Expr{
											&ast.BasicLit{
												Kind:  token.STRING,
												Value: fnNameLit,
											},
											&ast.BinaryExpr{
												X:  ast.NewIdent("__tracewrap_endThreads"),
												Op: token.SUB,
												Y:  ast.NewIdent("__tracewrap_startThreads"),
											},
										},
									},
								},
								&ast.AssignStmt{
									Lhs: []ast.Expr{&ast.Ident{Name: "__tracewrap_memStatsAfter"}},
									Tok: token.ASSIGN,
									Rhs: []ast.Expr{
										&ast.CompositeLit{
											Type: ast.NewIdent("runtime.MemStats"),
										},
									},
								},
								&ast.ExprStmt{
									X: &ast.CallExpr{
										Fun: &ast.SelectorExpr{
											X:   ast.NewIdent("runtime"),
											Sel: ast.NewIdent("ReadMemStats"),
										},
										Args: []ast.Expr{
											&ast.UnaryExpr{
												Op: token.AND,
												X:  ast.NewIdent("__tracewrap_memStatsAfter"),
											},
										},
									},
								},
								&ast.ExprStmt{
									X: &ast.CallExpr{
										Fun: &ast.SelectorExpr{
											X:   ast.NewIdent("tracer"),
											Sel: ast.NewIdent("RecordGCActivity"),
										},
										Args: []ast.Expr{
											&ast.BasicLit{
												Kind:  token.STRING,
												Value: fnNameLit,
											},
											&ast.BinaryExpr{
												X: &ast.SelectorExpr{
													X:   ast.NewIdent("__tracewrap_memStatsAfter"),
													Sel: ast.NewIdent("NumGC"),
												},
												Op: token.SUB,
												Y: &ast.SelectorExpr{
													X:   ast.NewIdent("__tracewrap_memStatsBefore"),
													Sel: ast.NewIdent("NumGC"),
												},
											},
										},
									},
								},
								&ast.ExprStmt{
									X: &ast.CallExpr{
										Fun: &ast.SelectorExpr{
											X:   ast.NewIdent("tracer"),
											Sel: ast.NewIdent("RecordHeapUsage"),
										},
										Args: []ast.Expr{
											&ast.BasicLit{
												Kind:  token.STRING,
												Value: fnNameLit,
											},
											&ast.BinaryExpr{
												X: &ast.CallExpr{
													Fun: ast.NewIdent("int64"),
													Args: []ast.Expr{
														&ast.SelectorExpr{
															X:   ast.NewIdent("__tracewrap_memStatsAfter"),
															Sel: ast.NewIdent("HeapAlloc"),
														},
													},
												},
												Op: token.SUB,
												Y: &ast.CallExpr{
													Fun: ast.NewIdent("int64"),
													Args: []ast.Expr{
														&ast.SelectorExpr{
															X:   ast.NewIdent("__tracewrap_memStatsBefore"),
															Sel: ast.NewIdent("HeapAlloc"),
														},
													},
												},
											},
											&ast.BasicLit{Kind: token.INT, Value: "0"},
										},
									},
								},
								&ast.AssignStmt{
									Lhs: []ast.Expr{&ast.Ident{Name: "__tracewrap_endNetUsage"}},
									Tok: token.ASSIGN,
									Rhs: []ast.Expr{
										&ast.CallExpr{
											Fun: &ast.SelectorExpr{
												X:   ast.NewIdent("tracer"),
												Sel: ast.NewIdent("GetNetworkUsage"),
											},
											Args: []ast.Expr{},
										},
									},
								},
								&ast.AssignStmt{
									Lhs: []ast.Expr{&ast.Ident{Name: "__tracewrap_endDiskUsage"}},
									Tok: token.ASSIGN,
									Rhs: []ast.Expr{
										&ast.CallExpr{
											Fun: &ast.SelectorExpr{
												X:   ast.NewIdent("tracer"),
												Sel: ast.NewIdent("GetDiskUsage"),
											},
											Args: []ast.Expr{},
										},
									},
								},
								&ast.ExprStmt{
									X: &ast.CallExpr{
										Fun: &ast.SelectorExpr{
											X:   ast.NewIdent("tracer"),
											Sel: ast.NewIdent("RecordIOUsage"),
										},
										Args: []ast.Expr{
											&ast.BasicLit{
												Kind:  token.STRING,
												Value: fnNameLit,
											},
											&ast.BinaryExpr{
												X:  ast.NewIdent("__tracewrap_endNetUsage"),
												Op: token.SUB,
												Y:  ast.NewIdent("__tracewrap_startNetUsage"),
											},
											&ast.BinaryExpr{
												X:  ast.NewIdent("__tracewrap_endDiskUsage"),
												Op: token.SUB,
												Y:  ast.NewIdent("__tracewrap_startDiskUsage"),
											},
										},
									},
								},
								&ast.ExprStmt{
									X: &ast.CallExpr{
										Fun: &ast.SelectorExpr{
											X:   ast.NewIdent("tracer"),
											Sel: ast.NewIdent("RecordExecutionFrequency"),
										},
										Args: []ast.Expr{
											&ast.BasicLit{
												Kind:  token.STRING,
												Value: fnNameLit,
											},
										},
									},
								},
							},
						},
					},
					Args: []ast.Expr{},
				},
			}

			newStmts := []ast.Stmt{
				recoverStmt,
				startTimeDecl,
				startCPUTimeDecl,
				startGoroutinesDecl,
				startThreadsDecl,
				memStatsBeforeDecl,
				readMemStatsBefore,
				startNetUsageDecl,
				startDiskUsageDecl,
				deferExit,
				newDefer,
				recordEntryCall,
			}
			newStmts = append(newStmts, paramLogs...)
			fn.Body.List = append(newStmts, fn.Body.List...)
			fn.Body = transformReturnsInBlock(fn.Body, fn.Name.Name)
		}
	}
	if strings.HasSuffix(filePath, "main.go") {
		dummyDecl := &ast.GenDecl{
			Tok: token.VAR,
			Specs: []ast.Spec{
				&ast.ValueSpec{
					Names: []*ast.Ident{ast.NewIdent("_dummy")},
					Values: []ast.Expr{&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   ast.NewIdent("fmt"),
							Sel: ast.NewIdent("Sprintf"),
						},
						Args: []ast.Expr{
							&ast.BasicLit{Kind: token.STRING, Value: "\"\""},
						},
					}},
				},
			},
		}
		f.Decls = append(f.Decls, dummyDecl)
	}

	outFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer outFile.Close()
	if err := printer.Fprint(outFile, fset, f); err != nil {
		return fmt.Errorf("error printing file: %v", err)
	}
	return nil
}

// transformReturnsInBlock recursively processes all statements within a block to transform return statements.
// It updates return statements by inserting instrumentation code that records return values.
//
// Parameters:
//   - block (*ast.BlockStmt): pointer to the AST block statement.
//   - functionName (string): the name of the function containing the block.
//
// Returns:
//   - *ast.BlockStmt: the transformed block statement.
func transformReturnsInBlock(block *ast.BlockStmt, functionName string) *ast.BlockStmt {
	for i, stmt := range block.List {
		block.List[i] = transformReturnsInStmt(stmt, functionName)
	}
	return block
}

// transformReturnsInStmt recursively processes an AST statement to transform return statements
// by wrapping them with instrumentation for recording return values.
//
// Parameters:
//   - stmt (ast.Stmt): the statement to process.
//   - functionName (string): the name of the function containing the statement.
//
// Returns:
//   - ast.Stmt: the transformed statement.
func transformReturnsInStmt(stmt ast.Stmt, functionName string) ast.Stmt {
	switch s := stmt.(type) {
	case *ast.BlockStmt:
		return transformReturnsInBlock(s, functionName)
	case *ast.IfStmt:
		s.Body = transformReturnsInBlock(s.Body, functionName)
		if s.Else != nil {
			s.Else = transformReturnsInStmt(s.Else, functionName)
		}
		return s
	case *ast.ForStmt:
		s.Body = transformReturnsInBlock(s.Body, functionName)
		return s
	case *ast.ReturnStmt:
		for _, expr := range s.Results {
			if _, ok := expr.(*ast.CallExpr); ok {
				return s
			}
			if ident, ok := expr.(*ast.Ident); ok && ident.Name == "nil" {
				return s
			}
		}
		return transformReturnStmt(s, functionName)
	default:
		return s
	}
}

// transformReturnStmt transforms a return statement by assigning its return values
// to temporary variables, recording these values with the tracer, and then returning the variables.
// This ensures that return values are logged before the function exits.
//
// Parameters:
//   - ret (*ast.ReturnStmt): pointer to the original return statement.
//   - functionName (string): the name of the function containing the return.
//
// Returns:
//   - ast.Stmt: a new block statement containing assignments, tracer recording, and the new return.
func transformReturnStmt(ret *ast.ReturnStmt, functionName string) ast.Stmt {
	var assignments []ast.Stmt
	var newIdents []ast.Expr
	for i, expr := range ret.Results {
		varName := fmt.Sprintf("_ret%d", i)
		assignStmt := &ast.AssignStmt{
			Lhs: []ast.Expr{&ast.Ident{Name: varName}},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{expr},
		}
		assignments = append(assignments, assignStmt)
		newIdents = append(newIdents, &ast.Ident{Name: varName})
	}
	recordCall := &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "tracer"},
				Sel: &ast.Ident{Name: "RecordReturn"},
			},
			Args: append([]ast.Expr{
				&ast.BasicLit{Kind: token.STRING, Value: "\"" + functionName + "\""},
			}, newIdents...),
		},
	}
	newReturn := &ast.ReturnStmt{
		Results: newIdents,
	}
	block := &ast.BlockStmt{
		List: append(assignments, recordCall, newReturn),
	}
	return block
}
