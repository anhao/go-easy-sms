package message_test

import (
	"testing"

	"github.com/anhao/go-easy-sms/message"
)

func TestPhoneNumber(t *testing.T) {
	// 测试国内号码
	phone := message.NewPhoneNumber("13800138000")

	if phone.GetNumber() != "13800138000" {
		t.Errorf("Expected number to be 13800138000, got: %s", phone.GetNumber())
	}

	if phone.GetIDDCode() != 0 {
		t.Errorf("Expected IDDCode to be 0, got: %d", phone.GetIDDCode())
	}

	if phone.GetUniversalNumber() != "13800138000" {
		t.Errorf("Expected universal number to be 13800138000, got: %s", phone.GetUniversalNumber())
	}

	if phone.GetZeroPrefixedNumber() != "13800138000" {
		t.Errorf("Expected zero prefixed number to be 13800138000, got: %s", phone.GetZeroPrefixedNumber())
	}

	if !phone.InChineseMainland() {
		t.Errorf("Expected InChineseMainland() to return true")
	}

	if phone.String() != "13800138000" {
		t.Errorf("Expected String() to return 13800138000, got: %s", phone.String())
	}

	// 测试国际号码
	intlPhone := message.NewPhoneNumber("13800138000", 31)

	if intlPhone.GetNumber() != "13800138000" {
		t.Errorf("Expected number to be 13800138000, got: %s", intlPhone.GetNumber())
	}

	if intlPhone.GetIDDCode() != 31 {
		t.Errorf("Expected IDDCode to be 31, got: %d", intlPhone.GetIDDCode())
	}

	if intlPhone.GetUniversalNumber() != "+3113800138000" {
		t.Errorf("Expected universal number to be +3113800138000, got: %s", intlPhone.GetUniversalNumber())
	}

	if intlPhone.GetZeroPrefixedNumber() != "003113800138000" {
		t.Errorf("Expected zero prefixed number to be 003113800138000, got: %s", intlPhone.GetZeroPrefixedNumber())
	}

	if intlPhone.InChineseMainland() {
		t.Errorf("Expected InChineseMainland() to return false")
	}

	if intlPhone.String() != "+3113800138000" {
		t.Errorf("Expected String() to return +3113800138000, got: %s", intlPhone.String())
	}

	// 测试中国大陆号码（明确指定区号）
	cnPhone := message.NewPhoneNumber("13800138000", 86)

	if !cnPhone.InChineseMainland() {
		t.Errorf("Expected InChineseMainland() to return true for phone with IDDCode 86")
	}
}
