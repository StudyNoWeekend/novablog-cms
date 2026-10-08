package logic

// 本文件定义个人资料「星座 / 性格」选项的展示配置，数据源与社交平台配置（social_platforms.go）一致：
// 管理后台选择器与博客前台均使用公开接口返回的 meta 渲染，前端渲染约定为
// `<svg viewBox="0 0 24 24" v-html="image">`，image 为 24×24 viewBox 内的自包含 SVG 片段。

// 星座按四象限配色：火象 / 土象 / 风象 / 水象
const (
	zodiacColorFire  = "#E2574C"
	zodiacColorEarth = "#5B8C5A"
	zodiacColorAir   = "#4A90D9"
	zodiacColorWater = "#7B68EE"
)

// MBTI 按四大气质类型配色：分析家 NT / 外交家 NF / 守护者 SJ / 探险家 SP
const (
	mbtiColorNT = "#6C5CE7"
	mbtiColorNF = "#00B894"
	mbtiColorSJ = "#0984E3"
	mbtiColorSP = "#F0932B"
)

// profile meta 校验种类（供白名单校验函数区分星座/性格）。
const (
	metaKindZodiac      = "zodiac"
	metaKindPersonality = "personality"
)

// ZodiacConfig 描述一个星座的展示配置。
type ZodiacConfig struct {
	Key       string // 星座 key（如 aries）
	Name      string // 星座名称（如 白羊座）
	Image     string // 星座连线图 SVG 片段（24×24 viewBox 内）
	DateRange string // 日期范围（如 3.21-4.19）
	Element   string // 星座类型（火象/土象/风象/水象）
}

// PersonalityConfig 描述一个性格（MBTI 16 型）的展示配置。
type PersonalityConfig struct {
	Key         string // MBTI 类型 key（如 INTJ）
	Name        string // 类型名称（如 建筑师）
	Image       string // 四字母徽章 SVG 片段（24×24 viewBox 内）
	Description string // 一句话介绍
}

// zodiacOrder 星座展示顺序（按日期先后）。
var zodiacOrder = []string{
	"aries", "taurus", "gemini", "cancer",
	"leo", "virgo", "libra", "scorpio",
	"sagittarius", "capricorn", "aquarius", "pisces",
}

// personalityOrder 性格展示顺序（四气质分组，组内按字母序）。
var personalityOrder = []string{
	"INTJ", "INTP", "ENTJ", "ENTP",
	"INFJ", "INFP", "ENFJ", "ENFP",
	"ISTJ", "ISFJ", "ESTJ", "ESFJ",
	"ISTP", "ISFP", "ESTP", "ESFP",
}

// zodiacMap 存储所有星座配置，键为星座 key。
var zodiacMap = map[string]ZodiacConfig{
	"aries": {
		Key:  "aries",
		Name: "白羊座",
		Image: `<path d="M4.5 17L10 10.5L15.5 6.5M10 10.5L8.5 16.5" fill="none" stroke="` + zodiacColorFire + `" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round"/><circle cx="4.5" cy="17" r="1.4" fill="` + zodiacColorFire + `"/><circle cx="10" cy="10.5" r="1.4" fill="` + zodiacColorFire + `"/><circle cx="8.5" cy="16.5" r="1.2" fill="` + zodiacColorFire + `"/><circle cx="15.5" cy="6.5" r="2" fill="` + zodiacColorFire + `"/>`,
		DateRange: "3.21-4.19",
		Element:   "火象",
	},
	"taurus": {
		Key:  "taurus",
		Name: "金牛座",
		Image: `<path d="M5 6L9.5 12L12 17.5M9.5 12L15 13.5L19 9.5" fill="none" stroke="` + zodiacColorEarth + `" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round"/><circle cx="5" cy="6" r="1.4" fill="` + zodiacColorEarth + `"/><circle cx="9.5" cy="12" r="2" fill="` + zodiacColorEarth + `"/><circle cx="12" cy="17.5" r="1.4" fill="` + zodiacColorEarth + `"/><circle cx="15" cy="13.5" r="1.4" fill="` + zodiacColorEarth + `"/><circle cx="19" cy="9.5" r="1.4" fill="` + zodiacColorEarth + `"/>`,
		DateRange: "4.20-5.20",
		Element:   "土象",
	},
	"gemini": {
		Key:  "gemini",
		Name: "双子座",
		Image: `<path d="M7.5 4.5C6.5 9 6.5 15 7.5 19.5M16.5 4.5C17.5 9 17.5 15 16.5 19.5M4.5 9.5L10.5 8M13.5 8L19.5 9.5M4.5 14.5L10.5 16M13.5 16L19.5 14.5" fill="none" stroke="` + zodiacColorAir + `" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round"/><circle cx="7.5" cy="4.5" r="1.3" fill="` + zodiacColorAir + `"/><circle cx="16.5" cy="4.5" r="1.3" fill="` + zodiacColorAir + `"/><circle cx="7.5" cy="19.5" r="1.3" fill="` + zodiacColorAir + `"/><circle cx="16.5" cy="19.5" r="1.3" fill="` + zodiacColorAir + `"/>`,
		DateRange: "5.21-6.21",
		Element:   "风象",
	},
	"cancer": {
		Key:  "cancer",
		Name: "巨蟹座",
		Image: `<path d="M12 5.5L8.5 11M12 5.5L15.5 11M8.5 11L6.5 18.5M15.5 11L17.5 18.5" fill="none" stroke="` + zodiacColorWater + `" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round"/><circle cx="12" cy="5.5" r="2" fill="` + zodiacColorWater + `"/><circle cx="8.5" cy="11" r="1.4" fill="` + zodiacColorWater + `"/><circle cx="15.5" cy="11" r="1.4" fill="` + zodiacColorWater + `"/><circle cx="6.5" cy="18.5" r="1.2" fill="` + zodiacColorWater + `"/><circle cx="17.5" cy="18.5" r="1.2" fill="` + zodiacColorWater + `"/>`,
		DateRange: "6.22-7.22",
		Element:   "水象",
	},
	"leo": {
		Key:  "leo",
		Name: "狮子座",
		Image: `<path d="M6 17.5L9.5 13M9.5 13L8.5 8.5L11 5.5M9.5 13L15 15.5L19.5 11.5" fill="none" stroke="` + zodiacColorFire + `" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round"/><circle cx="6" cy="17.5" r="2" fill="` + zodiacColorFire + `"/><circle cx="9.5" cy="13" r="1.4" fill="` + zodiacColorFire + `"/><circle cx="8.5" cy="8.5" r="1.3" fill="` + zodiacColorFire + `"/><circle cx="11" cy="5.5" r="1.3" fill="` + zodiacColorFire + `"/><circle cx="15" cy="15.5" r="1.4" fill="` + zodiacColorFire + `"/><circle cx="19.5" cy="11.5" r="1.4" fill="` + zodiacColorFire + `"/>`,
		DateRange: "7.23-8.22",
		Element:   "火象",
	},
	"virgo": {
		Key:  "virgo",
		Name: "处女座",
		Image: `<path d="M4.5 9L8.5 6.5L11.5 10.5L15.5 7.5L19.5 12M11.5 10.5L13 17.5" fill="none" stroke="` + zodiacColorEarth + `" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round"/><circle cx="4.5" cy="9" r="1.3" fill="` + zodiacColorEarth + `"/><circle cx="8.5" cy="6.5" r="1.4" fill="` + zodiacColorEarth + `"/><circle cx="11.5" cy="10.5" r="1.4" fill="` + zodiacColorEarth + `"/><circle cx="15.5" cy="7.5" r="1.3" fill="` + zodiacColorEarth + `"/><circle cx="19.5" cy="12" r="1.4" fill="` + zodiacColorEarth + `"/><circle cx="13" cy="17.5" r="1.3" fill="` + zodiacColorEarth + `"/>`,
		DateRange: "8.23-9.22",
		Element:   "土象",
	},
	"libra": {
		Key:  "libra",
		Name: "天秤座",
		Image: `<path d="M12 5L7.5 13H16.5L12 5ZM7.5 13L6.5 18.5M16.5 13L17.5 18.5" fill="none" stroke="` + zodiacColorAir + `" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round"/><circle cx="12" cy="5" r="2" fill="` + zodiacColorAir + `"/><circle cx="7.5" cy="13" r="1.4" fill="` + zodiacColorAir + `"/><circle cx="16.5" cy="13" r="1.4" fill="` + zodiacColorAir + `"/><circle cx="6.5" cy="18.5" r="1.2" fill="` + zodiacColorAir + `"/><circle cx="17.5" cy="18.5" r="1.2" fill="` + zodiacColorAir + `"/>`,
		DateRange: "9.23-10.23",
		Element:   "风象",
	},
	"scorpio": {
		Key:  "scorpio",
		Name: "天蝎座",
		Image: `<path d="M5 8C5.5 13.5 8.5 17.5 13.5 18L18.5 15.5M5 8L3 5.5M5 8L8 5" fill="none" stroke="` + zodiacColorWater + `" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round"/><circle cx="5" cy="8" r="1.4" fill="` + zodiacColorWater + `"/><circle cx="3" cy="5.5" r="1.2" fill="` + zodiacColorWater + `"/><circle cx="8" cy="5" r="1.2" fill="` + zodiacColorWater + `"/><circle cx="13.5" cy="18" r="2" fill="` + zodiacColorWater + `"/><circle cx="18.5" cy="15.5" r="1.4" fill="` + zodiacColorWater + `"/>`,
		DateRange: "10.24-11.22",
		Element:   "水象",
	},
	"sagittarius": {
		Key:  "sagittarius",
		Name: "射手座",
		Image: `<path d="M4.5 18L9.5 14L14.5 15L19 9.5M9.5 14L11.5 8" fill="none" stroke="` + zodiacColorFire + `" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round"/><circle cx="4.5" cy="18" r="1.4" fill="` + zodiacColorFire + `"/><circle cx="9.5" cy="14" r="2" fill="` + zodiacColorFire + `"/><circle cx="14.5" cy="15" r="1.4" fill="` + zodiacColorFire + `"/><circle cx="19" cy="9.5" r="1.4" fill="` + zodiacColorFire + `"/><circle cx="11.5" cy="8" r="1.3" fill="` + zodiacColorFire + `"/>`,
		DateRange: "11.23-12.21",
		Element:   "火象",
	},
	"capricorn": {
		Key:  "capricorn",
		Name: "摩羯座",
		Image: `<path d="M4 9.5C5.5 15 10 17.5 14 15L19.5 10.5" fill="none" stroke="` + zodiacColorEarth + `" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round"/><circle cx="4" cy="9.5" r="1.4" fill="` + zodiacColorEarth + `"/><circle cx="8" cy="15.4" r="1.3" fill="` + zodiacColorEarth + `"/><circle cx="14" cy="15" r="1.4" fill="` + zodiacColorEarth + `"/><circle cx="19.5" cy="10.5" r="2" fill="` + zodiacColorEarth + `"/>`,
		DateRange: "12.22-1.19",
		Element:   "土象",
	},
	"aquarius": {
		Key:  "aquarius",
		Name: "水瓶座",
		Image: `<path d="M4 8L7 11L10 8M12 13.5L15 16.5L18 13.5" fill="none" stroke="` + zodiacColorAir + `" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round"/><circle cx="7" cy="11" r="1.6" fill="` + zodiacColorAir + `"/><circle cx="15" cy="16.5" r="1.6" fill="` + zodiacColorAir + `"/><circle cx="4" cy="8" r="1.2" fill="` + zodiacColorAir + `"/><circle cx="10" cy="8" r="1.2" fill="` + zodiacColorAir + `"/><circle cx="12" cy="13.5" r="1.2" fill="` + zodiacColorAir + `"/><circle cx="18" cy="13.5" r="1.2" fill="` + zodiacColorAir + `"/>`,
		DateRange: "1.20-2.18",
		Element:   "风象",
	},
	"pisces": {
		Key:  "pisces",
		Name: "双鱼座",
		Image: `<path d="M5 4.5C10 9 10 15 5 19.5M19 4.5C14 9 14 15 19 19.5M5.5 12H18.5" fill="none" stroke="` + zodiacColorWater + `" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round"/><circle cx="5" cy="4.5" r="1.3" fill="` + zodiacColorWater + `"/><circle cx="19" cy="4.5" r="1.3" fill="` + zodiacColorWater + `"/><circle cx="5" cy="19.5" r="1.3" fill="` + zodiacColorWater + `"/><circle cx="19" cy="19.5" r="1.3" fill="` + zodiacColorWater + `"/>`,
		DateRange: "2.19-3.20",
		Element:   "水象",
	},
}

// mbtiBadge 生成 MBTI 四字母徽章 SVG 片段。
func mbtiBadge(key, color string) string {
	return `<rect x="2.5" y="4" width="19" height="16" rx="4" fill="` + color + `"/><text x="12" y="15.3" text-anchor="middle" font-size="6.4" font-weight="700" font-family="ui-monospace, SFMono-Regular, Menlo, monospace" fill="#FFFFFF" letter-spacing="0.5">` + key + `</text>`
}

// personalityMap 存储所有性格配置，键为 MBTI 类型 key。
var personalityMap = map[string]PersonalityConfig{
	"INTJ": {Key: "INTJ", Name: "建筑师", Image: mbtiBadge("INTJ", mbtiColorNT), Description: "沉静而富有想象力，善于将洞察转化为长远策略与体系化方案。"},
	"INTP": {Key: "INTP", Name: "逻辑学家", Image: mbtiBadge("INTP", mbtiColorNT), Description: "热衷探索概念与原理，以严谨的逻辑推演理解世界的运行规律。"},
	"ENTJ": {Key: "ENTJ", Name: "指挥官", Image: mbtiBadge("ENTJ", mbtiColorNT), Description: "意志坚定、天生具有领导力，擅长统筹资源并果断达成目标。"},
	"ENTP": {Key: "ENTP", Name: "辩论家", Image: mbtiBadge("ENTP", mbtiColorNT), Description: "思维敏捷、喜欢挑战常规，在观点交锋与头脑风暴中乐此不疲。"},
	"INFJ": {Key: "INFJ", Name: "提倡者", Image: mbtiBadge("INFJ", mbtiColorNF), Description: "安静而理想主义，洞察人心并坚定地以行动践行心中的意义。"},
	"INFP": {Key: "INFP", Name: "调停者", Image: mbtiBadge("INFP", mbtiColorNF), Description: "内心温柔且理想主义，以共情与创造力守护自己珍视的价值。"},
	"ENFJ": {Key: "ENFJ", Name: "主人公", Image: mbtiBadge("ENFJ", mbtiColorNF), Description: "热情而有感染力，善于凝聚他人并引导大家共同成长。"},
	"ENFP": {Key: "ENFP", Name: "竞选者", Image: mbtiBadge("ENFP", mbtiColorNF), Description: "热情洋溢、充满好奇心，总能用乐观与想象力点燃周围的人。"},
	"ISTJ": {Key: "ISTJ", Name: "物流师", Image: mbtiBadge("ISTJ", mbtiColorSJ), Description: "严谨务实、值得信赖，用条理与坚持把每件事做到位。"},
	"ISFJ": {Key: "ISFJ", Name: "守卫者", Image: mbtiBadge("ISFJ", mbtiColorSJ), Description: "温和细致、乐于奉献，默默守护身边人的周全与安宁。"},
	"ESTJ": {Key: "ESTJ", Name: "总经理", Image: mbtiBadge("ESTJ", mbtiColorSJ), Description: "果敢干练、重视秩序，擅长建立规则并高效推动事情落地。"},
	"ESFJ": {Key: "ESFJ", Name: "执政官", Image: mbtiBadge("ESFJ", mbtiColorSJ), Description: "热心周到、重视和谐，天生善于关照他人并营造融洽氛围。"},
	"ISTP": {Key: "ISTP", Name: "鉴赏家", Image: mbtiBadge("ISTP", mbtiColorSP), Description: "冷静灵活的实干家，热衷动手拆解与探索事物的运行原理。"},
	"ISFP": {Key: "ISFP", Name: "探险家", Image: mbtiBadge("ISFP", mbtiColorSP), Description: "安静而有艺术气质，用审美与温柔感知并拥抱生活的每个瞬间。"},
	"ESTP": {Key: "ESTP", Name: "企业家", Image: mbtiBadge("ESTP", mbtiColorSP), Description: "精力充沛、敢想敢做，在行动与挑战中享受当下的刺激。"},
	"ESFP": {Key: "ESFP", Name: "表演者", Image: mbtiBadge("ESFP", mbtiColorSP), Description: "天生的乐观派，热情有趣，享受把快乐带给身边的每一个人。"},
}

// GetZodiacConfig 获取星座配置，第二个返回值表示 key 是否存在。
func GetZodiacConfig(key string) (ZodiacConfig, bool) {
	cfg, ok := zodiacMap[key]
	return cfg, ok
}

// GetPersonalityConfig 获取性格配置，第二个返回值表示 key 是否存在。
func GetPersonalityConfig(key string) (PersonalityConfig, bool) {
	cfg, ok := personalityMap[key]
	return cfg, ok
}

// ListZodiacConfigs 按展示顺序返回全部星座配置。
func ListZodiacConfigs() []ZodiacConfig {
	list := make([]ZodiacConfig, 0, len(zodiacOrder))
	for _, key := range zodiacOrder {
		if cfg, ok := zodiacMap[key]; ok {
			list = append(list, cfg)
		}
	}
	return list
}

// ListPersonalityConfigs 按展示顺序返回全部性格配置。
func ListPersonalityConfigs() []PersonalityConfig {
	list := make([]PersonalityConfig, 0, len(personalityOrder))
	for _, key := range personalityOrder {
		if cfg, ok := personalityMap[key]; ok {
			list = append(list, cfg)
		}
	}
	return list
}
