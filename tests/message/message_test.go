package message_test

import (
	"testing"

	"github.com/anhao/go-easy-sms/message"
)

func TestMessage(t *testing.T) {
	// 创建一个新的消息
	msg := message.NewMessage()

	// 测试默认值
	if msg.GetType() != message.TextMessage {
		t.Errorf("Expected message type to be TextMessage, got: %s", msg.GetType())
	}

	if len(msg.GetData()) != 0 {
		t.Errorf("Expected empty data, got: %v", msg.GetData())
	}

	if len(msg.GetGateways()) != 0 {
		t.Errorf("Expected empty gateways, got: %v", msg.GetGateways())
	}

	// 测试设置内容
	content := "Test content"
	msg.SetContent(content)
	if msg.GetContent() != content {
		t.Errorf("Expected content to be %s, got: %s", content, msg.GetContent())
	}

	// 测试设置模板
	template := "test_template"
	msg.SetTemplate(template)
	if msg.GetTemplate() != template {
		t.Errorf("Expected template to be %s, got: %s", template, msg.GetTemplate())
	}

	// 测试设置数据
	data := map[string]interface{}{
		"code": "123456",
		"time": "5分钟",
	}
	msg.SetData(data)
	gotData := msg.GetData()
	if gotData["code"] != "123456" || gotData["time"] != "5分钟" {
		t.Errorf("Expected data to be %v, got: %v", data, gotData)
	}

	// 测试设置网关
	gateways := []string{"aliyun", "yunpian"}
	msg.SetGateways(gateways)
	gotGateways := msg.GetGateways()
	if len(gotGateways) != 2 || gotGateways[0] != "aliyun" || gotGateways[1] != "yunpian" {
		t.Errorf("Expected gateways to be %v, got: %v", gateways, gotGateways)
	}

	// 测试设置类型
	msgType := message.VoiceMessage
	msg.SetType(msgType)
	if msg.GetType() != msgType {
		t.Errorf("Expected type to be %s, got: %s", msgType, msg.GetType())
	}
}
