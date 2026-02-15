package loglint

import (
	"go/ast"
	"go/token"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis"
)

var allowedMethods = map[string]bool{
	"Info":    true,
	"Error":   true,
	"Warn":    true,
	"Debug":   true,
	"Print":   true,
	"Printf":  true,
	"Println": true,
}

var allowedLibs = map[string]bool{
	"log":  true,
	"slog": true,
	"zap":  true,
}

var Analyzer = &analysis.Analyzer{
	Name: "loglint",                                         // имя линтера
	Doc:  "finds invalid args in log funs based on 4 rules", // описание линтера
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			return ins(n, pass)
		})
	}
	return nil, nil
}

func ins(n ast.Node, pass *analysis.Pass) bool {
	call, ok := n.(*ast.CallExpr)
	if !ok {
		return true
	}
	funcName := call.Fun
	selector, ok := funcName.(*ast.SelectorExpr)
	if !ok {
		return true
	}
	var needX any
	switch x := selector.X.(type) {
	case *ast.Ident:
		needX = x
	case *ast.CallExpr:
		subFunName := x.Fun
		subSelector, ok := subFunName.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		subX := subSelector.X
		subXIdent, ok := subX.(*ast.Ident)
		if !ok {
			return true
		}
		needX = subXIdent
	default:
		return true
	}
	x, ok := needX.(*ast.Ident)
	if !ok {
		return true
	}
	sel := selector.Sel
	if !allowedLibs[x.Name] {
		return true
	}
	if !allowedMethods[sel.Name] {
		return true
	}
	// дальнейшая логика
	args, ok := extractMessage(call)
	if !ok {
		return true
	}
	switch needArgs := args.(type) {
	case string:
		isLower := checkLowerCase(needArgs)
		if !isLower {
			pass.Reportf(call.Pos(), "log message must not contain upper case letters")
		}
		isEnglish := checkEnglish(needArgs)
		if !isEnglish {
			pass.Reportf(call.Pos(), "log message must consist only of English letters")
		}
		isNoSpecial := checkNoSpecialChars(needArgs)
		if !isNoSpecial {
			pass.Reportf(call.Pos(), "log message must not contain special symbols")
		}
	case []ast.Expr:
		basicLit, ok := needArgs[0].(*ast.BasicLit)
		if !ok {
			return true
		}
		if basicLit.Kind != token.STRING {
			return true
		}
		basicLitValue, err := strconv.Unquote(basicLit.Value)
		if err != nil {
			return true
		}
		isLower := checkLowerCase(basicLitValue)
		if !isLower {
			pass.Reportf(call.Pos(), "log message must not contain upper case letters")
		}
		isEnglish := checkEnglish(basicLitValue)
		if !isEnglish {
			pass.Reportf(call.Pos(), "log message must consist only of English letters")
		}
		isNoSpecial := checkNoSpecialChars(basicLitValue)
		if !isNoSpecial {
			pass.Reportf(call.Pos(), "log message must not contain special symbols")
		}
		isSensitive := checkSensitive(needArgs)
		if !isSensitive {
			pass.Reportf(call.Pos(), "log message must not contain important data")
		}
	default:
		return true
	}
	return true
}

/*func sendReport(pos token.Pos, msg string, pass *analysis.Pass, args ...any) {
	pass.Reportf(pos, msg)
}*/

// получения содержимого лога
func extractMessage(call *ast.CallExpr) (any, bool) {
	args := call.Args
	if len(args) == 0 {
		return "", false
	}
	if len(args) > 1 {
		args := extractArgs(call)
		return args, true
	}
	basicLit, ok := args[0].(*ast.BasicLit)
	if !ok {
		return "", false
	}
	if basicLit.Kind != token.STRING {
		return "", false
	}
	basicLitValue, err := strconv.Unquote(basicLit.Value)
	if err != nil {
		return "", false
	}
	return basicLitValue, true
}

// проверка на регистр
func checkLowerCase(msg string) bool {
	for _, letter := range msg {
		if unicode.IsUpper(letter) {
			return false
		}
	}
	return true
}

// проверка на язык
func checkEnglish(msg string) bool {
	isEnglishLetter := func(letter rune) bool {
		if letter >= 'a' && letter <= 'z' {
			return true
		}
		return false
	}
	for _, letter := range msg {
		if unicode.IsLetter(letter) {
			if !isEnglishLetter(letter) {
				return false
			}
		}
	}
	return true
}

// проверка на спец символы
func checkNoSpecialChars(msg string) bool {
	isAllowed := func(letter rune) bool {
		if letter >= 'a' && letter <= 'z' || letter >= '0' && letter <= '9' || letter == ' ' || letter == '_' || letter == '-' {
			return true
		}
		return false
	}
	for _, letter := range msg {
		if !isAllowed(letter) {
			return false
		}
	}
	return true
}

// проверка на важные данные
func checkSensitive(args []ast.Expr) bool {
	for _, arg := range args {
		switch needArg := arg.(type) {
		case *ast.BasicLit:
			if needArg.Kind != token.STRING {
				continue
			}
			needValueUnquoted, err := strconv.Unquote(needArg.Value)
			if err != nil {
				continue
			}
			if blackList[strings.ToLower(needValueUnquoted)] {
				return false
			}
		case *ast.Ident:
			if blackList[strings.ToLower(needArg.Name)] {
				return false
			}
		case *ast.SelectorExpr:
			if blackList[strings.ToLower(needArg.Sel.Name)] {
				return false
			}
		default:
			continue
		}
	}
	return true
}

var blackList = map[string]bool{
	// passwords
	"password":     true,
	"passwd":       true,
	"pwd":          true,
	"pass":         true,
	"userpassword": true,
	"dbpassword":   true,
	"rootpassword": true,

	// tokens / auth
	"token":         true,
	"accesstoken":   true,
	"refreshtoken":  true,
	"jwt":           true,
	"jwttoken":      true,
	"bearer":        true,
	"authorization": true,
	"auth":          true,
	"session":       true,
	"sessionid":     true,

	// keys
	"secret":       true,
	"secretkey":    true,
	"privatekey":   true,
	"apikey":       true,
	"api_key":      true,
	"accesskey":    true,
	"clientsecret": true,

	// payments
	"card":          true,
	"cardnumber":    true,
	"cvv":           true,
	"cvc":           true,
	"iban":          true,
	"accountnumber": true,

	// PII
	"email":       true,
	"phone":       true,
	"phonenumber": true,
	"passport":    true,
	"ssn":         true,
	"inn":         true,
	"snils":       true,
	"address":     true,

	// DB
	"dsn":              true,
	"connectionstring": true,
	"connstring":       true,
	"databaseurl":      true,
	"dburl":            true,
}

func extractArgs(call *ast.CallExpr) []ast.Expr {
	args := call.Args
	return args
}
