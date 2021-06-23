package internal

import (
	"context"
	"github.com/yuchanns/yuchanns/pre-render/internal/chromedp"
	"github.com/yuchanns/yuchanns/pre-render/internal/utils"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sync"
)

// ListFiles scan the given directories
// then list files filtered by extension
// if the extension is empty string, all files are listed
// files are listed with the absolute path
func ListFiles(dirs []string, ext string) ([]string, error) {
	var files []string
	for _, dir := range dirs {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			return nil, err
		}
		if err := filepath.Walk(absDir, buildFileWalkFunc(&files, ext)); err != nil {
			return nil, err
		}
	}
	return files, nil
}

// OverwriteFiles trim script tags of each content
// then overwrite it into each file
// concurrency at most 10 goroutines
func OverwriteFiles(ctx context.Context, pages []*chromedp.Page) error {
	grs := len(pages)

	wg := &sync.WaitGroup{}
	wg.Add(grs)

	if grs > 10 {
		grs = 10
	}

	errGroup := utils.NewErrGroup()

	reg := regexp.MustCompile("<script(([\\s\\S])*?)</script>")

	for i := range pages {
		go func(ctx context.Context, page *chromedp.Page) {
			defer wg.Done()
			page.Content = reg.ReplaceAllString(page.Content, "")
			fd, err := os.Create(page.Path)
			if err != nil {
				errGroup.Append(err)
				return
			}
			defer fd.Close()
			if _, err := fd.WriteString(page.Content); err != nil {
				errGroup.Append(err)
				return
			}
			log.Printf("overwrite %s...\tDone", page.Path)
		}(ctx, pages[i])
	}

	wg.Wait()

	if errGroup.HasErr() {
		return errGroup
	}

	return nil
}

func buildFileWalkFunc(files *[]string, ext string) filepath.WalkFunc {
	return func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if ext != "" && filepath.Ext(path) != ext {
			return nil
		}

		*files = append(*files, path)

		return nil
	}
}
