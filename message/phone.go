package message

import (
	"fmt"
)

// PhoneNumber 表示电话号码
type PhoneNumber struct {
	Number  string // 电话号码
	IDDCode int    // 国际区号
}

// NewPhoneNumber 创建一个新的电话号码
func NewPhoneNumber(number string, iddCode ...int) *PhoneNumber {
	var code int
	if len(iddCode) > 0 {
		code = iddCode[0]
	}
	return &PhoneNumber{
		Number:  number,
		IDDCode: code,
	}
}

// GetNumber 获取电话号码
func (p *PhoneNumber) GetNumber() string {
	return p.Number
}

// GetIDDCode 获取国际区号
func (p *PhoneNumber) GetIDDCode() int {
	return p.IDDCode
}

// GetUniversalNumber 获取带国际区号的电话号码 (+8618888888888)
func (p *PhoneNumber) GetUniversalNumber() string {
	if p.IDDCode > 0 {
		return fmt.Sprintf("+%d%s", p.IDDCode, p.Number)
	}
	return p.Number
}

// GetZeroPrefixedNumber 获取带零前缀的国际电话号码 (008618888888888)
func (p *PhoneNumber) GetZeroPrefixedNumber() string {
	if p.IDDCode > 0 {
		return fmt.Sprintf("00%d%s", p.IDDCode, p.Number)
	}
	return p.Number
}

// InChineseMainland 判断是否是中国大陆号码
func (p *PhoneNumber) InChineseMainland() bool {
	return p.IDDCode == 0 || p.IDDCode == 86
}

// String 实现 Stringer 接口
func (p *PhoneNumber) String() string {
	if p.IDDCode > 0 {
		return fmt.Sprintf("+%d%s", p.IDDCode, p.Number)
	}
	return p.Number
}
