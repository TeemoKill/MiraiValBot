package misc

import (
	"encoding/base64"
	"fmt"
	"github.com/eric2788/common-utils/request"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/eric2788/MiraiValBot/internal/qq"
)

var logger = utils.GetModuleLogger("valbot.misc")

func NewRandomMessage() (*message.SendingMessage, error) {
	random, err := qq.GetRandomGroupMessage(qq.ValGroupInfo.Code)
	if err != nil {
		return nil, err
	}

	send := message.NewSendingMessage()

	for _, ele := range random.Elements {

		switch ele.(type) {
		case *message.ReplyElement:
			continue
		case *message.ForwardElement:
			continue
		default:
			break
		}
		send.Append(ele)
	}

	// 没有元素也略过
	if len(send.Elements) == 0 {
		return nil, fmt.Errorf("讯息元素为空。")
	}

	return send, nil
}

func NewRandomImage() (*message.SendingMessage, error) {
	rand.Seed(time.Now().UnixMicro())
	imgs := qq.GetImageList()

	if len(imgs) == 0 {
		return nil, fmt.Errorf("群图片缓存列表为空。")
	}

	logger.Debugf("成功索取 %d 张群图片缓存。", len(imgs))

	chosen := imgs[rand.Intn(len(imgs))]

	b, err := qq.GetCacheImage(chosen)
	if err != nil {
		return nil, err
	}
	img, err := qq.NewImageByByte(b)
	if err != nil {
		return nil, err
	}
	return message.NewSendingMessage().Append(img), nil
}

func NewRandomDragon() (*message.SendingMessage, error) {
	backup := "https://phqghume.github.io/img/"
	rand.Seed(time.Now().UnixMicro())
	random := rand.Intn(58) + 1
	ext := ".jpg"
	if random > 48 {
		ext = ".gif"
	}
	imgLink := fmt.Sprintf("%slong%%20(%d)%s", backup, random, ext)
	img, err := qq.NewImageByUrl(imgLink)
	if err != nil {
		return nil, err
	}
	return message.NewSendingMessage().Append(img), nil
}

func ShuffleText(content string) string {
	lcrune := []rune(content)
	rand.Shuffle(len(lcrune), func(i, j int) {
		lcrune[i], lcrune[j] = lcrune[j], lcrune[i]
	})
	return string(lcrune)
}

func FetchImageByteToForward(forwarder *message.ForwardMessage, b []byte, wg *sync.WaitGroup) {
	defer wg.Done()
	msg := message.NewSendingMessage()
	img, err := qq.NewImageByByte(b)
	if err != nil {
		logger.Errorf("上傳圖片失败: %v", err)
		msg.Append(message.NewText("[圖片获取失败]"))
	} else {
		msg.Append(img)
	}
	forwarder.AddNode(qq.NewForwardNode(msg))
}

func FetchImageToForward(forwarder *message.ForwardMessage, url string, wg *sync.WaitGroup) {
	defer wg.Done()
	msg := message.NewSendingMessage()
	img, err := qq.NewImageByUrl(url)
	if err != nil {
		logger.Errorf("上傳圖片失败: %v", err)
		msg.Append(message.NewText("[圖片获取失败]"))
	} else {
		msg.Append(img)
	}
	forwarder.AddNode(qq.NewForwardNode(msg))
}

func TrimPrefixes(s string, prefixes ...string) string {
	for _, prefix := range prefixes {
		s = strings.TrimPrefix(s, prefix)
	}
	return s
}

// ReadURLToSrcData return base64, type, error
func ReadURLToSrcData(url string) (string, string, error) {
	b, err := request.GetBytesByUrl(url)
	if err != nil {
		return "", "", fmt.Errorf("图片下载失败: %v", err)
	}
	t := http.DetectContentType(b)
	return fmt.Sprintf("data:%s;base64,", t) + base64.StdEncoding.EncodeToString(b), t, nil
}
