package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"path"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/spf13/cast"

	"github.com/fyne-csvspliter/cutheme"
)

var (
	mainApp fyne.App
)

func init() {
	mainApp = app.New()
	mainApp.Settings().SetTheme(&cutheme.MyTheme{})
}

func main() {
	w := mainApp.NewWindow("CSV文件礼包码拆分工具")
	w.CenterOnScreen()
	w.Resize(fyne.NewSize(1000, 500))

	var (
		filePathLabel     = widget.NewLabel("")
		sourceColumnIndex = widget.NewEntry()
		maxLinesPerFile   = widget.NewEntry()
		outputFolderLabel = widget.NewLabel("")
	)

	// set default value
	maxLinesPerFile.SetText("10000")
	sourceColumnIndex.SetText("4")

	form := &widget.Form{
		Items: []*widget.FormItem{ // we can specify items in the constructor
			{Text: "原csv文件路径", Widget: filePathLabel},
			{Text: "", Widget: widget.NewButton("请选择原csv文件", func() {
				dialog.ShowFileOpen(func(closer fyne.URIReadCloser, err error) {
					if err != nil {
						dialog.ShowError(err, w)
						return
					}
					if closer == nil {
						return
					}
					defer closer.Close()
					filePathLabel.SetText(fmt.Sprintf("%s", closer.URI().Path()))
				}, w)
			})},
			{Text: "原csv文件导出列号", Widget: sourceColumnIndex},
			{Text: "切分文件最大行", Widget: maxLinesPerFile},
			{Text: "输出文件夹", Widget: outputFolderLabel},
			{Text: "", Widget: widget.NewButton("请选择输出文件夹", func() {
				dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
					if err != nil {
						dialog.ShowError(err, w)
						return
					}
					if uri == nil {
						return
					}
					outputFolderLabel.SetText(fmt.Sprintf("%s", uri.Path()))
				}, w)
			})},
		},
		SubmitText: "执行", // optional, defaults to "Submit"
		OnSubmit: func() { // optional, handle form submission
			log.Println("File Path:", filePathLabel.Text)
			log.Println("Source Column Index:", sourceColumnIndex.Text)
			log.Println("Max Line Per File:", maxLinesPerFile.Text)
			log.Println("Output Folder:", outputFolderLabel.Text)
			dialog.ShowConfirm("请确认", "是否确认执行？", func(b bool) {
				if b {
					infinite := widget.NewProgressBarInfinite()
					customDialog := dialog.NewCustom("提示", "执行中...", infinite, w)
					customDialog.Show()
					time.Sleep(time.Second * 2)
					process(context.Background(), strings.TrimSpace(filePathLabel.Text), cast.ToInt(sourceColumnIndex.Text), cast.ToInt(maxLinesPerFile.Text), outputFolderLabel.Text)
					infinite.Stop()
					customDialog.Hide()
					dialog.ShowInformation("提示", "执行完成！请去目标文件夹查看拆分后文件", w)
				}
			}, w)

		},
	}

	w.SetContent(form)

	w.ShowAndRun()
}

func process(ctx context.Context, filePath string, takeColumnIndex int, maxLinesPerFile int, outputFolderPath string) {
	if takeColumnIndex <= 0 {
		fmt.Println("执行失败: 请填写正确的列数")
		return
	}

	if maxLinesPerFile <= 0 {
		fmt.Println("执行失败: 请填写正确的最大行数")
		return
	}

	uri := storage.NewFileURI(filePath)

	// Open the CSV file
	file, err := storage.LoadResourceFromURI(uri)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}

	// Use the filepath.Base function to get the file name
	fileName := filepath.Base(filePath)

	// Use the filepath.Ext function to get the extension of the file
	fileExt := filepath.Ext(fileName)
	if fileExt != ".csv" {
		fmt.Println("执行失败: 请上传正确的csv文件")
		return
	}

	// Remove the extension from the file name
	fileNameWithoutExt := fileName[0 : len(fileName)-len(fileExt)]

	fmt.Printf("文件名: 【%s】\n", fileName)
	fmt.Printf("截取第【%d】列数据\n", takeColumnIndex)
	fmt.Printf("每个文件包含的最大行数: 【%d】\n", maxLinesPerFile)

	// Create a new CSV reader
	b := new(bytes.Buffer)
	b.Write(file.Content())
	reader := csv.NewReader(b)
	reader.FieldsPerRecord = -1
	// Slice to hold the second column data
	var (
		isFirstRow = true
		lineMap    = make(map[string]string)
	)

	// Read the CSV file line by line
	for {
		// Read one record (line) at a time
		record, err := reader.Read()
		if err != nil {
			// Stop reading on any error including EOF
			// This example does not differentiate between EOF and other errors
			// for simplicity. You may wish to handle io.EOF specifically.
			break
		}

		if isFirstRow {
			isFirstRow = false
			continue
		}

		// record is a slice of strings holding each field in the current line
		if len(record) >= takeColumnIndex { // Make sure the line has at least two columns
			// 去重
			codeItem := record[takeColumnIndex-1]
			codeKey := strings.ToLower(codeItem)
			lineMap[codeKey] = codeItem
		} else {
			// Handle the case where there are not enough columns
			fmt.Println("Warning: line with not enough columns encountered")
		}
	}

	// Print the second column data

	var (
		outputStringBuffer *bytes.Buffer
		outputFileIndex    = 1
		lineNumber         = 0
	)

	for _, val := range lineMap {
		if lineNumber%maxLinesPerFile == 0 {
			if outputStringBuffer != nil {
				newFilePath := path.Join(outputFolderPath, fmt.Sprintf("%s_%d.csv", fileNameWithoutExt, outputFileIndex))
				newUri := storage.NewFileURI(newFilePath)
				uriWriteCloser, err := storage.Writer(newUri)
				if err != nil {
					fmt.Println("Error creating new CSV file:", err)
					return
				}
				_, err = uriWriteCloser.Write([]byte(outputStringBuffer.String()))
				if err != nil {
					fmt.Println("Error writing to CSV file:", err)
					return
				}
				uriWriteCloser.Close()
				outputFileIndex++
			}

			// create new buffer
			outputStringBuffer = new(bytes.Buffer)
			outputStringBuffer.WriteString("名称,礼包码\n")
			if err != nil {
				fmt.Println("Error creating new CSV file:", err)
				return
			}

		}

		_, err = outputStringBuffer.WriteString(fmt.Sprintf("%s,%s\n", fileNameWithoutExt, val))
		if err != nil {
			fmt.Println("Error writing to CSV file:", err)
			return
		}
		lineNumber++
	}

	if outputStringBuffer != nil {
		newFilePath := path.Join(outputFolderPath, fmt.Sprintf("%s_%d.csv", fileNameWithoutExt, outputFileIndex))
		newUri := storage.NewFileURI(newFilePath)
		uriWriteCloser, err := storage.Writer(newUri)
		if err != nil {
			fmt.Println("Error creating new CSV file:", err)
			return
		}
		_, err = uriWriteCloser.Write([]byte(outputStringBuffer.String()))
		if err != nil {
			fmt.Println("Error writing to CSV file:", err)
			return
		}
		uriWriteCloser.Close()
		outputFileIndex++
	}
}
