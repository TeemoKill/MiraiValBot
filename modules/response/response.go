package response

import (
	"crypto/md5"
	"encoding/binary"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/eric2788/MiraiValBot/internal/eventhook"
	"github.com/eric2788/MiraiValBot/internal/file"
	"github.com/eric2788/MiraiValBot/internal/qq"
	"github.com/eric2788/MiraiValBot/modules/chat_reply"
	"github.com/eric2788/MiraiValBot/services/copywriting"
	"github.com/eric2788/MiraiValBot/utils/misc"
)

const Tag = "valbot.response"

var (
	logger   = utils.GetModuleLogger(Tag)
	instance = &response{
		res: new(chat_reply.AIChatResponse),
	}
	YesNoPattern         = regexp.MustCompile(`^.+是.+吗[\?？]*$`)
	questionMarkReplacer = strings.NewReplacer("?", "", "？", "")

	longWongTalks = []string{
		"恭迎龙王 %s (跪拜)",
		"恭喜话痨 %s 成龙王咯",
		"口水多还得是你, %s",
		"%s, YOU 👆 ARE 👆 KING 👑",
		"你就是龙王 %s 吗, 不错",
	}
)

type response struct {
	res *chat_reply.AIChatResponse
}

func (r *response) HookEvent(bot *bot.Bot) {
	bot.GroupMessageEvent.Subscribe(func(c *client.QQClient, msg *message.GroupMessage) {
		content := msg.ToString()

		if res, ok := file.DataStorage.Responses[content]; ok {
			m := message.NewSendingMessage().Append(message.NewText(res))
			_ = qq.SendGroupMessageByGroup(msg.GroupCode, m)
		} else if YesNoPattern.MatchString(content) {
			m := message.NewSendingMessage()
			if ans, ok := file.DataStorage.Answers[content]; ok {
				logger.Infof("此问题已被手动设置，因此使用被设置的回答")
				m.Append(message.NewText(getResponse(ans)))
			} else {
				ans = getQuestionAns(content)
				logger.Infof("自动回答问题 %s 为 %t", content, ans)
				m.Append(message.NewText(getResponse(ans)))
			}
			_ = qq.SendGroupMessageByGroup(msg.GroupCode, m)
		} else {

			rand.Seed(time.Now().UnixNano())

			// 1/50 (2%) 机率会回复
			if rand.Intn(50) == 25 {

				// 没有文字信息，随机发送龙图?
				if len(qq.ParseMsgContent(msg.Elements).Texts) == 0 {
					send, err := misc.NewRandomDragon()

					if err != nil {
						logger.Errorf("获取龙图失败: %v, 改为发送随机群图片", err)
						send, err = misc.NewRandomImage()
					}

					// 依然失败
					if err != nil {
						logger.Errorf("获取图片失败: %v, 放弃发送。", err)
						return
					}

					_ = qq.SendGroupMessageByGroup(msg.GroupCode, send)
					return
				}

				// 透过 AI 回复信息
				reply, err := r.res.Response(msg)
				if err != nil {
					logger.Errorf("透过 AI 回复对话时出现错误: %v", err)
				} else {

					// create a message with no reply element
					send := message.NewSendingMessage()

					for _, r := range reply.Elements {

						// skip reply and at
						if _, ok := r.(*message.ReplyElement); ok {
							continue
						} else if _, ok = r.(*message.AtElement); ok {
							continue
						}

						send.Append(r)
					}

					_ = qq.SendGroupMessageByGroup(msg.GroupCode, send)
				}
			}

		}
	})

	bot.GroupNotifyEvent.Subscribe(func(c *client.QQClient, event client.INotifyEvent) {

		// 非瓦群无视
		if event.From() != qq.ValGroupInfo.Uin {
			return
		}

		rand.Seed(time.Now().UnixNano())

		switch notify := event.(type) {
		case *client.GroupPokeNotifyEvent:

			msg := message.NewSendingMessage()
			sender := qq.FindGroupMember(notify.Sender)

			// 非机器人
			if notify.Receiver != c.Uin {

				receivier := qq.FindGroupMember(notify.Receiver)

				// 5% 触发CP
				if rand.Intn(100)+1 > 95 {

					list, atk, def, err := copywriting.GetCPList()
					if err != nil {
						logger.Errorf("获取CP列表失败: %v", err)
					} else {
						random := list[rand.Intn(len(list))]
						replacer := strings.NewReplacer(atk, sender.DisplayName(), def, receivier.DisplayName())
						msg.Append(message.NewText(replacer.Replace(random)))
						_ = qq.SendGroupMessage(msg)
					}

				}

				return
			}

			if rand.Intn(100)+1 > 10 {
				msg.Append(qq.NewTextfLn("戳你妹戳戳戳, %s!", sender.DisplayName()))
				// 戳回去咯
				c.SendGroupPoke(qq.ValGroupInfo.Code, notify.Sender)
			} else { // 10% 机率触发发病

				if success := sendFabing(msg, sender); !success {
					return
				}

			}

			_ = qq.SendGroupMessage(msg)

		case *client.MemberHonorChangedNotifyEvent:

			msg := message.NewSendingMessage()

			if notify.Uin == c.Uin {

				msg.Append(qq.NewTextf("机器人也能成 %s, 你群是不是该好好反思一下", qq.GetHonorString(notify.Honor)))
				msg.Append(message.NewFace(15))

			} else {

				user := qq.FindGroupMember(notify.Uin)

				// 80% 随机祝贺, 20% 发病
				if rand.Intn(100)+1 > 20 {
					if notify.Honor == client.Talkative {
						random := longWongTalks[rand.Intn(len(longWongTalks))]
						msg.Append(qq.NewTextf(random, user.DisplayName()))
					}
				} else {
					if success := sendFabing(msg, user); !success {
						return
					}
				}
			}

			_ = qq.SendGroupMessage(msg)

		}

	})
}

func sendFabing(msg *message.SendingMessage, sender *client.GroupMemberInfo) bool {
	var getter func() ([]string, string, error)
	if rand.Intn(2) == 1 {
		getter = copywriting.GetFabingList
	} else {
		getter = copywriting.GetFadianList
	}
	if list, replace, err := getter(); err != nil {
		logger.Errorf("获取发病模板失败: %v", err)
		return false
	} else {
		random := list[rand.Intn(len(list))]
		msg.Append(message.NewText(strings.ReplaceAll(random, replace, sender.DisplayName())))
		return true
	}
}

func getQuestionAns(content string) bool {
	hasher := md5.New()
	question := questionMarkReplacer.Replace(content)
	hashed := hasher.Sum([]byte(question))
	u64 := binary.BigEndian.Uint64(hashed)
	rand.Seed(int64(u64))
	return rand.Intn(2) == 1
}

func getResponse(is bool) string {
	if is {
		return "确实"
	} else {
		return "并不是"
	}
}

func init() {
	eventhook.RegisterAsModule(instance, "自定義回應", Tag, logger)
}
