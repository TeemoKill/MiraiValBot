package valorant

import "github.com/eric2788/MiraiValBot/valorant"

type MatchMetaDataSub struct {
	DisplayName string                  `json:"display_name"`
	Data        *valorant.MatchMetaData `json:"data"`
}
