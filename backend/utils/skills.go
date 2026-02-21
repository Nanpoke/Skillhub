package utils

import (
	"fmt"
	"strings"
)

// SkillsLeaderboard 技能排行榜数据
type SkillsLeaderboard struct {
	Type    string
	HasMore bool
	Items   []SkillsLeaderboardItem
}

type SkillsLeaderboardItem struct {
	Rank        int
	Name        string
	Author      string
	Installs    string
	InstallsInt int
	URL         string
	Change      string // Hot 排行榜的变化量，如 "+423"
}

// SkillsClient skills.sh 客户端
type SkillsClient struct {
	baseURL string
}

// NewSkillsClient 创建 skills.sh 客户端
func NewSkillsClient() *SkillsClient {
	return &SkillsClient{
		baseURL: "https://skills.sh",
	}
}

// FetchLeaderboard 获取排行榜数据
// 注意：skills.sh API 不公开，使用静态数据
// 数据更新时间：2025-02-10
func (c *SkillsClient) FetchLeaderboard(listType string) (*SkillsLeaderboard, error) {
	var items []SkillsLeaderboardItem

	switch listType {
	case "trending":
		items = getTrendingData()
	case "hot":
		items = getHotData()
	default: // all
		items = getAllTimeData()
	}

	return &SkillsLeaderboard{
		Type:    listType,
		HasMore: false,
		Items:   items,
	}, nil
}

// getAllTimeData All Time 排行榜数据（数据更新时间：2025-02-10）
func getAllTimeData() []SkillsLeaderboardItem {
	data := `1 ### find-skills vercel-labs/skills 173.1K
2 ### vercel-react-best-practices vercel-labs/agent-skills 114.4K
3 ### web-design-guidelines vercel-labs/agent-skills 86.5K
4 ### remotion-best-practices remotion-dev/skills 79.4K
5 ### frontend-design anthropics/skills 56.2K
6 ### vercel-composition-patterns vercel-labs/agent-skills 31.9K
7 ### agent-browser vercel-labs/agent-browser 28.4K
8 ### skill-creator anthropics/skills 28.0K
9 ### browser-use browser-use/browser-use 27.6K
10 ### vercel-react-native-skills vercel-labs/agent-skills 23.1K
11 ### ui-ux-pro-max nextlevelbuilder/ui-ux-pro-max-skill 20.1K
12 ### audit-website squirrelscan/skills 16.3K
13 ### seo-audit coreyhaines31/marketingskills 15.7K
14 ### brainstorming obra/superpowers 14.8K
15 ### supabase-postgres-best-practices supabase/agent-skills 14.4K
16 ### pdf anthropics/skills 11.9K
17 ### copywriting coreyhaines31/marketingskills 11.2K
18 ### pptx anthropics/skills 9.9K
19 ### better-auth-best-practices better-auth/skills 9.5K
20 ### docx anthropics/skills 9.2K
21 ### xlsx anthropics/skills 9.2K
22 ### next-best-practices vercel-labs/next-skills 9.0K
23 ### building-native-ui expo/skills 8.7K
24 ### marketing-psychology coreyhaines31/marketingskills 8.6K
25 ### systematic-debugging obra/superpowers 8.2K
26 ### webapp-testing anthropics/skills 8.1K
27 ### mcp-builder anthropics/skills 7.4K
28 ### programmatic-seo coreyhaines31/marketingskills 7.4K
29 ### writing-plans obra/superpowers 7.1K
30 ### test-driven-development obra/superpowers 7.0K
31 ### marketing-ideas coreyhaines31/marketingskills 6.7K
32 ### canvas-design anthropics/skills 6.4K
33 ### executing-plans obra/superpowers 6.3K
34 ### social-content coreyhaines31/marketingskills 6.2K
35 ### pricing-strategy coreyhaines31/marketingskills 6.2K
36 ### requesting-code-review obra/superpowers 5.8K
37 ### copy-editing coreyhaines31/marketingskills 5.7K
38 ### native-data-fetching expo/skills 5.6K
39 ### upgrading-expo expo/skills 5.6K
40 ### page-cro coreyhaines31/marketingskills 5.5K
41 ### launch-strategy coreyhaines31/marketingskills 5.5K
42 ### vue-best-practices hyf0/vue-skills 5.5K
43 ### subagent-driven-development obra/superpowers 5.4K
44 ### doc-coauthoring anthropics/skills 5.4K`
	return parseData(data, "")
}

// getTrendingData Trending 排行榜数据（数据更新时间：2025-02-10）
func getTrendingData() []SkillsLeaderboardItem {
	data := `1 ### find-skills vercel-labs/skills 11.8K
2 ### vercel-react-best-practices vercel-labs/agent-skills 4.1K
3 ### frontend-design anthropics/skills 3.1K
4 ### web-design-guidelines vercel-labs/agent-skills 3.0K
5 ### remotion-best-practices remotion-dev/skills 2.4K
6 ### vercel-composition-patterns vercel-labs/agent-skills 1.9K
7 ### browser-use browser-use/browser-use 1.9K
8 ### skill-creator anthropics/skills 1.6K
9 ### ui-ux-pro-max nextlevelbuilder/ui-ux-pro-max-skill 1.4K
10 ### agent-browser vercel-labs/agent-browser 1.3K
11 ### vercel-react-native-skills vercel-labs/agent-skills 1.3K
12 ### brainstorming obra/superpowers 1.2K
13 ### agent-tools 1nfsh/skills 880
14 ### supabase-postgres-best-practices supabase/agent-skills 740
15 ### audit-website squirrelscan/skills 738
16 ### seo-audit coreyhaines31/marketingskills 726
17 ### agent-browser 1nfsh/skills 691
18 ### next-best-practices vercel-labs/next-skills 688
19 ### pdf anthropics/skills 651
20 ### systematic-debugging obra/superpowers 643
21 ### agent-tools 1nference-sh/skills 589
22 ### copywriting coreyhaines31/marketingskills 576
23 ### pptx anthropics/skills 548
24 ### writing-plans obra/superpowers 547
25 ### docx anthropics/skills 503
26 ### test-driven-development obra/superpowers 477
27 ### xlsx anthropics/skills 473
28 ### requesting-code-review obra/superpowers 462
29 ### executing-plans obra/superpowers 462
30 ### webapp-testing anthropics/skills 442
31 ### marketing-psychology coreyhaines31/marketingskills 419
32 ### tailwind-design-system wshobson/agents 389
33 ### mcp-builder anthropics/skills 388
34 ### programmatic-seo coreyhaines31/marketingskills 386
35 ### receiving-code-review obra/superpowers 380
36 ### better-auth-best-practices better-auth/skills 380
37 ### interface-design dammyjay93/interface-design 368
38 ### agent-browser 1nference-sh/skills 361
39 ### using-superpowers obra/superpowers 356
40 ### writing-skills obra/superpowers 352
41 ### design-md google-labs-code/stitch-skills 349
42 ### react:components google-labs-code/stitch-skills 348
43 ### social-content coreyhaines31/marketingskills 345
44 ### using-git-worktrees obra/superpowers 344`
	return parseData(data, "")
}

// getHotData Hot 排行榜数据（数据更新时间：2025-02-10）
func getHotData() []SkillsLeaderboardItem {
	data := `1 ### agent-tools 1nference-sh/skills 423
2 ### agent-browser 1nference-sh/skills 261
3 ### python-sdk 1nference-sh/skills 51
4 ### ai-image-generation 1nference-sh/skills 47
5 ### vercel-react-best-practices vercel-labs/agent-skills 201
6 ### agent-ui 1nference-sh/skills 42
7 ### python-executor 1nference-sh/skills 36
8 ### vercel-composition-patterns vercel-labs/agent-skills 108
9 ### frontend-design anthropics/skills 160
10 ### javascript-sdk 1nference-sh/skills 30
11 ### ai-video-generation 1nference-sh/skills 28
12 ### twitter-automation 1nference-sh/skills 27
13 ### remotion-best-practices remotion-dev/skills 126
14 ### web-search 1nference-sh/skills 24
15 ### talking-head-production 1nference-sh/skills 23
16 ### product-hunt-launch 1nference-sh/skills 22
17 ### ai-voice-cloning 1nference-sh/skills 22
18 ### prompt-engineering 1nference-sh/skills 21
19 ### widgets-ui 1nference-sh/skills 21
20 ### pitch-deck-visuals 1nference-sh/skills 21
21 ### ai-product-photography 1nference-sh/skills 21
22 ### vercel-react-native-skills vercel-labs/agent-skills 68
23 ### agent-browser vercel-labs/agent-browser 66
24 ### video-ad-specs 1nference-sh/skills 20
25 ### newsletter-curation 1nference-sh/skills 20
26 ### ai-marketing-videos 1nference-sh/skills 20
27 ### social-media-carousel 1nference-sh/skills 20
28 ### customer-persona 1nference-sh/skills 19
29 ### youtube-thumbnail-design 1nference-sh/skills 19
30 ### character-design-sheet 1nference-sh/skills 18
31 ### twitter-thread-creation 1nference-sh/skills 18
32 ### seo-content-brief 1nference-sh/skills 18
33 ### content-repurposing 1nference-sh/skills 17
34 ### speech-to-text 1nference-sh/skills 17
35 ### app-store-screenshots 1nference-sh/skills 17
36 ### product-photography 1nference-sh/skills 16
37 ### book-cover-design 1nference-sh/skills 16
38 ### tools-ui 1nference-sh/skills 16
39 ### linkedin-content 1nference-sh/skills 16
40 ### chat-ui 1nference-sh/skills 15
41 ### landing-page-design 1nference-sh/skills 15
42 ### email-design 1nference-sh/skills 15
43 ### image-upscaling 1nference-sh/skills 15
44 ### data-visualization 1nference-sh/skills 15`
	return parseData(data, "hot")
}

// parseData 解析排行榜数据
func parseData(content string, dataType string) []SkillsLeaderboardItem {
	var items []SkillsLeaderboardItem
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// 格式: rank ### name author installs
		parts := strings.Fields(trimmed)
		if len(parts) >= 4 {
			rank := parseInt(parts[0])
			name := parts[2]
			author := parts[3]
			installs := parts[len(parts)-1] // 最后一列是安装数

			items = append(items, SkillsLeaderboardItem{
				Rank:        rank,
				Name:        name,
				Author:      author,
				Installs:    installs,
				InstallsInt: parseInstallsInt(installs),
				URL:         fmt.Sprintf("https://skills.sh/%s/%s", author, name),
			})
		}
	}

	return items
}

// parseInt 解析整数
func parseInt(s string) int {
	var result int
	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			result = result*10 + int(ch-'0')
		} else {
			break
		}
	}
	return result
}

// parseInstallsInt 将安装数字符串转换为整数
func parseInstallsInt(s string) int {
	multiplier := 1
	if strings.HasSuffix(s, "M") || strings.HasSuffix(s, "m") {
		multiplier = 1000000
		s = strings.TrimSuffix(s, "M")
		s = strings.TrimSuffix(s, "m")
	} else if strings.HasSuffix(s, "K") || strings.HasSuffix(s, "k") {
		multiplier = 1000
		s = strings.TrimSuffix(s, "K")
		s = strings.TrimSuffix(s, "k")
	}

	var value float64
	var afterDecimal bool
	var decimalDivisor float64 = 1
	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			if afterDecimal {
				decimalDivisor *= 10
			}
			value = value*10 + float64(ch-'0')
		} else if ch == '.' {
			afterDecimal = true
		}
	}

	value = value / decimalDivisor
	return int(value * float64(multiplier))
}

// FetchSkillDetail 获取 Skill 详情（返回 skills.sh URL）
func (c *SkillsClient) FetchSkillDetail(author, skillName string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/%s/%s", c.baseURL, author, skillName)

	return map[string]interface{}{
		"url":     url,
		"htmlUrl": url,
		"note":    "请使用浏览器查看完整详情",
	}, nil
}
