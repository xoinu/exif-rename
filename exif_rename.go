//
// This is public domain software.
//
package main

import (
	"crypto/md5"
	"fmt"
	"github.com/rwcarlsen/goexif/exif"
	"io/ioutil"
	"log"
	"os"
	"time"
	"path/filepath"
	"strings"
)
//-----------------------------------------------------------------------------
// getDateTimeFromExif reads EXIF in Jpeg file and get the date when the photo
// was taken.
func getDateTimeFromExif(fname string) (tm time.Time, err error) {
	f, err := os.Open(fname)
	if err != nil {
		return
	}

	x, err := exif.Decode(f)
    f.Close()
	if err != nil {
		return
	}

	tm, err = x.DateTime()
    return    
}
//-----------------------------------------------------------------------------
// getDateTimeFromJpen first try to get the date when the photo was taken using
// EXIF information. Physical file timestamp is used as fallback.
func getDateTimeFromJpeg(fname string) (tm time.Time, err error) {
    tm, err = getDateTimeFromExif(fname)
    if err == nil {
        return
    }
    
    log.Println(
        "Failed to get timestamp from EXIF. " +
        "Try to use physical time stamp instead...")
    
    info, err := os.Stat(fname)
    if err != nil {
        return
    }
    
    tm = info.ModTime()
    return
}
//-----------------------------------------------------------------------------
// renameJpeg renames JPEG file into standard name that consists of timestamp
// and MD5 hash value using EXIF data in following format:
// "YYYY-MM-DD-HHmmss-<MD5 hash>.jpg"
// If EXIF is not available, physical file timestamp is used instead.
func renameJpeg(fname string) (err error) {
    tm, err := getDateTimeFromJpeg(fname)
	if err != nil {
		return err
	}

	buff, err := ioutil.ReadFile(fname)
	if err != nil {
		return err
	}
    
    newName := fmt.Sprintf(
		"%04d-%02d-%02d-%02d%02d%02d-%x.jpg",
		tm.Year(),
		tm.Month(),
		tm.Day(),
		tm.Hour(),
		tm.Minute(),
		tm.Second(),
		md5.Sum(buff))
        
    if newName == filepath.Base(fname) {
        // Exact same file. There's nothing to do.
        log.Println("No need to rename", fname)
        return
    }
    
    newPath := filepath.Join(filepath.Dir(fname), newName)

    log.Println(fname, "=>", newPath)
	err = os.Rename(fname, newPath)
	return err
}
//-----------------------------------------------------------------------------
// main function is the entry point of this command line tool.
func main() {
    count := 0
    
    for i, arg := range os.Args[1:] {
        log.Printf("ARGV[%d] = %s\n", i, arg)
        if !strings.HasSuffix(strings.ToUpper(arg), ".JPG") {
            log.Println("Skipped unexpected input file type:", arg)
            continue
        }

        if _, err := os.Stat(arg); os.IsNotExist(err) {
            log.Println("Input file not found:", arg)
            continue
        }
        
        if err := renameJpeg(arg); err != nil {
            log.Println(err)
            continue
        }
        
        count++
    }
    
    if count > 0 {
        log.Println(count, "file(s) are successfully processed.")
    }
}
