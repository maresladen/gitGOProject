package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/robertkrimen/otto"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var strFileServerPath string
var strPath string
var strComicID string
var strVolID string
var hostbase string
var importUrl string
var mPath string = "ComicBook"
var intfloder int = 1

//文件存放服务器地址
// "http://www.hhcool.com/script/ds.js"

func main() {
	importUrl = readForGetUrl("url.txt")
	jmjs := readForGetUrl("jmjs.txt")
	if importUrl == "" || jmjs == "" {
		return
	}
	// fmt.Println(importUrl, jmjs)

	createFloder("ComicBook")

	strFileServerPath = getFileServerAdd(jmjs)
	// fmt.Println(strFileServerPath)

	hostbase = getHostName(importUrl)

	getScrent(importUrl)
}

func readForGetUrl(txtPath string) string {

	result := ""
	fi, err := os.Open(txtPath)
	if err != nil {
		fmt.Println("呃...")
	} else {
		temp, _ := ioutil.ReadAll(fi)
		result = string(temp)
	}
	return result
}

func createFloder(fName string) {
	err := os.Chdir(fName)
	if err != nil {
		os.Mkdir(fName, 0)
	}
}

//获取文件存放地址
func getFileServerAdd(s string) string {
	res, err := http.Get(s)
	if err != nil {
		fmt.Println("get错误")
	}
	body, err := ioutil.ReadAll(res.Body) //转换byte数组
	if err != nil {
		fmt.Println("read错误")
	}
	defer res.Body.Close()
	//io.Copy(os.Stdout, res.Body)//写到输出流，
	bodystr := string(body)

	vm := otto.New()
	vm.Run(bodystr)
	value, err := vm.Get("sDS")
	tempStr, _ := value.ToString()
	temps := strings.Split(tempStr, "|")
	return temps[1]
}

//用于返回host网址
func getHostName(s string) string {
	if strings.Contains(s, "https") {
		return GetStrBeginWithStart(s, "https://", "/")
	}
	return GetStrBeginWithStart(s, "http://", "/")
}

//主方法
func getScrent(url string) {

	// res, err := http.Get(urls)
	// if err != nil {
	// 	fmt.Println("get错误")
	// }
	// body, err := ioutil.ReadAll(res.Body) //转换byte数组
	// if err != nil {
	// 	fmt.Println("read错误")
	// }
	// defer res.Body.Close()
	// //io.Copy(os.Stdout, res.Body)//写到输出流，
	// bodystr := string(body)

	doc, err := goquery.NewDocument(url)

	if err != nil {
		log.Fatal(err)
	}

	strComicID, _ = doc.Find("#hdInfoID").Attr("value")
	strVolID, _ = doc.Find("#hdID").Attr("value")

	bodystr, _ := doc.Html()

	strFiles := GetBetweenStr(bodystr, `sFiles="`, `";var sPath`, len(`sFiles="`))
	strPath = GetBetweenStr(bodystr, `var sPath="`, `";</script>`, len(`var sPath="`))

	runJS := `
    var x = s.substring(s.length-1);
    var xi="abcdefghijklmnopqrstuvwxyz".indexOf(x)+1;
    var sk = s.substring(s.length-xi-12,s.length-xi-1);
    s=s.substring(0,s.length-xi-12);
    var k=sk.substring(0,sk.length-1);
    var f=sk.substring(sk.length-1);
    var k=sk.substring(0,sk.length-1);
    var f=sk.substring(sk.length-1);
    for(i=0;i<k.length;i++) {
        eval("s=s.replace(/"+ k.substring(i,i+1) +"/g,'"+ i +"')");
    }
    var ss = s.split(f);
    s="";
    for(i=0;i<ss.length;i++) {
        s+=String.fromCharCode(ss[i]);
    }
    `

	imgpaths, err := runJSGetAddress(strFiles, runJS)
	if err != nil {
		fmt.Println("js解析错误")
	}
	// fmt.Println(imgpaths)
	// fmt.Println(mPath + "/" + strconv.Itoa(intfloder))
	createFloder(mPath + "/" + strconv.Itoa(intfloder))

	for index, s := range imgpaths {
		// fmt.Println(index, strFileServerPath+strPath+s)
		fmt.Println("正在下载第" + strconv.Itoa(intfloder) + "集,第" + strconv.Itoa(index+1) + "页")
		downloadFiles(strFileServerPath+strPath+s, index+1)
	}

	getNextUrls()
}

//获得下一集的地址
func getNextUrls() {
	strNextUrl := hostbase + "/app/getNextVolUrl.aspx?ComicID=" + strComicID + "&VolID=" + strVolID + "&t=N"
	res, err := http.Get(strNextUrl)
	if err != nil {
		fmt.Println("get错误")
	}
	body, err := ioutil.ReadAll(res.Body) //转换byte数组
	if err != nil {
		fmt.Println("read错误")
	}
	defer res.Body.Close()
	//io.Copy(os.Stdout, res.Body)//写到输出流，
	bodystr := string(body)
	if strings.HasPrefix(bodystr, "Err_没有") {
		fmt.Println("下载完成")
		return
	} else {
		intfloder += 1
		getScrent(bodystr)
	}
}

//下载图片
func downloadFiles(urls string, index int) {

	res, _ := http.Get(urls)
	defer res.Body.Close()
	file, _ := os.Create(mPath + "/" + strconv.Itoa(intfloder) + "/" + strconv.Itoa(index) + ".jpg")
	io.Copy(file, res.Body)
}

//运行js方法,用汗汗的加密方式得到真正的图片地址
func runJSGetAddress(s string, js string) ([]string, error) {
	vm := otto.New()
	vm.Set("s", s)
	vm.Run(js)
	value, err := vm.Get("s")
	if err != nil {
		return nil, err
	}
	tempStr, _ := value.ToString()
	return strings.Split(tempStr, "|"), nil
}

//以起始点和结束点截取字符串
func GetBetweenStr(str, start, end string, offset int) string {
	n := strings.Index(str, start) + offset
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

func GetStrBeginWithStart(str, start, end string) string {
	n := strings.Index(str, start)
	if n == -1 {
		n = 0
	}
	str = string([]byte(str)[n+len(start):])
	m := strings.Index(str, end)
	if m == -1 {
		m = len(str)
	}
	str = start + string([]byte(str)[:m])
	return str
}

//以起始点和长度截取字符串
func Substr(str string, start, length int) string {
	rs := []rune(str)
	rl := len(rs)
	end := 0

	if start < 0 {
		start = rl - 1 + start
	}
	end = start + length

	if start > end {
		start, end = end, start
	}

	if start < 0 {
		start = 0
	}
	if start > rl {
		start = rl
	}
	if end < 0 {
		end = 0
	}
	if end > rl {
		end = rl
	}

	return string(rs[start:end])
}
