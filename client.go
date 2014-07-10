package dao

import (
	"errors"
	"reflect"
)

type ClientCall struct {
	Receiver string
	Method   string
	Params   []interface{}
}

// {"receiver": "World", "method": "RegisterAccount", "params": ["wiwi", "wiwi"]}
// {"receiver": "World", "method": "LoginAccount", "params": ["wiwi", "wiwi"]}
// {"receiver": "Account", "method": "Logout", "params": []}
// {"receiver": "Account", "method": "CreateChar", "params": ["dodo"]}
// {"receiver": "Account", "method": "LoginChar", "params": [0]}
// {"receiver": "Char", "method": "Logout", "params": []}

func (c *ClientCall) CastJSON(f reflect.Value) ([]reflect.Value, error) {
	numIn := f.Type().NumIn()
	if len(c.Params) != numIn {
		return nil, errors.New("not match params length")
	}
	in := make([]reflect.Value, numIn)
	var ftype reflect.Type
	for i, param := range c.Params {
		ftype = f.Type().In(i)
		switch ftype.Kind() {
		case reflect.Int:
			switch param.(type) {
			case float64:
				in[i] = reflect.ValueOf(int(param.(float64)))
			default:
				return nil, errors.New("not match params type")
			}
		case reflect.String:
			switch param.(type) {
			case string:
				in[i] = reflect.ValueOf(param)
			default:
				return nil, errors.New("not match params type")
			}
		case reflect.Float32:
			switch param.(type) {
			case float64:
				in[i] = reflect.ValueOf(float32(param.(float64)))
			default:
				return nil, errors.New("not match params type")
			}
		case reflect.Float64:
			switch param.(type) {
			case float64:
				in[i] = reflect.ValueOf(param)
			default:
				return nil, errors.New("not match params type")
			}
		case reflect.Ptr:
			switch param.(type) {
			case (*wsConn):
				in[i] = reflect.ValueOf(param)
			default:
				return nil, errors.New("not match params type")
			}
		case reflect.Slice:
			switch ftype.String() {
			case "[]float64":
				switch param.(type) {
				case []float64:
					in[i] = reflect.ValueOf(param)
				default:
					return nil, errors.New("not match params type")
				}
			case "[]float32":
				switch v := param.(type) {
				case []float64:
					f32s := make([]float32, len(v))
					for i, f64 := range v {
						f32s[i] = float32(f64)
					}
					in[i] = reflect.ValueOf(f32s)
				default:
					return nil, errors.New("not match params type")
				}
			case "[]string":
				switch param.(type) {
				case []string:
					in[i] = reflect.ValueOf(param)
				default:
					return nil, errors.New("not match params type")
				}
			}
		}
	}
	return in, nil
}
