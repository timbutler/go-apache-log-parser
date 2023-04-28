package apachelogparser

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Line : Represents a line in standard Apache log
type Line struct {
	RemoteHost string
	Time       time.Time
	Request    string
	Status     int
	Bytes      int
	Referer    string
	UserAgent  string
	URL        string
	Method     string
	Protocol   string
}

func (li *Line) String() string {
	return fmt.Sprintf(
		"%s\t%s\t%s\t%d\t%d\t%s\t%s\t%s\t%s\t%s",
		li.RemoteHost,
		li.Time,
		li.Request,
		li.Status,
		li.Bytes,
		li.Referer,
		li.UserAgent,
		li.URL,
		li.Method,
		li.Protocol,
	)
}

func scanreader(ioreader io.Reader) ([]string, error) {
	var lines []string
	var scanner *bufio.Scanner

	// Create a bufio scanner to read the file line by line
	scanner = bufio.NewScanner(ioreader)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return lines, err
	}

	return lines, nil
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	var tarReader *tar.Reader

	// .tar.gz and.tgz are generally all gzipped tar archives
	if strings.HasSuffix(path, ".gz") || strings.HasSuffix(path, ".tgz") {
		gzipReader, err := gzip.NewReader(file)
		if err != nil {
			return lines, err
		}
		defer gzipReader.Close()

		tarReader = tar.NewReader(gzipReader)

		if tarReader != nil {
			// Loop through each file in the tar archive
			for {
				header, err := tarReader.Next()
				if err == io.EOF {
					// End of archive
					break
				}
				if err != nil {
					return lines, err
				}

				// Only parse if it's a file
				if header.Typeflag != tar.TypeReg {
					continue
				}

				return scanreader(tarReader)
			}
		} else {
			// if it's not a tar file, assume it's just gzipped
			return scanreader(gzipReader)
		}
	} else {
		// If it's not a tar.gz, assume it's a plain text file
		return scanreader(file)
	}
	return lines, errors.New("Unable to parse file")
}

func parse(file string) ([]Line, error) {
	var items []Line

	lines, err := readLines(file)
	if err != nil {
		log.Fatalf("readLines: %s", err)
	}
	for _, line := range lines {
		var buffer bytes.Buffer
		buffer.WriteString(`^(\S+)\s`)                  // 1) IP
		buffer.WriteString(`\S+\s+`)                    // remote logname
		buffer.WriteString(`(?:\S+\s+)+`)               // remote user
		buffer.WriteString(`\[([^]]+)\]\s`)             // 2) date
		buffer.WriteString(`"(\S*)\s?`)                 // 3) method
		buffer.WriteString(`(?:((?:[^"]*(?:\\")?)*)\s`) // 4) URL
		buffer.WriteString(`([^"]*)"\s|`)               // 5) protocol
		buffer.WriteString(`((?:[^"]*(?:\\")?)*)"\s)`)  // 6) or, possibly URL with no protocol
		buffer.WriteString(`(\S+)\s`)                   // 7) status code
		buffer.WriteString(`(\S+)\s`)                   // 8) bytes
		buffer.WriteString(`"((?:[^"]*(?:\\")?)*)"\s`)  // 9) referrer
		buffer.WriteString(`"(.*)"$`)                   // 10) user agent

		re1, err := regexp.Compile(buffer.String())
		if err != nil {
			log.Fatalf("regexp: %s", err)
		}
		result := re1.FindStringSubmatch(line)

		lineItem := new(Line)
		lineItem.RemoteHost = result[1]
		// [05/Oct/2014:04:06:21 -0500]
		value := result[2]
		layout := "02/Jan/2006:15:04:05 -0700"
		t, _ := time.Parse(layout, value)
		lineItem.Time = t
		lineItem.Request = result[3] + " " + result[4] + " " + result[5]
		lineItem.Method = result[3]
		lineItem.Protocol = result[5]
		status, err := strconv.Atoi(result[7])
		if err != nil {
			status = 0
		}
		bytes, err := strconv.Atoi(result[8])
		if err != nil {
			bytes = 0
		}
		lineItem.Status = status
		lineItem.Bytes = bytes
		lineItem.Referer = result[9]
		lineItem.UserAgent = result[10]
		url := result[4]
		altURL := result[6]
		if url == "" && altURL != "" {
			url = altURL
		}
		lineItem.URL = url
		items = append(items, *lineItem)
		//for k, v := range result {
		//	fmt.Printf("%d. %s\n", k, v)
		//}
	}
	return items, nil
}

// Parse : Parse the log file
func Parse(file string) ([]Line, error) {
	lines, err := parse(file)
	if err != nil {
		return nil, err
	}
	return lines, nil
}
