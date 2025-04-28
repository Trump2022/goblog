package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"text/template"
	"time"
	"unicode/utf8"

	"github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

var router = mux.NewRouter()
var db *sql.DB

func initDB() {

	var err error
	config := mysql.Config{
		User:                 "homestead",
		Passwd:               "secret#GO123'",
		Addr:                 "127.0.0.1:3306",
		Net:                  "tcp",
		DBName:               "goblog",
		AllowNativePasswords: true,
	}

	// 准备数据库连接池
	db, err = sql.Open("mysql", config.FormatDSN())
	checkError(err)

	// 设置最大连接数
	db.SetMaxOpenConns(25)
	// 设置最大空闲连接数
	db.SetMaxIdleConns(25)
	// 设置每个链接的过期时间
	db.SetConnMaxLifetime(5 * time.Minute)

	// 尝试连接，失败会报错
	err = db.Ping()
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>Hello, 欢迎来到 goblog！</h1>")
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "此博客是用以记录编程笔记，如您有反馈或建议，请联系 "+
		"<a href=\"mailto:summer@example.com\">summer@example.com</a>")
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "<h1>请求页面未找到 :(</h1><p>如有疑惑，请联系我们。</p>")
}

func articlesShowHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	fmt.Fprint(w, "文章 ID："+id)
}

func articlesIndexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "访问文章列表")
}

// ArticlesFormData 创建博文表单数据
type ArticlesFormData struct {
	Title, Body string
	URL         *url.URL
	Errors      map[string]string
}

func articlesStoreHandler(w http.ResponseWriter, r *http.Request) {
	// fmt.Fprint(w, "创建新的文章")
	// err := r.ParseForm()
	// if err != nil {
	// 	// 解析错误，这里应该有错误处理
	// 	fmt.Fprint(w, "请提供正确的数据")
	// 	return
	// }

	// title := r.PostForm.Get("title")
	// fmt.Fprintf(w, "POST PostForm: %v <br>", r.PostForm)
	// fmt.Fprintf(w, "POST Form: %v <br>", r.Form)
	// fmt.Fprintf(w, "title 的值为: %v", title)

	// fmt.Fprintf(w, "r.Form 中 title 的值为: %v <br>", r.FormValue("title"))
	// fmt.Fprintf(w, "r.PostForm 中 title 的值为: %v <br>", r.PostFormValue("title"))
	// fmt.Fprintf(w, "r.Form 中 test 的值为: %v <br>", r.FormValue("test"))
	// fmt.Fprintf(w, "r.PostForm 中 test 的值为: %v <br>", r.PostFormValue("test"))

	title := r.PostFormValue("title")
	body := r.PostFormValue("body")

	errors := make(map[string]string)

	// 验证标题
	if title == "" {
		errors["title"] = "标题不能为空"
	} else if utf8.RuneCountInString(title) < 3 || utf8.RuneCountInString(title) > 40 {
		errors["title"] = "标题长度需介于 3-40"
	}

	//验证内容
	if body == "" {
		errors["body"] = "内容不能为空"
	} else if utf8.RuneCountInString(body) < 10 {
		errors["body"] = "内容长度需大于或等于10个字符"
	}

	//检查是否有错误
	if len(errors) == 0 {
		fmt.Fprint(w, "验证通过!<br>")
		fmt.Fprintf(w, "title的值为: %v <br>", title)
		fmt.Fprintf(w, "title的长度为: %v <br>", utf8.RuneCountInString(title))
		fmt.Fprintf(w, "body的值为: %v <br>", body)
		fmt.Fprintf(w, "body的长度为: %v <br>", utf8.RuneCountInString(body))
	} else {
		// fmt.Fprintf(w, "有错误发生, errors的值为: %v <br>", errors)

		// 		html := `
		// <!DOCTYPE html>
		// <html lang="en">
		// <head>
		//     <title>创建文章 —— 我的技术博客</title>
		//     <style type="text/css">.error {color: red;}</style>
		// </head>
		// <body>
		//     <form action="{{ .URL }}" method="post">
		//         <p><input type="text" name="title" value="{{ .Title }}"></p>
		//         {{ with .Errors.title }}
		//         <p class="error">{{ . }}</p>
		//         {{ end }}
		//         <p><textarea name="body" cols="30" rows="10">{{ .Body }}</textarea></p>
		//         {{ with .Errors.body }}
		//         <p class="error">{{ . }}</p>
		//         {{ end }}
		//         <p><button type="submit">提交</button></p>
		//     </form>
		// </body>
		// </html>
		// `
		storeURL, _ := router.Get("articles.store").URL()

		data := ArticlesFormData{
			Title:  title,
			Body:   body,
			URL:    storeURL,
			Errors: errors,
		}
		// tmpl, err := template.New("create-form").Parse(html)
		tmpl, err := template.ParseFiles("resources/views/articles/create.gohtml")
		if err != nil {
			panic(err)
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			panic(err)
		}
	}
}

func forceHTMLMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. 设置标头
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// 2. 继续处理请求
		next.ServeHTTP(w, r)
	})
}

func removeTrailingSlash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. 除首页以外，移除所有请求路径后面的斜杆
		if r.URL.Path != "/" {
			r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
		}

		// 2. 将请求传递下去
		next.ServeHTTP(w, r)
	})
}

func articlesCreateHandler(w http.ResponseWriter, r *http.Request) {
	// 	html := `
	// <!DOCTYPE html>
	// <html lang="en">
	// <head>
	//     <title>创建文章 —— 我的技术博客</title>
	// </head>
	// <body>
	// 	<form action="%s?test=data" method="post">
	//         <p><input type="text" name="title"></p>
	//         <p><textarea name="body" cols="30" rows="10"></textarea></p>
	//         <p><button type="submit">提交</button></p>
	//     </form>
	// </body>
	// </html>
	// `
	// storeURL, _ := router.Get("articles.store").URL()
	// fmt.Fprintf(w, html, storeURL)

	storeURL, _ := router.Get("articles.store").URL()
	data := ArticlesFormData{
		Title:  "",
		Body:   "",
		URL:    storeURL,
		Errors: nil,
	}
	tmpl, err := template.ParseFiles("resources/views/articles/create.gohtml")
	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		panic(err)
	}

}

func createTables() {
	createArticlesSQL := `CREATE TABLE IF NOT EXISTS articles(
    id bigint(20) PRIMARY KEY AUTO_INCREMENT NOT NULL,
    title varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
    body longtext COLLATE utf8mb4_unicode_ci
); `

	_, err := db.Exec(createArticlesSQL)
	checkError(err)
}

type Result interface {
	LastInsertId() (int64, error) // 使用 INSERT 向数据插入记录，数据表有自增 id 时，该函数有返回值
	RowsAffected() (int64, error) // 表示影响的数据表行数
}

func main() {
	initDB()
	createTables()
	fmt.Println(config.FormatDSN())
	// router := mux.NewRouter()

	router.HandleFunc("/", homeHandler).Methods("GET").Name("home")
	router.HandleFunc("/about", aboutHandler).Methods("GET").Name("about")

	router.HandleFunc("/articles/{id:[0-9]+}", articlesShowHandler).Methods("GET").Name("articles.show")
	router.HandleFunc("/articles", articlesIndexHandler).Methods("GET").Name("articles.index")
	router.HandleFunc("/articles", articlesStoreHandler).Methods("POST").Name("articles.store")
	router.HandleFunc("/articles/create", articlesCreateHandler).Methods("GET").Name("articles.create")

	// 自定义 404 页面
	router.NotFoundHandler = http.HandlerFunc(notFoundHandler)

	// 中间件：强制内容类型为 HTML
	router.Use(forceHTMLMiddleware)

	// // 通过命名路由获取 URL 示例
	// homeURL, _ := router.Get("home").URL()
	// fmt.Println("homeURL: ", homeURL)
	// articleURL, _ := router.Get("articles.show").URL("id", "23")
	// fmt.Println("articleURL: ", articleURL)

	http.ListenAndServe(":3000", removeTrailingSlash(router))
}
