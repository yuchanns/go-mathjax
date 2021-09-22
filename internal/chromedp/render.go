package chromedp

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/chromedp/chromedp"
	"github.com/pkufergus/goroutinepool"
	"github.com/yuchanns/hugo-pre-render/internal/utils"
)

type Page struct {
	Path    string
	Content string
}

type PagesManager struct {
	locker *sync.Mutex
	pages  []*Page
}

func NewPagesManager(n int) *PagesManager {
	if n <= 0 {
		n = 1
	}
	return &PagesManager{
		locker: &sync.Mutex{},
		pages:  make([]*Page, 0, n),
	}
}

func (p *PagesManager) Append(page *Page) {
	p.locker.Lock()
	defer p.locker.Unlock()
	p.pages = append(p.pages, page)
}

func (p *PagesManager) GetPages() []*Page {
	return p.pages
}

// Render render every page of given files
// concurrency at most 10 goroutines
func Render(ctx context.Context, files []string) ([]*Page, error) {
	grs := len(files)
	pageManager := NewPagesManager(grs)
	errGroup := utils.NewErrGroup()
	if grs > 20 {
		grs = 20
	}

	p := goroutinepool.NewRoutinePool(grs)

	for i := range files {
		file := files[i]
		p.AddJob(func() {
			content, err := render(ctx, fmt.Sprintf("file:///%s", file))
			if err != nil {
				log.Printf("render %s...\tErr: %s\n", file, err)
				errGroup.Append(err)
				return
			}
			page := &Page{
				Path:    file,
				Content: content,
			}
			pageManager.Append(page)
			log.Printf("render %s...\tDone\n", file)

		})
	}

	p.WaitAll()

	if errGroup.HasErr() {
		return nil, errGroup
	}

	return pageManager.GetPages(), nil
}

func render(ctx context.Context, url string) (string, error) {
	chromeCtx, chromeCancel := chromedp.NewContext(ctx, chromedp.WithLogf(log.Printf))
	defer chromeCancel()

	var html string

	actions := []chromedp.Action{
		chromedp.Navigate(url),
		chromedp.WaitReady("#MJX-SVG-global-cache"),
		chromedp.OuterHTML("html", &html, chromedp.NodeVisible,
			chromedp.ByQuery),
	}

	if err := chromedp.Run(chromeCtx, actions...); err != nil {
		return "", err
	}

	return html, nil
}
