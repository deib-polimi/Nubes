package parser

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"
)

func getErrorCheckExpr(fn *ast.FuncDecl, errorVariableName string) ast.IfStmt {
	ifStmt := ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  &ast.Ident{Name: errorVariableName},
			Op: token.NEQ,
			Y:  &ast.Ident{Name: "nil"},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{},
		},
	}

	// given faas handler reqs, there can be one optional return type apart from "error"
	if fn.Type.Results != nil && fn.Type.Results.List != nil {
		returnTypeName := types.ExprString(fn.Type.Results.List[0].Type)
		if returnTypeName != "error" {
			ifStmt.Body.List = []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.StarExpr{
							X: &ast.CallExpr{
								Args: []ast.Expr{&ast.Ident{Name: returnTypeName}},
								Fun:  &ast.Ident{Name: "new"},
							},
						},
						&ast.Ident{
							Name: LibErrorVariableName,
						},
					},
				},
			}

			return ifStmt
		}
	}

	ifStmt.Body.List = []ast.Stmt{
		&ast.ReturnStmt{
			Results: []ast.Expr{
				&ast.Ident{
					Name: LibErrorVariableName,
				},
			},
		},
	}
	return ifStmt
}

type getDBStmtsParam struct {
	idFieldName          string
	typeName             string
	receiverVariableName string
	fieldName            string
	fieldType            string
}

func getGetterDBStmts(fn *ast.FuncDecl, input getDBStmtsParam) ast.IfStmt {
	isInitializedCheck := getIsInitializedCheck(fn.Recv.List[0].Names[0].Name)
	getFieldFromLib := ast.AssignStmt{
		Tok: token.DEFINE,
		Lhs: []ast.Expr{
			&ast.Ident{Name: "fieldValue"},
			&ast.Ident{Name: LibErrorVariableName},
		},
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.IndexExpr{
					Index: &ast.Ident{Name: input.fieldType},
					X: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "lib"},
						Sel: &ast.Ident{Name: LibraryGetFieldOfType},
					},
				},
				Args: []ast.Expr{
					&ast.CompositeLit{
						Type: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "lib"},
							Sel: &ast.Ident{Name: GetStateParamType},
						},
						Elts: []ast.Expr{
							&ast.KeyValueExpr{
								Key: &ast.Ident{Name: Id},
								Value: &ast.SelectorExpr{
									X:   &ast.Ident{Name: input.receiverVariableName},
									Sel: &ast.Ident{Name: input.idFieldName},
								},
							},
							&ast.KeyValueExpr{
								Key: &ast.Ident{Name: TypeName},
								Value: &ast.BasicLit{
									Kind:  token.STRING,
									Value: "\"" + input.typeName + "\"",
								},
							},
							&ast.KeyValueExpr{
								Key: &ast.Ident{Name: FieldName},
								Value: &ast.BasicLit{
									Kind:  token.STRING,
									Value: "\"" + input.fieldName + "\"",
								},
							},
						},
					},
				},
			},
		}}
	errorCheck := getErrorCheckExpr(fn, LibErrorVariableName)
	fieldAssign := ast.AssignStmt{
		Tok: token.ASSIGN,
		Lhs: []ast.Expr{
			&ast.SelectorExpr{
				X:   &ast.Ident{Name: input.receiverVariableName},
				Sel: &ast.Ident{Name: input.fieldName},
			},
		},
		Rhs: []ast.Expr{
			&ast.Ident{Name: "fieldValue"},
		},
	}

	isInitializedCheck.Body.List = []ast.Stmt{&getFieldFromLib, &errorCheck, &fieldAssign}
	return isInitializedCheck
}

func getReferenceNavigationListDBStmts(fn *ast.FuncDecl, input getDBStmtsParam) ast.IfStmt {

	isUnInitializedCheck := ast.IfStmt{
		Cond: &ast.UnaryExpr{
			Op: token.NOT,
			X: &ast.SelectorExpr{
				X:   &ast.Ident{Name: fn.Recv.List[0].Names[0].Name},
				Sel: &ast.Ident{Name: IsInitializedFieldName},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{&ast.ReturnStmt{
				Results: []ast.Expr{
					&ast.Ident{Name: "nil"},
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "fmt"},
							Sel: &ast.Ident{Name: "Errorf"},
						},
						Args: []ast.Expr{&ast.BasicLit{
							Kind:  token.STRING,
							Value: "`fields of type ReferenceNavigationList can be used only after instance initialization. \n\t\t\tUse lib.Load or lib.Export from the Nubes library to create initialized instances`",
						}},
					},
				},
			}},
		},
	}
	return isUnInitializedCheck
}

func getSetterDBStmts(fn *ast.FuncDecl, input getDBStmtsParam) ast.IfStmt {
	isInitializedCheck := getIsInitializedCheck(fn.Recv.List[0].Names[0].Name)
	getFieldFromLib := ast.AssignStmt{
		Tok: token.DEFINE,
		Lhs: []ast.Expr{
			&ast.Ident{Name: LibErrorVariableName},
		},
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "lib"},
					Sel: &ast.Ident{Name: SetField},
				},
				Args: []ast.Expr{
					&ast.CompositeLit{
						Type: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "lib"},
							Sel: &ast.Ident{Name: SetFieldParam},
						},
						Elts: []ast.Expr{
							&ast.KeyValueExpr{
								Key: &ast.Ident{Name: Id},
								Value: &ast.SelectorExpr{
									X:   &ast.Ident{Name: input.receiverVariableName},
									Sel: &ast.Ident{Name: input.idFieldName},
								},
							},
							&ast.KeyValueExpr{
								Key: &ast.Ident{Name: TypeName},
								Value: &ast.BasicLit{
									Kind:  token.STRING,
									Value: "\"" + input.typeName + "\"",
								},
							},
							&ast.KeyValueExpr{
								Key: &ast.Ident{Name: FieldName},
								Value: &ast.BasicLit{
									Kind:  token.STRING,
									Value: "\"" + input.fieldName + "\"",
								},
							},
							&ast.KeyValueExpr{
								Key: &ast.Ident{Name: "Value"},
								Value: &ast.SelectorExpr{
									X:   &ast.Ident{Name: input.receiverVariableName},
									Sel: &ast.Ident{Name: input.fieldName},
								},
							},
						},
					},
				},
			}}}
	errorCheck := getErrorCheckExpr(fn, LibErrorVariableName)

	isInitializedCheck.Body.List = []ast.Stmt{&getFieldFromLib, &errorCheck}
	return isInitializedCheck
}

func getNobjectStateConditionalRetrieval(fn *ast.FuncDecl, parsedPackage ParsedPackage) ast.IfStmt {
	isInitializedCheck := getIsInitializedCheck(fn.Recv.List[0].Names[0].Name)
	readFromLibExpr, isPointerReceiver := getReadFromLibExpr(fn, parsedPackage.TypesWithCustomId)
	errorCheck := getErrorCheckExpr(fn, LibErrorVariableName)

	tempRecvAssignment := getTempRecvAssignStmt(fn.Recv.List[0].Names[0].Name, isPointerReceiver)
	initCall := getInitCall(fn.Recv.List[0].Names[0].Name)
	isInitializedCheck.Body.List = []ast.Stmt{&readFromLibExpr, &errorCheck, &tempRecvAssignment, &initCall}

	return isInitializedCheck
}

func getNobjectStateConditionalUpsert(fn *ast.FuncDecl, parsedPackage ParsedPackage) ast.IfStmt {
	isInitializedCheck := getIsInitializedCheck(fn.Recv.List[0].Names[0].Name)
	saveExpr := getUpsertInLibExpr(fn, parsedPackage.TypesWithCustomId)
	erorCheck := getErrorCheckExpr(fn, LibErrorVariableName)

	isInitializedCheck.Body.List = []ast.Stmt{&saveExpr, &erorCheck}
	return isInitializedCheck
}

func getReadFromLibExpr(fn *ast.FuncDecl, typesWithCustomId map[string]string) (ast.AssignStmt, bool) {
	typeName := types.ExprString(fn.Recv.List[0].Type)
	isPointerReceiver := strings.Contains(typeName, "*")
	typeName = strings.TrimPrefix(typeName, "*")

	idFieldName := ""
	if idField, isPresent := typesWithCustomId[typeName]; isPresent {
		idFieldName = idField
	} else {
		idFieldName = Id
	}

	assignStmt := ast.AssignStmt{
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.IndexExpr{
					Index: &ast.Ident{Name: typeName},
					X: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "lib"},
						Sel: &ast.Ident{Name: LibraryGetObjectStateMethod},
					},
				},
				Args: []ast.Expr{
					&ast.SelectorExpr{
						X:   &ast.Ident{Name: fn.Recv.List[0].Names[0].Name},
						Sel: &ast.Ident{Name: idFieldName},
					},
				},
			},
		},
	}

	assignStmt.Lhs = []ast.Expr{
		&ast.Ident{Name: TemporaryReceiverName},
		&ast.Ident{Name: LibErrorVariableName},
	}

	return assignStmt, isPointerReceiver
}

func getUpsertInLibExpr(fn *ast.FuncDecl, typesWithCustomId map[string]string) ast.AssignStmt {
	typeName := getFunctionReceiverTypeAsString(fn.Recv)
	idFieldName := ""
	if idField, isPresent := typesWithCustomId[typeName]; isPresent {
		idFieldName = idField
	} else {
		idFieldName = Id
	}

	return ast.AssignStmt{
		Tok: token.DEFINE,
		Lhs: []ast.Expr{
			&ast.Ident{Name: LibErrorVariableName},
		},
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "lib"},
					Sel: &ast.Ident{Name: Upsert},
				},
				Args: []ast.Expr{
					&ast.Ident{Name: fn.Recv.List[0].Names[0].Name},
					&ast.SelectorExpr{
						X:   &ast.Ident{Name: fn.Recv.List[0].Names[0].Name},
						Sel: &ast.Ident{Name: idFieldName},
					}},
			},
		},
	}
}

func getTempRecvAssignStmt(receiverName string, isPointerReceiver bool) ast.AssignStmt {
	assign := ast.AssignStmt{
		Tok: token.ASSIGN,
		Lhs: []ast.Expr{
			&ast.Ident{Name: receiverName},
		},
	}

	if isPointerReceiver {
		assign.Rhs = []ast.Expr{&ast.Ident{Name: TemporaryReceiverName}}
	} else {
		assign.Rhs = []ast.Expr{
			&ast.StarExpr{
				X: &ast.Ident{Name: TemporaryReceiverName},
			},
		}
	}

	return assign
}

func getIsInitializedCheck(receiverVariableName string) ast.IfStmt {
	ifStmt := ast.IfStmt{
		Cond: &ast.SelectorExpr{
			X:   &ast.Ident{Name: receiverVariableName},
			Sel: &ast.Ident{Name: IsInitializedFieldName},
		},
		Body: &ast.BlockStmt{},
	}

	return ifStmt
}

func getInitCall(receiverVariableName string) ast.ExprStmt {
	return ast.ExprStmt{X: &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   &ast.Ident{Name: receiverVariableName},
			Sel: &ast.Ident{Name: InitFunctionName},
		},
		Args: []ast.Expr{},
	}}
}
