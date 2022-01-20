package cmd

import (
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/eric2788/MiraiValBot/file"
	"github.com/eric2788/MiraiValBot/modules/command"
	"github.com/eric2788/MiraiValBot/sites/twitter"
	"github.com/eric2788/MiraiValBot/utils/qq"
)

func tListen(args []string, source *command.MessageSource) error {
	screenId := args[0]
	reply := qq.CreateReply(source.Message)

	result, err := twitter.StartListen(screenId)

	if err != nil {
		reply.Append(qq.NewTextf("启动监听时出现错误: %v", err))
	} else if result {
		reply.Append(qq.NewTextf("开始监听推特用户 %s", screenId))
	} else {
		reply.Append(qq.NewTextf("该用户 %s 已启动监听", screenId))
	}

	return qq.SendGroupMessage(reply)

}

func tTerminate(args []string, source *command.MessageSource) error {
	screenId := args[0]
	reply := qq.CreateReply(source.Message)

	result, err := twitter.StopListen(screenId)

	if err != nil {
		reply.Append(qq.NewTextf("中止监听时出现错误: %v", err))
	} else if result {
		reply.Append(qq.NewTextf("已中止监听推特用户 %s", screenId))
	} else {
		reply.Append(qq.NewTextf("你尚未开始监听此推特用户。"))
	}

	return qq.SendGroupMessage(reply)
}

func tListening(args []string, source *command.MessageSource) error {
	listening := file.DataStorage.Listening.Twitter
	reply := qq.CreateReply(source.Message)

	if listening.Size() > 0 {
		reply.Append(qq.NewTextf("正在监听的推特用户: %v", listening.ToArr()))
	} else {
		reply.Append(message.NewText("没有在监听的推特用户"))
	}

	return qq.SendGroupMessage(reply)
}

var (
	tListenCommand    = command.NewNode([]string{"listen", "监听"}, "启动监听用户", true, tListen, "<用户ID>")
	tTerminateCommand = command.NewNode([]string{"terminate", "中止", "中止监听"}, "中止监听用户", true, tTerminate, "<用户ID>")
	tListeningCommand = command.NewNode([]string{"listening", "正在监听", "监听列表"}, "获取监听列表", false, tListening)
)

var twitterCommand = command.NewParent([]string{"twitter", "推特"}, "推特指令",
	tListenCommand,
	tTerminateCommand,
	tListeningCommand,
)

func init() {
	command.AddCommand(twitterCommand)
}
