package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

//converts string of shape "dd/MM/yy hh:mm:ss" into a time.Time object.
func StringDateToTime(date string) time.Time {
	dat := strings.Split(date, " ")
	dmy := strings.Split(dat[0], "/")
	hms := strings.Split(dat[1], ":")

	return time.Date(AtoiWrap(dmy[2]), time.Month(AtoiWrap(dmy[1])), AtoiWrap(dmy[0]),
		AtoiWrap(hms[0]), AtoiWrap(hms[1]), AtoiWrap(hms[2]), 0, time.UTC)
}

// wrapper for strconv.Atoi, handles errors via panic
func AtoiWrap(s string) int {
	i, err := strconv.Atoi(s)
	errCheck(err)
	return i
}

// difference returns the elements in `a` that aren't in `b`.
func difference(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

// checks if a certain time (check) is in a certain timeframe between [start, end]
func inTimeSpan(start, end, check time.Time) bool {
	if start.Before(end) {
		return !check.Before(start) && !check.After(end)
	}

	if start.Equal(end) {
		return check.Equal(start)
	}

	return !start.After(check) || !end.Before(check)
}

// delete element from array at index i, waiting for new beta features introducig generics to do this function in a non-disgusting way.
func deleteAtIndex(a interface{}) {
	fmt.Print("hi :)")
}

// pass any error to throw a panic if you do not want to handle the error.
func errCheck(e error) {
	if e != nil {
		panic(e)
	}
}
