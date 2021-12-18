package my

import (
	"errors"
	"reflect"
	"testing"
)

func TestDump(t *testing.T) {
	Dump("hi")
}
func TestPanicIf(t *testing.T) {
	if false {
		PanicIf(errors.New("test"))
	}
	if false {
		Must(errors.New("test"))
	}
}
func TestRevert(t *testing.T) {
	slice := []int{1, 2, 3, 4}
	expected := []int{4, 3, 2, 1}
	actual := Revert(slice).([]int)
	if !reflect.DeepEqual(actual, expected) {
		t.Error(expected, actual)
	}
}
func TestRemove(t *testing.T) {
	slice := []int{1, 2, 3, 4}
	slice = Remove(slice, 2).([]int)
	if !reflect.DeepEqual(slice, []int{1, 2, 4}) { t.Error(slice) }
	slice = Remove(slice, 2).([]int)
	if !reflect.DeepEqual(slice, []int{1, 2}) { t.Error(slice) }
	slice = Remove(slice, 0).([]int)
	if !reflect.DeepEqual(slice, []int{2}) { t.Error(slice) }
	slice = Remove(slice, 0).([]int)
	if !reflect.DeepEqual(slice, []int{}) { t.Error(slice) }
}
func TestInArray(t *testing.T) {
	if !InArray(3, []int{1, 2, 3}) { t.Error() }
	if InArray("3", []string{"1", "2"}) { t.Error() }
	//InArray("1", []int{1, 2, 3})
}
