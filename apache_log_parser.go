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
	"path/filepath"
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

func scanreader(ioreader io.Reader) (lines []string, err error) {
	// Create a bufio scanner to read the file line by line
	scanner := bufio.NewScanner(ioreader)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return lines, err
	}

	return lines, nil
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(filepath.Clean(path))
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

		if strings.HasSuffix(path, ".tar.gz") || strings.HasSuffix(path, ".tgz") {
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
			}
		} else {
			// if it's not a tar file, assume it's just gzipped
			return scanreader(gzipReader)
		}
	} else {
		// If it's not a tar.gz, assume it's a plain text file
		return scanreader(file)
	}
	return lines, errors.New("unable to parse file")
}

// ParseLine - Parses a single line of the Apache log
func ParseLine(inputline string) *Line {
	resultline := &Line{}

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

	re, err := regexp.Compile(buffer.String())
	if err != nil {
		log.Fatalf("Failed to compile regex: %s", err)
	}

	result := re.FindStringSubmatch(inputline)

	resultline.RemoteHost = result[1]
	// [05/Oct/2014:04:06:21 -0500]
	value := result[2]
	layout := "02/Jan/2006:15:04:05 -0700"
	t, _ := time.Parse(layout, value)
	resultline.Time = t
	resultline.Request = result[3] + " " + result[4] + " " + result[5]
	resultline.Method = result[3]
	resultline.Protocol = result[5]
	status, err := strconv.Atoi(result[7])
	if err != nil {
		status = 0
	}
	bytessent, err := strconv.Atoi(result[8])
	if err != nil {
		bytessent = 0
	}
	resultline.Status = status
	resultline.Bytes = bytessent
	resultline.Referer = result[9]
	resultline.UserAgent = result[10]
	url := result[4]
	altURL := result[6]
	if url == "" && altURL != "" {
		url = altURL
	}
	resultline.URL = url

	return resultline
}

// Parse : Parse the log file
func Parse(file string) ([]Line, error) {
	var items []Line

	lines, err := readLines(file)
	if err != nil {
		log.Fatalf("readLines: %s", err)
	}
	for _, line := range lines {
		lineItem := ParseLine(line)
		items = append(items, *lineItem)
	}
	return items, nil
}
