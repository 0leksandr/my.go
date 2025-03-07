package my

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

//const indentStr = "    "
//const indentLen = len(indentStr)
const indentStr = "\t"
const indentLen = 4
const lineLen = 80

type formatter struct {
	addresses map[uintptr]struct{}
}

func (formatter) New() formatter {
	return formatter{make(map[uintptr]struct{})}
}
func (formatter formatter) Format(value any) string {
	return formatter.format(value, 0)
}
func (formatter formatter) format(value any, indent int) string {
	if indent > 300 { panic("fuck") }
	if value == nil { return "nil" }
	rv := reflect.ValueOf(value)
	rt := rv.Type()
	rk := rt.Kind()
	indentFn := func(s string, firstLineIndent bool) string {
		var firstIndent string
		if firstLineIndent { firstIndent += indentStr }
		return firstIndent + regexp.MustCompile("\n").ReplaceAllString(s, "\n" + indentStr)
	}
	switch rk {
		case reflect.Chan,
		reflect.Func,
		reflect.Interface,
		reflect.Map,
		reflect.Pointer,
		reflect.Slice:
			if rv.IsNil() {
				return fmt.Sprintf("(%s)(nil)", formatter.formatType(rt))
			}
		default:
	}
	switch rk {
		case reflect.Struct:
			str := formatter.formatType(rt)
			if rv.NumField() == 0 {
				str += "{}"
			} else {
				str += "{\n"
				getter := ReflectStructGetter{}.New(value)
				for i := 0; i < getter.NumField(); i++ {
					str += indentFn(
						fmt.Sprintf(
							"%s: %s,",
							rt.Field(i).Name,
							formatter.format(getter.Get(i), indent + indentLen),
						),
						true,
					) + "\n"
				}
				str += "}"
			}
			return str
		case reflect.Pointer:
			addr := rv.Pointer()
			if _, ok := formatter.addresses[addr]; ok {
				return fmt.Sprintf("[circular reference %#x]", addr)
			} else {
				formatter.addresses[addr] = struct{}{}
				return "&" + formatter.format(rv.Elem().Interface(), indent)
			}
		case reflect.Bool:
			if rv.Bool() {
				return "true"
			} else {
				return "false"
			}
		case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64:
			return strconv.Itoa(int(rv.Int()))
		case reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64:
			return strconv.FormatUint(rv.Uint(), 10)
		case reflect.Float32,
		reflect.Float64:
			res := strconv.FormatFloat(rv.Float(), 'g', -1, 64)
			if regexp.MustCompile(`^\d+$`).MatchString(res) { res += "." }
			return res
		case reflect.Array,
		reflect.Slice:
			_type := "["
			if rk == reflect.Array { _type += strconv.Itoa(rv.Len()) }
			_type += "]" + formatter.formatType(rv.Type().Elem())

			if bytes, ok := value.([]byte); ok {
				return fmt.Sprintf("[]byte(%s)", formatter.format(string(bytes), indent))
			}

			str := ""
			str += _type + "{"
			l := rv.Len()
			if l > 0 { str += "\n" }
			for i := 0; i < l; i++ {
				str += indentFn(formatter.format(rv.Index(i).Interface(), indent + indentLen), true) + ",\n"
			}
			str += "}"
			return str
		case reflect.Chan:
			return fmt.Sprintf("(chan %s)(nil)", formatter.formatType(rt.Elem()))
		case reflect.Func:
			var args []string
			for i := 0; i < rv.Type().NumIn(); i++ {
				args = append(args, formatter.formatType(rv.Type().In(i)))
			}
			var ret []string
			for i := 0; i < rv.Type().NumOut(); i++ {
				ret = append(ret, "_ " + formatter.formatType(rv.Type().Out(i)))
			}
			return fmt.Sprintf(
				"func(%s) (%s) { return }",
				strings.Join(args, ", "),
				strings.Join(ret, ", "),
			)
		case reflect.Interface:
			return formatter.format(rv.Elem().Interface(), indent + indentLen)
		case reflect.Map:
			type key struct {
				key   any
				value   reflect.Value
				_string string
				_float  float64
				isFloat bool
			}
			var keys []key
			var toFloat func(v reflect.Value) (float64, bool)
			toFloat = func(v reflect.Value) (float64, bool) {
				switch v.Type().Kind() {
					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
						return float64(v.Int()), true
					case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
						return float64(v.Uint()), true
					case reflect.Float32, reflect.Float64:
						return v.Float(), true
					case reflect.Interface:
						return toFloat(v.Elem())
					default:
						return 0, false
				}
			}
			for _, _key := range rv.MapKeys() {
				_float, isFloat := toFloat(_key)
				keys = append(keys, key{
					key:     _key.Interface(),
					value:   _key,
					_string: formatter.format(_key.Interface(), indent + indentLen),
					_float:  _float,
					isFloat: isFloat,
				})
			}
			sort.Slice(keys, func(i, j int) bool {
				if keys[i].isFloat && keys[j].isFloat {
					return keys[i]._float < keys[j]._float
				} else if keys[i].isFloat {
					return true
				} else if keys[j].isFloat {
					return false
				} else {
					return keys[i]._string < keys[j]._string
				}
			})
			l := rv.Len()
			str := ""
			str += fmt.Sprintf("map[%s]%s", formatter.formatType(rt.Key()), formatter.formatType(rt.Elem())) + "{"
			if l > 0 { str += "\n" }
			for _, _key := range keys {
				str += indentFn(
					fmt.Sprintf(
						"%s: %s,",
						formatter.format(_key.key, indent + indentLen),
						formatter.format(rv.MapIndex(_key.value).Interface(), indent + indentLen),
					),
					true,
				) + "\n"
			}
			str += "}"
			return str
		case reflect.String:
			str := rv.String()
			availableLength := lineLen - indent - 2
			isMultiline := strings.ContainsAny(str, "\n")
			if !isMultiline {
				if availableLength < len(str) {
					isMultiline = true
				}
			}
			escapeWrap := func(symbol string) {
				for _, symbol2 := range []string{"\\", symbol} {
					str = strings.ReplaceAll(str, symbol2, "\\" + symbol2)
				}
				str = symbol + str + symbol
			}
			if isMultiline {
				lines := strings.Split(str, "\n")
				joinLists := func(lists... []string) []string {
					var joined []string
					for _, list := range lists {
						joined = append(joined, list...)
					}
					return joined
				}
				for i := 0; i < len(lines); i++ {
					line := lines[i]
					if len(line) > availableLength {
						lines = joinLists(
							lines[:i],
							[]string{line[:availableLength], line[availableLength:]},
							lines[i+1:],
						)
					}
				}
				str = strings.Join(lines, "\n")
				escapeWrap("`")
			} else {
				escapeWrap("\"")
			}
			return str
		case reflect.Uintptr:
			return formatter.formatUintptr(value.(uintptr))

		case reflect.Invalid,
		reflect.UnsafePointer,
		reflect.Complex64,
		reflect.Complex128:
			panic(formatter.formatType(rt))
		default: panic("unreachable: " + rk.String())
	}
}
func (formatter) formatType(t reflect.Type) string {
	return t.String()
}
func (formatter) formatUintptr(v uintptr) string {
	return fmt.Sprintf("%#x", v)
}
