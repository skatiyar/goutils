package goutils_test

import (
	"fmt"
	"sort"

	"github.com/skatiyar/goutils"
)

type student struct {
	Name  string
	Grade int
}

var exampleData map[int]student = map[int]student{
	1: {Name: "Quincy", Grade: 96},
	2: {Name: "Jason", Grade: 84},
	3: {Name: "Alexis", Grade: 100},
	4: {Name: "Sam", Grade: 65},
	5: {Name: "Katie", Grade: 90},
}

// A whole file example is a file that ends in `_test.go` and contains exactly one
// example function, no test or benchmark functions, and at least one other
// package-level declaration.
func Example_grades() {
	passedStudents, passedStudentsErr := goutils.FilterMap(exampleData, func(key int, student student) (bool, error) {
		return student.Grade > 40, nil
	})
	if passedStudentsErr != nil {
		fmt.Println(passedStudentsErr)
		return
	}

	grades, gradesErr := goutils.Map(passedStudents, func(key int, student student) (int, string, error) {
		switch {
		case student.Grade >= 80:
			return key, "A", nil
		case student.Grade >= 60:
			return key, "B", nil
		default:
			return key, "C", nil
		}
	})
	if gradesErr != nil {
		fmt.Println(gradesErr)
		return
	}

	gradesGroup, gradesGroupErr := goutils.GroupByMap(grades, func(key int, grade string) (string, int, error) {
		return grade, key, nil
	})
	if gradesGroupErr != nil {
		fmt.Println(gradesGroupErr)
		return
	}

	for key, val := range gradesGroup {
		sort.Ints(val)
		fmt.Println("Students grouped by grades", key, val)
	}
	// Unordered output:
	// Students grouped by grades A [1 2 3 5]
	// Students grouped by grades B [4]
}
