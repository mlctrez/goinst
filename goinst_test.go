package main

import (
	"testing"
)

func TestVersion_RcBoth(t *testing.T) {
	// testing 1.1.rc1 < 1.1.rc2
	one := &Version{Major: 1, Minor: 1, Patch: 1, ReleaseCandidate: true}
	two := &Version{Major: 1, Minor: 1, Patch: 2, ReleaseCandidate: true}
	if !one.LessThan(two) {
		t.Error("rc patch level comparison failed forward")
	}
	if two.LessThan(one) {
		t.Error("rc patch level comparison failed reverse")
	}

}

func TestVersion_RcOne(t *testing.T) {
	// testing for 1.1.rc1 < 1.1.1
	one := &Version{Major: 1, Minor: 1, Patch: 1, ReleaseCandidate: true}
	two := &Version{Major: 1, Minor: 1, Patch: 1, ReleaseCandidate: false}

	if !one.LessThan(two) {
		t.Error("rc should be less than non-rc at same major, minor, and patch level")
	}
	if two.LessThan(one) {
		t.Error("rc should be less than non-rc at same major, minor, and patch level")
	}
}

func TestVersion_RcOne_DiffPatch(t *testing.T) {
	// testing for 1.1.rc2 < 1.1.1
	one := &Version{Major: 1, Minor: 1, Patch: 2, ReleaseCandidate: true}
	two := &Version{Major: 1, Minor: 1, Patch: 1, ReleaseCandidate: false}

	if !one.LessThan(two) {
		t.Error("rc should be less than non-rc at same major and minor but different patch level")
	}
	if two.LessThan(one) {
		t.Error("rc should be less than non-rc at same major and minor but different patch level")
	}
}

func TestVersion_RcNone_DiffPatch(t *testing.T) {
	// testing for 1.11.0 < 1.11.1
	one := &Version{Major: 1, Minor: 11, Patch: 0, ReleaseCandidate: false}
	two := &Version{Major: 1, Minor: 11, Patch: 1, ReleaseCandidate: false}

	if two.LessThan(one) {
		t.Errorf("rc false with patch diff test one failed")
	}

	if !one.LessThan(two) {
		t.Errorf("rc false with patch diff test two failed")
	}
}
