package workers

import (
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/golang/glog"
	"github.com/hu17889/go_spider/core/common/page"
	"github.com/hu17889/go_spider/core/pipeline"
	"github.com/hu17889/go_spider/core/spider"
	"github.com/spider-docker/events"
	"github.com/spider-docker/models"
	"os"
	"regexp"
	"strconv"
	"strings"
	"math/rand"
)

const ()

var (
	/*
	** 存储当前爬取页面的url和layer
	 */
	crawlUrl models.CrawlUrl

	/*
	 **用来存储要扒取这页中产生的url以及url对应的等级
	 */
	Urls []string
	UrlsLevel []int64
	i int


	headerJson = [10]string {
		"./header_1.json",
		"./header_2.json",
		"./header_3.json",
		"./header_4.json",
		"./header_5.json",
		"./header_6.json",
		"./header_7.json",
		"./header_8.json",
		"./header_9.json",
		"./header_10.json", 
	}


)

type (
	MyPageProcesser struct {
		w *SocialWorker
	}
	User struct {
		uid      string
		friendid string
		uname    string
		usex     string
	}
	SocialWorker struct {
		Queue             *sqs.SQS
		UserQueueUrl      *string
		ContainerQueueUrl *string
		Dynamo            *dynamodb.DynamoDB
		/**
		 * message process
		 */
	}
)

func NewSocialWorker(queue *sqs.SQS, userQueueUrl *string, containerQueueUrl *string, dynamo *dynamodb.DynamoDB) *SocialWorker {
	var worker *SocialWorker
	worker = new(SocialWorker)
	worker.Queue = queue
	worker.UserQueueUrl = userQueueUrl
	worker.ContainerQueueUrl = containerQueueUrl
	worker.Dynamo = dynamo
	return worker
}

func (w *SocialWorker) Start() {
	w.run()
}

func (w *SocialWorker) run() {
	/*
	** 这个位置应该用CrawlUrl函数进行传输
	** 从container的环境变量中取得参数
	** 同时初始化本页的crawlUrl
	 */
	url := os.Getenv("CRAWL_URL")
	id := os.Getenv("CRAWL_ID")
	fatherId := os.Getenv("CRAWL_FID")
	layer_string := os.Getenv("CRAWL_LAYER")
	layer, err := strconv.ParseInt(layer_string, 10, 0)
	if err != nil {
		glog.Errorln("error on parseInt layer: ", err.Error())
	}
	/**
	 * 设置要爬取页面的Url，用户id，以及本页的layer
	 */
	crawlUrl.Url = url
	crawlUrl.Id = id
	crawlUrl.FatherId = fatherId
	crawlUrl.Layer = layer
	glog.Infoln(crawlUrl)

	Urls = append(Urls, url)
	UrlsLevel = append(UrlsLevel, layer)
	i = 0
	w.SpiderMain()
}

func NewMyPageProcesser(w *SocialWorker) *MyPageProcesser {
	processer := &MyPageProcesser{}
	processer.w = w
	return processer
}

func (w *SocialWorker) SpiderMain() {
	spider.NewSpider(NewMyPageProcesser(w), "TaskName").
		AddUrlWithHeaderFile(crawlUrl.Url, "html", "./header_1.json"). // Start url, html is the responce type ("html" or "json" or "jsonp" or "text")
		AddPipeline(pipeline.NewPipelineConsole()).                  // Print result on screen
		SetThreadnum(1).                                             // Crawl request by three Coroutines
		Run()
}

/*
 ** 解析页面，把粉丝的信息存入dynamodb，同时把接下来要爬取的url存入sqs
 */
func (this *MyPageProcesser) Process(p *page.Page) {
	if !p.IsSucc() {
		glog.Errorln(p.Errormsg())
		return
	}
	/*
	 ** 打印爬取得页面
	 */
	glog.Infoln(p)
	query := p.GetHtmlParser()
	
	if Urls[i] == "weibo.cn" {
		i = i + 1
	}

	if UrlsLevel[i] == 0 {
		glog.Infoln("layer:", crawlUrl.Layer)
		this.w.GetNextPageUrl(query, p)
		this.w.GetFriendsUrl(query,p)
	}else if UrlsLevel[i] == 1 {
		this.w.GetFriendsInfo(query)
	}
	// if crawlUrl.Layer == 0 {
	// } else if crawlUrl.Layer == 1 {
	// 	glog.Infoln("layer:", crawlUrl.Layer)
	// 	this.w.GetNextPageUrl(query, p)
	// 	this.w.GetFFUrl(query)
	// } else if crawlUrl.Layer == 2 {
	// 	glog.Infoln("layer:", crawlUrl.Layer)
	// 	this.w.GetFFInfo(query)
	// }
	// 
	


	header_num := rand.Intn(9)
	header_json := headerJson[header_num]	
	i = i + 1
	p.AddTargetRequestWithHeaderFile(Urls[i], "html", header_json) 

}

func (this *MyPageProcesser) Finish() {
	go this.w.sendContainerMessage(crawlUrl.Id)
	glog.Infoln("TODO:before end spider \r\n")
}

/*
 **get next page url
 */
func (w *SocialWorker) GetNextPageUrl(query *goquery.Document, p *page.Page) {
	pageText := query.Find("div#pagelist").Find("form").Find("div").Find("a:first-child").Text()
	glog.Infoln(pageText)
	if pageText == "上页" {
		return
	} else {
		nextPageUrlString, _ := query.Find("div#pagelist").Find("form").Find("div").Find("a:first-child").Attr("href")
		nextPageUrl := "http://weibo.cn" + nextPageUrlString
		Urls = append(Urls, nextPageUrl)
		UrlsLevel = append(UrlsLevel, UrlsLevel[i])
	}
}

/*
**get friends url
 */
func (w *SocialWorker) GetFriendsUrl(query *goquery.Document, p *page.Page) {
	var str_1 string
	// newCrawlUrl := models.CrawlUrl{}
	query.Find("div.c").Find("table").Find("tbody").Find("tr").Find("a:last-child").Each(func(j int, s *goquery.Selection) {
		if j%2 != 0 {
			friendsUrlString, _ := s.Attr("href")
			var digitsRegexp = regexp.MustCompile(`(^|&|\?)uid=([^&]*)(&|$)`)
			str := digitsRegexp.FindStringSubmatch(friendsUrlString)
			if str == nil {
				str_1 = "1"
			} else {
				str_1 = str[2]
			}
			friendsInfoUrl := "http://weibo.cn/" + str_1 + "/info"
			// newCrawlUrl.Url = "http://weibo.cn/" + str_1 + "/fans"
			// p.AddTargetRequestWithHeaderFile(friendsInfoUrl, "html", "./header.json")
			// newCrawlUrl.Id = str_1
			// newCrawlUrl.Layer = crawlUrl.Layer + 1
			// newCrawlUrl.FatherId = crawlUrl.Id
			// w.SendMessageToSQS(newCrawlUrl)
			
			Urls = append(Urls, friendsInfoUrl)
			UrlsLevel = append(UrlsLevel, UrlsLevel[i] + 1)
		}
	})
}

/*
**get friends info
 */
func (w *SocialWorker) GetFriendsInfo(query *goquery.Document) {
	var user User
	var uid string
	var usex string
	/*
	 ** 获取红人粉丝的的uid（str)
	 */
	uidString, _ := query.Find("div.c").Eq(1).Find("a").Attr("href")
	var digitsRegexp = regexp.MustCompile(`(^|&|\?)uid=([^&]*)(&|$)`)
	str := digitsRegexp.FindStringSubmatch(uidString)
	uid = str[2]
	/*
	 ** 获取红人粉丝的基本信息
	 */
	uStr := query.Find("div.c").Eq(2).Text()
	nameStr_1 := GetBetweenStr(uStr, ":", "性别")
	nameStr_2 := GetBetweenStr(nameStr_1, ":", "认证")
	nameStr_3 := strings.Split(nameStr_2, ":")
	uname := nameStr_3[1]
	sexStr_1 := GetBetweenStr(uStr, "性别", "地区")
	sexStr_2 := strings.Split(sexStr_1, ":")
	if sexStr_2[1] == "男" {
		usex = "male"
	} else {
		usex = "famale"
	}

	user.uid = crawlUrl.Id
	user.friendid = uid
	user.uname = uname
	user.usex = usex
	glog.Infoln(user)
	w.putItems(user)
}




/*
**	get friends' friends url
 */
func (w *SocialWorker) GetFFUrl(query *goquery.Document) {
	var str_1 string
	newCrawlUrl := models.CrawlUrl{}
	query.Find("div.c").Find("table").Find("tbody").Find("tr").Find("a:last-child").Each(func(j int, s *goquery.Selection) {
		if j%2 == 0 {
			friendsUrlString, _ := s.Attr("href")
			var digitsRegexp = regexp.MustCompile(`(^|&|\?)uid=([^&]*)(&|$)`)
			str := digitsRegexp.FindStringSubmatch(friendsUrlString)
			if str == nil {
				return
			} else {
				str_1 = str[2]
			}
			friendsUrl := "http://weibo.cn/" + str_1 + "/info"
			newCrawlUrl.Url = friendsUrl
			newCrawlUrl.Id = str_1
			newCrawlUrl.FatherId = crawlUrl.Id
			newCrawlUrl.Layer = crawlUrl.Layer + 1
			w.SendMessageToSQS(newCrawlUrl)
		}
	})
}

/*
** get friends' friends info
 */
func (w *SocialWorker) GetFFInfo(query *goquery.Document) {
	var user User
	// var uid string
	var usex string
	// var usersId []string
	// var usersName []string
	// uidString, _ := query.Find("div.c").Eq(1).Find("a").Attr("href")
	// var digitsRegexp = regexp.MustCompile(`(^|&|\?)uid=([^&]*)(&|$)`)
	/*
	 ** 获取粉丝的粉丝的uid（str)
	 */
	// str := digitsRegexp.FindStringSubmatch(uidString)
	// uid = crawlUrl.Id
	// usersId = append(usersId, uid)
	uStr := query.Find("div.c").Eq(2).Text()
	nameStr_1 := GetBetweenStr(uStr, ":", "性别")
	nameStr_2 := GetBetweenStr(nameStr_1, ":", "认证")
	nameStr_3 := strings.Split(nameStr_2, ":")
	uname := nameStr_3[1]
	sexStr_1 := GetBetweenStr(uStr, "性别", "地区")
	sexStr_2 := strings.Split(sexStr_1, ":")
	if sexStr_2[1] == "男" {
		usex = "male"
	} else {
		usex = "famale"
	}

	user.uid = crawlUrl.FatherId
	user.friendid = crawlUrl.Id
	user.uname = uname
	user.usex = usex
	glog.Infoln(user)
	w.putItems(user)
}
func GetBetweenStr(str, start, end string) string {
	n := strings.Index(str, start)
	if n == -1 {
		n = 0
	}
	str = string([]byte(str)[n:])
	m := strings.Index(str, end)
	if m == -1 {
		m = len(str)
	}
	str = string([]byte(str)[:m])
	return str
}

/*
**send message to sqs
 */
func (w *SocialWorker) SendMessageToSQS(crawlUrl models.CrawlUrl) {

	e := new(events.SocialEvent)
	e.CrawlUrl = &crawlUrl
	e.EventId = events.DOCKER_RUN
	messageBody, err := json.Marshal(e)

	if err != nil {
		glog.Errorln("error on json crawl:", err.Error())
	}

	message := string(messageBody)

	var sendMessageInput = &sqs.SendMessageInput{
		MessageBody: aws.String(message),
		QueueUrl:    w.UserQueueUrl,
	}
	_, err = w.Queue.SendMessage(sendMessageInput)
	if err != nil {
		glog.Errorln("Error on receive message: ", err.Error())
	}
}

/*
**put items into dynamodb
 */
func (w *SocialWorker) putItems(user User) {
	if len(user.usex) == 0 {
		user.usex = "null"
	}
	glog.Infoln("user:", user)
	params := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"uid":      {S: aws.String(user.uid)},
			"friendid": {S: aws.String(user.friendid)},
			"uname":    {S: aws.String(user.uname)},
			"usex":     {S: aws.String(user.usex)},
		},
		TableName: aws.String("friends"),
	}
	_, err := w.Dynamo.PutItem(params)
	if err != nil {
		glog.Errorln(err.Error())
	}
}

func (w *SocialWorker) sendContainerMessage(crawlId string) {
	e := new(events.ContainerEvent)
	e.CrawlId = crawlId
	messageBody, err := json.Marshal(e)

	if err != nil {
		glog.Errorln("error on json crawl:", err.Error())
	}

	message := string(messageBody)

	var sendMessageInput = &sqs.SendMessageInput{
		MessageBody: aws.String(message),
		QueueUrl:    w.ContainerQueueUrl,
	}
	_, err = w.Queue.SendMessage(sendMessageInput)
	if err != nil {
		glog.Errorln("Error on receive message: ", err.Error())
	}
}
