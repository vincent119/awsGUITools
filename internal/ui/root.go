// Package ui 提供主應用程式的 UI 管理。
package ui

import (
	"context"
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/vincent119/awsGUITools/internal/app/config"
	"github.com/vincent119/awsGUITools/internal/app/state"
	"github.com/vincent119/awsGUITools/internal/aws/profile"
	"github.com/vincent119/awsGUITools/internal/i18n"
	"github.com/vincent119/awsGUITools/internal/models"
	"github.com/vincent119/awsGUITools/internal/search"
	"github.com/vincent119/awsGUITools/internal/service/resource"
	"github.com/vincent119/awsGUITools/internal/theme"
	"github.com/vincent119/awsGUITools/internal/ui/detail"
	"github.com/vincent119/awsGUITools/internal/ui/keymap"
	"github.com/vincent119/awsGUITools/internal/ui/list"
	"github.com/vincent119/awsGUITools/internal/ui/modals"
	"github.com/vincent119/awsGUITools/internal/ui/widgets"
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
		SetLabel(i18n.T("search.label")).
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
	// 檢查當前焦點是否在主頁面元件
	focus := r.app.GetFocus()
	isMainPage := focus == r.searchBox ||
		focus == r.listView.Primitive() ||
		focus == r.detailView.Primitive() ||
		focus == r.statusBar.Primitive()

	// 如果不在主頁面（例如在 Modal 中），讓元件自己處理按鍵
	if !isMainPage {
		return event
	}

	// 如果焦點在搜尋欄，處理特殊鍵
	if focus == r.searchBox {
		switch event.Key() {
		case tcell.KeyEnter:
			// 執行搜尋
			r.app.SetFocus(r.listView.Primitive())
			go r.reload()
			return nil
		case tcell.KeyEscape:
			// 離開搜尋欄（不清除內容）
			r.app.SetFocus(r.listView.Primitive())
			return nil
		}
		// 其他按鍵讓搜尋欄自己處理
		return event
	}

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
		case '5':
			r.changeKind(resource.KindRoute53)
			return nil
		case '/':
			r.app.SetFocus(r.searchBox)
			return nil
		case 'p':
			r.showProfilePicker()
			return nil
		case 't':
			r.toggleTheme()
			return nil
		case 'l', 'L':
			r.toggleLanguage()
			return nil
		case 'a':
			r.showActionPanel()
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
	case tcell.KeyEnter:
		r.handleEnter()
		return nil
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		r.handleBackspace()
		return nil
	case tcell.KeyEscape:
		r.handleEscape()
		return nil
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
	r.setStatus(i18n.T("app.loading"))

	query := r.searchBox.GetText()
	matcher := search.NewMatcher(query)
	ctx, cancel := context.WithTimeout(r.ctx, 20*time.Second)
	defer cancel()

	items, err := r.service.ListItems(ctx, r.currentKind, matcher)
	r.app.QueueUpdateDraw(func() {
		if err != nil {
			r.setStatus(fmt.Sprintf("[red]%v[-]", err))
			return
		}
		r.listView.SetItems(items)
		count := r.listView.Count()
		r.setStatus(i18n.Tf("app.loaded", count))
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
		r.setStatus(fmt.Sprintf("[red]%s[-]", i18n.Tf("theme.switch.failed", err)))
		return
	}
	r.state.SetTheme(next)

	// 強制刷新所有 UI 元件的背景色
	r.refreshTheme()
	r.app.Sync()

	r.setStatus(i18n.Tf("theme.switched", next))
}

// refreshTheme 刷新所有 UI 元件以套用新主題。
func (r *Root) refreshTheme() {
	bg := tview.Styles.PrimitiveBackgroundColor
	contrastBg := tview.Styles.ContrastBackgroundColor
	text := tview.Styles.PrimaryTextColor
	secondary := tview.Styles.SecondaryTextColor
	border := tview.Styles.BorderColor

	// 更新搜尋列
	r.searchBox.SetBackgroundColor(bg)
	r.searchBox.SetFieldBackgroundColor(contrastBg)
	r.searchBox.SetLabelColor(secondary)
	r.searchBox.SetFieldTextColor(text)

	// 更新列表（Table 繼承 Box）
	listTable := r.listView.Primitive()
	listTable.SetBackgroundColor(bg)
	listTable.SetBorderColor(border)
	listTable.SetTitleColor(text)

	// 更新詳情頁
	detailView := r.detailView.Primitive()
	detailView.SetBackgroundColor(bg)
	detailView.SetBorderColor(border)
	detailView.SetTitleColor(text)
	detailView.SetTextColor(text)

	// 更新狀態列
	statusView := r.statusBar.Primitive()
	statusView.SetBackgroundColor(contrastBg)
	statusView.SetTextColor(text)

	// 更新 layout
	r.layout.SetBackgroundColor(bg)
	r.pages.SetBackgroundColor(bg)
}

func (r *Root) toggleLanguage() {
	next := i18n.NextLanguage()
	i18n.SetLanguage(next)
	r.state.SetLanguage(string(next))

	// 更新所有 UI 元件的文字
	r.refreshLabels()

	r.app.ForceDraw()
	r.setStatus(i18n.Tf("language.switched", i18n.LanguageDisplayName(next)))
}

// refreshLabels 刷新所有 UI 元件的文字標籤以套用新語言。
func (r *Root) refreshLabels() {
	// 更新搜尋列標籤
	r.searchBox.SetLabel(i18n.T("search.label"))

	// 更新列表標題與欄位名稱
	r.listView.RefreshLabels()

	// 更新詳情頁標題
	r.detailView.Primitive().SetTitle(i18n.T("ui.resource_detail"))
}

func (r *Root) showHelp() {
	modal := tview.NewModal().
		SetText(keymap.GetHelpText()).
		AddButtons([]string{i18n.T("action.close")}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			r.pages.RemovePage("help")
		})
	r.pages.AddAndSwitchToPage("help", modal, true)
}

// showActionPanel 顯示操作面板。
func (r *Root) showActionPanel() {
	item, ok := r.listView.CurrentItem()
	if !ok {
		r.setStatus("[yellow]No resource selected[-]")
		return
	}

	actions := modals.AvailableActions(item.Type)
	if len(actions) == 0 {
		r.setStatus(fmt.Sprintf("[yellow]No actions available for %s[-]", item.Type))
		return
	}

	panel := modals.NewActionPanel()
	panel.SetActions(actions, func(action string) {
		r.pages.RemovePage("action-panel")
		if action == "" {
			return
		}
		r.executeAction(item, action)
	})

	// 建立置中的 modal
	modal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(panel.Primitive(), 10, 0, true).
			AddItem(nil, 0, 1, false), 40, 0, true).
		AddItem(nil, 0, 1, false)

	r.pages.AddAndSwitchToPage("action-panel", modal, true)
}

// executeAction 執行操作（帶確認）。
func (r *Root) executeAction(item models.ListItem, action string) {
	// 顯示確認對話框
	confirm := modals.NewConfirmModal()
	confirm.Show(
		i18n.T("action.confirm"),
		fmt.Sprintf("%s %s: %s?", action, item.Type, item.Name),
		func(confirmed bool) {
			r.pages.RemovePage("confirm")
			if !confirmed {
				return
			}
			r.setStatus(fmt.Sprintf("Executing %s on %s...", action, item.Name))
			// TODO: 實際執行操作（呼叫 ops 層）
		},
	)
	r.pages.AddAndSwitchToPage("confirm", confirm.Primitive(), true)
}

// showProfilePicker 顯示 AWS Profile 選擇器。
// 選擇 Profile 後會自動切換到該 Profile 對應的 Region。
func (r *Root) showProfilePicker() {
	profiles := r.state.Profiles()
	if profiles == nil || !profiles.HasProfiles() {
		r.setStatus(fmt.Sprintf("[yellow]%s[-]", i18n.T("profile.not_found")))
		return
	}

	picker := modals.NewProfilePicker()
	picker.UpdateTitle(len(profiles.Profiles))
	picker.SetProfiles(profiles.Profiles, r.state.Profile())

	picker.SetOnSelect(func(info profile.Info) {
		r.pages.RemovePage("profile-picker")
		// SetProfile 會自動切換 Region
		r.state.SetProfile(info.Name)
		r.setStatus(i18n.Tf("profile.switched", info.Name, r.state.Region()))
		go r.reload()
	})

	picker.SetOnCancel(func() {
		r.pages.RemovePage("profile-picker")
	})

	r.pages.AddAndSwitchToPage("profile-picker", picker.Primitive(), true)
}

// handleEnter 處理 Enter 鍵 - 進入 S3 bucket/目錄、Route53 Zone，或顯示詳情。
func (r *Root) handleEnter() {
	item, ok := r.listView.CurrentItem()
	if !ok {
		return
	}

	switch r.currentKind {
	case resource.KindS3:
		// 進入 bucket 瀏覽物件
		r.service.SetCurrentBucket(item.ID)
		r.currentKind = resource.KindS3Objects
		go r.reload()
	case resource.KindS3Objects:
		// 如果是目錄，進入子目錄
		if item.Type == "Dir" {
			r.service.SetCurrentPrefix(item.ID)
			go r.reload()
		} else {
			// 檔案：顯示詳情
			r.showDetail(item)
		}
	case resource.KindRoute53:
		// 進入 hosted zone 查看 records
		r.service.SetCurrentZone(item.ID, item.Name)
		r.currentKind = resource.KindRoute53Records
		go r.reload()
	default:
		// EC2, RDS, Lambda, Route53Records：顯示詳情
		r.showDetail(item)
	}
}

// handleBackspace 處理 Backspace 鍵 - 返回上層。
func (r *Root) handleBackspace() {
	switch r.currentKind {
	case resource.KindS3Objects:
		if r.service.NavigateUp() {
			// 還在 bucket 內，刷新物件列表
			go r.reload()
		} else {
			// 回到 bucket 列表
			r.currentKind = resource.KindS3
			go r.reload()
		}
	case resource.KindRoute53Records:
		r.service.ClearCurrentZone()
		r.currentKind = resource.KindRoute53
		go r.reload()
	}
}

// handleEscape 處理 Escape 鍵 - 返回主資源列表。
func (r *Root) handleEscape() {
	switch r.currentKind {
	case resource.KindS3Objects:
		r.service.SetCurrentBucket("")
		r.currentKind = resource.KindS3
		go r.reload()
	case resource.KindRoute53Records:
		r.service.ClearCurrentZone()
		r.currentKind = resource.KindRoute53
		go r.reload()
	}
}
