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
	"strings"
	"time"

	"github.com/nfnt/resize"
)

var binary = "C:\\Users\\jof4002\\AppData\\Local\\Bandizip\\Bandizip.exe"

// func RemoveContents(dir string) error {
// 	d, err := os.Open(dir)
// 	if err != nil {
// 		return err
// 	}
// 	defer d.Close()
// 	names, err := d.Readdirnames(-1)
// 	if err != nil {
// 		return err
// 	}
// 	for _, name := range names {
// 		err = os.RemoveAll(filepath.Join(dir, name))
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

func processFile(imgpath, basedir string) {
	fmt.Println(imgpath)

	filename := filepath.Base(imgpath)
	filenamelen := len(filename) - len(filepath.Ext(imgpath))
	if filenamelen < 3 {
		filename = strings.Repeat("0", 3-filenamelen) + filename
	}

	newpath := basedir + "_re\\" + filename
	//fmt.Println(newpath)
	os.MkdirAll(filepath.Dir(newpath), os.ModePerm)

	imgfile, err := os.Open(imgpath)
	if err != nil {
		fmt.Println(err)
		time.Sleep(10 * time.Second)
		return
	}
	defer imgfile.Close()

	// get image type
	//image, imageType, err := image.Decode(imgfile)
	image, imageType, err := image.Decode(imgfile)
	if imageType == "gif" || err != nil {
		// just copy
		imgfile.Seek(0, 0)

		to, err := os.OpenFile(newpath, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			log.Fatal(err)
			time.Sleep(10 * time.Second)
		}
		defer to.Close()

		_, err = io.Copy(to, imgfile)
		if err != nil {
			log.Fatal(err)
			time.Sleep(10 * time.Second)
		}
		return
	}

	//m := resize.Thumbnail(2048, 2048, image, resize.Lanczos3)
	m := resize.Thumbnail(2048, 2048, image, resize.Lanczos3)

	ext := path.Ext(newpath)
	outfilepath := newpath[:len(newpath)-len(ext)] + ".jpg"

	out, err := os.Create(outfilepath)
	if err != nil {
		log.Fatal(err)
		time.Sleep(10 * time.Second)
	}
	defer out.Close()

	// write new image to file
	jpeg.Encode(out, m, nil)
}

func processDirectory(dirpath string) {

	fmt.Println(dirpath)

	err := filepath.Walk(dirpath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		processFile(path, dirpath)
		return nil
	})

	if err != nil {
		fmt.Printf("walk error [%v]\n", err)
		time.Sleep(10 * time.Second)
		return
	}

	// compress
	candiname := dirpath
	ext := ".zip"
	for {
		if _, err := os.Stat(candiname + ext); os.IsNotExist(err) {
			break
		}
		candiname += "_re"
	}

	cmd := exec.Command(binary, "a", candiname+ext, dirpath+"_re")
	execErr := cmd.Run()
	if execErr != nil {
		panic(execErr)
	}
	//	C:\Users\jof4002\AppData\Local\Bandizip\Bandizip.exe a "%arg3%_re.zip" "%arg3%\re\"

	os.RemoveAll(dirpath + "_re")
}

func processArchive(path string) {
	ext := filepath.Ext(path)

	// -o:outputpath
	extractpath := strings.TrimSpace(path[:len(path)-len(ext)])

	cmd := exec.Command(binary, "x", "-o:"+extractpath, path)

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
