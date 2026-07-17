package difficulty

import "time"

type Profile struct {
	Level       int           `json:"level"`
	Name        string        `json:"name"`
	MoveTime    time.Duration `json:"-"`
	MoveTimeMs  int64         `json:"moveTimeMs"`
	MaxDepth    int           `json:"maxDepth"`
	MaxNodes    uint64        `json:"maxNodes"`
	MultiPV     int           `json:"multiPV"`
	Description string        `json:"description"`
}

func Profiles() []Profile {
	return []Profile{
		profile(1, "初识棋盘", 80*time.Millisecond, 1, 800, "只看眼前一着，适合熟悉规则"),
		profile(2, "略懂攻守", 120*time.Millisecond, 2, 2_000, "开始避免直接丢子"),
		profile(3, "稳步入门", 180*time.Millisecond, 2, 5_000, "具备基础吃子与应将能力"),
		profile(4, "街巷棋手", 260*time.Millisecond, 3, 12_000, "能看到短程战术"),
		profile(5, "棋社常客", 380*time.Millisecond, 3, 25_000, "攻守较均衡"),
		profile(6, "沉着应战", 550*time.Millisecond, 4, 55_000, "搜索更深并减少明显失误"),
		profile(7, "战术敏锐", 750*time.Millisecond, 4, 110_000, "更重视强制手段"),
		profile(8, "布局有方", 950*time.Millisecond, 5, 220_000, "较稳定的中短程计算"),
		profile(9, "棋坛强手", 1300*time.Millisecond, 5, 450_000, "更高节点预算与稳定性"),
		profile(10, "境界求真", 1800*time.Millisecond, 6, 900_000, "内置引擎最高资源档，不标注虚构 Elo"),
	}
}

func Get(level int) Profile {
	profiles := Profiles()
	if level < 1 {
		level = 1
	}
	if level > len(profiles) {
		level = len(profiles)
	}
	return profiles[level-1]
}

func profile(level int, name string, moveTime time.Duration, depth int, nodes uint64, description string) Profile {
	return Profile{
		Level: level, Name: name, MoveTime: moveTime, MoveTimeMs: moveTime.Milliseconds(),
		MaxDepth: depth, MaxNodes: nodes, MultiPV: 1, Description: description,
	}
}
