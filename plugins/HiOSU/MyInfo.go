package HiOSU

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/RicheyJang/PaimengBot/manager"
	"github.com/RicheyJang/PaimengBot/utils"
	"github.com/RicheyJang/PaimengBot/utils/client"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
)

type User struct {
	UserID       string `json:"user_id"`         // 数字ID  0
	UserName     string `json:"username"`        // 名称 1
	JoinDate     string `json:"join_date"`       // 加入时间  2
	Country      string `json:"country"`         // 国家   18
	GlobalRank   string `json:"pp_rank"`         // 国际PP排名  9
	CountryRank  string `json:"pp_country_rank"` // 国内的PP排名  20
	PP           string `json:"pp_raw"`          // PP总数
	Accuracy     string `json:"accuracy"`        // 准确率
	CountRankSs  string `json:"count_rank_ss"`
	CountRankSSH string `json:"count_rank_ssh"`
	CountRankS   string `json:"count_rank_s"`
	CountRankSh  string `json:"count_rank_sh"`
	CountRankA   string `json:"count_rank_a"`
}

func MineInfoHandler(ctx *zero.Ctx) {
	appid := proxy.GetConfigInt64("appid")
	secret := proxy.GetConfigString("secret")
	if appid == 0 || secret == "" {
		ctx.Send("管理员尚未配置API KEY，快去催他！")
		return
	}
	// 查询数据表中用户绑定信息
	OSUid := GetOsuid(ctx.Event.UserID)
	if len(OSUid) == 0 {
		ctx.Send("没有绑定OSU账号的说\n(○｀ 3′○)")
		return
	}
	// 获取用户要查询的模式
	model := strings.TrimSpace(utils.GetArgs(ctx))
	if model != "1" && model != "2" && model != "3" {
		model = "0"
	}
	// Model := GetModel(model)
	var user ApiUser
	var scores []Score
	c := client.NewHttpClient(&client.HttpOptions{
		TryTime: 3,
	})
	token := refreshToken(appid, secret)
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://osu.ppy.sh/api/v2/users/%s/scores/best", OSUid), nil)
	// req.URL.RawQuery = "mode=" + Model
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	response, err := c.Do(req)
	if err != nil {
		log.Errorf("GetBest err: %v", err)
		ctx.Send("失败了...")
		return
	}
	defer response.Body.Close()
	all, _ := io.ReadAll(response.Body)
	if err := json.NewDecoder(bytes.NewReader(all)).Decode(&scores); err != nil {
		log.Errorf("GetBest err: %v %v", err, string(all))
		ctx.Send("失败了...")
		return
	}
	req, _ = http.NewRequest("GET", fmt.Sprintf("https://osu.ppy.sh/api/v2/users/%s", OSUid), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	response, err = c.Do(req)
	if err != nil {
		log.Errorf("GetBest err: %v", err)
		ctx.Send("失败了...")
		return
	}
	all, _ = io.ReadAll(response.Body)
	if err := json.NewDecoder(bytes.NewReader(all)).Decode(&user); err != nil {
		log.Errorf("GetBest err: %v %v", err, string(all))
		ctx.Send("失败了...")
		return
	}
	defer response.Body.Close()
	img, err := drawUserInfo(user, scores)
	if err != nil {
		log.Errorf("GetBest err: %v", err)
		ctx.Send("失败了...")
		return
	}
	auto, err := img.GenMessageAuto()
	if err != nil {
		log.Errorf("GetBest err: %v", err)
		ctx.Send("失败了...")
		return
	}
	ctx.Send(auto)
}

var cache = struct {
	Token  string
	Expire time.Time
}{
	Token:  "",
	Expire: time.Now(),
}

func refreshToken(appid int64, secret string) string {
	if cache.Token != "" && cache.Expire.After(time.Now()) {
		return cache.Token
	}
	c := client.NewHttpClient(nil)
	r, err := c.PostJson("https://osu.ppy.sh/oauth/token", map[string]interface{}{
		"client_id":     appid,
		"client_secret": secret,
		"grant_type":    "client_credentials",
		"scope":         "public",
	})
	if err != nil {
		return ""
	}

	cache.Token = r.Get("access_token").Str
	cache.Expire = time.Now().Add(time.Duration(r.Get("expires_in").Int()) * time.Second)
	err = proxy.GetLevelDB().Put([]byte("osu_token"), []byte(cache.Token), nil)
	err = proxy.GetLevelDB().Put([]byte("osu_expire"), []byte(strconv.FormatInt(cache.Expire.UnixMilli(), 10)), nil)
	if err != nil {
		log.Errorf("PutLevelDB err: %v", err)
	}
	return r.Get("access_token").Str
}

type ApiUser struct {
	AvatarURL     string      `json:"avatar_url"`
	CountryCode   string      `json:"country_code"`
	DefaultGroup  string      `json:"default_group"`
	ID            int         `json:"id"`
	IsActive      bool        `json:"is_active"`
	IsBot         bool        `json:"is_bot"`
	IsDeleted     bool        `json:"is_deleted"`
	IsOnline      bool        `json:"is_online"`
	IsSupporter   bool        `json:"is_supporter"`
	LastVisit     time.Time   `json:"last_visit"`
	PmFriendsOnly bool        `json:"pm_friends_only"`
	ProfileColour interface{} `json:"profile_colour"`
	Username      string      `json:"username"`
	CoverURL      string      `json:"cover_url"`
	Discord       interface{} `json:"discord"`
	HasSupported  bool        `json:"has_supported"`
	Interests     interface{} `json:"interests"`
	JoinDate      time.Time   `json:"join_date"`
	Location      interface{} `json:"location"`
	MaxBlocks     int         `json:"max_blocks"`
	MaxFriends    int         `json:"max_friends"`
	Occupation    interface{} `json:"occupation"`
	Playmode      string      `json:"playmode"`
	Playstyle     []string    `json:"playstyle"`
	PostCount     int         `json:"post_count"`
	ProfileHue    interface{} `json:"profile_hue"`
	ProfileOrder  []string    `json:"profile_order"`
	Title         interface{} `json:"title"`
	TitleURL      interface{} `json:"title_url"`
	Twitter       interface{} `json:"twitter"`
	Website       interface{} `json:"website"`
	Country       struct {
		Code string `json:"code"`
		Name string `json:"name"`
	} `json:"country"`
	Cover struct {
		CustomURL interface{} `json:"custom_url"`
		URL       string      `json:"url"`
		ID        string      `json:"id"`
	} `json:"cover"`
	Kudosu struct {
		Available int `json:"available"`
		Total     int `json:"total"`
	} `json:"kudosu"`
	AccountHistory          []interface{} `json:"account_history"`
	ActiveTournamentBanner  interface{}   `json:"active_tournament_banner"`
	ActiveTournamentBanners []interface{} `json:"active_tournament_banners"`
	Badges                  []interface{} `json:"badges"`
	BeatmapPlaycountsCount  int           `json:"beatmap_playcounts_count"`
	CommentsCount           int           `json:"comments_count"`
	DailyChallengeUserStats struct {
		DailyStreakBest     int       `json:"daily_streak_best"`
		DailyStreakCurrent  int       `json:"daily_streak_current"`
		LastUpdate          time.Time `json:"last_update"`
		LastWeeklyStreak    time.Time `json:"last_weekly_streak"`
		Playcount           int       `json:"playcount"`
		Top10PPlacements    int       `json:"top_10p_placements"`
		Top50PPlacements    int       `json:"top_50p_placements"`
		UserID              int       `json:"user_id"`
		WeeklyStreakBest    int       `json:"weekly_streak_best"`
		WeeklyStreakCurrent int       `json:"weekly_streak_current"`
	} `json:"daily_challenge_user_stats"`
	FavouriteBeatmapsetCount int           `json:"favourite_beatmapset_count"`
	FollowerCount            int           `json:"follower_count"`
	GraveyardBeatmapsetCount int           `json:"graveyard_beatmapset_count"`
	Groups                   []interface{} `json:"groups"`
	GuestBeatmapsetCount     int           `json:"guest_beatmapset_count"`
	LovedBeatmapsetCount     int           `json:"loved_beatmapset_count"`
	MappingFollowerCount     int           `json:"mapping_follower_count"`
	MonthlyPlaycounts        []struct {
		StartDate string `json:"start_date"`
		Count     int    `json:"count"`
	} `json:"monthly_playcounts"`
	NominatedBeatmapsetCount int `json:"nominated_beatmapset_count"`
	Page                     struct {
		HTML string `json:"html"`
		Raw  string `json:"raw"`
	} `json:"page"`
	PendingBeatmapsetCount int           `json:"pending_beatmapset_count"`
	PreviousUsernames      []interface{} `json:"previous_usernames"`
	RankHighest            struct {
		Rank      int       `json:"rank"`
		UpdatedAt time.Time `json:"updated_at"`
	} `json:"rank_highest"`
	RankedBeatmapsetCount int           `json:"ranked_beatmapset_count"`
	ReplaysWatchedCounts  []interface{} `json:"replays_watched_counts"`
	ScoresBestCount       int           `json:"scores_best_count"`
	ScoresFirstCount      int           `json:"scores_first_count"`
	ScoresPinnedCount     int           `json:"scores_pinned_count"`
	ScoresRecentCount     int           `json:"scores_recent_count"`
	Statistics            struct {
		Count100  int `json:"count_100"`
		Count300  int `json:"count_300"`
		Count50   int `json:"count_50"`
		CountMiss int `json:"count_miss"`
		Level     struct {
			Current  int `json:"current"`
			Progress int `json:"progress"`
		} `json:"level"`
		GlobalRank             int         `json:"global_rank"`
		GlobalRankExp          interface{} `json:"global_rank_exp"`
		Pp                     float64     `json:"pp"`
		PpExp                  int         `json:"pp_exp"`
		RankedScore            int         `json:"ranked_score"`
		HitAccuracy            float64     `json:"hit_accuracy"`
		PlayCount              int         `json:"play_count"`
		PlayTime               int         `json:"play_time"`
		TotalScore             int         `json:"total_score"`
		TotalHits              int         `json:"total_hits"`
		MaximumCombo           int         `json:"maximum_combo"`
		ReplaysWatchedByOthers int         `json:"replays_watched_by_others"`
		IsRanked               bool        `json:"is_ranked"`
		GradeCounts            struct {
			Ss  int `json:"ss"`
			SSH int `json:"ssh"`
			S   int `json:"s"`
			Sh  int `json:"sh"`
			A   int `json:"a"`
		} `json:"grade_counts"`
		CountryRank int `json:"country_rank"`
		Rank        struct {
			Country int `json:"country"`
		} `json:"rank"`
	} `json:"statistics"`
	SupportLevel     int         `json:"support_level"`
	Team             interface{} `json:"team"`
	UserAchievements []struct {
		AchievedAt    time.Time `json:"achieved_at"`
		AchievementID int       `json:"achievement_id"`
	} `json:"user_achievements"`
	RankHistory struct {
		Mode string `json:"mode"`
		Data []int  `json:"data"`
	} `json:"rank_history"`
	RankedAndApprovedBeatmapsetCount int `json:"ranked_and_approved_beatmapset_count"`
	UnrankedBeatmapsetCount          int `json:"unranked_beatmapset_count"`
}
type Score struct {
	Accuracy   float64       `json:"accuracy"`
	BestID     int64         `json:"best_id"`
	CreatedAt  time.Time     `json:"created_at"`
	ID         int64         `json:"id"`
	MaxCombo   int           `json:"max_combo"`
	Mode       string        `json:"mode"`
	ModeInt    int           `json:"mode_int"`
	Mods       []interface{} `json:"mods"`
	Passed     bool          `json:"passed"`
	Perfect    bool          `json:"perfect"`
	Pp         float64       `json:"pp"`
	Rank       string        `json:"rank"`
	Replay     bool          `json:"replay"`
	Score      int           `json:"score"`
	Statistics struct {
		Count100  int         `json:"count_100"`
		Count300  int         `json:"count_300"`
		Count50   int         `json:"count_50"`
		CountGeki interface{} `json:"count_geki"`
		CountKatu interface{} `json:"count_katu"`
		CountMiss int         `json:"count_miss"`
	} `json:"statistics"`
	Type                  string `json:"type"`
	UserID                int    `json:"user_id"`
	CurrentUserAttributes struct {
		Pin interface{} `json:"pin"`
	} `json:"current_user_attributes"`
	Beatmap struct {
		BeatmapsetID     int     `json:"beatmapset_id"`
		DifficultyRating float64 `json:"difficulty_rating"`
		ID               int     `json:"id"`
		Mode             string  `json:"mode"`
		Status           string  `json:"status"`
		TotalLength      int     `json:"total_length"`
		UserID           int     `json:"user_id"`
		Version          string  `json:"version"`
		Accuracy         float64 `json:"accuracy"`
		Ar               float64 `json:"ar"`
		Bpm              float64 `json:"bpm"`
		Convert          bool    `json:"convert"`
		CountCircles     int     `json:"count_circles"`
		CountSliders     int     `json:"count_sliders"`
		CountSpinners    int     `json:"count_spinners"`
		HitLength        int     `json:"hit_length"`
		IsScoreable      bool    `json:"is_scoreable"`
		ModeInt          int     `json:"mode_int"`
		Ranked           int     `json:"ranked"`
	} `json:"beatmap"`
	Beatmapset struct {
		Title         string `json:"title"`
		Artist        string `json:"artist"`
		ArtistUnicode string `json:"artist_unicode"`
		Covers        struct {
			Cover       string `json:"cover"`
			Cover2X     string `json:"cover@2x"`
			Card        string `json:"card"`
			Card2X      string `json:"card@2x"`
			List        string `json:"list"`
			List2X      string `json:"list@2x"`
			Slimcover   string `json:"slimcover"`
			Slimcover2X string `json:"slimcover@2x"`
		} `json:"covers"`
		ID int `json:"id"`
	} `json:"beatmapset"`
	User struct {
		AvatarURL     string      `json:"avatar_url"`
		CountryCode   string      `json:"country_code"`
		DefaultGroup  string      `json:"default_group"`
		ID            int         `json:"id"`
		IsActive      bool        `json:"is_active"`
		IsBot         bool        `json:"is_bot"`
		IsDeleted     bool        `json:"is_deleted"`
		IsOnline      bool        `json:"is_online"`
		IsSupporter   bool        `json:"is_supporter"`
		LastVisit     time.Time   `json:"last_visit"`
		PmFriendsOnly bool        `json:"pm_friends_only"`
		ProfileColour interface{} `json:"profile_colour"`
		Username      string      `json:"username"`
	} `json:"user"`
	Weight struct {
		Percentage float64 `json:"percentage"`
		Pp         float64 `json:"pp"`
	} `json:"weight"`
}

func GetModel(ModelNumber string) string {
	switch ModelNumber {
	case "1":
		return "taiko"
	case "2":
		return "fruits"
	case "3":
		return "mania"
	default:
		return "osu"
	}
}

func getModelImage(Model string) (image.Image, error) {
	switch Model {
	case "Osu!Mania":
		return manager.DecodeStaticImage("HiOSU/Model/mania_20x20.png")
	case "Taiko":
		return manager.DecodeStaticImage("HiOSU/Model/taiko_20x20.png")
	case "CtB":
		return manager.DecodeStaticImage("HiOSU/Model/catch_20x20.png")
	default:
		return manager.DecodeStaticImage("HiOSU/Model/std_20x20.png")
	}
}
