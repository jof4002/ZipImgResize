package main

import (
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/nfnt/resize"
)

//var zipApp = "C:\\Users\\jof4002\\AppData\\Local\\Bandizip\\Bandizip.exe"
var zipApp = "C:\\Users\\jof4002\\AppData\\Local\\Bandizip\\bc.exe"

// imgpath = basdir/subdir/imgname.ext
func processFile(imgpath, basedir string, wg *sync.WaitGroup) {
	defer wg.Done()

	// imgname.ext
	imgname := filepath.Base(imgpath)
	fmt.Println(imgname)
	// /subdir/
	subdir := imgpath[len(basedir) : len(imgpath)-len(imgname)]
	// append 0 to imgname.ext
	imgnamelen := len(imgname) - len(filepath.Ext(imgpath))
	if imgnamelen < 3 {
		imgname = strings.Repeat("0", 3-imgnamelen) + imgname
	}

	newpath := basedir + "_re\\" + subdir + imgname
	//fmt.Println(newpath)
	// create folder basedir_re/subdir/
	os.MkdirAll(filepath.Dir(newpath), os.ModePerm)

	imgfile, err := os.Open(imgpath)
	if err != nil {
		log.Fatal(err)
		time.Sleep(10 * time.Second)
		return
	}
	defer imgfile.Close()

	// decode image
	image, imageType, err := image.Decode(imgfile)
	// gif or not image - just copy
	if imageType == "gif" || err != nil {
		imgfile.Seek(0, 0)

		to, err := os.OpenFile(newpath, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			log.Fatal(err)
			time.Sleep(10 * time.Second)
			return
		}
		defer to.Close()

		_, err = io.Copy(to, imgfile)
		if err != nil {
			log.Fatal(err)
			time.Sleep(10 * time.Second)
			return
		}
		return
	}

	//m := resize.Thumbnail(2048, 2048, image, resize.Lanczos3)
	m := resize.Thumbnail(2048, 2048, image, resize.Lanczos3)

	ext := path.Ext(newpath)
	// imgname.ext to imgname.jpg
	outfilepath := newpath[:len(newpath)-len(ext)] + ".jpg"

	out, err := os.Create(outfilepath)
	if err != nil {
		log.Fatal(err)
		time.Sleep(10 * time.Second)
		return
	}
	defer out.Close()

	// write new image to file
	jpeg.Encode(out, m, nil)
}

// convert folder A to folder A_re, compress A_re to A(_re)*.zip delete folder A_re
func processDirectory(dirpath string) {

	var wg sync.WaitGroup

	err := filepath.Walk(dirpath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		wg.Add(1)
		processFile(path, dirpath, &wg)
		return nil
	})

	wg.Wait()

	if err != nil {
		log.Fatal(err)
		time.Sleep(10 * time.Second)
		return
	}

	// compress
	candiname := dirpath
	ext := ".zip"
	// add _re until not exist file
	for {
		if _, err := os.Stat(candiname + ext); os.IsNotExist(err) {
			break
		}
		candiname += "_re"
	}

	fmt.Println(dirpath + "_re")
	// bandizip.exe doesn't save top folder name, but bc.exe does, so use *.* with -r
	cmd := exec.Command(zipApp, "c", "-r", candiname+ext, dirpath+"_re\\*.*")
	execErr := cmd.Run()
	if execErr != nil {
		panic(execErr)
	}

	// delete folder A_re
	os.RemoveAll(dirpath + "_re")
}

// extract zip A to folder A, process folder A, delete folder A
func processArchive(path string) {
	ext := filepath.Ext(path)

	// -o:outputpath
	extractpath := strings.TrimSpace(path[:len(path)-len(ext)])

	cmd := exec.Command(zipApp, "x", "-o:"+extractpath, path)

	execErr := cmd.Run()
	if execErr != nil {
		panic(execErr)
	}

	processDirectory(extractpath)

	os.RemoveAll(extractpath)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("No file")
		return
	}
	runtime.GOMAXPROCS(runtime.NumCPU() - 1)
	lenarg := len(os.Args)
	for i := 1; i < lenarg; i++ {
		path := os.Args[i]
		fmt.Println(path)
		fileInfo, err := os.Stat(path)
		if err != nil {
			fmt.Println(err)
			time.Sleep(10 * time.Second)
			continue
		}
		if fileInfo.IsDir() {
			processDirectory(path)
		} else {
			processArchive(path)
		}
	}
}
