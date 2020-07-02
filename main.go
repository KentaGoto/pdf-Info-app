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
	exe, _ := os.Executable()    // 実行ファイルのフルパス
	rootDir := filepath.Dir(exe) // 実行ファイルのあるディレクトリ

	r := gin.Default()
	r.Static("/results", "./results") // 静的ディレクトリとしておかないとHTMLのダウンロードリンクからアクセスできない
	r.LoadHTMLGlob("html/**/*.tmpl")

	// アクセスされたらこれを表示
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "html/index.tmpl", gin.H{
			"title": "PDF Info",
		})
	})

	// uploadされたらこれ
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
		fmt.Fprintln(csvFile, "File Name" + "," + "Creator" + "," + "Producer")

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
				var creator string
				creatorRe := regexp.MustCompile(`Creator:(\s)+(.+)`)
				var producer string
				producerRe := regexp.MustCompile(`Producer:(\s)+(.+)`)

				for _, s := range sArray {
					if regexp.MustCompile(`Creator:(\s)+(.+)`).MatchString(s) == true {
						creator = creatorRe.ReplaceAllString(s, "$2")
						creator = strings.TrimRight(creator, "\n\r")
						log.Println(creator)
					} else if regexp.MustCompile(`Producer:(\s)+(.+)`).MatchString(s) == true {
						producer = producerRe.ReplaceAllString(s, "$2")
						producer = strings.TrimRight(producer, "\n\r")
						log.Println(producer)
					}
				}

				// pathから不要な文字を削除する
				replacedPath := strings.Replace(path, rootDir+"\\out", "", 1)

				// csvに書き込み
				fmt.Fprintln(csvFile, replacedPath + "," + creator + "," + producer)
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

	r.Run(":10")
}

// 再帰
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
