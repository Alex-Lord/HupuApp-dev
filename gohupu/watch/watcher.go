package watch

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/wudizhangzhi/HupuApp"
	"github.com/wudizhangzhi/HupuApp/gohupu/api"
	"github.com/wudizhangzhi/HupuApp/gohupu/api_utils"
	"github.com/wudizhangzhi/HupuApp/gohupu/logger"
	"github.com/wudizhangzhi/HupuApp/gohupu/message"
)

type Watcher struct {
	GameType        api.GameType
	TeamName        string
	PlayerName      string
	NotifyCmd       string
	liveActivityKey string
	Match           message.Match
	LastCommentId   string
	LastCommentTime string
	Th              *time.Ticker
	ThQueryMatch    *time.Ticker
	InterruptCh     chan os.Signal
	found           bool
}

func New(gameType api.GameType, teamName string, playerName string, notifyCmd string) *Watcher {
	return &Watcher{
		GameType:   gameType,
		TeamName:   teamName,
		PlayerName: playerName,
		NotifyCmd:  notifyCmd,
	}
}

// FindMatch 在今日比赛中查找包含指定球队名的正在进行的比赛
func (w *Watcher) FindMatch() error {
	matches, err := api_utils.GetMatchesFromDate(w.GameType)
	if err != nil {
		return fmt.Errorf("获取比赛列表失败: %w", err)
	}

	for _, m := range matches {
		if strings.Contains(m.HomeTeamName, w.TeamName) || strings.Contains(m.AwayTeamName, w.TeamName) {
			// 优先选择正在进行的比赛
			if m.MatchStatus != "NOT_STARTED" && m.MatchStatus != "COMPLETED" {
				w.Match = m
				return nil
			}
		}
	}

	// 没有正在进行的，选未开始的
	for _, m := range matches {
		if strings.Contains(m.HomeTeamName, w.TeamName) || strings.Contains(m.AwayTeamName, w.TeamName) {
			if m.MatchStatus == "NOT_STARTED" {
				w.Match = m
				return nil
			}
		}
	}

	// 都没有，选任何包含该队的比赛
	for _, m := range matches {
		if strings.Contains(m.HomeTeamName, w.TeamName) || strings.Contains(m.AwayTeamName, w.TeamName) {
			w.Match = m
			return nil
		}
	}

	return fmt.Errorf("未找到包含球队 %s 的比赛", w.TeamName)
}

func (w *Watcher) init() error {
	liveActivityKey, err := api_utils.GetLiveActivityKey(w.Match.MatchId)
	if err != nil {
		return fmt.Errorf("获取 liveActivityKey 失败: %w", err)
	}
	w.liveActivityKey = liveActivityKey
	logger.Info.Printf("liveActivityKey: %s\n", liveActivityKey)
	return nil
}

// notify 执行通知命令
func (w *Watcher) notify(msg string) {
	boldGreen := color.New(color.FgGreen, color.Bold).SprintFunc()
	fmt.Fprintf(color.Output, "\n🔔 %s\n\n", boldGreen(msg))

	if w.NotifyCmd != "" {
		cmd := exec.Command("sh", "-c", fmt.Sprintf("%s %s", w.NotifyCmd, fmt.Sprintf("\"%s\"", msg)))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			logger.Error.Printf("执行通知命令失败: %v", err)
		}
	}
}

// checkPlayer 检查消息中是否包含球员名（重点关注换人消息）
func (w *Watcher) checkPlayer(msg message.LiveMsg) bool {
	content := msg.Content
	if strings.Contains(content, w.PlayerName) {
		// 换人消息通常包含"换上"、"替换"、"上场"等关键词
		if strings.Contains(content, "换上") || strings.Contains(content, "替换") ||
			strings.Contains(content, "上场") || strings.Contains(content, "换下") {
			return true
		}
		// 即使不是典型换人消息，只要包含球员名也触发
		return true
	}
	return false
}

func (w *Watcher) isNewComment(msg message.LiveMsg) bool {
	if w.LastCommentTime == "" {
		return true
	}
	t, _ := HupuApp.GetTimeFromString(msg.Time)
	t_last, _ := HupuApp.GetTimeFromString(w.LastCommentTime)
	if t.After(t_last) {
		return true
	} else if t.Equal(t_last) {
		if msg.PreviousCommentId == w.LastCommentId {
			return true
		}
	}
	return false
}

func (w *Watcher) printLiveMsg(msg message.LiveMsg) {
	awayScore := fmt.Sprintf("%v", w.Match.AwayScore)
	homeScore := fmt.Sprintf("%v", w.Match.HomeScore)
	fmt.Fprintf(color.Output, "%s %s:%s %s | %s: %s\n",
		w.Match.AwayTeamName,
		awayScore,
		homeScore,
		w.Match.HomeTeamName,
		msg.NickName,
		msg.Content,
	)
}

func (w *Watcher) OnMatchUpdate() {
	match, err := api_utils.GetSingleMatch(w.Match.MatchId)
	if err != nil {
		logger.Error.Printf("更新比赛状态失败: %v", err)
		return
	}
	if match.MatchStatus == "COMPLETED" {
		if !w.found {
			w.notify(fmt.Sprintf("比赛结束，%s 未上场", w.PlayerName))
		} else {
			w.notify("比赛结束")
		}
		w.End()
	} else {
		w.Match = match
	}
}

func (w *Watcher) OnLiveMessage() {
	matchTextMsgs, err := api_utils.GetLiveMsgList(w.Match.MatchId, w.liveActivityKey, "")
	if err != nil {
		logger.Error.Printf("获取直播消息失败: %v", err)
		return
	}

	if len(matchTextMsgs) == 0 {
		w.OnMatchUpdate()
	} else {
		for _, msg := range matchTextMsgs {
			if w.isNewComment(msg) {
				w.printLiveMsg(msg)
				if !w.found && w.checkPlayer(msg) {
					w.found = true
					w.notify(fmt.Sprintf("%s 上场了！", w.PlayerName))
				}
				w.LastCommentId = msg.CommentId
				w.LastCommentTime = msg.Time
			}
		}
	}
}

func (w *Watcher) End() {
	logger.Info.Println("watch 结束")
	fmt.Println("----- 监控结束 -----")
	w.InterruptCh <- syscall.SIGQUIT
}

func (w *Watcher) Start() error {
	// 查找比赛
	if err := w.FindMatch(); err != nil {
		return err
	}

	bold := color.New(color.Bold).SprintFunc()
	fmt.Fprintf(color.Output, "----- 开始监控 %s 比赛: %s vs %s -----\n", bold(w.TeamName), w.Match.AwayTeamName, w.Match.HomeTeamName)
	fmt.Fprintf(color.Output, "等待球员 %s 上场...\n\n", bold(w.PlayerName))

	// 初始化
	w.InterruptCh = make(chan os.Signal, 1)
	w.Th = time.NewTicker(HupuApp.LIVE_HEART_BEAT_PERIOD * time.Second)
	w.ThQueryMatch = time.NewTicker(HupuApp.LIVE_HEART_BEAT_PERIOD * 2 * time.Second)

	signal.Notify(w.InterruptCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	if err := w.init(); err != nil {
		return err
	}

OutLoop:
	for {
		select {
		case <-w.InterruptCh:
			logger.Info.Println("收到中断信号，退出监控")
			break OutLoop
		case <-w.Th.C:
			w.OnLiveMessage()
		case <-w.ThQueryMatch.C:
			w.OnMatchUpdate()
		}
	}

	close(w.InterruptCh)
	return nil
}
