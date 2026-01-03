package work

// case *PipeExpr:
// 		val, err := e.eval(n.Base)
// 		if err != nil {
// 			return reflect.Value{}, err
// 		}
// 		for _, call := range n.Pipes {
// 			fn, ok := e.Pipes[call.Name]
// 			if !ok {
// 				return reflect.Value{}, fmt.Errorf("pipe not found: %s", call.Name)
// 			}
// 			fv := reflect.ValueOf(fn)
// 			if fv.Kind() != reflect.Func {
// 				return reflect.Value{}, fmt.Errorf("pipe %s is not func", call.Name)
// 			}
// 			args := []reflect.Value{}
// 			for _, a := range call.Args {
// 				av, err := e.eval(a)
// 				if err != nil {
// 					return reflect.Value{}, err
// 				}
// 				args = append(args, reflect.ValueOf(av))
// 			}
// 			args = append(args, reflect.ValueOf(val))
// 			out := fv.Call(args)
// 			if len(out) == 0 || !out[0].IsValid() {
// 				return reflect.Value{}, fmt.Errorf("pipe %s returned invalid value", call.Name)
// 			}
// 			val = out[0].Interface().(reflect.Value)
// 		}
// 		return val, nil
