package utils

import (
	"kiv_zos/utils"
	"testing"
)

func TestHasBit(t *testing.T) {
	b := byte(0xB5)

	if !utils.HasBit(b, 7) {
		t.Errorf("8th bit should be set for b %b", b)
	}
	if utils.HasBit(b, 6) {
		t.Errorf("7th bit should not be set for b %b", b)
	}
	if !utils.HasBit(b, 5) {
		t.Errorf("6th bit should be set for b %b", b)
	}
	if !utils.HasBit(b, 4) {
		t.Errorf("5th bit should be set for b %b", b)
	}
	if utils.HasBit(b, 3) {
		t.Errorf("4th bit should not be set for b %b", b)
	}
	if !utils.HasBit(b, 2) {
		t.Errorf("3rd bit should be set for b %b", b)
	}
	if utils.HasBit(b, 1) {
		t.Errorf("2nd bit should not be set for b %b", b)
	}
	if !utils.HasBit(b, 0) {
		t.Errorf("1st bit should be set for b %b", b)
	}
}

func TestSetBit(t *testing.T) {
	b := byte(0)

	b = utils.SetBit(b, 0)
	b = utils.SetBit(b, 2)
	b = utils.SetBit(b, 4)
	b = utils.SetBit(b, 5)
	b = utils.SetBit(b, 7)

	if !utils.HasBit(b, 7) {
		t.Errorf("8th bit should be set for b %b", b)
	}
	if utils.HasBit(b, 6) {
		t.Errorf("7th bit should not be set for b %b", b)
	}
	if !utils.HasBit(b, 5) {
		t.Errorf("6th bit should be set for b %b", b)
	}
	if !utils.HasBit(b, 4) {
		t.Errorf("5th bit should be set for b %b", b)
	}
	if utils.HasBit(b, 3) {
		t.Errorf("4th bit should not be set for b %b", b)
	}
	if !utils.HasBit(b, 2) {
		t.Errorf("3rd bit should be set for b %b", b)
	}
	if utils.HasBit(b, 1) {
		t.Errorf("2nd bit should not be set for b %b", b)
	}
	if !utils.HasBit(b, 0) {
		t.Errorf("1st bit should be set for b %b", b)
	}
}

func TestClearBit(t *testing.T) {
	b := byte(0xFF)

	b = utils.ClearBit(b, 1)
	b = utils.ClearBit(b, 3)
	b = utils.ClearBit(b, 6)

	if !utils.HasBit(b, 7) {
		t.Errorf("8th bit should be set for b %b", b)
	}
	if utils.HasBit(b, 6) {
		t.Errorf("7th bit should not be set for b %b", b)
	}
	if !utils.HasBit(b, 5) {
		t.Errorf("6th bit should be set for b %b", b)
	}
	if !utils.HasBit(b, 4) {
		t.Errorf("5th bit should be set for b %b", b)
	}
	if utils.HasBit(b, 3) {
		t.Errorf("4th bit should not be set for b %b", b)
	}
	if !utils.HasBit(b, 2) {
		t.Errorf("3rd bit should be set for b %b", b)
	}
	if utils.HasBit(b, 1) {
		t.Errorf("2nd bit should not be set for b %b", b)
	}
	if !utils.HasBit(b, 0) {
		t.Errorf("1st bit should be set for b %b", b)
	}
}
