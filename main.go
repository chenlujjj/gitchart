package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

const leftPaddingWidth = 4 // 显示"周几"的宽度
const cellWidth = 3        // 每个格子的宽度

func main() {
	months := flag.Int("month", 6, "since how many months ago")
	username := flag.String("username", "", "count commits by the 'username' author. If not set, would count all authors' commits")
	self := flag.Bool("self", false, "only count commits by myself. This flag would override 'username' flag")

	flag.Parse()
	repo, err := os.Getwd()
	if err != nil {
		fmt.Printf("ERROR: Getwd failed: %v", err)
		os.Exit(1)
	}

	now := time.Now()
	ago := now.AddDate(0, -1**months, 0)

	// 找到最近的周日作为起点
	gap := time.Sunday - ago.Local().Weekday()
	startDay := ago.AddDate(0, 0, int(gap))

	dayCommits, err := getDayCommits(repo, *months, *username, *self)
	if err != nil {
		fmt.Printf("ERROR: get commits of repository %s: %v", repo, err)
		os.Exit(1)
	}

	// 打印月份
	fmt.Printf(strings.Repeat(" ", leftPaddingWidth))
	var lastMonth time.Month
	for t := startDay; !t.After(now); t = t.AddDate(0, 0, 7) {
		if t.Month() != lastMonth {
			printMonth(t.Month())
			lastMonth = t.Month()
			continue
		}
		fmt.Printf(strings.Repeat(" ", cellWidth))
	}
	fmt.Println()

	// 打印每天的commits数量
	for i := time.Sunday; i < 7; i++ {
		printWeekday(i)
		for t := startDay.AddDate(0, 0, int(i)); !t.After(now); t = t.AddDate(0, 0, 7) {
			day := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
			commits := dayCommits[day]
			printCommits(commits)
		}
		fmt.Println()
	}
}

func printWeekday(d time.Weekday) {
	switch d {
	case time.Monday, time.Wednesday, time.Friday:
		fmt.Printf("%-4v", shortDayNames[d])
	default:
		fmt.Printf(strings.Repeat(" ", leftPaddingWidth))
	}
}

var shortDayNames = []string{
	"Sun",
	"Mon",
	"Tue",
	"Wed",
	"Thu",
	"Fri",
	"Sat",
}

func printMonth(m time.Month) {
	fmt.Printf(shortMonthNames[m-1])
}

var shortMonthNames = []string{
	"Jan",
	"Feb",
	"Mar",
	"Apr",
	"May",
	"Jun",
	"Jul",
	"Aug",
	"Sep",
	"Oct",
	"Nov",
	"Dec",
}

func getDayCommits(path string, lastMonths int, username string, self bool) (map[time.Time]int, error) {
	r, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return nil, err
	}
	if self {
		conf, err := r.Config()
		if err != nil {
			return nil, err
		}
		username = conf.User.Name
	}

	// ... retrieves the commit history
	since := time.Now().AddDate(0, -1*lastMonths, 0)
	cIter, err := r.Log(&git.LogOptions{Since: &since})
	if err != nil {
		return nil, err
	}

	// ... just iterates over the commits
	// 统计每天交了多少个commit
	dayCommits := make(map[time.Time]int)
	err = cIter.ForEach(func(c *object.Commit) error {
		if username != "" && c.Author.Name != username {
			return nil
		}
		when := c.Author.When
		day := time.Date(when.Year(), when.Month(), when.Day(), 0, 0, 0, 0, time.Local)
		if _, ok := dayCommits[day]; !ok {
			dayCommits[day] = 1
		}
		dayCommits[day]++
		return nil
	})
	if err != nil {
		return nil, err
	}

	return dayCommits, nil
}

func printCommits(commits int) {
	// fmt.Printf("%3d", commits)
	colorize(commits).Printf("%3d", commits)
}

// 格子的颜色
var (
	// 白色底色
	white = color.New(color.BgWhite)
	// 绿色
	boldGreen = color.New(color.BgGreen).Add(color.Bold)
	// 亮绿色
	boldHiGreen = color.New(color.BgHiGreen).Add(color.Bold)
	// 青色
	boldCyan = color.New(color.BgCyan).Add(color.Bold)
	// 亮青色
	boldHiCyan = color.New(color.BgHiCyan).Add(color.Bold)
)

func colorize(commits int) *color.Color {
	switch commits {
	case 0:
		return white
	case 1:
		return boldHiCyan
	case 2:
		return boldCyan
	case 3:
		return boldHiGreen
	default:
		return boldGreen
	}
}
