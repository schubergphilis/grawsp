package errors

import "fmt"

type CacheMissError struct {
	ObjectName string
}

func (e *CacheMissError) Error() string {
	return fmt.Sprintf("Object %s missed in cache\n", e.ObjectName)
}
