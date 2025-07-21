package controllers

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
)

// ArticlesController 文章相关页面
type ArticlesController struct {
}

// Show 文章详情页面
func (*ArticlesController) Show(w http.ResponseWriter, r *http.Request) {
	// vars := mux.Vars(r)
	// id := vars["id"]
	// fmt.Fprint(w, "文章 ID："+id)

	// 1.获取 URL 参数
	id := getRouteVariable("id", r)
	// vars := mux.Vars(r)
	// id := vars["id"]

	// 2. 读取对应的文章数据
	article, err := getArticleByID(id)
	// article := Article{}
	// query := "SELECT * From articles WHERE id = ?"
	// err := db.QueryRow(query, id).Scan(&article.ID, &article.Title, &article.Body)

	// stmt, err := db.Prepare(query)
	// checkError(err)
	// defer stmt.Close()
	// err = stmt.QueryRow(id).Scan(&article.ID, &article.Title, &article.Body)

	// 3. 如果出现错误
	if err != nil {
		if err == sql.ErrNoRows {
			// 3.1 数据未找到
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "404 文章未找到")
		} else {
			// 3.2 数据库错误
			checkError(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "500服务器内部错误")
		}
	} else {
		// 4. 读取成功, 显示文章
		// fmt.Fprint(w, "读取成功, 文章标题 --"+article.Title)
		// tmpl, err := template.ParseFiles("resources/views/articles/show.gohtml")

		tmpl, err := template.New("show.gohtml").Funcs(template.FuncMap{
			"RouteName2URL": RouteName2URL,
			"Int64ToString": Int64ToString,
		}).ParseFiles("resources/views/articles/show.gohtml")
		checkError(err)
		err = tmpl.Execute(w, article)
		checkError(err)
	}
}
