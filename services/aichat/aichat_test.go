package aichat

import (
	"github.com/eric2788/MiraiValBot/utils/test"
	"os"
	"strings"
	"testing"
)

var chats = map[string]AIReply{
	"xiaoai":    &XiaoAi{},
	"qingyunke": &QingYunKe{},
	"tianxing":  &TianXing{},
}

func TestGetXiaoAi(t *testing.T) {

	t.Skip("xiaoai is dead")

	aichat := chats["xiaoai"]

	msg, err := aichat.Reply("你好，你叫什么？")
	if err != nil {
		if strings.Contains(err.Error(), "timeout") || err.Error() == "无法获取回复讯息" {
			t.Skip(err)
		}
		t.Fatal(err)
	}

	t.Logf("Reply: %s", msg)
}

func TestQingYunKe(t *testing.T) {
	aichat := chats["qingyunke"]

	msg, err := aichat.Reply("你好，你叫什么？")
	if err != nil {
		if strings.Contains(err.Error(), "timeout") {
			t.Skip(err)
		}
		t.Fatal(err)
	}

	t.Logf("Reply: %s", msg)
}

func TestTianXing_Reply(t *testing.T) {
	aichat := chats["tianxing"]

	if os.Getenv("TIAN_API_KEY") == "" {
		t.Skip("TIAN_API_KEY is empty")
	}

	msg, err := aichat.Reply("你好，你叫什么？")
	if err != nil {
		if strings.Contains(err.Error(), "timeout") {
			t.Skip(err)
		}
		t.Fatal(err)
	}

	t.Logf("Reply: %s", msg)
}

func init() {
	test.InitTesting()
}
