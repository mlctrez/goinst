package main

import (
	"testing"
)

func TestVersion_LessThan(t *testing.T) {

	one := &Version{Major: 1, Minor: 1, Patch: 1, ReleaseCandidate: true}
	two := &Version{Major: 1, Minor: 1, Patch: 2, ReleaseCandidate: true}

	if !one.LessThan(two) {
		t.Error("rc patch level comparison failed")
	}

	one = &Version{Major: 1, Minor: 1, Patch: 1, ReleaseCandidate: true}
	two = &Version{Major: 1, Minor: 1, Patch: 1, ReleaseCandidate: false}

	if !one.LessThan(two) {
		t.Error("rc should be less than non-rc at same major, minor, and patch level")
	}
	if two.LessThan(one) {
		t.Error("rc should be less than non-rc at same major, minod, and patch level")
	}

	one = &Version{Major: 1, Minor: 1, Patch: 2, ReleaseCandidate: true}
	two = &Version{Major: 1, Minor: 1, Patch: 1, ReleaseCandidate: false}

	if !one.LessThan(two) {
		t.Error("rc should be less than non-rc at same major and minor but different patch level")
	}
	if two.LessThan(one) {
		t.Error("rc should be less than non-rc at same major and minor but different patch level")
	}

}
