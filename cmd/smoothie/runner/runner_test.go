package runner

import (
	"reflect"
	"testing"
)

func TestParseDrivers(t *testing.T) {
	type expected struct {
		drivers []string
		args    map[string][]string
	}
	tests := map[string]struct {
		args     []string
		expected expected
	}{
		"single driver without args": {
			[]string{"github:events"}, expected{
				[]string{"github:events"}, map[string][]string{"github:events": []string{}},
			},
		},
		"single driver with args": {
			[]string{"github:events:arg"}, expected{
				[]string{"github:events"}, map[string][]string{"github:events": []string{"arg"}},
			},
		},
		"multi drivers without args": {
			[]string{"github:events", "gmail"}, expected{
				[]string{"github:events", "gmail"}, map[string][]string{"github:events": []string{}, "gmail": []string{}},
			},
		},
		"multi drivers with args": {
			[]string{"github:events:arg", "reddit:arg"}, expected{
				[]string{"github:events", "reddit"}, map[string][]string{"github:events": []string{"arg"}, "reddit": []string{"arg"}},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cnf := &config{
				args: make(map[string][]string),
			}
			cnf.parseDrivers(test.args)

			if !reflect.DeepEqual(cnf.drivers, test.expected.drivers) {
				t.Errorf("unexpected driver of config: got %v, expect %v\n", cnf.drivers, test.expected.drivers)
			}
			if !reflect.DeepEqual(cnf.args, test.expected.args) {
				t.Errorf("unexpected driver of config: got %v, expect %v\n", cnf.args, test.expected.args)
			}
		})
	}
}
