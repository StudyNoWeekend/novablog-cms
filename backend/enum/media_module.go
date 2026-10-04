// 媒体库模块文件夹常量：业务模块 key 与中文文件夹名的映射（后端唯一权威）。
// 前端上传时仅传模块 key，文件夹中文名由此处决定；模块顶级文件夹在启动时播种、
// 上传时按需自动创建，保证前后端无需同步维护名称。
package enum

// 媒体模块 key（前端上传 module 参数的合法取值）。
const (
	MediaModuleArticle   = "article"
	MediaModuleRecipe    = "recipe"
	MediaModuleBook      = "book"
	MediaModuleGame      = "game"
	MediaModuleTechStack = "tech_stack"
	MediaModuleTravel    = "travel"
	MediaModuleEquipment = "equipment"
	MediaModuleProject   = "project"
	MediaModuleVideo     = "video"
	MediaModuleMusic     = "music"
	MediaModulePortfolio = "portfolio"
	MediaModuleProfile   = "profile"
)

// MediaModuleFolders 模块 key → 中文文件夹名。
var MediaModuleFolders = map[string]string{
	MediaModuleArticle:   "文章",
	MediaModuleRecipe:    "美食菜谱",
	MediaModuleBook:      "读书书架",
	MediaModuleGame:      "游戏库",
	MediaModuleTechStack: "技术栈",
	MediaModuleTravel:    "旅行攻略",
	MediaModuleEquipment: "个人设备",
	MediaModuleProject:   "项目经历",
	MediaModuleVideo:     "视频作品",
	MediaModuleMusic:     "音乐",
	MediaModulePortfolio: "摄影作品集",
	MediaModuleProfile:   "博主资料",
}

// MediaModuleOrder 模块文件夹的播种/自动归类顺序。
var MediaModuleOrder = []string{
	MediaModuleArticle,
	MediaModuleRecipe,
	MediaModuleBook,
	MediaModuleGame,
	MediaModuleTechStack,
	MediaModuleTravel,
	MediaModuleEquipment,
	MediaModuleProject,
	MediaModuleVideo,
	MediaModuleMusic,
	MediaModulePortfolio,
	MediaModuleProfile,
}
