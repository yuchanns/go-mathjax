package chromedp

import (
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
	"github.com/yuchanns/yuchanns/pre-render/internal/utils"
	"log"
	"sync"
)

type Page struct {
	Path    string
	Content string
}

type PagesManager struct {
	locker *sync.Mutex
	pages  []*Page
}

func NewPagesManager() *PagesManager {
	return &PagesManager{
		locker: &sync.Mutex{},
		pages:  make([]*Page, 0, 20),
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

	wg := &sync.WaitGroup{}
	wg.Add(grs)

	if grs > 10 {
		grs = 10
	}

	pageManager := NewPagesManager()
	errGroup := utils.NewErrGroup()

	for i := range files {
		go func(ctx context.Context, file string) {
			defer wg.Done()

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

		}(ctx, files[i])
	}
	wg.Wait()

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
