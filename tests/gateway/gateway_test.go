package gateway_test

import (
	"testing"

	"github.com/anhao/go-easy-sms/gateway"
)

func TestBaseGateway(t *testing.T) {
	// 创建一个基础网关
	config := map[string]interface{}{
		"string_value":  "test",
		"int_value":     123,
		"float_value":   123.45,
		"bool_value":    true,
		"string_int":    "456",
		"string_float":  "456.78",
		"string_bool":   "true",
		"missing_value": nil,
	}

	g := gateway.NewBaseGateway("test", config)

	// 测试获取名称
	if g.GetName() != "test" {
		t.Errorf("Expected name to be test, got: %s", g.GetName())
	}

	// 测试获取配置
	if g.GetConfig()["string_value"] != "test" {
		t.Errorf("Expected config[string_value] to be test, got: %v", g.GetConfig()["string_value"])
	}

	// 测试获取字符串配置
	if g.GetConfigString("string_value") != "test" {
		t.Errorf("Expected GetConfigString(string_value) to return test, got: %s", g.GetConfigString("string_value"))
	}

	// 测试获取不存在的字符串配置，使用默认值
	if g.GetConfigString("non_existent", "default") != "default" {
		t.Errorf("Expected GetConfigString(non_existent, default) to return default, got: %s", g.GetConfigString("non_existent", "default"))
	}

	// 测试获取整数配置
	if g.GetConfigInt("int_value") != 123 {
		t.Errorf("Expected GetConfigInt(int_value) to return 123, got: %d", g.GetConfigInt("int_value"))
	}

	// 测试获取字符串形式的整数配置
	if g.GetConfigInt("string_int") != 456 {
		t.Errorf("Expected GetConfigInt(string_int) to return 456, got: %d", g.GetConfigInt("string_int"))
	}

	// 测试获取不存在的整数配置，使用默认值
	if g.GetConfigInt("non_existent", 789) != 789 {
		t.Errorf("Expected GetConfigInt(non_existent, 789) to return 789, got: %d", g.GetConfigInt("non_existent", 789))
	}

	// 测试获取浮点数配置
	if g.GetConfigFloat("float_value") != 123.45 {
		t.Errorf("Expected GetConfigFloat(float_value) to return 123.45, got: %f", g.GetConfigFloat("float_value"))
	}

	// 测试获取字符串形式的浮点数配置
	if g.GetConfigFloat("string_float") != 456.78 {
		t.Errorf("Expected GetConfigFloat(string_float) to return 456.78, got: %f", g.GetConfigFloat("string_float"))
	}

	// 测试获取不存在的浮点数配置，使用默认值
	if g.GetConfigFloat("non_existent", 789.01) != 789.01 {
		t.Errorf("Expected GetConfigFloat(non_existent, 789.01) to return 789.01, got: %f", g.GetConfigFloat("non_existent", 789.01))
	}

	// 测试获取布尔配置
	if !g.GetConfigBool("bool_value") {
		t.Errorf("Expected GetConfigBool(bool_value) to return true")
	}

	// 测试获取字符串形式的布尔配置
	if !g.GetConfigBool("string_bool") {
		t.Errorf("Expected GetConfigBool(string_bool) to return true")
	}

	// 测试获取不存在的布尔配置，使用默认值
	if g.GetConfigBool("non_existent", true) != true {
		t.Errorf("Expected GetConfigBool(non_existent, true) to return true")
	}

	if g.GetConfigBool("non_existent", false) != false {
		t.Errorf("Expected GetConfigBool(non_existent, false) to return false")
	}
}
