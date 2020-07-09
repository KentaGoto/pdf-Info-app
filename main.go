package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	exe, _ := os.Executable()
	rootDir := filepath.Dir(exe)

	r := gin.Default()
	r.Static("/results", "./results")
	r.LoadHTMLGlob("html/**/*.tmpl")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "html/index.tmpl", gin.H{
			"title": "PDF Info",
		})
	})

	r.POST("/", func(c *gin.Context) {
		zipFile, err := c.FormFile("upload")
		if err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
			return
		}
		log.Println(zipFile.Filename)

		// 特定のディレクトリにファイルをアップロードする
		dst := rootDir + "\\uploaded" + "\\" + filepath.Base(zipFile.Filename)
		log.Println(dst)
		if err := c.SaveUploadedFile(zipFile, dst); err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("upload file err: %s", err.Error()))
			return
		}

		// 時刻オブジェクト
		t := time.Now()
		const layout = "2006-01-02_15-04-05"
		tFormat := t.Format(layout)

		// 結果が記載されるcsvのファイル名
		resultFile := tFormat + ".csv"
		resultFile = "results\\" + resultFile

		// outフォルダを削除する
		if err := os.RemoveAll("out"); err != nil {
			fmt.Println(err)
		}

		// outフォルダを作る
		if err := os.Mkdir("out", 0777); err != nil {
			fmt.Println(err)
		}

		// unzipする
		out, err3 := exec.Command("7z.exe", "x", "-y", "-o"+rootDir+"\\out", dst).CombinedOutput()
		log.Println("7z.exe", "x", "-y", "-o"+rootDir+"\\out", dst)
		if err3 != nil {
			fmt.Println("7zip command Exec Error")
		}
		fmt.Printf("ls result: \n%s", string(out))

		// resultFileを作成してオープンする
		csvFile, err := os.OpenFile(resultFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal(err)
		}
		defer csvFile.Close()

		// resultFileのヘッダー
		fmt.Fprintln(csvFile, "File Name"+","+"Author"+","+"Creator"+","+"Producer"+","+"CreationDate"+","+"ModDate"+","+"Page size"+","+"Pages"+","+"File size"+","+"PDF version")

		// 再帰でPDFを処理する
		paths := dirwalk(rootDir + `\out`)

		flag := 0
		for _, path := range paths {
			ext := filepath.Ext(path) // ファイルの拡張子を得る
			if ext == ".pdf" {
				flag++
				log.Println("Processing... " + path)

				// pdfinfoコマンドの出力をゲットする
				pdfinfoOut, err := exec.Command("pdfinfo.exe", path).CombinedOutput()
				if err != nil {
					fmt.Println("pdfinfo command Exec Error")
				}

				s := string(pdfinfoOut)
				sArray := strings.Split(s, "\n") // 改行でスプリットして配列にプッシュ

				// pdfinfoコマンドの結果から抽出する項目
				var author string
				authorRe := regexp.MustCompile(`Author:(\s)+(.+)`)
				var creator string
				creatorRe := regexp.MustCompile(`Creator:(\s)+(.+)`)
				var producer string
				producerRe := regexp.MustCompile(`Producer:(\s)+(.+)`)
				var creationDate string
				creationDateRe := regexp.MustCompile(`CreationDate:(\s)+(.+)`)
				var modDate string
				modDateRe := regexp.MustCompile(`ModDate:(\s)+(.+)`)
				var pagesize string
				pagesizeRe := regexp.MustCompile(`Page size:(\s)+(.+)`)
				var pages string
				pagesRe := regexp.MustCompile(`Pages:(\s)+(.+)`)
				var fileSize string
				fileSizeRe := regexp.MustCompile(`File size:(\s)+(\d+)\s.+`)
				var pdfVersion string
				pdfVersionRe := regexp.MustCompile(`PDF version:(\s)+(.+)`)

				for _, s := range sArray {
					s = strings.Replace(s, ",", "_", -1) // カンマはアンスコに置換

					if authorRe.MatchString(s) == true {
						author = authorRe.ReplaceAllString(s, "$2")
						author = strings.TrimRight(author, "\n\r")
					} else if creatorRe.MatchString(s) == true {
						creator = creatorRe.ReplaceAllString(s, "$2")
						creator = strings.TrimRight(creator, "\n\r")
					} else if producerRe.MatchString(s) == true {
						producer = producerRe.ReplaceAllString(s, "$2")
						producer = strings.TrimRight(producer, "\n\r")
					} else if creationDateRe.MatchString(s) == true {
						creationDate = creationDateRe.ReplaceAllString(s, "$2")
						creationDate = strings.TrimRight(creationDate, "\n\r")
					} else if modDateRe.MatchString(s) == true {
						modDate = modDateRe.ReplaceAllString(s, "$2")
						modDate = strings.TrimRight(modDate, "\n\r")
					} else if pagesizeRe.MatchString(s) == true {
						pagesize = pagesizeRe.ReplaceAllString(s, "$2")
						pagesize = strings.TrimRight(pagesize, "\n\r")
					} else if pagesRe.MatchString(s) == true {
						pages = pagesRe.ReplaceAllString(s, "$2")
						pages = strings.TrimRight(pages, "\n\r")
					} else if fileSizeRe.MatchString(s) == true {
						fileSize = fileSizeRe.ReplaceAllString(s, "$2")
						fileSize = strings.TrimRight(fileSize, "\n\r")
					} else if pdfVersionRe.MatchString(s) == true {
						pdfVersion = pdfVersionRe.ReplaceAllString(s, "$2")
						pdfVersion = strings.TrimRight(pdfVersion, "\n\r")
					}
				}

				// pathから不要な文字を削除する
				replacedPath := strings.Replace(path, rootDir+"\\out", "", 1)

				// csvに書き込み
				fmt.Fprintln(csvFile, replacedPath+","+author+","+creator+","+producer+","+creationDate+","+modDate+","+pagesize+","+pages+","+fileSize+","+pdfVersion)
			}
		}

		csvFile.Close()

		// nkfコマンドでBOM付きにする
		errNkf := exec.Command("nkf.exe", "-w8", "--overwrite", rootDir+"\\"+resultFile).Run()
		log.Println("nkf.exe", "-w8", "--overwrite", rootDir+"\\"+resultFile)
		if errNkf != nil {
			fmt.Println("nkf command Exec Error")
		}

		if flag == 0 {
			// pdfがなかった場合はこれを返す
			c.String(http.StatusOK, "There is no pdf in the uploaded zip.")
		} else {
			// pdfがあった場合はcsvを返す
			// index.tmplを書き換えて、HTMLからダウンロードさせる
			c.HTML(http.StatusOK, "html/index.tmpl", gin.H{
				"title":           "PDF Info",
				"downloadMessage": "Download: ",
				"downloadfile":    tFormat + ".csv",
			})
		}
	})

	r.Run(":12")
}

func dirwalk(dir string) []string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	var paths []string
	for _, file := range files {
		if file.IsDir() {
			paths = append(paths, dirwalk(filepath.Join(dir, file.Name()))...)
			continue
		}
		paths = append(paths, filepath.Join(dir, file.Name()))
	}

	return paths
}

func newCsvWriter(w io.Writer, bom bool) *csv.Writer {
	bw := bufio.NewWriter(w)
	if bom {
		bw.Write([]byte{0xEF, 0xBB, 0xBF})
	}
	return csv.NewWriter(bw)
}

