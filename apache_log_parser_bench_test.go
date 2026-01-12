package apachelogparser

import (
	"testing"
)

var testLogLine = `127.0.0.1 - - [05/Oct/2014:04:34:35 -0500] "GET /test.html HTTP/1.1" 200 6776 "http://www.example.com/index.html" "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:32.0) Gecko/20100101 Firefox/32.0"`

var testLogLines = []string{
	`127.0.0.1 - - [05/Oct/2014:04:34:35 -0500] "GET /test.html HTTP/1.1" 200 6776 "http://www.example.com/index.html" "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:32.0) Gecko/20100101 Firefox/32.0"`,
	`192.168.1.1 - - [10/Oct/2014:13:55:36 -0700] "POST /api/data HTTP/1.1" 201 1234 "-" "curl/7.35.0"`,
	`10.0.0.1 - - [15/Oct/2014:08:22:15 +0000] "PUT /upload/file.txt HTTP/1.1" 204 0 "https://example.com/upload" "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_9_5)"`,
	`172.16.0.1 - - [20/Oct/2014:16:45:22 -0500] "DELETE /resource/123 HTTP/1.1" 404 512 "-" "Python-urllib/3.4"`,
	`192.168.0.100 - - [25/Oct/2014:09:12:33 +0100] "GET /images/logo.png HTTP/1.0" 304 0 "http://www.site.com/" "Mozilla/5.0 (iPhone; CPU iPhone OS 8_0)"`,
}

func BenchmarkParseLine(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseLine(testLogLine)
	}
}

func BenchmarkParseLineMultiple(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, line := range testLogLines {
			ParseLine(line)
		}
	}
}

func BenchmarkParseLineParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ParseLine(testLogLine)
		}
	})
}
