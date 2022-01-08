package bilibili

import (
	"fmt"
	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/eric2788/MiraiValBot/file"
	bc "github.com/eric2788/MiraiValBot/modules/broadcaster"
	"github.com/eric2788/MiraiValBot/utils/array"
	"github.com/eric2788/MiraiValBot/utils/request"
)

type RoomInfo struct {
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	Message string `json:"message"`
}

const Host = "https://api.live.bilibili.com/room/v1/Room/get_info"

var (
	roomInfoCache = make(map[int64]*RoomInfo)
	listening     = file.DataStorage.Listening
	topic         = func(room int64) string { return fmt.Sprintf("blive:%d", room) }
)

func StartListen(room int64) (bool, error) {

	if info, err := GetRoomInfo(room); err != nil {
		return false, err
	} else if info.Code != 0 {
		return false, fmt.Errorf("房間不存在")
	}

	file.UpdateStorage(func() {
		listening.Bilibili = array.AddInt64(listening.Bilibili, room)
	})

	info, _ := bot.GetModule(bc.Tag)

	broadcaster := info.Instance.(*bc.Broadcaster)

	return broadcaster.Subscribe(topic(room), MessageHandler)
}

func StopListen(room int64) (bool, error) {

	index := array.IndexOfInt64(listening.Bilibili, room)

	if index == -1 {
		return false, nil
	}

	file.UpdateStorage(func() {
		listening.Bilibili = array.RemoveInt64(listening.Bilibili, index)
	})

	info, _ := bot.GetModule(bc.Tag)

	broadcaster := info.Instance.(*bc.Broadcaster)

	result := broadcaster.UnSubscribe(topic(room))

	return result, nil
}

func GetRoomInfo(room int64) (*RoomInfo, error) {
	if info, ok := roomInfoCache[room]; ok {
		return info, nil
	}

	var info = &RoomInfo{}
	if err := request.Get(fmt.Sprintf("%s?room_id=%d", Host, room), info); err != nil {
		return nil, err
	}

	roomInfoCache[room] = info
	return info, nil
}

func ClearRoomInfo(room int64) bool {
	if room != -1 {
		if _, ok := roomInfoCache[room]; !ok {
			return false
		}
		delete(roomInfoCache, room)
		return true
	} else {
		roomInfoCache = make(map[int64]*RoomInfo)
		return true
	}
}
