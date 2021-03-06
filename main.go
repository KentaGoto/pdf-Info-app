package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
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

		// Uploading files to a specific directory.
		dst := rootDir + "\\uploaded" + "\\" + filepath.Base(zipFile.Filename)
		log.Println(dst)
		if err := c.SaveUploadedFile(zipFile, dst); err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("upload file err: %s", err.Error()))
			return
		}

		// Time object
		t := time.Now()
		const layout = "2006-01-02_15-04-05"
		tFormat := t.Format(layout)

		// Result csv file name
		resultFile := tFormat + ".csv"
		resultFile = "results\\" + resultFile

		// Delete out folder
		if err := os.RemoveAll("out"); err != nil {
			fmt.Println(err)
		}

		// Make out folder
		if err := os.Mkdir("out", 0777); err != nil {
			fmt.Println(err)
		}

		// unzip
		out, err3 := exec.Command("7z.exe", "x", "-y", "-o"+rootDir+"\\out", dst).CombinedOutput()
		log.Println("7z.exe", "x", "-y", "-o"+rootDir+"\\out", dst)
		if err3 != nil {
			fmt.Println("7zip command Exec Error")
		}
		fmt.Printf("ls result: \n%s", string(out))

		// Create and open a resultFile.
		csvFile, err := os.OpenFile(resultFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal(err)
		}
		defer csvFile.Close()

		// Header of resultFile.
		fmt.Fprintln(csvFile, "File Name"+","+"Author"+","+"Creator"+","+"Producer"+","+"CreationDate"+","+"ModDate"+","+"Page size"+","+"JavaScript"+","+"Pages"+","+"Encrypted"+","+"Page rot"+","+"File size(MB)"+","+"PDF version")

		// Processing PDFs with Recursion.
		paths := dirwalk(rootDir + `\out`)

		flag := 0
		for _, path := range paths {
			ext := filepath.Ext(path) // Get the file extension.
			if ext == ".pdf" {
				flag++
				log.Println("Processing... " + path)

				// Get the output of the pdfinfo command.
				pdfinfoOut, err := exec.Command("pdfinfo", "-isodates", path).CombinedOutput()
				if err != nil {
					fmt.Println("pdfinfo command Exec Error")
				}

				s := string(pdfinfoOut)
				fmt.Println(s)
				sArray := strings.Split(s, "\n") // Split on a line break and push to an array.

				// Items to be extracted from the results of the pdfinfo command.
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
				var javaScript string
				javaScriptRe := regexp.MustCompile(`JavaScript:(\s)+(.+)`)
				var pages string
				pagesRe := regexp.MustCompile(`Pages:(\s)+(.+)`)
				var encrypted string
				encryptedRe := regexp.MustCompile(`Encrypted:(\s)+(.+)`)
				var pagesize string
				pagesizeRe := regexp.MustCompile(`Page size:(\s)+(.+)`)
				var pageRot string
				pageRotRe := regexp.MustCompile(`Page rot:(\s)+(.+)`)
				var fileSize string
				fileSizeRe := regexp.MustCompile(`File size:(\s)+(\d+)\s.+`)
				var pdfVersion string
				pdfVersionRe := regexp.MustCompile(`PDF version:(\s)+(.+)`)

				for _, s := range sArray {
					s = strings.Replace(s, ",", "_", -1) // Replace comma to underscore

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
					} else if javaScriptRe.MatchString(s) == true {
						javaScript = javaScriptRe.ReplaceAllString(s, "$2")
						javaScript = strings.TrimRight(javaScript, "\n\r")
					} else if pagesRe.MatchString(s) == true {
						pages = pagesRe.ReplaceAllString(s, "$2")
						pages = strings.TrimRight(pages, "\n\r")
					} else if encryptedRe.MatchString(s) == true {
						encrypted = encryptedRe.ReplaceAllString(s, "$2")
						encrypted = strings.TrimRight(encrypted, "\n\r")
					} else if pageRotRe.MatchString(s) == true {
						pageRot = pageRotRe.ReplaceAllString(s, "$2")
						pageRot = strings.TrimRight(pageRot, "\n\r")
					} else if fileSizeRe.MatchString(s) == true {
						fileSize = fileSizeRe.ReplaceAllString(s, "$2")
						fileSize = strings.TrimRight(fileSize, "\n\r")
						convertedStrFloat64, _ := strconv.ParseFloat(fileSize, 64)
						fileSize = convertByte2MB(convertedStrFloat64)
					} else if pdfVersionRe.MatchString(s) == true {
						pdfVersion = pdfVersionRe.ReplaceAllString(s, "$2")
						pdfVersion = strings.TrimRight(pdfVersion, "\n\r")
					}
				}

				// Remove unnecessary characters from "path".
				replacedPath := strings.Replace(path, rootDir+"\\out", "", 1)

				// Write to csv.
				fmt.Fprintln(csvFile, replacedPath+","+author+","+creator+","+producer+","+creationDate+","+modDate+","+pagesize+","+javaScript+","+pages+","+encrypted+","+pageRot+","+fileSize+","+pdfVersion)
			}
		}

		csvFile.Close()

		// Use the nkf command to add a BOM.
		errNkf := exec.Command("nkf", "-w8", "--overwrite", rootDir+"\\"+resultFile).Run()
		log.Println("nkf", "-w8", "--overwrite", rootDir+"\\"+resultFile)
		if errNkf != nil {
			fmt.Println("nkf command Exec Error")
		}

		if flag == 0 {
			// If there is no pdf, return this.
			c.String(http.StatusOK, "There is no pdf in the uploaded zip.")
		} else {
			// If there is a pdf, return a csv.
			// Rewrite index.tmpl to make it download from HTML.
			c.HTML(http.StatusOK, "html/index.tmpl", gin.H{
				"title":           "PDF Info",
				"downloadMessage": "Download: ",
				"downloadfile":    tFormat + ".csv",
			})
		}
	})

	r.Run(":14")
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

// Convert bytes to megabytes
func convertByte2MB(fis float64) string {
	fisMB := fis / 1048576
	fisMBstr := strconv.FormatFloat(fisMB, 'f', 2, 64)
	return fisMBstr
}
