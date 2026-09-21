import {
  TravelStatus,
  getContinent,
  toTravelGuide,
  type TravelAttraction,
  type TravelFilters,
  type TravelGuide,
  type TravelGuideFormData,
  type TravelItineraryDay,
  type TravelReview,
} from '../types/travel'

const unsplash = (id: string): string =>
  `https://images.unsplash.com/photo-${id}?w=800&q=80`

const avatar = (id: string): string =>
  `https://images.unsplash.com/photo-${id}?w=100&h=100&fit=crop&q=80`

function createReview(
  index: number,
  username: string,
  rating: number,
  content: string,
  date: string,
): TravelReview {
  return {
    id: `review-${index}`,
    username,
    avatar: avatar('1438761681033-6461ffad8d80'),
    rating,
    content,
    date,
  }
}

function createAttraction(
  id: string,
  name: string,
  description: string,
  imageId: string,
  duration: string,
  location: string,
  latitude?: number,
  longitude?: number,
): TravelAttraction {
  return {
    id,
    name,
    description,
    image: unsplash(imageId),
    duration,
    location,
    latitude,
    longitude,
  }
}

function createDay(
  day: number,
  title: string,
  description: string,
  attractions: TravelAttraction[],
): TravelItineraryDay {
  return {
    day,
    title,
    description,
    attractions,
    attractionIds: attractions.map((attraction) => attraction.id),
  }
}

// 旧版 mock 数据未包含分类字段，导出前统一补默认值
type RawTravelGuide = Omit<TravelGuide, 'categoryId' | 'categoryName'>

const rawTravelGuides: RawTravelGuide[] = [
  {
    id: 'tg-001',
    title: '西藏拉萨：雪域圣城8日朝圣之旅',
    summary:
      '从布达拉宫到大昭寺，从纳木错到羊卓雍措，感受世界屋脊的信仰与自然之美。',
    coverImage: unsplash('1545569341-9eb8b30979d9'),
    status: TravelStatus.Published,
    destination: '西藏拉萨',
    region: 'china',
    days: 8,
    bestMonth: '5月-10月',
    viewCount: 12580,
    likeCount: 3420,
    rating: 4.8,
    reviewCount: 128,
    createdAt: '2024-03-15T08:30:00Z',
    attractions: [
      createAttraction('a-001', '布达拉宫', '世界文化遗产，藏传佛教的圣地', '1599700651', '3小时', '西藏拉萨市城关区北京中路35号', 29.6577, 91.1170),
      createAttraction('a-002', '大昭寺', '拉萨最古老的寺庙，供奉释迦牟尼12岁等身像', '1564507591', '2小时', '西藏拉萨市城关区八廓街2号', 29.6520, 91.1310),
      createAttraction('a-003', '纳木错', '西藏三大圣湖之一，海拔最高的咸水湖', '1500534311', '全天', '西藏拉萨市当雄县纳木错乡', 30.7333, 90.5333),
      createAttraction('a-004', '羊卓雍措', '高原蓝宝石，雪山环绕的绝美湖泊', '1476514521', '半天', '西藏山南市浪卡子县羊卓雍措景区', 29.1167, 90.4500),
      createAttraction('a-005', '八廓街', '拉萨最繁华的商业街，感受藏地市井生活', '1528164344', '2小时', '西藏拉萨市城关区八廓街', 29.6500, 91.1310),
      createAttraction('a-006', '色拉寺', '格鲁派六大寺庙之一', '1506748686', '3小时', '西藏拉萨市城关区色拉路1号', 29.6700, 91.1200),
      createAttraction('a-007', '甘丹寺', '格鲁派祖寺', '1469474964', '4小时', '西藏拉萨市达孜区章多乡强措村', 29.6400, 91.4800),
    ],
    itinerary: [
      createDay(1, '抵达拉萨', '适应高原气候，游览布达拉宫广场夜景', [
        createAttraction('a-001', '布达拉宫广场', '世界海拔最高的城市广场', '1599700651', '1小时', '西藏拉萨市城关区北京中路35号', 29.6577, 91.1170),
      ]),
      createDay(2, '圣城核心', '参观布达拉宫与大昭寺', [
        createAttraction('a-001', '布达拉宫', '世界文化遗产', '1599700651', '3小时', '西藏拉萨市城关区北京中路35号', 29.6577, 91.1170),
        createAttraction('a-002', '大昭寺', '藏传佛教圣地', '1564507591', '2小时', '西藏拉萨市城关区八廓街2号', 29.6520, 91.1310),
      ]),
      createDay(3, '八廓街转经', '体验藏族人文与手工艺品', [
        createAttraction('a-005', '八廓街', '藏地商业街', '1528164344', '3小时', '西藏拉萨市城关区八廓街', 29.6500, 91.1310),
      ]),
      createDay(4, '纳木错一日游', '朝圣天湖，远眺念青唐古拉山', [
        createAttraction('a-003', '纳木错', '天湖圣境', '1500534311', '全天', '西藏拉萨市当雄县纳木错乡', 30.7333, 90.5333),
      ]),
      createDay(5, '羊卓雍措', '探访高原蓝宝石', [
        createAttraction('a-004', '羊卓雍措', '绝美圣湖', '1476514521', '半天', '西藏山南市浪卡子县羊卓雍措景区', 29.1167, 90.4500),
      ]),
      createDay(6, '色拉寺辩经', '观看僧人辩经，了解藏传佛教', [
        createAttraction('a-006', '色拉寺', '格鲁派六大寺庙之一', '1506748686', '3小时', '西藏拉萨市城关区色拉路1号', 29.6700, 91.1200),
      ]),
      createDay(7, '甘丹寺朝圣', '前往拉萨东郊的甘丹寺', [
        createAttraction('a-007', '甘丹寺', '格鲁派祖寺', '1469474964', '4小时', '西藏拉萨市达孜区章多乡强措村', 29.6400, 91.4800),
      ]),
      createDay(8, '返程', '带着信仰与回忆离开拉萨', []),
    ],
    reviews: [
      createReview(1, 'wanderlust_猫', 5, '一生必去的地方，布达拉宫的震撼无法用语言形容。', '2024-04-02T10:00:00Z'),
      createReview(2, '背包客阿明', 4, '高反有点严重，但风景值得。纳木错太美了！', '2024-04-10T14:20:00Z'),
      createReview(3, '旅行达人Lisa', 5, '行程安排合理，文化体验很丰富。', '2024-05-01T09:15:00Z'),
      createReview(4, '高原行者', 4, '建议多留两天适应海拔。', '2024-05-18T16:45:00Z'),
    ],
  },
  {
    id: 'tg-002',
    title: '日本京都：古都风情5日漫游',
    summary:
      '穿梭于千本鸟居与古刹庭院，品味抹茶与怀石料理，体验千年古都的四季风雅。',
    coverImage: unsplash('1493976040374'),
    status: TravelStatus.Published,
    destination: '日本京都',
    region: 'japan',
    days: 5,
    bestMonth: '3月-5月、10月-11月',
    viewCount: 18920,
    likeCount: 5100,
    rating: 4.9,
    reviewCount: 215,
    createdAt: '2024-02-20T06:00:00Z',
    attractions: [
      createAttraction('a-008', '伏见稻荷大社', '千本鸟居绵延山间，京都最具代表性的神社', '1493976040374', '3小时', '日本京都府京都市伏见区深草薮之内町68番地', 34.9671, 135.7727),
      createAttraction('a-009', '清水寺', '悬空舞台俯瞰京都，春樱秋枫绝美', '1524413860', '2小时', '日本京都府京都市东山区清水1丁目294', 34.9949, 135.7850),
      createAttraction('a-010', '金阁寺', '金碧辉煌的禅宗寺院，倒映在镜湖池中', '1545562081', '1.5小时', '日本京都府京都市北区金阁寺町1番地', 35.0394, 135.7292),
      createAttraction('a-011', '岚山竹林', '幽静竹林小径，感受京都自然之美', '1470252649', '2小时', '日本京都府京都市右京区嵯峨野竹林ノ小径', 35.0170, 135.6716),
      createAttraction('a-012', '祗园', '艺伎文化发源地，传统茶屋林立', '1554791846', '2小时', '日本京都府京都市东山区祇园町北侧', 35.0036, 135.7776),
      createAttraction('a-013', '二条城', '德川家康府邸，日本国宝级城堡', '1505567748', '2小时', '日本京都府京都市中京区二条通堀川西入二条城町541', 35.0142, 135.7481),
      createAttraction('a-014', '天龙寺', '世界文化遗产庭园', '1518098261', '2小时', '日本京都府京都市右京区嵯峨天龙寺芒ノ马场町68', 35.0170, 135.6760),
    ],
    itinerary: [
      createDay(1, '抵达京都', '入住祗园附近，漫步花见小路', [
        createAttraction('a-012', '祗园', '艺伎文化区', '1554791846', '2小时', '日本京都府京都市东山区祇园町北侧', 35.0036, 135.7776),
      ]),
      createDay(2, '东山古寺', '清水寺与二年坂三年坂', [
        createAttraction('a-009', '清水寺', '悬空舞台', '1524413860', '3小时', '日本京都府京都市东山区清水1丁目294', 34.9949, 135.7850),
      ]),
      createDay(3, '伏见稻荷', '千本鸟居徒步登山', [
        createAttraction('a-008', '伏见稻荷大社', '千本鸟居', '1493976040374', '4小时', '日本京都府京都市伏见区深草薮之内町68番地', 34.9671, 135.7727),
      ]),
      createDay(4, '岚山一日游', '竹林、渡月桥与天龙寺', [
        createAttraction('a-011', '岚山竹林', '竹林小径', '1470252649', '3小时', '日本京都府京都市右京区嵯峨野竹林ノ小径', 35.0170, 135.6716),
        createAttraction('a-014', '天龙寺', '世界文化遗产庭园', '1518098261', '2小时', '日本京都府京都市右京区嵯峨天龙寺芒ノ马场町68', 35.0170, 135.6760),
      ]),
      createDay(5, '金阁寺与返程', '北山文化巡礼后返回', [
        createAttraction('a-010', '金阁寺', '金碧辉煌', '1545562081', '2小时', '日本京都府京都市北区金阁寺町1番地', 35.0394, 135.7292),
        createAttraction('a-013', '二条城', '幕府城堡', '1505567748', '2小时', '日本京都府京都市中京区二条通堀川西入二条城町541', 35.0142, 135.7481),
      ]),
    ],
    reviews: [
      createReview(5, '樱花控小雅', 5, '春天去真的太美了，清水寺的樱花如梦如幻。', '2024-03-25T08:30:00Z'),
      createReview(6, '日本通老陈', 5, '京都永远去不腻，每个季节都有不同的美。', '2024-04-05T11:00:00Z'),
      createReview(7, '美食家小美', 4, '怀石料理很精致，但价格不菲。', '2024-04-12T19:20:00Z'),
      createReview(8, '摄影爱好者', 5, '岚山竹林拍照超出片！', '2024-04-28T06:45:00Z'),
      createReview(9, '自由行玩家', 4, '交通很方便，建议买巴士一日券。', '2024-05-03T13:10:00Z'),
    ],
  },
  {
    id: 'tg-003',
    title: '泰国清迈：小城故事7日慢生活',
    summary:
      '在泰北玫瑰感受慢节奏生活，逛夜市、学泰菜、访寺庙，享受悠闲度假时光。',
    coverImage: unsplash('1506665532115-35661ca690c5'),
    status: TravelStatus.Published,
    destination: '泰国清迈',
    region: 'southeast-asia',
    days: 7,
    bestMonth: '11月-2月',
    viewCount: 15630,
    likeCount: 4280,
    rating: 4.6,
    reviewCount: 167,
    createdAt: '2024-01-10T03:00:00Z',
    attractions: [
      createAttraction('a-015', '双龙寺', '素贴山上的金色佛塔，俯瞰清迈全景', '1506665532115', '3小时', '泰国清迈府清迈市素贴山', 18.8000, 98.9286),
      createAttraction('a-016', '塔佩门', '古城东门，网红打卡地与鸽子广场', '1523906834', '1小时', '泰国清迈府清迈市塔佩路Tha Phae Road', 18.7873, 98.9910),
      createAttraction('a-017', '周日夜市', '世界十大夜市之一，手工艺品与美食天堂', '1555392339', '3小时', '泰国清迈府清迈市拉差路Ratchadamnoen Road', 18.7870, 98.9850),
      createAttraction('a-018', '清迈古城', '护城河环绕的古城，寺庙与咖啡馆林立', '1516483638', '半天', '泰国清迈府清迈市古城内', 18.7883, 98.9853),
      createAttraction('a-019', '大象自然公园', 'ethical elephant sanctuary，与大象亲密接触', '1504754524', '全天', '泰国清迈府湄登县', 19.0780, 98.8900),
      createAttraction('a-020', '宁曼路', '文艺小清新街区，网红咖啡店聚集地', '1497930172', '2小时', '泰国清迈府清迈市宁曼路Nimmanahaeminda Road', 18.8030, 98.9690),
      createAttraction('a-021', '女子监狱按摩', '传统泰式按摩', '1519824149', '2小时', '泰国清迈府清迈市帕抛拷路Tha Phae Road附近', 18.7880, 98.9900),
    ],
    itinerary: [
      createDay(1, '抵达清迈', '入住古城，塔佩门喂鸽子', [
        createAttraction('a-016', '塔佩门', '古城东门', '1523906834', '1小时', '泰国清迈府清迈市塔佩路Tha Phae Road', 18.7873, 98.9910),
      ]),
      createDay(2, '古城寺庙巡礼', '契迪龙寺与帕辛寺', [
        createAttraction('a-018', '清迈古城', '古城巡礼', '1516483638', '4小时', '泰国清迈府清迈市古城内', 18.7883, 98.9853),
      ]),
      createDay(3, '素贴山一日游', '双龙寺与蒲屏皇宫', [
        createAttraction('a-015', '双龙寺', '金色佛塔', '1506665532115', '4小时', '泰国清迈府清迈市素贴山', 18.8000, 98.9286),
      ]),
      createDay(4, '大象保护营', '与大象共度有意义的一天', [
        createAttraction('a-019', '大象自然公园', 'ethical sanctuary', '1504754524', '全天', '泰国清迈府湄登县', 19.0780, 98.8900),
      ]),
      createDay(5, '泰菜体验', '上午学做泰菜，下午逛宁曼路', [
        createAttraction('a-020', '宁曼路', '文艺街区', '1497930172', '3小时', '泰国清迈府清迈市宁曼路Nimmanahaeminda Road', 18.8030, 98.9690),
      ]),
      createDay(6, '周日集市', '周日夜市淘宝与美食', [
        createAttraction('a-017', '周日夜市', '夜市淘宝', '1555392339', '4小时', '泰国清迈府清迈市拉差路Ratchadamnoen Road', 18.7870, 98.9850),
      ]),
      createDay(7, '返程', '按摩放松后离开清迈', [
        createAttraction('a-021', '女子监狱按摩', '传统泰式按摩', '1519824149', '2小时', '泰国清迈府清迈市帕抛拷路Tha Phae Road附近', 18.7880, 98.9900),
      ]),
    ],
    reviews: [
      createReview(10, '慢生活爱好者', 5, '清迈太适合躺平了，物价便宜又舒服。', '2024-02-01T09:00:00Z'),
      createReview(11, '背包客小王', 4, '夜市很好逛，但要注意砍价。', '2024-02-08T15:30:00Z'),
      createReview(12, '亲子游妈妈', 5, '大象营孩子很喜欢，比骑大象有意义多了。', '2024-02-15T11:20:00Z'),
    ],
  },
  {
    id: 'tg-004',
    title: '云南大理：风花雪月4日治愈之旅',
    summary:
      '洱海骑行、古城漫步、苍山远眺，在苍洱之间寻找内心的宁静。',
    coverImage: unsplash('1469854523086-cc02fe5d8800'),
    status: TravelStatus.Draft,
    destination: '云南大理',
    region: 'china',
    days: 4,
    bestMonth: '3月-5月、9月-11月',
    viewCount: 8920,
    likeCount: 2150,
    rating: 4.5,
    reviewCount: 86,
    createdAt: '2024-05-20T02:00:00Z',
    attractions: [
      createAttraction('a-022', '洱海', '高原明珠，环湖骑行是最佳体验方式', '1469854523086', '全天', '云南省大理白族自治州大理市洱海', 25.7500, 100.1833),
      createAttraction('a-023', '大理古城', '南诏古国都城，白族建筑与人文风情', '1477959859', '3小时', '云南省大理白族自治州大理市大理镇', 25.6940, 100.1580),
      createAttraction('a-024', '苍山', '十九峰十八溪，缆车登顶俯瞰洱海', '1500534311', '半天', '云南省大理白族自治州大理市苍山地质公园', 25.6750, 100.0833),
      createAttraction('a-025', '崇圣寺三塔', '大理地标，千年佛教圣迹', '1518098261', '2小时', '云南省大理白族自治州大理市三塔路', 25.7280, 100.1540),
      createAttraction('a-026', '喜洲古镇', '白族民居博物馆，品尝喜洲粑粑', '1506748686', '3小时', '云南省大理白族自治州大理市喜洲镇', 25.8470, 100.1630),
      createAttraction('a-027', '双廊古镇', '洱海风光', '1476514521', '3小时', '云南省大理白族自治州大理市双廊镇', 25.9460, 100.1830),
    ],
    itinerary: [
      createDay(1, '抵达大理', '入住古城，夜游人民路', [
        createAttraction('a-023', '大理古城', '古城夜游', '1477959859', '2小时', '云南省大理白族自治州大理市大理镇', 25.6940, 100.1580),
      ]),
      createDay(2, '洱海西线', '才村码头骑行至喜洲', [
        createAttraction('a-022', '洱海', '环湖骑行', '1469854523086', '全天', '云南省大理白族自治州大理市洱海', 25.7500, 100.1833),
        createAttraction('a-026', '喜洲古镇', '白族古镇', '1506748686', '2小时', '云南省大理白族自治州大理市喜洲镇', 25.8470, 100.1630),
      ]),
      createDay(3, '苍山与三塔', '索道登苍山，下午参观三塔', [
        createAttraction('a-024', '苍山', '缆车登顶', '1500534311', '4小时', '云南省大理白族自治州大理市苍山地质公园', 25.6750, 100.0833),
        createAttraction('a-025', '崇圣寺三塔', '千年圣迹', '1518098261', '2小时', '云南省大理白族自治州大理市三塔路', 25.7280, 100.1540),
      ]),
      createDay(4, '双廊慢时光', '海边咖啡馆发呆，返程', [
        createAttraction('a-027', '双廊古镇', '洱海风光', '1476514521', '3小时', '云南省大理白族自治州大理市双廊镇', 25.9460, 100.1830),
      ]),
    ],
    reviews: [
      createReview(13, '文艺青年阿杰', 5, '洱海骑行太治愈了，风花雪月名不虚传。', '2024-05-28T07:30:00Z'),
      createReview(14, '情侣游小溪', 4, '古城商业化有点重，但风景真的很美。', '2024-06-03T12:00:00Z'),
      createReview(15, '独自旅行者', 4, '适合一个人发呆的地方。', '2024-06-10T18:45:00Z'),
    ],
  },
  {
    id: 'tg-005',
    title: '马尔代夫：蓝色天堂3日浪漫假期',
    summary:
      '一岛一酒店的私密天堂，水上屋、浮潜、日落巡航，享受极致海岛浪漫。',
    coverImage: unsplash('1514282401197-3e9e6e9e9e9e'),
    status: TravelStatus.Published,
    destination: '马尔代夫',
    region: 'south-asia',
    days: 3,
    bestMonth: '11月-4月',
    viewCount: 22100,
    likeCount: 6800,
    rating: 4.9,
    reviewCount: 312,
    createdAt: '2023-12-05T04:00:00Z',
    attractions: [
      createAttraction('a-028', '水上屋', '直接下海的梦幻住宿体验', '1514282401197', '全天', '马尔代夫南马累环礁度假村', 4.2000, 73.5000),
      createAttraction('a-029', '浮潜', '与海龟和热带鱼共舞', '1507525428', '3小时', '马尔代夫北马累环礁浮潜点', 4.3000, 73.5500),
      createAttraction('a-030', '日落巡航', '乘船追逐海豚与壮丽日落', '1500375599', '2小时', '马尔代夫马累附近海域', 4.1750, 73.5090),
      createAttraction('a-031', '无人沙洲', '拖尾沙滩上的私人野餐', '1507525428', '半天', '马尔代夫私人岛屿沙洲', 4.2500, 73.6000),
      createAttraction('a-032', '海底餐厅', '边用餐边观赏海洋生物', '1514362549', '2小时', '马尔代夫伦格里岛康莱德度假村', 3.9500, 73.4500),
    ],
    itinerary: [
      createDay(1, '抵达度假村', '水飞上岛，入住水上屋', [
        createAttraction('a-028', '水上屋', '梦幻住宿', '1514282401197', '全天', '马尔代夫南马累环礁度假村', 4.2000, 73.5000),
      ]),
      createDay(2, '海洋探索', '浮潜与日落巡航', [
        createAttraction('a-029', '浮潜', '海底世界', '1507525428', '3小时', '马尔代夫北马累环礁浮潜点', 4.3000, 73.5500),
        createAttraction('a-030', '日落巡航', '追逐日落', '1500375599', '2小时', '马尔代夫马累附近海域', 4.1750, 73.5090),
      ]),
      createDay(3, '离岛前休闲', '无人沙洲野餐与SPA', [
        createAttraction('a-031', '无人沙洲', '拖尾沙滩', '1507525428', '3小时', '马尔代夫私人岛屿沙洲', 4.2500, 73.6000),
      ]),
    ],
    reviews: [
      createReview(16, '蜜月夫妻', 5, '度蜜月首选，太浪漫了！', '2024-01-15T10:00:00Z'),
      createReview(17, '海岛控', 5, '海水颜色层次太美了，浮潜就能看到很多鱼。', '2024-01-22T14:30:00Z'),
      createReview(18, '奢华游玩家', 4, '很贵但值得，服务一流。', '2024-02-05T09:15:00Z'),
      createReview(19, '潜水爱好者', 5, '珊瑚保护得很好，海龟随处可见。', '2024-02-18T16:00:00Z'),
    ],
  },
  {
    id: 'tg-006',
    title: '法国巴黎：浪漫之都3日精华游',
    summary:
      '埃菲尔铁塔、卢浮宫、塞纳河游船，三日尽览巴黎最经典的浪漫地标。',
    coverImage: unsplash('1502602898657-3e91760cbb34'),
    status: TravelStatus.Published,
    destination: '法国巴黎',
    region: 'western-europe',
    days: 3,
    bestMonth: '4月-6月、9月-10月',
    viewCount: 19850,
    likeCount: 5620,
    rating: 4.7,
    reviewCount: 203,
    createdAt: '2024-02-14T07:00:00Z',
    attractions: [
      createAttraction('a-033', '埃菲尔铁塔', '巴黎地标，浪漫象征', '1511739001486', '3小时', '法国法兰西岛大区巴黎第七区战神广场', 48.8584, 2.2945),
      createAttraction('a-034', '卢浮宫', '世界四大博物馆之首', '1499856871', '4小时', '法国法兰西岛大区巴黎第一区卢浮宫 rue de Rivoli', 48.8606, 2.3376),
      createAttraction('a-035', '凯旋门', '香榭丽舍大街尽头的雄壮 monument', '1502602898657', '1.5小时', '法国法兰西岛大区巴黎第八区戴高乐广场', 48.8738, 2.2950),
      createAttraction('a-036', '塞纳河游船', '从水上欣赏巴黎两岸风光', '1502602898657', '2小时', '法国法兰西岛大区巴黎塞纳河沿岸', 48.8566, 2.3522),
      createAttraction('a-037', '蒙马特高地', '艺术家聚集地，圣心大教堂所在地', '1499856871', '3小时', '法国法兰西岛大区巴黎第十八区蒙马特', 48.8867, 2.3431),
      createAttraction('a-038', '奥赛博物馆', '印象派艺术宝库', '1545562081', '2小时', '法国法兰西岛大区巴黎第七区奥赛街62号', 48.8600, 2.3266),
    ],
    itinerary: [
      createDay(1, '经典地标', '埃菲尔铁塔与凯旋门', [
        createAttraction('a-033', '埃菲尔铁塔', '巴黎地标', '1511739001486', '3小时', '法国法兰西岛大区巴黎第七区战神广场', 48.8584, 2.2945),
        createAttraction('a-035', '凯旋门', '雄壮 monument', '1502602898657', '2小时', '法国法兰西岛大区巴黎第八区戴高乐广场', 48.8738, 2.2950),
      ]),
      createDay(2, '艺术殿堂', '卢浮宫与奥赛博物馆', [
        createAttraction('a-034', '卢浮宫', '艺术殿堂', '1499856871', '4小时', '法国法兰西岛大区巴黎第一区卢浮宫 rue de Rivoli', 48.8606, 2.3376),
        createAttraction('a-038', '奥赛博物馆', '印象派宝库', '1545562081', '2小时', '法国法兰西岛大区巴黎第七区奥赛街62号', 48.8600, 2.3266),
      ]),
      createDay(3, '浪漫收尾', '蒙马特高地与塞纳河游船', [
        createAttraction('a-037', '蒙马特高地', '艺术家高地', '1499856871', '3小时', '法国法兰西岛大区巴黎第十八区蒙马特', 48.8867, 2.3431),
        createAttraction('a-036', '塞纳河游船', '水上巴黎', '1502602898657', '2小时', '法国法兰西岛大区巴黎塞纳河沿岸', 48.8566, 2.3522),
      ]),
    ],
    reviews: [
      createReview(20, '浪漫主义者', 5, '巴黎真的很浪漫，铁塔夜景超美。', '2024-03-01T20:00:00Z'),
      createReview(21, '艺术爱好者', 5, '卢浮宫一天都逛不完，太震撼了。', '2024-03-10T11:30:00Z'),
      createReview(22, '欧洲行者', 4, '小偷有点多，要注意财物安全。', '2024-03-18T15:00:00Z'),
    ],
  },
  {
    id: 'tg-007',
    title: '新疆伊犁：草原花海15日自驾行',
    summary:
      '独库公路、赛里木湖、那拉提草原，用车轮丈量天山脚下的壮美画卷。',
    coverImage: unsplash('1469474964'),
    status: TravelStatus.Draft,
    destination: '新疆伊犁',
    region: 'china',
    days: 15,
    bestMonth: '6月-8月',
    viewCount: 10250,
    likeCount: 3180,
    rating: 4.8,
    reviewCount: 95,
    createdAt: '2024-04-01T05:00:00Z',
    attractions: [
      createAttraction('a-039', '赛里木湖', '大西洋最后一滴眼泪', '1469474964', '全天', '新疆维吾尔自治区博尔塔拉蒙古自治州博乐市', 44.6000, 81.2000),
      createAttraction('a-040', '那拉提草原', '空中草原，哈萨克牧民家园', '1506748686', '全天', '新疆维吾尔自治区伊犁哈萨克自治州新源县那拉提镇', 43.3800, 84.4700),
      createAttraction('a-041', '独库公路', '中国最美景观公路', '1500534311', '全天', '新疆维吾尔自治区独山子至库车段', 44.3000, 84.2000),
      createAttraction('a-042', '喀拉峻草原', '人体草原，摄影天堂', '1476514521', '半天', '新疆维吾尔自治区伊犁哈萨克自治州特克斯县喀拉峻景区', 43.0500, 82.4000),
      createAttraction('a-043', '霍城薰衣草', '东方普罗旺斯', '1490750960', '3小时', '新疆维吾尔自治区伊犁哈萨克自治州霍城县芦草沟镇', 44.0700, 80.4000),
      createAttraction('a-044', '巴音布鲁克', '九曲十八弯日落', '1507525428', '半天', '新疆维吾尔自治区巴音郭楞蒙古自治州和静县巴音布鲁克镇', 42.9000, 84.1500),
      createAttraction('a-045', '夏塔古道', '徒步穿越天山古道', '1506748686', '全天', '新疆维吾尔自治区伊犁哈萨克自治州昭苏县夏塔柯尔克孜族乡', 42.6500, 80.6000),
      createAttraction('a-046', 'S101省道', '丹霞公路', '1500534311', '全天', '新疆维吾尔自治区昌吉回族自治州至巴音郭楞蒙古自治州S101省道', 44.0000, 86.0000),
      createAttraction('a-047', '喀赞其', '蓝色民俗村', '1523906834', '3小时', '新疆维吾尔自治区伊犁哈萨克自治州伊宁市喀赞其民俗旅游区', 43.9100, 81.3500),
      createAttraction('a-048', '八卦城', '八卦布局城市', '1516483638', '2小时', '新疆维吾尔自治区伊犁哈萨克自治州特克斯县', 43.2300, 81.8400),
      createAttraction('a-049', '天山神秘大峡谷', '红色峡谷', '1470252649', '3小时', '新疆维吾尔自治区阿克苏地区库车市天山神秘大峡谷', 42.4000, 83.2000),
      createAttraction('a-050', '库车王府', '龟兹古国', '1528164344', '2小时', '新疆维吾尔自治区阿克苏地区库车市老城林基路街', 41.7200, 82.9700),
    ],
    itinerary: [
      createDay(1, '抵达乌鲁木齐', '集合取车，准备物资', []),
      createDay(2, 'S101省道', '百里丹霞地貌', [
        createAttraction('a-046', 'S101省道', '丹霞公路', '1500534311', '全天', '新疆维吾尔自治区昌吉回族自治州至巴音郭楞蒙古自治州S101省道', 44.0000, 86.0000),
      ]),
      createDay(3, '赛里木湖', '环湖自驾', [
        createAttraction('a-039', '赛里木湖', '高山湖泊', '1469474964', '全天', '新疆维吾尔自治区博尔塔拉蒙古自治州博乐市', 44.6000, 81.2000),
      ]),
      createDay(4, '霍城薰衣草', '紫色花海', [
        createAttraction('a-043', '霍城薰衣草', '紫色花海', '1490750960', '3小时', '新疆维吾尔自治区伊犁哈萨克自治州霍城县芦草沟镇', 44.0700, 80.4000),
      ]),
      createDay(5, '伊宁老城', '喀赞其民俗村', [
        createAttraction('a-047', '喀赞其', '蓝色民俗村', '1523906834', '3小时', '新疆维吾尔自治区伊犁哈萨克自治州伊宁市喀赞其民俗旅游区', 43.9100, 81.3500),
      ]),
      createDay(6, '夏塔古道', '雪山森林徒步', [
        createAttraction('a-045', '夏塔古道', '古道徒步', '1506748686', '全天', '新疆维吾尔自治区伊犁哈萨克自治州昭苏县夏塔柯尔克孜族乡', 42.6500, 80.6000),
      ]),
      createDay(7, '特克斯八卦城', '世界最大八卦城', [
        createAttraction('a-048', '八卦城', '八卦布局城市', '1516483638', '2小时', '新疆维吾尔自治区伊犁哈萨克自治州特克斯县', 43.2300, 81.8400),
      ]),
      createDay(8, '喀拉峻草原', '人体草原摄影', [
        createAttraction('a-042', '喀拉峻草原', '人体草原', '1476514521', '全天', '新疆维吾尔自治区伊犁哈萨克自治州特克斯县喀拉峻景区', 43.0500, 82.4000),
      ]),
      createDay(9, '那拉提草原', '空中草原', [
        createAttraction('a-040', '那拉提草原', '空中草原', '1506748686', '全天', '新疆维吾尔自治区伊犁哈萨克自治州新源县那拉提镇', 43.3800, 84.4700),
      ]),
      createDay(10, '独库公路北段', '翻越天山', [
        createAttraction('a-041', '独库公路', '景观大道', '1500534311', '全天', '新疆维吾尔自治区独山子至库车段', 44.3000, 84.2000),
      ]),
      createDay(11, '巴音布鲁克', '九曲十八弯', [
        createAttraction('a-044', '巴音布鲁克', '草原湿地', '1507525428', '半天', '新疆维吾尔自治区巴音郭楞蒙古自治州和静县巴音布鲁克镇', 42.9000, 84.1500),
      ]),
      createDay(12, '独库公路南段', '天山神秘大峡谷', [
        createAttraction('a-049', '天山神秘大峡谷', '红色峡谷', '1470252649', '3小时', '新疆维吾尔自治区阿克苏地区库车市天山神秘大峡谷', 42.4000, 83.2000),
      ]),
      createDay(13, '库车老城', '龟兹文化', [
        createAttraction('a-050', '库车王府', '龟兹古国', '1528164344', '2小时', '新疆维吾尔自治区阿克苏地区库车市老城林基路街', 41.7200, 82.9700),
      ]),
      createDay(14, '返程乌鲁木齐', '沿途风光', []),
      createDay(15, '结束行程', '还车返程', []),
    ],
    reviews: [
      createReview(23, '自驾狂人', 5, '新疆自驾天花板，风景太壮阔了！', '2024-06-20T08:00:00Z'),
      createReview(24, '摄影师老李', 5, '每帧都是壁纸，赛里木湖蓝得不像话。', '2024-06-28T14:20:00Z'),
      createReview(25, '草原控', 4, '路程比较长，但风景绝对值得。', '2024-07-05T10:10:00Z'),
    ],
  },
  {
    id: 'tg-008',
    title: '四川九寨沟：人间仙境6日探秘',
    summary:
      '五彩池、诺日朗瀑布、长海，探寻童话世界般的九寨归来不看水。',
    coverImage: unsplash('1506748686'),
    status: TravelStatus.Archived,
    destination: '四川九寨沟',
    region: 'china',
    days: 6,
    bestMonth: '9月-10月',
    viewCount: 14560,
    likeCount: 3890,
    rating: 4.9,
    reviewCount: 178,
    createdAt: '2023-10-15T01:00:00Z',
    attractions: [
      createAttraction('a-051', '五花海', '九寨沟最美海子，色彩斑斓', '1506748686', '2小时', '四川省阿坝藏族羌族自治州九寨沟县漳扎镇九寨沟景区内', 33.1700, 103.9000),
      createAttraction('a-052', '诺日朗瀑布', '中国最宽钙华瀑布', '1507525428', '1.5小时', '四川省阿坝藏族羌族自治州九寨沟县漳扎镇九寨沟景区内', 33.1700, 103.8900),
      createAttraction('a-053', '长海', '九寨沟海拔最高最大的海子', '1500534311', '1小时', '四川省阿坝藏族羌族自治州九寨沟县漳扎镇九寨沟景区内', 33.2000, 103.9100),
      createAttraction('a-054', '五彩池', '最小却最艳丽的海子', '1506748686', '1小时', '四川省阿坝藏族羌族自治州九寨沟县漳扎镇九寨沟景区内', 33.1900, 103.9100),
      createAttraction('a-055', '镜海', '水面如镜，倒影清晰', '1469474964', '1小时', '四川省阿坝藏族羌族自治州九寨沟县漳扎镇九寨沟景区内', 33.1800, 103.9000),
      createAttraction('a-056', '树正群海', '海子与瀑布相连的奇观', '1476514521', '2小时', '四川省阿坝藏族羌族自治州九寨沟县漳扎镇九寨沟景区内', 33.1600, 103.8800),
      createAttraction('a-057', '珍珠滩瀑布', '西游记取景地', '1507525428', '1.5小时', '四川省阿坝藏族羌族自治州九寨沟县漳扎镇九寨沟景区内', 33.1700, 103.8900),
    ],
    itinerary: [
      createDay(1, '抵达成都', '集合休整', []),
      createDay(2, '成都至九寨沟', '沿途岷江峡谷风光', []),
      createDay(3, '九寨沟日则沟', '五花海、珍珠滩、镜海', [
        createAttraction('a-051', '五花海', '最美海子', '1506748686', '2小时', '四川省阿坝藏族羌族自治州九寨沟县漳扎镇九寨沟景区内', 33.1700, 103.9000),
        createAttraction('a-057', '珍珠滩瀑布', '西游记取景地', '1507525428', '1.5小时', '四川省阿坝藏族羌族自治州九寨沟县漳扎镇九寨沟景区内', 33.1700, 103.8900),
        createAttraction('a-055', '镜海', '水面如镜', '1469474964', '1小时', '四川省阿坝藏族羌族自治州九寨沟县漳扎镇九寨沟景区内', 33.1800, 103.9000),
      ]),
      createDay(4, '九寨沟则查洼沟', '长海与五彩池', [
        createAttraction('a-053', '长海', '最大海子', '1500534311', '1.5小时', '四川省阿坝藏族羌族自治州九寨沟县漳扎镇九寨沟景区内', 33.2000, 103.9100),
        createAttraction('a-054', '五彩池', '最艳丽海子', '1506748686', '1小时', '四川省阿坝藏族羌族自治州九寨沟县漳扎镇九寨沟景区内', 33.1900, 103.9100),
      ]),
      createDay(5, '九寨沟树正沟', '树正群海与诺日朗瀑布', [
        createAttraction('a-056', '树正群海', '海子瀑布相连', '1476514521', '2小时', '四川省阿坝藏族羌族自治州九寨沟县漳扎镇九寨沟景区内', 33.1600, 103.8800),
        createAttraction('a-052', '诺日朗瀑布', '最宽瀑布', '1507525428', '1.5小时', '四川省阿坝藏族羌族自治州九寨沟县漳扎镇九寨沟景区内', 33.1700, 103.8900),
      ]),
      createDay(6, '返程成都', '带着九寨的记忆返程', []),
    ],
    reviews: [
      createReview(26, '自然风光迷', 5, '九寨沟的水真的是人间仙境，秋季最美。', '2023-11-02T09:30:00Z'),
      createReview(27, '退休夫妇', 5, '景区设施完善，老人也能轻松游览。', '2023-11-10T11:00:00Z'),
      createReview(28, '摄影达人', 4, '人多的时候要早起才能拍到好照片。', '2023-11-18T15:40:00Z'),
      createReview(29, '四川土著', 5, '作为四川人，九寨依然是我心中的第一。', '2023-11-25T08:20:00Z'),
    ],
  },
  {
    id: 'tg-009',
    title: '印度尼西亚巴厘岛：海岛风情5日度假',
    summary:
      '乌布梯田、海神庙、金巴兰日落，体验印尼最富盛名的度假天堂。',
    coverImage: unsplash('1537994473'),
    status: TravelStatus.Published,
    destination: '印度尼西亚巴厘岛',
    region: 'southeast-asia',
    days: 5,
    bestMonth: '4月-10月',
    viewCount: 13200,
    likeCount: 3650,
    rating: 4.5,
    reviewCount: 142,
    createdAt: '2024-03-05T03:30:00Z',
    attractions: [
      createAttraction('a-058', '乌布梯田', '德格拉朗梯田，绿意盎然', '1537994473', '3小时', '印度尼西亚巴厘省吉安雅县德格拉朗村', -8.4280, 115.2780),
      createAttraction('a-059', '海神庙', '海上岩石神庙，日落绝美', '1514282401197', '2小时', '印度尼西亚巴厘省塔巴南县贝拉坦湖畔', -8.6210, 115.0870),
      createAttraction('a-060', '金巴兰海滩', '全球最美日落海滩之一', '1507525428', '3小时', '印度尼西亚巴厘省巴东县金巴兰村', -8.7830, 115.1670),
      createAttraction('a-061', '圣泉寺', '千年圣泉沐浴祈福', '1506665532115', '2小时', '印度尼西亚巴厘省吉安雅县马努克亚村', -8.4290, 115.3300),
      createAttraction('a-062', '情人崖', '悬崖海景与猴子共舞', '1470252649', '2小时', '印度尼西亚巴厘省巴东县乌鲁瓦图', -8.8290, 115.0870),
      createAttraction('a-063', '蓝梦岛', '清澈海水与恶魔眼泪', '1506748686', '全天', '印度尼西亚巴厘省克隆孔县蓝梦岛', -8.6720, 115.4700),
      createAttraction('a-064', '库塔海滩', '冲浪圣地', '1507525428', '2小时', '印度尼西亚巴厘省巴东县库塔', -8.7180, 115.1680),
    ],
    itinerary: [
      createDay(1, '抵达巴厘岛', '入住库塔或水明漾', [
        createAttraction('a-064', '库塔海滩', '冲浪圣地', '1507525428', '2小时', '印度尼西亚巴厘省巴东县库塔', -8.7180, 115.1680),
      ]),
      createDay(2, '乌布文化', '梯田、圣猴林与皇宫', [
        createAttraction('a-058', '乌布梯田', '梯田风光', '1537994473', '3小时', '印度尼西亚巴厘省吉安雅县德格拉朗村', -8.4280, 115.2780),
      ]),
      createDay(3, '海神庙与情人崖', '西海岸日落之旅', [
        createAttraction('a-059', '海神庙', '海上神庙', '1514282401197', '2小时', '印度尼西亚巴厘省塔巴南县贝拉坦湖畔', -8.6210, 115.0870),
        createAttraction('a-062', '情人崖', '悬崖海景', '1470252649', '2小时', '印度尼西亚巴厘省巴东县乌鲁瓦图', -8.8290, 115.0870),
      ]),
      createDay(4, '蓝梦岛一日游', '跳岛浮潜', [
        createAttraction('a-063', '蓝梦岛', '跳岛浮潜', '1506748686', '全天', '印度尼西亚巴厘省克隆孔县蓝梦岛', -8.6720, 115.4700),
      ]),
      createDay(5, '金巴兰日落', '海鲜烧烤与返程', [
        createAttraction('a-060', '金巴兰海滩', '日落海滩', '1507525428', '3小时', '印度尼西亚巴厘省巴东县金巴兰村', -8.7830, 115.1670),
      ]),
    ],
    reviews: [
      createReview(30, '海岛度假控', 5, '巴厘岛性价比很高，SPA和美食都很棒。', '2024-03-20T10:00:00Z'),
      createReview(31, '冲浪爱好者', 4, '库塔的浪很适合初学者。', '2024-03-28T14:30:00Z'),
      createReview(32, '蜜月旅行', 5, '乌布的稻田别墅太惬意了。', '2024-04-05T09:15:00Z'),
    ],
  },
  {
    id: 'tg-010',
    title: '云南香格里拉：心中的日月12日深度游',
    summary:
      '普达措国家公园、松赞林寺、梅里雪山，走进詹姆斯·希尔顿笔下的理想国。',
    coverImage: unsplash('1500534311'),
    status: TravelStatus.Draft,
    destination: '云南香格里拉',
    region: 'china',
    days: 12,
    bestMonth: '5月-10月',
    viewCount: 7650,
    likeCount: 1980,
    rating: 4.6,
    reviewCount: 72,
    createdAt: '2024-04-18T02:00:00Z',
    attractions: [
      createAttraction('a-065', '普达措国家公园', '原始森林与高原湖泊', '1500534311', '全天', '云南省迪庆藏族自治州香格里拉市建塘镇红坡村', 27.8700, 99.9500),
      createAttraction('a-066', '松赞林寺', '小布达拉宫，云南最大藏传佛教寺院', '1564507591', '3小时', '云南省迪庆藏族自治州香格里拉市尼旺路3号', 27.8500, 99.7500),
      createAttraction('a-067', '梅里雪山', '藏区八大神山之首，日照金山', '1469474964', '全天', '云南省迪庆藏族自治州德钦县梅里雪山景区', 28.4000, 98.7000),
      createAttraction('a-068', '独克宗古城', '月光之城，世界最大转经筒', '1528164344', '2小时', '云南省迪庆藏族自治州香格里拉市独克宗古城', 27.8300, 99.7000),
      createAttraction('a-069', '纳帕海', '季节性湖泊，黑颈鹤栖息地', '1506748686', '半天', '云南省迪庆藏族自治州香格里拉市纳帕海自然保护区', 27.8600, 99.6200),
      createAttraction('a-070', '雨崩村', '梅里雪山脚下的隐秘村落', '1476514521', '2天', '云南省迪庆藏族自治州德钦县云岭乡雨崩村', 28.4000, 98.8500),
      createAttraction('a-071', '虎跳峡', '世界上最深的峡谷之一', '1470252649', '半天', '云南省丽江市玉龙纳西族自治县虎跳峡镇', 27.2000, 100.1000),
    ],
    itinerary: [
      createDay(1, '抵达丽江', '集合适应', []),
      createDay(2, '虎跳峡', '高路徒步体验', [
        createAttraction('a-071', '虎跳峡', '峡谷徒步', '1470252649', '半天', '云南省丽江市玉龙纳西族自治县虎跳峡镇', 27.2000, 100.1000),
      ]),
      createDay(3, '独克宗古城', '月光之城', [
        createAttraction('a-068', '独克宗古城', '转经筒', '1528164344', '3小时', '云南省迪庆藏族自治州香格里拉市独克宗古城', 27.8300, 99.7000),
      ]),
      createDay(4, '松赞林寺', '藏传佛教圣地', [
        createAttraction('a-066', '松赞林寺', '小布达拉宫', '1564507591', '3小时', '云南省迪庆藏族自治州香格里拉市尼旺路3号', 27.8500, 99.7500),
      ]),
      createDay(5, '普达措国家公园', '高原湖泊森林', [
        createAttraction('a-065', '普达措国家公园', '原始森林', '1500534311', '全天', '云南省迪庆藏族自治州香格里拉市建塘镇红坡村', 27.8700, 99.9500),
      ]),
      createDay(6, '纳帕海', '草原湖泊风光', [
        createAttraction('a-069', '纳帕海', '季节性湖泊', '1506748686', '半天', '云南省迪庆藏族自治州香格里拉市纳帕海自然保护区', 27.8600, 99.6200),
      ]),
      createDay(7, '前往飞来寺', '梅里雪山观景点', [
        createAttraction('a-067', '梅里雪山', '日照金山', '1469474964', '全天', '云南省迪庆藏族自治州德钦县梅里雪山景区', 28.4000, 98.7000),
      ]),
      createDay(8, '雨崩村', '徒步进入世外桃源', [
        createAttraction('a-070', '雨崩村', '隐秘村落', '1476514521', '全天', '云南省迪庆藏族自治州德钦县云岭乡雨崩村', 28.4000, 98.8500),
      ]),
      createDay(9, '雨崩神瀑', '朝圣徒步', [
        createAttraction('a-070', '雨崩神瀑', '神瀑朝圣', '1476514521', '全天', '云南省迪庆藏族自治州德钦县云岭乡雨崩村', 28.4000, 98.8500),
      ]),
      createDay(10, '出雨崩', '返回飞来寺', []),
      createDay(11, '返程香格里拉', '途中白马雪山', []),
      createDay(12, '结束行程', '从香格里拉返程', []),
    ],
    reviews: [
      createReview(33, '徒步爱好者', 5, '雨崩徒步虽然辛苦，但看到神瀑那一刻都值得。', '2024-05-10T08:00:00Z'),
      createReview(34, 'spiritual seeker', 4, '松赞林寺很庄严，氛围很好。', '2024-05-18T13:30:00Z'),
      createReview(35, '风光摄影师', 5, '梅里雪山日照金山，此生难忘。', '2024-05-25T06:45:00Z'),
    ],
  },
  {
    id: 'tg-011',
    title: '日本北海道：雪国童话9日漫游',
    summary:
      '小樽运河、富良野花海、登别温泉，四季皆有不同风情的北国大地。',
    coverImage: unsplash('1542640244'),
    status: TravelStatus.Published,
    destination: '日本北海道',
    region: 'japan',
    days: 9,
    bestMonth: '7月-8月、12月-2月',
    viewCount: 11300,
    likeCount: 3240,
    rating: 4.7,
    reviewCount: 118,
    createdAt: '2024-01-25T05:00:00Z',
    attractions: [
      createAttraction('a-072', '小樽运河', '北海道浪漫地标，煤油灯与雪景', '1542640244', '2小时', '日本北海道小樽市港町5番地', 43.1907, 140.9947),
      createAttraction('a-073', '富田农场', '薰衣草花田与哈密瓜', '1490750960', '3小时', '日本北海道中富良野町宫町15番地', 43.4250, 142.4570),
      createAttraction('a-074', '登别地狱谷', '火山地热景观与温泉街', '1470252649', '2小时', '日本北海道登别市登别温泉町', 42.5200, 140.8500),
      createAttraction('a-075', '旭山动物园', '企鹅散步与极地动物', '1504754524', '4小时', '日本北海道旭川市东8条南11丁目', 43.7700, 142.4900),
      createAttraction('a-076', '函馆山夜景', '世界三大夜景之一', '1511739001486', '2小时', '日本北海道函馆市函馆山', 41.7600, 140.7000),
      createAttraction('a-077', '札幌钟楼', '北海道首府地标', '1555392339', '1小时', '日本北海道札幌市中央区北1条西2丁目', 43.0620, 141.3540),
      createAttraction('a-078', '洞爷湖', '火山湖与温泉', '1469474964', '半天', '日本北海道虻田郡洞爷湖町洞爷湖温泉', 42.5600, 140.8500),
      createAttraction('a-079', '美瑛丘陵', '田园骑行', '1506748686', '半天', '日本北海道上川郡美瑛町', 43.6800, 142.4500),
    ],
    itinerary: [
      createDay(1, '抵达札幌', '大通公园与札幌钟楼', [
        createAttraction('a-077', '札幌钟楼', '城市地标', '1555392339', '1小时', '日本北海道札幌市中央区北1条西2丁目', 43.0620, 141.3540),
      ]),
      createDay(2, '小樽一日游', '运河与音乐盒堂', [
        createAttraction('a-072', '小樽运河', '浪漫运河', '1542640244', '3小时', '日本北海道小樽市港町5番地', 43.1907, 140.9947),
      ]),
      createDay(3, '旭川与动物园', '企鹅散步', [
        createAttraction('a-075', '旭山动物园', '极地动物', '1504754524', '4小时', '日本北海道旭川市东8条南11丁目', 43.7700, 142.4900),
      ]),
      createDay(4, '富良野花海', '薰衣草田', [
        createAttraction('a-073', '富田农场', '薰衣草', '1490750960', '3小时', '日本北海道中富良野町宫町15番地', 43.4250, 142.4570),
      ]),
      createDay(5, '美瑛骑行', '超广角之路', [
        createAttraction('a-079', '美瑛丘陵', '田园骑行', '1506748686', '半天', '日本北海道上川郡美瑛町', 43.6800, 142.4500),
      ]),
      createDay(6, '登别温泉', '地狱谷与温泉旅馆', [
        createAttraction('a-074', '登别地狱谷', '地热景观', '1470252649', '2小时', '日本北海道登别市登别温泉町', 42.5200, 140.8500),
      ]),
      createDay(7, '洞爷湖', '火山湖游船', [
        createAttraction('a-078', '洞爷湖', '火山湖', '1469474964', '半天', '日本北海道虻田郡洞爷湖町洞爷湖温泉', 42.5600, 140.8500),
      ]),
      createDay(8, '函馆', '五棱郭与夜景', [
        createAttraction('a-076', '函馆山夜景', '世界三大夜景', '1511739001486', '2小时', '日本北海道函馆市函馆山', 41.7600, 140.7000),
      ]),
      createDay(9, '返程', '从函馆机场返程', []),
    ],
    reviews: [
      createReview(36, '雪景爱好者', 5, '冬天的小樽运河像童话世界一样。', '2024-02-10T09:00:00Z'),
      createReview(37, '花海控', 5, '夏天的富良野薰衣草太梦幻了。', '2024-07-20T10:30:00Z'),
      createReview(38, '温泉达人', 4, '登别的温泉很正宗，硫磺味有点重。', '2024-02-18T19:00:00Z'),
      createReview(39, '亲子游爸爸', 5, '旭山动物园孩子玩得很开心。', '2024-07-28T14:15:00Z'),
    ],
  },
  {
    id: 'tg-012',
    title: '希腊圣托里尼：爱琴海蓝白梦境3日浪漫游',
    summary:
      '伊亚日落、蓝顶教堂、火山岛巡游，在地中海最浪漫的岛屿许下心愿。',
    coverImage: unsplash('1570077188670'),
    status: TravelStatus.Archived,
    destination: '希腊圣托里尼',
    region: 'southern-europe',
    days: 3,
    bestMonth: '5月-6月、9月-10月',
    viewCount: 24600,
    likeCount: 7920,
    rating: 4.8,
    reviewCount: 289,
    createdAt: '2023-09-10T06:00:00Z',
    attractions: [
      createAttraction('a-080', '伊亚日落', '世界最美日落之一', '1570077188670', '3小时', '希腊圣托里尼岛伊亚镇Oia', 36.4618, 25.3753),
      createAttraction('a-081', '蓝顶教堂', '圣托里尼标志性建筑', '1514282401197', '1小时', '希腊圣托里尼岛费拉镇Fira', 36.4220, 25.4270),
      createAttraction('a-082', '费拉小镇', '悬崖边的白色小镇', '1502602898657', '3小时', '希腊圣托里尼岛费拉镇Fira', 36.4200, 25.4320),
      createAttraction('a-083', '红沙滩', '火山岩形成的红色海滩', '1470252649', '2小时', '希腊圣托里尼岛阿克罗蒂里村附近', 36.3500, 25.4000),
      createAttraction('a-084', '火山岛巡游', '活火山与温泉游泳', '1507525428', '半天', '希腊圣托里尼岛火山岛Nea Kameni', 36.4000, 25.3700),
      createAttraction('a-085', '阿克罗蒂里遗址', '米诺斯文明遗址', '1528164344', '2小时', '希腊圣托里尼岛阿克罗蒂里考古遗址', 36.3500, 25.3900),
    ],
    itinerary: [
      createDay(1, '抵达圣托里尼', '费拉小镇探索', [
        createAttraction('a-082', '费拉小镇', '悬崖小镇', '1502602898657', '3小时', '希腊圣托里尼岛费拉镇Fira', 36.4200, 25.4320),
        createAttraction('a-081', '蓝顶教堂', '地标建筑', '1514282401197', '1小时', '希腊圣托里尼岛费拉镇Fira', 36.4220, 25.4270),
      ]),
      createDay(2, '火山岛与海滩', '火山巡游与红沙滩', [
        createAttraction('a-084', '火山岛巡游', '活火山', '1507525428', '半天', '希腊圣托里尼岛火山岛Nea Kameni', 36.4000, 25.3700),
        createAttraction('a-083', '红沙滩', '红色海滩', '1470252649', '2小时', '希腊圣托里尼岛阿克罗蒂里村附近', 36.3500, 25.4000),
      ]),
      createDay(3, '伊亚日落', '蓝白梦境与浪漫日落', [
        createAttraction('a-080', '伊亚日落', '世界最美日落', '1570077188670', '3小时', '希腊圣托里尼岛伊亚镇Oia', 36.4618, 25.3753),
      ]),
    ],
    reviews: [
      createReview(40, '浪漫情侣', 5, '度蜜月完美目的地，伊亚日落终生难忘。', '2023-10-01T18:30:00Z'),
      createReview(41, '欧洲背包客', 4, '很美但游客太多，建议淡季去。', '2023-10-08T11:00:00Z'),
      createReview(42, '摄影爱好者', 5, '蓝白配色怎么拍都好看。', '2023-10-15T16:45:00Z'),
      createReview(43, '美食旅行者', 4, '海鲜很棒，但岛上消费不低。', '2023-10-22T13:20:00Z'),
      createReview(44, '夕阳追逐者', 5, '为了伊亚日落，值得飞这么远。', '2023-11-01T19:00:00Z'),
    ],
  },
]

export const travelGuides: TravelGuide[] = rawTravelGuides.map((guide) => ({
  ...guide,
  categoryId: 'cat-travel',
  categoryName: '旅行攻略',
}))

function matchDaysRange(days: number, range: TravelFilters['days']): boolean {
  if (!range || range === 'all') return true
  switch (range) {
    case '1-3':
      return days >= 1 && days <= 3
    case '4-7':
      return days >= 4 && days <= 7
    case '8-14':
      return days >= 8 && days <= 14
    case '15+':
      return days >= 15
    default:
      return true
  }
}

export function filterTravelGuides(
  guides: TravelGuide[],
  filters: TravelFilters,
): TravelGuide[] {
  const { keyword, region, days, sort } = filters

  const normalizedKeyword = keyword?.trim().toLowerCase()

  const result = guides.filter((guide) => {
    if (normalizedKeyword) {
      const searchable = `${guide.title} ${guide.destination} ${guide.summary}`.toLowerCase()
      if (!searchable.includes(normalizedKeyword)) return false
    }

    if (region && region !== 'all') {
      const guideContinent = getContinent(guide.region)
      if (region !== guideContinent && guide.region !== region) return false
    }

    if (!matchDaysRange(guide.days, days)) return false

    return true
  })

  switch (sort) {
    case 'views':
      result.sort((a, b) => b.viewCount - a.viewCount)
      break
    case 'rating':
      result.sort((a, b) => b.rating - a.rating)
      break
    case 'likes':
      result.sort((a, b) => b.likeCount - a.likeCount)
      break
    case 'latest':
    default:
      result.sort(
        (a, b) =>
          new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime(),
      )
      break
  }

  return result
}

export function createEmptyGuide(): TravelGuideFormData {
  return {
    id: undefined,
    title: '',
    summary: '',
    coverImage: null,
    status: TravelStatus.Draft,
    destination: '',
    region: 'china',
    categoryId: undefined,
    days: 3,
    bestMonth: '',
    viewCount: 0,
    likeCount: 0,
    rating: 4.8,
    reviewCount: 0,
    createdAt: undefined,
    attractions: [],
    itinerary: Array.from({ length: 3 }, (_, index) => ({
      day: index + 1,
      title: '',
      description: '',
      attractions: [],
      attractionIds: [],
    })),
    reviews: [],
  }
}

function generateNewGuideId(): string {
  const maxId = travelGuides.reduce((max, guide) => {
    const match = guide.id.match(/tg-(\d+)/)
    if (!match) return max
    return Math.max(max, Number.parseInt(match[1], 10))
  }, 0)
  return `tg-${String(maxId + 1).padStart(3, '0')}`
}

export function saveGuideToMock(guide: TravelGuideFormData): TravelGuide {
  const savedGuide = toTravelGuide({
    ...guide,
    createdAt: guide.createdAt ?? new Date().toISOString(),
    status: guide.status ?? TravelStatus.Draft,
  })

  if (guide.id) {
    const index = travelGuides.findIndex((item) => item.id === guide.id)
    if (index >= 0) {
      travelGuides[index] = savedGuide
      return savedGuide
    }
  }

  const newGuide: TravelGuide = {
    ...savedGuide,
    id: generateNewGuideId(),
  }
  travelGuides.push(newGuide)
  return newGuide
}
