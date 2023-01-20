package parser

import (
	"fmt"
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

func getNewCtorStmts(fn *ast.FuncDecl, typeName, idFieldName string) ([]ast.Stmt, error) {
	toInsertVariableName, err := getFirstParamVariableName(fn.Type.Params)
	if err != nil {
		return nil, err
	}

	insertInLib := ast.AssignStmt{
		Tok: token.DEFINE,
		Lhs: []ast.Expr{
			&ast.Ident{Name: "out"},
			&ast.Ident{Name: LibErrorVariableName},
		},
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "lib"},
					Sel: &ast.Ident{Name: "Insert"},
				},
				Args: []ast.Expr{
					&ast.Ident{Name: toInsertVariableName},
				},
			}}}
	errorCheck := getErrorCheckExpr(fn, LibErrorVariableName)
	idAssign := ast.AssignStmt{
		Tok: token.ASSIGN,
		Lhs: []ast.Expr{
			&ast.SelectorExpr{
				X:   &ast.Ident{Name: toInsertVariableName},
				Sel: &ast.Ident{Name: idFieldName},
			},
		},
		Rhs: []ast.Expr{
			&ast.Ident{Name: "out"},
		}}

	return []ast.Stmt{&insertInLib, &errorCheck, &idAssign}, nil
}

func getNewDestructorStmts(fn *ast.FuncDecl, typeName, idFieldName string) ([]ast.Stmt, error) {
	idVariableName, err := getFirstParamVariableName(fn.Type.Params)
	if err != nil {
		return nil, err
	}

	deleteLibCall := ast.AssignStmt{
		Tok: token.DEFINE,
		Lhs: []ast.Expr{
			&ast.Ident{Name: LibErrorVariableName},
		},
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.IndexExpr{
					Index: &ast.Ident{Name: typeName},
					X: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "lib"},
						Sel: &ast.Ident{Name: "Delete"},
					},
				},
				Args: []ast.Expr{
					&ast.Ident{Name: idVariableName},
				},
			},
		},
	}
	errorCheck := getErrorCheckExpr(fn, LibErrorVariableName)

	return []ast.Stmt{&deleteLibCall, &errorCheck}, nil
}

func getFirstParamVariableName(params *ast.FieldList) (string, error) {
	if params.List == nil || len(params.List) == 0 {
		return "", fmt.Errorf("object to be inserted not found in the parameters list")
	} else if len(params.List) > 1 {
		return "", fmt.Errorf("maximum allowed number of parameters is 1")
	}

	return params.List[0].Names[0].Name, nil
}

func getUpsertInLibExpr(fn *ast.FuncDecl, typesWithCustomId map[string]string) ast.AssignStmt {
	typeName := getFunctionReceiverTypeAsString(fn.Recv)
	idFieldName := ""
	if idField, isPresent := typesWithCustomId[typeName]; isPresent {
		idFieldName = idField
	} else {
		idFieldName = "Id"
	}

	return ast.AssignStmt{
		Tok: token.ASSIGN,
		Lhs: []ast.Expr{
			&ast.Ident{Name: LibErrorVariableName},
		},
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "lib"},
					Sel: &ast.Ident{Name: "Upsert"},
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

func getPointerAssignStmt(receiverName string) ast.AssignStmt {
	return ast.AssignStmt{
		Tok: token.ASSIGN,
		Lhs: []ast.Expr{
			&ast.Ident{Name: receiverName},
		},
		Rhs: []ast.Expr{
			&ast.StarExpr{
				X: &ast.Ident{Name: TemporaryReceiverName},
			},
		},
	}
}

func getReadFromLibExpr(fn *ast.FuncDecl, typesWithCustomId map[string]string) (ast.AssignStmt, bool) {
	typeName := types.ExprString(fn.Recv.List[0].Type)
	isPointerReceiver := strings.Contains(typeName, "*")
	typeName = strings.TrimPrefix(typeName, "*")

	idFieldName := ""
	if idField, isPresent := typesWithCustomId[typeName]; isPresent {
		idFieldName = idField
	} else {
		idFieldName = "Id"
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

	if isPointerReceiver {
		assignStmt.Lhs = []ast.Expr{
			&ast.Ident{Name: fn.Recv.List[0].Names[0].Name},
			&ast.Ident{Name: LibErrorVariableName},
		}
	} else {
		assignStmt.Lhs = []ast.Expr{
			&ast.Ident{Name: TemporaryReceiverName},
			&ast.Ident{Name: LibErrorVariableName},
		}
	}

	return assignStmt, isPointerReceiver
}

type getDBStmtsParam struct {
	idFieldName          string
	typeName             string
	receiverVariableName string
	fieldName            string
	fieldType            string
}

func getGetterDBStmts(fn *ast.FuncDecl, input getDBStmtsParam) []ast.Stmt {
	getFieldFromLib := ast.AssignStmt{
		Tok: token.DEFINE,
		Lhs: []ast.Expr{
			&ast.Ident{Name: "fieldValue"},
			&ast.Ident{Name: LibErrorVariableName},
		},
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "lib"},
					Sel: &ast.Ident{Name: "GetField"},
				},
				Args: []ast.Expr{
					&ast.SelectorExpr{
						X:   &ast.Ident{Name: input.receiverVariableName},
						Sel: &ast.Ident{Name: input.idFieldName},
					},
					&ast.CompositeLit{
						Type: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "lib"},
							Sel: &ast.Ident{Name: "GetFieldParam"},
						},
						Elts: []ast.Expr{
							&ast.KeyValueExpr{
								Key: &ast.Ident{Name: "TypeName"},
								Value: &ast.BasicLit{
									Kind:  token.STRING,
									Value: "\"" + input.typeName + "\"",
								},
							},
							&ast.KeyValueExpr{
								Key: &ast.Ident{Name: "FieldName"},
								Value: &ast.BasicLit{
									Kind:  token.STRING,
									Value: "\"" + input.fieldName + "\"",
								},
							},
						},
					},
				},
			}}}
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
			&ast.TypeAssertExpr{
				Type: &ast.Ident{Name: input.fieldType},
				X:    &ast.Ident{Name: "fieldValue"},
			},
		}}

	return []ast.Stmt{&getFieldFromLib, &errorCheck, &fieldAssign}
}

func getSetterDBStmts(fn *ast.FuncDecl, input getDBStmtsParam) []ast.Stmt {
	getFieldFromLib := ast.AssignStmt{
		Tok: token.DEFINE,
		Lhs: []ast.Expr{
			&ast.Ident{Name: LibErrorVariableName},
		},
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "lib"},
					Sel: &ast.Ident{Name: "SetField"},
				},
				Args: []ast.Expr{
					&ast.SelectorExpr{
						X:   &ast.Ident{Name: input.receiverVariableName},
						Sel: &ast.Ident{Name: input.idFieldName},
					},
					&ast.CompositeLit{
						Type: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "lib"},
							Sel: &ast.Ident{Name: "SetFieldParam"},
						},
						Elts: []ast.Expr{
							&ast.KeyValueExpr{
								Key: &ast.Ident{Name: "TypeName"},
								Value: &ast.BasicLit{
									Kind:  token.STRING,
									Value: "\"" + input.typeName + "\"",
								},
							},
							&ast.KeyValueExpr{
								Key: &ast.Ident{Name: "FieldName"},
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

	return []ast.Stmt{&getFieldFromLib, &errorCheck}
}

func getInitFunctionForType(typeName, idFieldName string, oneToMany []NavigationToField, manyToMany []ManyToManyRelationshipField) *ast.FuncDecl {
	receiverName := "receiver"
	function := &ast.FuncDecl{
		Name: &ast.Ident{Name: InitFunctionName},
		Recv: &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{{Name: receiverName}},
					Type:  &ast.StarExpr{X: &ast.Ident{Name: typeName}},
				},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.AssignStmt{
					Tok: token.ASSIGN,
					Lhs: []ast.Expr{
						&ast.SelectorExpr{
							X:   &ast.Ident{Name: receiverName},
							Sel: &ast.Ident{Name: IsInitializedFieldName},
						},
					},
					Rhs: []ast.Expr{
						&ast.Ident{Name: "true"},
					},
				},
			},
		},
		Type: &ast.FuncType{Params: &ast.FieldList{}},
	}

	for _, oneToManyRel := range oneToMany {
		initOneToManyRef := &ast.AssignStmt{
			Tok: token.ASSIGN,
			Lhs: []ast.Expr{
				&ast.SelectorExpr{
					X:   &ast.Ident{Name: receiverName},
					Sel: &ast.Ident{Name: oneToManyRel.FromFieldName},
				},
			},
			Rhs: []ast.Expr{
				&ast.StarExpr{X: &ast.CallExpr{
					Args: []ast.Expr{
						&ast.SelectorExpr{
							X:   &ast.Ident{Name: receiverName},
							Sel: &ast.Ident{Name: idFieldName},
						},
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   &ast.Ident{Name: receiverName},
								Sel: &ast.Ident{Name: NobjectImplementationMethod},
							},
						},
						&ast.BasicLit{
							Kind:  token.STRING,
							Value: "\"" + oneToManyRel.FieldName + "\"",
						},
						&ast.Ident{Name: "false"},
					},
					Fun: &ast.IndexExpr{
						Index: &ast.Ident{Name: oneToManyRel.TypeName},
						X: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "lib"},
							Sel: &ast.Ident{Name: ReferenceNavigationListCtor},
						},
					},
				}},
			},
		}

		function.Body.List = append(function.Body.List, initOneToManyRef)
	}

	for _, manyToManyRel := range manyToMany {
		// PartionKeyName and SortKeyName define the two types
		// used in many-to-many relationship
		// here, the type used in field declaration is different
		// the type different from the owner type
		relationshipType := manyToManyRel.PartionKeyName
		if typeName == manyToManyRel.PartionKeyName {
			relationshipType = manyToManyRel.SortKeyName
		}

		initManyToManyRef := &ast.AssignStmt{
			Tok: token.ASSIGN,
			Lhs: []ast.Expr{
				&ast.SelectorExpr{
					X:   &ast.Ident{Name: receiverName},
					Sel: &ast.Ident{Name: manyToManyRel.FromFieldName},
				},
			},
			Rhs: []ast.Expr{
				&ast.StarExpr{X: &ast.CallExpr{
					Args: []ast.Expr{
						&ast.SelectorExpr{
							X:   &ast.Ident{Name: receiverName},
							Sel: &ast.Ident{Name: idFieldName},
						},
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   &ast.Ident{Name: receiverName},
								Sel: &ast.Ident{Name: NobjectImplementationMethod},
							},
						},
						&ast.BasicLit{
							Kind:  token.STRING,
							Value: "\"\"",
						},
						&ast.Ident{Name: "true"},
					},
					Fun: &ast.IndexExpr{
						Index: &ast.Ident{Name: relationshipType},
						X: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "lib"},
							Sel: &ast.Ident{Name: ReferenceNavigationListCtor},
						},
					},
				}},
			},
		}

		function.Body.List = append(function.Body.List, initManyToManyRef)
	}

	return function
}