package strategy_test

import (
	"reflect"
	"testing"

	"github.com/anhao/go-easy-sms/strategy"
)

func TestOrderStrategy(t *testing.T) {
	s := strategy.NewOrderStrategy()

	// 测试空切片
	empty := []string{}
	result := s.Apply(empty)
	if len(result) != 0 {
		t.Errorf("Expected empty slice, got: %v", result)
	}

	// 测试单个元素
	single := []string{"aliyun"}
	result = s.Apply(single)
	if len(result) != 1 || result[0] != "aliyun" {
		t.Errorf("Expected [aliyun], got: %v", result)
	}

	// 测试多个元素
	multiple := []string{"aliyun", "yunpian", "custom"}
	result = s.Apply(multiple)
	if !reflect.DeepEqual(result, multiple) {
		t.Errorf("Expected %v, got: %v", multiple, result)
	}
}

func TestRandomStrategy(t *testing.T) {
	s := strategy.NewRandomStrategy()

	// 测试空切片
	empty := []string{}
	result := s.Apply(empty)
	if len(result) != 0 {
		t.Errorf("Expected empty slice, got: %v", result)
	}

	// 测试单个元素
	single := []string{"aliyun"}
	result = s.Apply(single)
	if len(result) != 1 || result[0] != "aliyun" {
		t.Errorf("Expected [aliyun], got: %v", result)
	}

	// 测试多个元素
	// 注意：由于随机性，我们只能测试结果的长度和元素是否相同，而不是顺序
	multiple := []string{"aliyun", "yunpian", "custom"}
	result = s.Apply(multiple)

	if len(result) != len(multiple) {
		t.Errorf("Expected result length %d, got: %d", len(multiple), len(result))
	}

	// 检查所有元素是否都在结果中
	for _, item := range multiple {
		found := false
		for _, r := range result {
			if r == item {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Element %s not found in result %v", item, result)
		}
	}
}
