package control

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	ErrorInvalidArgument     = "invalid argument at index %d: expected a function"
	ErrorInvalidReturnCount  = "invalid function at index %d: must return at least 2 values (structs..., error)"
	ErrorInvalidReturnType   = "invalid function at index %d: return value %d must be a struct, got %s"
	ErrorInvalidReturnError  = "invalid function at index %d: last return value must be error"
	ErrorInvalidInputType    = "invalid function at index %d: input parameter %d must be a struct, got %s"
	ErrorDuplicateOutputType = "duplicate output type %s: produced by functions at index %d and %d"
	ErrorCyclicDependency    = "cyclic dependency detected involving function at index %d"
	ErrorNilFunction         = "nil function at index %d"
	ErrorNoResultProducer    = "no function produces the result type %s"
	ErrorSeedTypeMismatch    = "Run: argument at index %d has type %s, which is not a struct"
	ErrorDuplicateSeedType   = "Run: duplicate seed type %s at index %d"
	ErrorMissingDependency   = "Run: function %s requires type %s, not produced by any function or seed"
	ErrorSeedConflict        = "Run: seed type %s conflicts with function output at index %d"
)

var errorInterface = reflect.TypeOf((*error)(nil)).Elem()

type autoNode struct {
	name           string
	index          int
	fn             reflect.Value
	fnType         reflect.Type
	inputs         []reflect.Type
	outputs        []reflect.Type
	deps           []int
	dependents     []int
	externalInputs []reflect.Type
}

type autoResult struct {
	nodeIndex int
	values    []reflect.Value
	err       error
	duration  time.Duration
}

// Auto is a generic DAG-based parallel function executor. It takes a set of functions,
// infers dependencies between them by matching struct input/output types, and executes
// them concurrently wherever the dependency graph allows.
//
// The type parameter T is the final result type. Exactly one function in the DAG must
// produce T as one of its return values.
//
// Each function must follow the signature: func(structs...) (structs..., error)
//   - Inputs: 0 or more struct parameters, representing dependencies on other functions' outputs.
//   - Outputs: 1 or more struct return values, plus an error as the last return value.
//
// Dependencies are resolved automatically: if function B takes TypeX as a parameter and
// function A returns TypeX, then B depends on A. Functions with no unsatisfied dependencies
// run in parallel.
//
// Example:
//
//	funcA := func() (TypeA, error) { ... }                        // root — no deps
//	funcB := func(a TypeA) (TypeB, TypeC, error) { ... }          // depends on A, produces B and C
//	funcC := func(b TypeB, c TypeC) (TypeD, error) { ... }        // depends on B and C
//
//	auto, err := NewAuto[TypeD](funcA, funcB, funcC)
//	result, err := auto.Run()  // executes A, then B, then C — returns TypeD
type Auto[T any] struct {
	nodes       []autoNode
	typeToNode  map[reflect.Type]int
	sortedOrder []int
	resultType  reflect.Type
	resultNode  int
}

// AutoMetrics contains timing information from an Auto execution.
// Duration is the total wall-clock time for the entire DAG execution.
// FuncMetrics maps each function's name to the time it took to execute.
type AutoMetrics struct {
	Duration    time.Duration
	FuncMetrics map[string]time.Duration
}

// NewAuto creates a new Auto instance by validating the provided functions, building a
// dependency graph, and computing a topological execution order.
//
// Each function must:
//   - Accept 0 or more struct parameters (dependencies).
//   - Return 1 or more struct values followed by an error as the last return value.
//
// NewAuto returns an error if:
//   - Any argument is nil or not a function.
//   - A function's inputs or outputs are not structs (except the final error return).
//   - Two functions produce the same struct type.
//   - The dependency graph contains a cycle.
//   - No function produces the result type T.
//
// Dependencies not produced by any function are treated as external and must be
// provided as seed values when calling Run or RunWithMetrics.
func NewAuto[T any](fns ...any) (Auto[T], error) {
	var zero Auto[T]

	if len(fns) == 0 {
		resultType := reflect.TypeOf((*T)(nil)).Elem()
		return zero, fmt.Errorf(ErrorNoResultProducer, resultType)
	}

	// Phase 1: Validate each function signature and build nodes
	nodes := make([]autoNode, 0, len(fns))
	for idx := range fns {
		fn := fns[idx]
		if fn == nil {
			return zero, fmt.Errorf(ErrorNilFunction, idx)
		}
		argType := reflect.TypeOf(fn)
		if argType.Kind() != reflect.Func {
			return zero, fmt.Errorf(ErrorInvalidArgument, idx)
		}
		if argType.NumOut() < 2 {
			return zero, fmt.Errorf(ErrorInvalidReturnCount, idx)
		}
		if !argType.Out(argType.NumOut() - 1).Implements(errorInterface) {
			return zero, fmt.Errorf(ErrorInvalidReturnError, idx)
		}
		outputs := make([]reflect.Type, 0, argType.NumOut()-1)
		for i := 0; i < argType.NumOut()-1; i++ {
			if argType.Out(i).Kind() != reflect.Struct {
				return zero, fmt.Errorf(ErrorInvalidReturnType, idx, i, argType.Out(i).Kind())
			}
			outputs = append(outputs, argType.Out(i))
		}
		inputs := make([]reflect.Type, 0, argType.NumIn())
		for i := 0; i < argType.NumIn(); i++ {
			if argType.In(i).Kind() != reflect.Struct {
				return zero, fmt.Errorf(ErrorInvalidInputType, idx, i, argType.In(i).Kind())
			}
			inputs = append(inputs, argType.In(i))
		}
		fnName := fmt.Sprintf("%d_%s", idx, runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name())
		nodes = append(nodes, autoNode{
			name:    fnName,
			index:   idx,
			fn:      reflect.ValueOf(fn),
			fnType:  argType,
			inputs:  inputs,
			outputs: outputs,
		})
	}

	// Phase 2: Build typeToNode map
	typeToNode := make(map[reflect.Type]int)
	for i, node := range nodes {
		for _, outType := range node.outputs {
			if existing, ok := typeToNode[outType]; ok {
				return zero, fmt.Errorf(ErrorDuplicateOutputType, outType, existing, i)
			}
			typeToNode[outType] = i
		}
	}

	// Phase 3: Validate result type T
	resultType := reflect.TypeOf((*T)(nil)).Elem()
	resultNode, ok := typeToNode[resultType]
	if !ok {
		return zero, fmt.Errorf(ErrorNoResultProducer, resultType)
	}

	// Phase 4: Resolve dependency edges
	for i := range nodes {
		for _, inType := range nodes[i].inputs {
			if producerIdx, ok := typeToNode[inType]; ok {
				nodes[i].deps = append(nodes[i].deps, producerIdx)
				nodes[producerIdx].dependents = append(nodes[producerIdx].dependents, i)
			} else {
				nodes[i].externalInputs = append(nodes[i].externalInputs, inType)
			}
		}
	}

	// Phase 5: Cycle detection + topological sort (Kahn's algorithm)
	inDegree := make([]int, len(nodes))
	for i := range nodes {
		inDegree[i] = len(nodes[i].deps)
	}
	queue := make([]int, 0)
	for i, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, i)
		}
	}
	sortedOrder := make([]int, 0, len(nodes))
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		sortedOrder = append(sortedOrder, n)
		for _, dep := range nodes[n].dependents {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				queue = append(queue, dep)
			}
		}
	}
	if len(sortedOrder) != len(nodes) {
		for i := range nodes {
			if inDegree[i] > 0 {
				return zero, fmt.Errorf(ErrorCyclicDependency, i)
			}
		}
	}

	return Auto[T]{
		nodes:       nodes,
		typeToNode:  typeToNode,
		sortedOrder: sortedOrder,
		resultType:  resultType,
		resultNode:  resultNode,
	}, nil
}

// Run executes the DAG and returns the result of type T.
// Optional seed args can be provided as struct values to satisfy external dependencies
// (inputs not produced by any function in the graph).
//
// Functions are executed in parallel wherever the dependency graph allows. If any function
// returns an error or panics, execution stops immediately and the error is returned.
func (a Auto[T]) Run(args ...any) (T, error) {
	result, _, err := a.execute(args, false)
	return result, err
}

// RunWithMetrics executes the DAG like Run, but additionally returns AutoMetrics
// containing the total execution duration and per-function timing.
func (a Auto[T]) RunWithMetrics(args ...any) (T, AutoMetrics, error) {
	return a.execute(args, true)
}

// Describe returns a human-readable description of the dependency graph in topological order.
// Each entry describes a function, its dependencies, and its output types.
func (a Auto[T]) Describe() []string {
	if len(a.nodes) == 0 {
		return []string{}
	}
	descriptions := make([]string, 0, len(a.nodes))
	for _, idx := range a.sortedOrder {
		node := a.nodes[idx]
		outNames := make([]string, 0, len(node.outputs))
		for _, outType := range node.outputs {
			outNames = append(outNames, outType.String())
		}
		if len(node.deps) == 0 && len(node.externalInputs) == 0 {
			descriptions = append(descriptions, fmt.Sprintf("%s() -> (%s)", node.name, strings.Join(outNames, ", ")))
		} else {
			depNames := make([]string, 0)
			for _, depIdx := range node.deps {
				depNames = append(depNames, a.nodes[depIdx].name)
			}
			for _, extType := range node.externalInputs {
				depNames = append(depNames, fmt.Sprintf("seed(%s)", extType))
			}
			descriptions = append(descriptions, fmt.Sprintf(
				"%s(%s) -> (%s)",
				node.name,
				strings.Join(depNames, ", "),
				strings.Join(outNames, ", "),
			))
		}
	}
	return descriptions
}

func (a Auto[T]) execute(args []any, trackMetrics bool) (T, AutoMetrics, error) {
	var zero T
	metrics := AutoMetrics{}

	if len(a.nodes) == 0 {
		return zero, metrics, nil
	}

	startTime := time.Now()

	// Validate seed args
	seedValues := make(map[reflect.Type]reflect.Value)
	for i, arg := range args {
		if arg == nil {
			return zero, metrics, fmt.Errorf(ErrorSeedTypeMismatch, i, "nil")
		}
		argType := reflect.TypeOf(arg)
		if argType.Kind() != reflect.Struct {
			return zero, metrics, fmt.Errorf(ErrorSeedTypeMismatch, i, argType)
		}
		if _, exists := seedValues[argType]; exists {
			return zero, metrics, fmt.Errorf(ErrorDuplicateSeedType, argType, i)
		}
		if nodeIdx, conflicts := a.typeToNode[argType]; conflicts {
			return zero, metrics, fmt.Errorf(ErrorSeedConflict, argType, nodeIdx)
		}
		seedValues[argType] = reflect.ValueOf(arg)
	}

	// Check completeness: every external input must be satisfied by seeds
	for _, node := range a.nodes {
		for _, extType := range node.externalInputs {
			if _, ok := seedValues[extType]; !ok {
				return zero, metrics, fmt.Errorf(ErrorMissingDependency, node.name, extType)
			}
		}
	}

	// Initialize values map and remaining deps count
	var mu sync.Mutex
	values := make(map[reflect.Type]reflect.Value)
	for t, v := range seedValues {
		values[t] = v
	}

	remaining := make(map[int]int, len(a.nodes))
	for i, node := range a.nodes {
		count := 0
		for _, depIdx := range node.deps {
			// Only count deps whose outputs aren't already in values (from seeds)
			satisfied := true
			for _, outType := range a.nodes[depIdx].outputs {
				if _, ok := values[outType]; !ok {
					satisfied = false
					break
				}
			}
			if !satisfied {
				count++
			}
		}
		remaining[i] = count
	}

	resultChan := make(chan autoResult, len(a.nodes))
	stop := make(chan struct{})
	var stopOnce sync.Once
	var firstErr error
	var errOnce sync.Once
	fnMetrics := make(map[string]time.Duration)

	launch := func(nodeIdx int) {
		node := a.nodes[nodeIdx]
		go func() {
			defer func() {
				if r := recover(); r != nil {
					var err error
					if rec, ok := r.(error); ok {
						err = rec
					} else {
						err = fmt.Errorf("panic in function %s: %v", node.name, r)
					}
					resultChan <- autoResult{nodeIndex: nodeIdx, err: err}
				}
			}()
			select {
			case <-stop:
				resultChan <- autoResult{nodeIndex: nodeIdx, err: fmt.Errorf("execution cancelled")}
				return
			default:
			}
			mu.Lock()
			inputArgs := make([]reflect.Value, len(node.inputs))
			for i, inType := range node.inputs {
				inputArgs[i] = values[inType]
			}
			mu.Unlock()

			var nodeStart time.Time
			if trackMetrics {
				nodeStart = time.Now()
			}
			results := node.fn.Call(inputArgs)
			var dur time.Duration
			if trackMetrics {
				dur = time.Since(nodeStart)
			}

			outValues := results[:len(results)-1]
			errVal := results[len(results)-1]
			var err error
			if !errVal.IsNil() {
				err = errVal.Interface().(error)
			}
			resultChan <- autoResult{
				nodeIndex: nodeIdx,
				values:    outValues,
				err:       err,
				duration:  dur,
			}
		}()
	}

	// Launch initially ready nodes
	active := 0
	for i, rem := range remaining {
		if rem == 0 {
			launch(i)
			active++
		}
	}

	completed := 0
	for completed < len(a.nodes) && active > 0 {
		res := <-resultChan
		active--
		completed++

		if res.err != nil {
			errOnce.Do(func() {
				firstErr = res.err
				stopOnce.Do(func() { close(stop) })
			})
			// Drain remaining active goroutines
			for active > 0 {
				<-resultChan
				active--
				completed++
			}
			break
		}

		if trackMetrics {
			fnMetrics[a.nodes[res.nodeIndex].name] = res.duration
		}

		// Store output values
		node := a.nodes[res.nodeIndex]
		mu.Lock()
		for i, outType := range node.outputs {
			values[outType] = res.values[i]
		}
		mu.Unlock()

		// Launch newly ready dependents
		for _, depIdx := range node.dependents {
			mu.Lock()
			remaining[depIdx]--
			ready := remaining[depIdx] == 0
			mu.Unlock()
			if ready {
				launch(depIdx)
				active++
			}
		}
	}

	if firstErr != nil {
		return zero, metrics, firstErr
	}

	if trackMetrics {
		metrics = AutoMetrics{
			Duration:    time.Since(startTime),
			FuncMetrics: fnMetrics,
		}
	}

	return values[a.resultType].Interface().(T), metrics, nil
}
