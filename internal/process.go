package internal

import (
	"context"
	"github.com/yuchanns/hugo-pre-render/internal/chromedp"
)

func Process(ctx context.Context, dirs []string, ext string) error {
	files, err := ListFiles(dirs, ext)
	if err != nil {
		return err
	}
	pages, err := chromedp.Render(ctx, files)
	if err != nil {
		return err
	}
	return OverwriteFiles(ctx, pages)
}
