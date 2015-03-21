package main

import(
    "fmt"
	"flag"
	"log"
	"os"
	"util"
	"strconv"
	"statStruct"
	"time"
)

const (
	DEFAULT_SMALLEST_TIME_STRING = "1000-03-20T08:38:17.428370Z"
)

var svnXmlFile *string = flag.String("f", "", "svn log with xml format")
var svnDir *string = flag.String("d", "", "code working directory")


func main() {
	flag.Parse()

	//判断有没有指定svnlog.xml文件
	if *svnXmlFile == "" {
		log.Fatal("-f cannot be empty, -f svnlog.xml")
		return
	}

	//判断有没有指定svnlog.xml文件
	if *svnDir == "" {
		log.Fatal("-d cannot be empty, -d svnWorkDir")
		return
	}

	//判断文件是否存在
	if _,err := os.Stat(*svnXmlFile); os.IsNotExist(err) {
		log.Fatalf("svn log file '%s' not exists.", *svnXmlFile)
	}

	//获取svn root目录
	svnRoot, err := util.GetSvnRoot(*svnDir);

	svnXmlLogs, err := util.ParaseSvnXmlLog(*svnXmlFile)
//	fmt.Printf("%v", svnXmlLogs)
	util.CheckErr(err)

	authorTimeStats := make(statStruct.AuthorTimeStats)

	AuthorStats := make(map[string]statStruct.AuthorStat)

	for _, svnXmlLog := range svnXmlLogs.Logentry {
		newRev, _ := strconv.Atoi(svnXmlLog.Revision)
		fmt.Printf("svn diff on r%d ,\n", newRev)
		for _, path := range svnXmlLog.Paths {
			if path.Action == "M" && path.Kind == "file" {
				stdout, err := util.CallSvnDiff(newRev-1, newRev, svnRoot+path.Path)
				if err == nil {
					//fmt.Println("stdout ",stdout)
				} else {
					fmt.Println("err ", err.Error())
				}
				appendLines, removeLines, err := util.GetLineDiff(stdout)
				fmt.Printf("\t%s on r%d +%d -%d,\n", path.Path, newRev, appendLines, removeLines)
				if err == nil {
					//综合统计
					Author, ok := AuthorStats[svnXmlLog.Author]
					if ok {
						Author.AppendLines += appendLines
						Author.RemoveLines += removeLines
					} else {
						Author.AppendLines = appendLines
						Author.RemoveLines = removeLines
					}
					AuthorStats[svnXmlLog.Author] = Author
					//分时统计
					Author.AppendLines = appendLines
					Author.RemoveLines = removeLines
					authorTimeStat := make(statStruct.AuthorTimeStat)
					authorTimeStat[svnXmlLog.Date] = Author
					authorTimeStats[svnXmlLog.Author] = append(authorTimeStats[svnXmlLog.Author], authorTimeStat)
					//fmt.Println(appendLines, removeLines, AuthorStats)
				}
			}
		}
	}
	//输出结果
	ConsoleOutPutTable(AuthorStats)
	//输出按小时统计结果
	ConsoleOutPutHourTable(authorTimeStats)

}

//console输出结果
func ConsoleOutPutTable(AuthorStats map[string]statStruct.AuthorStat) {/*{{{*/
	fmt.Printf(" ==user== \t==lines==\n")
	for author, val := range AuthorStats {
		fmt.Printf("%10s\t%5d\n", author, val.AppendLines+val.RemoveLines)
	}
}/*}}}*/

//console按小时输出结果
func ConsoleOutPutHourTable(authorTimeStats statStruct.AuthorTimeStats) {/*{{{*/
	defaultSmallestTime, _ := time.Parse("2006-01-02T15:04:05Z", DEFAULT_SMALLEST_TIME_STRING)
	fmt.Printf(" ==user== \t==hour==\t==lines==\n")
	//先取到时间的区间值
	for authorName, Author := range authorTimeStats {
		var minTime time.Time
		var maxTime time.Time
		for _, t := range Author {
			for sTime,_ := range t {
				fmtTime, err := time.Parse("2006-01-02T15:04:05Z", sTime)
				util.CheckErr(err)
				if minTime.Before(defaultSmallestTime) || minTime.After(fmtTime) {
					minTime = fmtTime
				}
				if maxTime.Before(defaultSmallestTime) || maxTime.Before(fmtTime) {
					maxTime = fmtTime
				}
			}
		}
		//Todo 用户按时合并,去重
		//输出单个用户的数据
		for _, t := range Author {
			for sTime,Sval := range t {
				fmtTime, err := time.Parse("2006-01-02T15:04:05Z", sTime)
				util.CheckErr(err)
				fmt.Printf("%10s\t%5d\t%12d\n", authorName, fmtTime.Hour(), Sval.AppendLines+Sval.RemoveLines)
			}
		}
	}
}/*}}}*/
