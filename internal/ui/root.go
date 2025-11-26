package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/vin/ck123gogo/internal/app/config"
	"github.com/vin/ck123gogo/internal/app/state"
	"github.com/vin/ck123gogo/internal/models"
	"github.com/vin/ck123gogo/internal/search"
	"github.com/vin/ck123gogo/internal/service/resource"
	"github.com/vin/ck123gogo/internal/theme"
	"github.com/vin/ck123gogo/internal/ui/detail"
	"github.com/vin/ck123gogo/internal/ui/keymap"
	"github.com/vin/ck123gogo/internal/ui/list"
	"github.com/vin/ck123gogo/internal/ui/widgets"
)

// Root 管理 tview Application 與主要畫面。
type Root struct {
	app      *tview.Application
	pages    *tview.Pages
	layout   *tview.Flex
	state    *state.Store
	themeMgr *theme.Manager
	config   config.Config
	service  *resource.Service

	listView   *list.View
	detailView *detail.View
	statusBar  *widgets.StatusBar
	searchBox  *tview.InputField

	ctx         context.Context
	currentKind resource.Kind
	themeCycle  []string
	lastMessage string
}

// NewRoot 建立 Root，並套用預設主題與內容。
func NewRoot(cfg config.Config, mgr *theme.Manager, st *state.Store, svc *resource.Service) (*Root, error) {
	if mgr == nil {
		return nil, fmt.Errorf("theme manager is nil")
	}
	if svc == nil {
		return nil, fmt.Errorf("resource service is nil")
	}
	app := tview.NewApplication()
	if err := mgr.Apply(cfg.Theme); err != nil {
		return nil, err
	}

	r := &Root{
		app:         app,
		state:       st,
		themeMgr:    mgr,
		config:      cfg,
		service:     svc,
		currentKind: resource.KindEC2,
		themeCycle:  []string{"dark", "light", "high-contrast"},
	}
	r.initLayout()
	return r, nil
}

func (r *Root) initLayout() {
	r.listView = list.NewView()
	r.detailView = detail.NewView()
	r.statusBar = widgets.NewStatusBar()
	r.searchBox = tview.NewInputField().
		SetLabel("搜尋: ").
		SetFieldWidth(30)

	columns := tview.NewFlex().
		AddItem(r.listView.Primitive(), 0, 2, true).
		AddItem(r.detailView.Primitive(), 0, 3, false)

	r.layout = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(r.searchBox, 1, 0, false).
		AddItem(columns, 0, 1, true).
		AddItem(r.statusBar.Primitive(), 1, 0, false)

	r.pages = tview.NewPages()
	r.pages.AddPage("main", r.layout, true, true)

	r.listView.SetOnSelect(func(item models.ListItem) {
		r.showDetail(item)
	})

	r.searchBox.SetDoneFunc(func(key tcell.Key) {
		r.app.SetFocus(r.listView.Primitive())
		r.reload()
	})

	r.app.SetInputCapture(r.handleKeys)
}

// Run 開啟 UI loop，並可透過 context 取消。
func (r *Root) Run(ctx context.Context) error {
	if r.app == nil {
		return fmt.Errorf("tview application is nil")
	}
	r.ctx = ctx

	go r.reload()

	errCh := make(chan error, 1)
	go func() {
		errCh <- r.app.SetRoot(r.pages, true).Run()
	}()

	select {
	case <-ctx.Done():
		r.app.Stop()
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

// Stop 停止 UI。
func (r *Root) Stop() {
	if r.app != nil {
		r.app.Stop()
	}
}

func (r *Root) handleKeys(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyRune:
		switch event.Rune() {
		case '1':
			r.changeKind(resource.KindEC2)
			return nil
		case '2':
			r.changeKind(resource.KindRDS)
			return nil
		case '3':
			r.changeKind(resource.KindS3)
			return nil
		case '4':
			r.changeKind(resource.KindLambda)
			return nil
		case '/':
			r.app.SetFocus(r.searchBox)
			return nil
		case 'p':
			r.promptInput("切換 AWS Profile", r.state.Profile(), func(val string) {
				r.state.SetProfile(strings.TrimSpace(val))
				r.reload()
			})
			return nil
		case 'r':
			r.promptInput("切換 AWS Region", r.state.Region(), func(val string) {
				r.state.SetRegion(strings.TrimSpace(val))
				r.reload()
			})
			return nil
		case 't':
			r.toggleTheme()
			return nil
		case 'g':
			go r.reload()
			return nil
		case '?':
			r.showHelp()
			return nil
		case 'q':
			r.app.Stop()
			return nil
		}
	case tcell.KeyCtrlC:
		r.app.Stop()
		return nil
	}
	return event
}

func (r *Root) changeKind(kind resource.Kind) {
	if r.currentKind == kind {
		return
	}
	r.currentKind = kind
	go r.reload()
}

func (r *Root) reload() {
	r.setStatus("載入中...")

	query := r.searchBox.GetText()
	matcher := search.NewMatcher(query)
	ctx, cancel := context.WithTimeout(r.ctx, 20*time.Second)
	defer cancel()

	items, err := r.service.ListItems(ctx, r.currentKind, matcher)
	r.app.QueueUpdateDraw(func() {
		if err != nil {
			r.setStatus(fmt.Sprintf("[red]%v", err))
			return
		}
		r.listView.SetItems(items)
		count := r.listView.Count()
		r.setStatus(fmt.Sprintf("載入完成 (%d)", count))
		if count > 0 {
			if item, ok := r.listView.CurrentItem(); ok {
				go r.loadDetail(item)
			}
		} else {
			r.detailView.SetDetail(models.DetailView{})
		}
	})
}

func (r *Root) showDetail(item models.ListItem) {
	go r.loadDetail(item)
}

func (r *Root) loadDetail(item models.ListItem) {
	ctx, cancel := context.WithTimeout(r.ctx, 20*time.Second)
	defer cancel()
	detail, err := r.service.Detail(ctx, r.currentKind, item.ID)
	r.app.QueueUpdateDraw(func() {
		if err != nil {
			r.detailView.SetDetail(models.DetailView{
				Overview: map[string]string{
					"Error": err.Error(),
				},
			})
			return
		}
		r.detailView.SetDetail(detail)
	})
}

func (r *Root) setStatus(message string) {
	profile, region, theme, _ := r.state.Snapshot()
	r.lastMessage = message
	r.statusBar.SetStatus(profile, region, theme, r.currentKind, r.listView.Count(), message)
}

func (r *Root) toggleTheme() {
	current := r.state.Theme()
	idx := 0
	for i, name := range r.themeCycle {
		if name == current {
			idx = i
			break
		}
	}
	next := r.themeCycle[(idx+1)%len(r.themeCycle)]
	if err := r.themeMgr.Apply(next); err != nil {
		r.setStatus(fmt.Sprintf("[red]套用主題失敗: %v", err))
		return
	}
	r.state.SetTheme(next)
	r.app.ForceDraw()
	r.setStatus(fmt.Sprintf("已切換主題為 %s", next))
}

func (r *Root) showHelp() {
	modal := tview.NewModal().
		SetText(keymap.HelpText).
		AddButtons([]string{"關閉"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			r.pages.RemovePage("help")
		})
	r.pages.AddAndSwitchToPage("help", modal, true)
}

func (r *Root) promptInput(title, initial string, onSubmit func(string)) {
	input := tview.NewInputField().SetText(initial)
	form := tview.NewForm().
		AddFormItem(input).
		AddButton("確定", func() {
			r.pages.RemovePage("prompt")
			text := input.GetText()
			onSubmit(text)
		}).
		AddButton("取消", func() {
			r.pages.RemovePage("prompt")
		})
	form.SetBorder(true).SetTitle(title).SetTitleAlign(tview.AlignLeft)
	modal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(form, 10, 1, true).
			AddItem(nil, 0, 1, false), 60, 0, true).
		AddItem(nil, 0, 1, false)

	r.pages.AddAndSwitchToPage("prompt", modal, true)
	r.app.SetFocus(input)
}
