package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/graniticio/inifile"
)

var (
	dir string
)

func main() {
	flag.StringVar(&dir, "dir", "", "Absolute path for target directory")

	flag.Parse()
	if len(dir) < 1 {
		log.Fatal("No --dir is given")
	}

	urlFiles := Scan(dir, ".url")
	urlFilesLen := len(urlFiles)
	if urlFilesLen < 1 {
		log.Fatal("No .url file found")
	}
	fmt.Printf("There are %d url files\n", len(urlFiles))

	file, err := os.Create(fmt.Sprintf("%s.txt", getFolderName(dir)))
	errExit(err)
	defer file.Close()
	w := bufio.NewWriter(file)

	for _, s := range urlFiles {
		ic, err := inifile.NewIniConfigFromPath(s)
		errExit(err)
		url, err := ic.Value("InternetShortcut", "URL")
		errExit(err)
		fmt.Println("checking", url, ", in", s, "...")
		protocol := url[0:strings.Index(url, `://`)]
		if protocol == `http` || protocol == `https` {
			title, err := getTitle(url)
			errExit(err)
			fmt.Fprintf(w, "- [%s](%s)\n", title, url)
		} else {
			fmt.Fprintf(w, "- [%s](%s)\n", url, url)
		}
	}
	errExit(w.Flush())

	// for _, s := range urlFiles {
	// 	errExit(os.Remove(s))
	// }
}

func errExit(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func Scan(root, ext string) []string {
	var a []string
	filepath.WalkDir(root, func(s string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if filepath.Ext(d.Name()) == ext {
			a = append(a, s)
		}
		return nil
	})
	return a
}

func getTitle(urlstr string) (string, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	var title string

	chromedp.ListenTarget(ctx, func(ev interface{}) {

		switch ev := ev.(type) {

		case *network.EventResponseReceived:
			resp := ev.Response
			if resp.URL == urlstr {
				log.Printf("received headers: %s %s", resp.URL, resp.MimeType)
				if resp.MimeType != "text/html" {
					chromedp.Cancel(ctx)
				}

				if strings.Contains(resp.URL, "youtube.com") {
					log.Printf("YT!!")
				}

				// may be redirected
				switch ContentType := resp.Headers["Content-Type"].(type) {
				case string:
					// here v has type T
					if !strings.Contains(ContentType, "text/html") {
						chromedp.Cancel(ctx)
					}
				}

				switch ContentType := resp.Headers["content-type"].(type) {
				case string:
					// here v has type T
					if !strings.Contains(ContentType, "text/html") {
						chromedp.Cancel(ctx)
					}
				}
			}
		}
	})

	req := `
(async () => new Promise((resolve, reject) => {
	var handle = NaN;

	(function animate() {
		if (!isNaN(handle)) {
			clearTimeout(handle);
		}

		if (document.title.length > 0 && !document.title.startsWith("http")) {
			resolve(document.title);
		} else {
			handle = setTimeout(animate, 1000);
		}
	}());
}));
`
	err := chromedp.Run(ctx,
		chromedp.Navigate(urlstr),
		//chromedp.Evaluate(`window.location.href`, &res),
		chromedp.Evaluate(req, nil, func(p *runtime.EvaluateParams) *runtime.EvaluateParams {
			return p.WithAwaitPromise(true)
		}),
		chromedp.Title(&title),
	)
	if err == context.Canceled {
		// url as title
		log.Printf("Cancel!!")
		return urlstr, nil
	}

	return title, err
}

// fmt.Println(getFolderName(`P`))           //P
// fmt.Println(getFolderName(`P:`))          //P
// fmt.Println(getFolderName(`P:\`))         //P
// fmt.Println(getFolderName(`P:\testing`))  //testing
// fmt.Println(getFolderName(`P:\testing\`)) //testing
func getFolderName(input string) string {
	var folderName string
	lastIndex := strings.LastIndex(input, `\`)
	length := len(input)
	if lastIndex+1 == length {
		lastIndex = strings.LastIndex(input[0:length-1], `\`)
		folderName = input[lastIndex+1 : length-1]
	} else {
		folderName = input[lastIndex+1 : length]
	}

	if folderName[len(folderName)-1:] == ":" {
		return folderName[0 : len(folderName)-1]
	} else {
		return folderName
	}
}
