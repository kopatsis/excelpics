package main

import (
	"bufio"
	"bytes"
	"fmt"

	"image"
	_ "image/jpeg"
	"image/png"
	"io"

	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"unicode"

	"github.com/xuri/excelize/v2"
	"golang.org/x/image/draw"
)

const HEIGHT = 100.0
const scaleYL = 0.75
const scaleXL = 0.124696673

func getNums(cell string) int {
	var str string
	for _, char := range cell {
		if unicode.IsDigit(char) {
			str += string(char)
		}
	}
	if str == "" {
		return 0
	}
	ret, _ := strconv.Atoi(str)
	return ret
}

func getInputs() map[string]any {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter file name (include fill path if not in WORKING): ")
	filename, err := reader.ReadString('\n')
	filename = filename[:len(filename)-2]
	if err != nil {
		fmt.Println("error: " + err.Error())
		return nil
	}
	if !strings.Contains(filename, "C:/") {
		filename = "C:/Users/demet/Dropbox/Product Files/WORKING/" + filename
	}
	ret := map[string]any{"filename": filename}

	fmt.Print(`Enter sheet name and range for image URLs (ex. "Sheet1!A1:A3): `)
	urlRange, err := reader.ReadString('\n')
	urlRange = urlRange[:len(urlRange)-2]
	if err != nil {
		fmt.Println("error: " + err.Error())
		return nil
	}
	if !strings.Contains(urlRange, "!") || !strings.Contains(urlRange, ":") {
		fmt.Println("Error, incorrect format")
		return nil
	}
	urlSheet := urlRange[:strings.Index(urlRange, "!")]
	urlRange = urlRange[1+strings.Index(urlRange, "!"):]
	urlStart := urlRange[:strings.Index(urlRange, ":")]
	urlEnd := urlRange[1+strings.Index(urlRange, ":"):]
	ret["urlSheet"] = urlSheet
	ret["urlEnd"] = urlEnd
	ret["urlStart"] = urlStart

	fmt.Print(`Enter sheet name and range for cells where image will be added (ex. "Sheet2!B1:B3): `)
	imgRange, err := reader.ReadString('\n')
	imgRange = imgRange[:len(imgRange)-2]
	if err != nil {
		fmt.Println("error: " + err.Error())
		return nil
	}
	if !strings.Contains(imgRange, "!") || !strings.Contains(imgRange, ":") {
		fmt.Println("Error, incorrect format")
		return nil
	}
	imgSheet := imgRange[:strings.Index(imgRange, "!")]
	imgRange = imgRange[1+strings.Index(imgRange, "!"):]
	imgStart := imgRange[:strings.Index(imgRange, ":")]
	imgEnd := imgRange[1+strings.Index(imgRange, ":"):]
	ret["imgSheet"] = imgSheet
	ret["imgEnd"] = imgEnd
	ret["imgStart"] = imgStart

	uStartNum, uEndNum, iStartNum, iEndNum := getNums(urlStart), getNums(urlEnd), getNums(imgStart), getNums(imgEnd)

	if (uEndNum-uStartNum) != (iEndNum-iStartNum) || uStartNum == 0 || uEndNum == 0 || iStartNum == 0 || iEndNum == 0 {
		fmt.Println("Error in ranges")
		return nil
	}

	ret["rangelen"] = 1 + uEndNum - uStartNum

	return ret
}

func PictureInfoMW(url string, it int) (int, error) {
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	imgData, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	srcImg, format, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		return 0, err
	}

	bounds := srcImg.Bounds()
	origWidth := bounds.Dx()
	origHeight := bounds.Dy()

	if origHeight == 0 {
		return 0, fmt.Errorf("invalid image height")
	}

	maxHeight := 200
	scale := float64(maxHeight) / float64(origHeight)
	newWidth := int(float64(origWidth) * scale)
	newHeight := maxHeight

	dstImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.CatmullRom.Scale(dstImg, dstImg.Bounds(), srcImg, bounds, draw.Over, nil)

	outFile, err := os.Create("current.png")
	if err != nil {
		return 0, err
	}
	defer outFile.Close()

	err = png.Encode(outFile, dstImg)
	if err != nil {
		return 0, err
	}

	fmt.Printf("Successfully got image %d (%s)\n", it+1, format)

	pScale := float64(HEIGHT) / float64(origHeight)
	width := int(scaleXL * math.Round(pScale*float64(origWidth)))

	return width, nil
}

func PictureInfo(url string, it int) (int, error) {

	if strings.HasPrefix(url, "INT::") {
		url = strings.TrimPrefix(url, "INT::")
		cwd, err := os.Getwd()
		if err != nil {
			return 0, err
		}
		filePath := fmt.Sprintf("%s\\%s", cwd, url)
		file, err := os.Open(filePath)
		if err != nil {
			return 0, err
		}
		defer file.Close()

		config, _, err := image.DecodeConfig(file)
		if err != nil {
			return 0, err
		}

		width := int(scaleXL * math.Round(float64(HEIGHT)/float64(config.Height)*float64(config.Width)))
		fmt.Printf("Successfully got image %d\n", it+1)

		file.Seek(0, 0) // Reset the file pointer to the beginning
		fileCopy, err := os.Create("current.jpg")
		if err != nil {
			return 0, err
		}
		defer fileCopy.Close()

		_, err = io.Copy(fileCopy, file)
		if err != nil {
			return 0, err
		}

		return width, nil
	}

	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	name := "current.jpg"
	file, err := os.Create(name)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	var buffer bytes.Buffer
	_, err = io.Copy(&buffer, resp.Body)
	if err != nil {
		fmt.Println("1")
		return 0, err
	}

	reader1 := bytes.NewReader(buffer.Bytes())
	reader2 := bytes.NewReader(buffer.Bytes())

	_, err = io.Copy(file, reader1)
	if err != nil {
		fmt.Println("2")
		return 0, err
	}

	config, _, err := image.DecodeConfig(reader2)
	if err != nil {
		fmt.Println("3")
		return 0, err
	}

	fmt.Printf("Successfully got image %d\n", it+1)

	pScale := float64(HEIGHT) / float64(config.Height)
	width := int(scaleXL * math.Round(pScale*float64(config.Width)))

	return width, nil
}

func main() {

	inputs := getInputs()

	if inputs == nil {
		return
	}

	f, err := excelize.OpenFile(inputs["filename"].(string))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	maxWidth := 0
	for i := 0; i < inputs["rangelen"].(int); i++ {
		col, row, _ := excelize.CellNameToCoordinates(inputs["urlStart"].(string))
		currentURLCell, _ := excelize.CoordinatesToCellName(col, row+i)

		URL, err := f.GetCellValue(inputs["urlSheet"].(string), currentURLCell)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if URL == "" || URL == "0" {
			fmt.Println("Blank URL")
			continue
		}

		width, err := PictureInfo(URL, i)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if width > maxWidth {
			maxWidth = width
		}

		icol, irow, _ := excelize.CellNameToCoordinates(inputs["imgStart"].(string))
		currentIMGCell, _ := excelize.CoordinatesToCellName(icol, irow+i)

		if err := f.SetRowHeight(inputs["imgSheet"].(string), irow+i, float64(HEIGHT)*float64(scaleYL)); err != nil {
			fmt.Println(err)
			continue
		}

		enable := true
		if err := f.AddPicture(inputs["imgSheet"].(string), currentIMGCell, "current.jpg", &excelize.GraphicOptions{
			PrintObject:     &enable,
			LockAspectRatio: true,
			AutoFit:         true,
			Positioning:     "oneCell",
		}); err != nil {
			fmt.Println(err)
			continue
		}

		if err := os.Remove("current.jpg"); err != nil {
			fmt.Println(err)
			continue
		}

	}

	if err := f.SaveAs(inputs["filename"].(string)); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Images saved to xlsx file")
	}

}
