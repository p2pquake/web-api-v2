package userquake

func GetAreaName(code int) ([]string, bool) {
	areaMapping := map[int][]string{
		0:   {"未設定", "未設定", "地域未設定"},
		901: {"不明", "不明", "地域不明"},
		905: {"外国", "外国", "日本以外"},
		10:  {"北海道", "北海道", "北海道 石狩"},
		15:  {"北海道", "北海道", "北海道 渡島"},
		20:  {"北海道", "北海道", "北海道 檜山"},
		25:  {"北海道", "北海道", "北海道 後志"},
		30:  {"北海道", "北海道", "北海道 空知"},
		35:  {"北海道", "北海道", "北海道 上川"},
		40:  {"北海道", "北海道", "北海道 留萌"},
		45:  {"北海道", "北海道", "北海道 宗谷"},
		50:  {"北海道", "北海道", "北海道 網走"},
		55:  {"北海道", "北海道", "北海道 胆振"},
		60:  {"北海道", "北海道", "北海道 日高"},
		65:  {"北海道", "北海道", "北海道 十勝"},
		70:  {"北海道", "北海道", "北海道 釧路"},
		75:  {"北海道", "北海道", "北海道 根室"},
		100: {"東北", "青森", "青森津軽"},
		105: {"東北", "青森", "青森三八上北"},
		106: {"東北", "青森", "青森下北"},
		110: {"東北", "岩手", "岩手沿岸北部"},
		111: {"東北", "岩手", "岩手沿岸南部"},
		115: {"東北", "岩手", "岩手内陸"},
		120: {"東北", "宮城", "宮城北部"},
		125: {"東北", "宮城", "宮城南部"},
		130: {"東北", "秋田", "秋田沿岸"},
		135: {"東北", "秋田", "秋田内陸"},
		140: {"東北", "山形", "山形庄内"},
		141: {"東北", "山形", "山形最上"},
		142: {"東北", "山形", "山形村山"},
		143: {"東北", "山形", "山形置賜"},
		150: {"東北", "福島", "福島中通り"},
		151: {"東北", "福島", "福島浜通り"},
		152: {"東北", "福島", "福島会津"},
		200: {"関東", "茨城", "茨城北部"},
		205: {"関東", "茨城", "茨城南部"},
		210: {"関東", "栃木", "栃木北部"},
		215: {"関東", "栃木", "栃木南部"},
		220: {"関東", "群馬", "群馬北部"},
		225: {"関東", "群馬", "群馬南部"},
		230: {"関東", "埼玉", "埼玉北部"},
		231: {"関東", "埼玉", "埼玉南部"},
		232: {"関東", "埼玉", "埼玉秩父"},
		240: {"関東", "千葉", "千葉北東部"},
		241: {"関東", "千葉", "千葉北西部"},
		242: {"関東", "千葉", "千葉南部"},
		250: {"関東", "東京", "東京"},
		255: {"関東", "東京", "伊豆諸島北部"},
		260: {"関東", "東京", "伊豆諸島南部"},
		265: {"関東", "東京", "小笠原"},
		270: {"関東", "神奈川", "神奈川東部"},
		275: {"関東", "神奈川", "神奈川西部"},
		300: {"北陸", "新潟", "新潟上越"},
		301: {"北陸", "新潟", "新潟中越"},
		302: {"北陸", "新潟", "新潟下越"},
		305: {"北陸", "新潟", "新潟佐渡"},
		310: {"北陸", "富山", "富山東部"},
		315: {"北陸", "富山", "富山西部"},
		320: {"北陸", "石川", "石川能登"},
		325: {"北陸", "石川", "石川加賀"},
		330: {"北陸", "福井", "福井嶺北"},
		335: {"北陸", "福井", "福井嶺南"},
		340: {"甲信", "山梨", "山梨東部"},
		345: {"甲信", "山梨", "山梨中・西部"},
		350: {"甲信", "長野", "長野北部"},
		351: {"甲信", "長野", "長野中部"},
		355: {"甲信", "長野", "長野南部"},
		400: {"東海", "岐阜", "岐阜飛騨"},
		405: {"東海", "岐阜", "岐阜美濃"},
		410: {"東海", "静岡", "静岡伊豆"},
		411: {"東海", "静岡", "静岡東部"},
		415: {"東海", "静岡", "静岡中部"},
		416: {"東海", "静岡", "静岡西部"},
		420: {"東海", "愛知", "愛知東部"},
		425: {"東海", "愛知", "愛知西部"},
		430: {"東海", "三重", "三重北中部"},
		435: {"東海", "三重", "三重南部"},
		440: {"近畿", "滋賀", "滋賀北部"},
		445: {"近畿", "滋賀", "滋賀南部"},
		450: {"近畿", "京都", "京都北部"},
		455: {"近畿", "京都", "京都南部"},
		460: {"近畿", "大阪", "大阪北部"},
		465: {"近畿", "大阪", "大阪南部"},
		470: {"近畿", "兵庫", "兵庫北部"},
		475: {"近畿", "兵庫", "兵庫南部"},
		480: {"近畿", "奈良", "奈良"},
		490: {"近畿", "和歌山", "和歌山北部"},
		495: {"近畿", "和歌山", "和歌山南部"},
		500: {"中国", "鳥取", "鳥取東部"},
		505: {"中国", "鳥取", "鳥取中・西部"},
		510: {"中国", "島根", "島根東部"},
		515: {"中国", "島根", "島根西部"},
		514: {"中国", "島根", "島根隠岐"},
		520: {"中国", "岡山", "岡山北部"},
		525: {"中国", "岡山", "岡山南部"},
		530: {"中国", "広島", "広島北部"},
		535: {"中国", "広島", "広島南部"},
		540: {"中国", "山口", "山口北部"},
		545: {"中国", "山口", "山口中・東部"},
		541: {"中国", "山口", "山口西部"},
		550: {"四国", "徳島", "徳島北部"},
		555: {"四国", "徳島", "徳島南部"},
		560: {"四国", "香川", "香川"},
		570: {"四国", "愛媛", "愛媛東予"},
		575: {"四国", "愛媛", "愛媛中予"},
		576: {"四国", "愛媛", "愛媛南予"},
		580: {"四国", "高知", "高知東部"},
		581: {"四国", "高知", "高知中部"},
		582: {"四国", "高知", "高知西部"},
		600: {"九州北", "福岡", "福岡福岡"},
		601: {"九州北", "福岡", "福岡北九州"},
		602: {"九州北", "福岡", "福岡筑豊"},
		605: {"九州北", "福岡", "福岡筑後"},
		610: {"九州北", "佐賀", "佐賀北部"},
		615: {"九州北", "佐賀", "佐賀南部"},
		620: {"九州北", "長崎", "長崎北部"},
		625: {"九州北", "長崎", "長崎南部"},
		630: {"九州北", "長崎", "長崎壱岐・対馬"},
		635: {"九州北", "長崎", "長崎五島"},
		640: {"九州北", "熊本", "熊本阿蘇"},
		641: {"九州北", "熊本", "熊本熊本"},
		645: {"九州北", "熊本", "熊本球磨"},
		646: {"九州北", "熊本", "熊本天草・芦北"},
		650: {"九州北", "大分", "大分北部"},
		651: {"九州北", "大分", "大分中部"},
		655: {"九州北", "大分", "大分西部"},
		656: {"九州北", "大分", "大分南部"},
		660: {"九州南", "宮崎", "宮崎北部平野部"},
		661: {"九州南", "宮崎", "宮崎北部山沿い"},
		665: {"九州南", "宮崎", "宮崎南部平野部"},
		666: {"九州南", "宮崎", "宮崎南部山沿い"},
		670: {"九州南", "鹿児島", "鹿児島薩摩"},
		675: {"九州南", "鹿児島", "鹿児島大隅"},
		680: {"九州南", "鹿児島", "種子島・屋久島"},
		685: {"九州南", "鹿児島", "鹿児島奄美"},
		700: {"沖縄", "沖縄", "沖縄本島北部"},
		701: {"沖縄", "沖縄", "沖縄本島中南部"},
		702: {"沖縄", "沖縄", "沖縄久米島"},
		710: {"沖縄", "沖縄", "沖縄大東島"},
		706: {"沖縄", "沖縄", "沖縄宮古島"},
		705: {"沖縄", "沖縄", "沖縄八重山"},
	}

	a, b := areaMapping[code]
	return a, b
}
