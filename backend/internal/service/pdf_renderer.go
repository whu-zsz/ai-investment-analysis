package service

import (
	"context"
	"fmt"
	"net/url"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

type PDFRenderer interface {
	RenderHTMLToPDF(ctx context.Context, html string) ([]byte, error)
}

type chromedpPDFRenderer struct{}

func NewChromedpPDFRenderer() PDFRenderer {
	return &chromedpPDFRenderer{}
}

func (r *chromedpPDFRenderer) RenderHTMLToPDF(ctx context.Context, html string) ([]byte, error) {
	allocatorCtx, cancelAllocator := chromedp.NewExecAllocator(ctx,
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Headless,
			chromedp.DisableGPU,
			chromedp.NoSandbox,
		)...,
	)
	defer cancelAllocator()

	renderCtx, cancelRender := chromedp.NewContext(allocatorCtx)
	defer cancelRender()

	var pdfBytes []byte
	dataURL := "data:text/html;charset=utf-8," + url.PathEscape(html)

	err := chromedp.Run(renderCtx,
		chromedp.Navigate(dataURL),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().
				WithPrintBackground(true).
				WithPreferCSSPageSize(true).
				Do(ctx)
			if err != nil {
				return fmt.Errorf("print html to pdf: %w", err)
			}
			pdfBytes = buf
			return nil
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("render pdf with chromedp: %w", err)
	}

	return pdfBytes, nil
}
