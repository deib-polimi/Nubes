package parser

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"strings"
)

type StateChangingHandler struct {
	OrginalPackage      string
	OrginalPackageAlias string
	Imports             string
	MethodName          string
	ReceiverType        string
	ReceiverIdFieldName string
	OptionalReturnType  string
	Invocation          string
}

type detectedFunction struct {
	Function *ast.FuncDecl
	Imports  []*ast.ImportSpec
}

func (t *TypeSpecParser) prepareHandleres() {
	fileFunctionsMap := t.detectedFunctions

	handlerFuncs := []StateChangingHandler{}
	for _, functions := range fileFunctionsMap {
		for _, detectedFunction := range functions {
			f := detectedFunction.Function

			if f.Name.Name == NobjectImplementationMethod || f.Name.Name == CustomIdImplementationMethod || strings.HasPrefix(f.Name.Name, SetPrefix) || strings.HasPrefix(f.Name.Name, GetPrefix) {
				continue
			} else if f.Recv == nil {

				// TODO: detect & generate handler(s) with CTORs

			} else {
				receiverTypeName := getFunctionReceiverTypeAsString(f.Recv)
				if isNobject := t.Output.IsNobjectInOrginalPackage[receiverTypeName]; !isNobject {
					fmt.Println("Member type does not implement Nobject interface. Handler generation for " + f.Name.Name + "skipped")
					continue
				}

				newHandler := StateChangingHandler{
					OrginalPackage:      t.Output.ImportPath,
					OrginalPackageAlias: OrginalPackageAlias,
					MethodName:          f.Name.Name,
					ReceiverType:        receiverTypeName,
					ReceiverIdFieldName: "Id",
					Imports:             getImportsAsString(t.tokenSet, detectedFunction.Imports),
				}

				if customIdFieldName, hasCustomId := t.Output.TypesWithCustomId[receiverTypeName]; hasCustomId {
					newHandler.ReceiverIdFieldName = customIdFieldName
				}

				if retParamsVerifier.Check(f) {
					if len(f.Type.Results.List) > 1 {
						newHandler.OptionalReturnType = types.ExprString(f.Type.Results.List[0].Type)
						if _, isPresent := t.Output.IsNobjectInOrginalPackage[newHandler.OptionalReturnType]; isPresent {
							newHandler.OptionalReturnType = newHandler.OrginalPackageAlias + "." + newHandler.OptionalReturnType
						}
					}
				}

				parameters, err := getStateChangingFuncParams(f.Type.Params, t.Output.IsNobjectInOrginalPackage)
				if err != nil {
					fmt.Println("Maximum allowed number of parameters is 1. Handler generation for " + f.Name.Name + "skipped")
					continue
				}
				newHandler.Invocation = f.Name.Name + "(" + parameters + ")"
				t.Handlers = append(handlerFuncs, newHandler)
			}
		}
	}

}

func getImportsAsString(fset *token.FileSet, imports []*ast.ImportSpec) string {
	var buf bytes.Buffer
	for _, imp := range imports {
		err := printer.Fprint(&buf, fset, imp)
		buf.WriteString("\n")
		if err != nil {
			fmt.Println(err)
		}
	}

	return buf.String()
}

func getStateChangingFuncParams(params *ast.FieldList, isNobjectInOrgPkg map[string]bool) (string, error) {
	if params.List == nil || len(params.List) == 0 {
		return "", nil
	} else if len(params.List) > 1 {
		return "", fmt.Errorf("maximum allowed number of parameters is 1")
	}

	inputParamType := types.ExprString(params.List[0].Type)
	if _, isPresent := isNobjectInOrgPkg[inputParamType]; isPresent {
		inputParamType = OrginalPackageAlias + "." + inputParamType
	}
	return HandlerInputParameterName + "." + HandlerInputParameterFieldName + ".(" + inputParamType + ")", nil
}