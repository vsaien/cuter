package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsEmail(t *testing.T) {
	cases := []struct {
		email string
		out   bool
	}{
		{"test@example.com", true},
		{"!k@example.com", true},
		{".k@example.com", true},
		{"11@example.com", true},
		{"test.@example.com", true},
		{"test@@example.com", false},
		{"test@test@example.com", false},
		{"NoEmail", false},
		{"@NoEmail", false},
		{"@NoEmail.com", false},
		{"测试@example.com", false},
		{"test @example.com", false},
	}

	for _, each := range cases {
		actual := IsEmail(each.email)
		assert.Equal(t, each.out, actual)
	}
}

func TestIsCellphone(t *testing.T) {
	cases := []struct {
		cellphone string
		out       bool
	}{
		{"111111", false},
		{"aaaaaa", false},
		{"13120972717", true},
	}

	for _, each := range cases {
		actual := IsCellphone(each.cellphone)
		assert.Equal(t, each.out, actual)
	}
}

func TestIsTelephone(t *testing.T) {
	cases := []struct {
		telphone string
		out      bool
	}{
		{"2457343", true},
		{"24573431", true},
		{"245734311", false},
		{"A245734", false},
		{"111111", false},
	}

	for _, each := range cases {
		actual := IsTelephone(each.telphone)
		assert.Equal(t, each.out, actual)
	}
}
