// Command desktop is the Fyne-based Sanskrit dictionary application.
// Design inspired by Apple Dictionary: clean white, sidebar with word list, main content area.
package main

import (
	"flag"
	"fmt"
	"image/color"
	"log"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/licht1stein/sanskrit-upaya/pkg/download"
	"github.com/licht1stein/sanskrit-upaya/pkg/search"
	"github.com/licht1stein/sanskrit-upaya/pkg/state"
	"github.com/licht1stein/sanskrit-upaya/pkg/transliterate"
	"github.com/licht1stein/sanskrit-upaya/pkg/version"
)

var testDownload = flag.Bool("test-download", false, "Simulate download flow for testing")
var showVersion = flag.Bool("version", false, "Print version and exit")

// Version is set at build time via ldflags
var Version = "dev"

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Println(Version)
		return
	}

	log.Printf("Starting Sanskrit Dictionary (version: %s)...", Version)

	// Open state/settings store
	settings, err := state.Open()
	if err != nil {
		log.Printf("Warning: Could not open settings: %v", err)
	} else {
		defer settings.Close()
	}

	// Create app first (needed for download dialog)
	a := app.New()

	// Check if database exists, download if not
	dbPath, err := download.GetDatabasePath()
	if err != nil {
		log.Printf("Warning: Could not get database path: %v", err)
		dbPath = "sanskrit.db" // Fallback to local
	}

	// Database pointer - may be set after download
	var db *search.DB

	// Check database status
	dbStatus := download.CheckDatabase()
	if *testDownload {
		dbStatus = download.DatabaseMissing // Force download flow for testing
	}

	// Try to open database if valid
	if dbStatus == download.DatabaseValid {
		db, err = search.Open(dbPath)
		if err != nil {
			log.Printf("Warning: Could not open database: %v", err)
		} else {
			log.Println("Database opened successfully")
			defer db.Close()
		}
	}

	// Load zoom from settings (default 100%)
	zoomPercent := 100
	if settings != nil {
		if zoomStr := settings.Get("zoom"); zoomStr != "" {
			if z, err := strconv.Atoi(zoomStr); err == nil && z >= 50 && z <= 200 {
				zoomPercent = z
			}
		}
	}

	// Apply scaled theme
	applyZoom := func(percent int) {
		scale := float32(percent) / 100.0
		a.Settings().SetTheme(newScaledTheme(theme.LightTheme(), scale))
	}
	applyZoom(zoomPercent)

	w := a.NewWindow("Sanskrit Upāya")
	w.Resize(fyne.NewSize(1100, 700))

	// If database doesn't exist or needs update, download it first
	if db == nil {
		// Create download UI with appropriate message
		progressBar := widget.NewProgressBar()
		statusLabel := widget.NewLabel("Preparing to download dictionary database...")

		var titleLabel, subtitleLabel *widget.Label
		if dbStatus == download.DatabaseNeedsUpdate {
			titleLabel = widget.NewLabel("Dictionary database needs to be re-downloaded (~670 MB).")
			subtitleLabel = widget.NewLabel("This may be due to an app update or corrupted data.")
		} else {
			titleLabel = widget.NewLabel("Sanskrit Upāya needs to download the dictionary database (~670 MB).")
			subtitleLabel = widget.NewLabel("This only happens once.")
		}

		downloadContent := container.NewVBox(
			titleLabel,
			subtitleLabel,
			widget.NewLabel(""),
			statusLabel,
			progressBar,
		)

		w.SetContent(container.NewCenter(downloadContent))

		// Channel to receive download result
		type downloadResult struct {
			db  *search.DB
			err error
		}
		done := make(chan downloadResult, 1)

		// Start download in background
		go func() {
			var downloadErr error

			if *testDownload {
				// Simulate download for testing
				log.Println("Simulating download...")
				total := int64(670 * 1024 * 1024) // 670 MB
				for i := 0; i <= 10; i++ {
					downloaded := total * int64(i) / 10
					percent := float64(i) / 10.0
					mb := downloaded / (1024 * 1024)
					totalMb := total / (1024 * 1024)
					fyne.Do(func() {
						progressBar.SetValue(percent)
						statusLabel.SetText(fmt.Sprintf("Downloading... %d / %d MB", mb, totalMb))
					})
					time.Sleep(200 * time.Millisecond)
				}
			} else {
				// Real download
				downloadErr = download.Download(func(downloaded, total int64) {
					if total > 0 {
						percent := float64(downloaded) / float64(total)
						fyne.Do(func() {
							progressBar.SetValue(percent)
							statusLabel.SetText(fmt.Sprintf("Downloading... %d / %d MB", downloaded/(1024*1024), total/(1024*1024)))
						})
					}
				})
			}

			if downloadErr != nil {
				done <- downloadResult{nil, downloadErr}
				return
			}

			fyne.Do(func() {
				statusLabel.SetText("Download complete! Loading database...")
				progressBar.SetValue(1.0)
			})

			// Open the database
			newDB, openErr := search.Open(dbPath)
			done <- downloadResult{newDB, openErr}
		}()

		// Wait for download result in another goroutine, then build main UI
		go func() {
			result := <-done
			fyne.Do(func() {
				if result.err != nil {
					dialog.ShowError(fmt.Errorf("Failed: %v", result.err), w)
					statusLabel.SetText("Failed. Please restart the app to try again.")
					return
				}

				log.Println("Database opened successfully after download")
				db = result.db
				// Build and show the main UI
				buildMainUI(w, a, db, settings, zoomPercent, applyZoom)
			})
		}()

		w.ShowAndRun()
		if db != nil {
			db.Close()
		}
		return
	}

	// Build and run main UI
	buildMainUI(w, a, db, settings, zoomPercent, applyZoom)
	w.ShowAndRun()
}

// buildMainUI constructs the main application interface
func buildMainUI(w fyne.Window, a fyne.App, db *search.DB, settings *state.Store, zoomPercent int, applyZoom func(int)) {
	// Search state
	currentMode := search.ModeExact // Default to exact
	var groupedResults []GroupedResult
	var cachedResults []search.Result // Cache raw results for re-grouping

	// Lazy loading - only show first N results, load more on demand
	const initialDisplayLimit = 100
	const loadMoreIncrement = 100
	displayLimit := initialDisplayLimit

	// Limit articles rendered per dictionary to avoid UI freeze
	const maxArticlesPerDict = 10

	// Content cache for prefetched articles
	contentCache := make(map[int64]string)
	var contentCacheMu sync.RWMutex

	// Check if content is cached
	isContentCached := func(articleID int64) bool {
		contentCacheMu.RLock()
		_, ok := contentCache[articleID]
		contentCacheMu.RUnlock()
		return ok
	}

	// Helper to get content (from cache or fetch)
	getContent := func(articleID int64) (string, error) {
		contentCacheMu.RLock()
		if content, ok := contentCache[articleID]; ok {
			contentCacheMu.RUnlock()
			return content, nil
		}
		contentCacheMu.RUnlock()

		// Not in cache, fetch from DB
		content, err := db.GetArticleContent(articleID)
		if err != nil {
			return "", err
		}

		// Store in cache
		contentCacheMu.Lock()
		contentCache[articleID] = content
		contentCacheMu.Unlock()

		return content, nil
	}

	// Prefetch content for visible results in background (batch fetch)
	prefetchContent := func(results []GroupedResult, limit int) {
		go func() {
			// Collect article IDs to fetch
			var ids []int64
			for _, gr := range results {
				if len(ids) >= limit {
					break
				}
				for _, entry := range gr.Entries {
					if len(ids) >= limit {
						break
					}
					for _, article := range entry.Articles {
						if len(ids) >= limit {
							break
						}
						// Skip if already cached
						contentCacheMu.RLock()
						_, exists := contentCache[article.ArticleID]
						contentCacheMu.RUnlock()
						if !exists {
							ids = append(ids, article.ArticleID)
						}
					}
				}
			}

			if len(ids) == 0 {
				return
			}

			// Batch fetch all at once
			contents, err := db.GetArticleContents(ids)
			if err != nil {
				return
			}

			// Store in cache
			contentCacheMu.Lock()
			for id, content := range contents {
				contentCache[id] = content
			}
			contentCacheMu.Unlock()
		}()
	}

	// Load group setting from state (default false)
	groupResultsSetting := false
	if settings != nil {
		groupResultsSetting = settings.GetBool("group_results", false)
	}

	// Dictionary selection state
	var allDicts []search.Dict
	selectedDicts := make(map[string]bool)
	if db != nil {
		var err error
		allDicts, err = db.GetDicts()
		if err != nil {
			log.Printf("Warning: Could not load dictionaries: %v", err)
		}
	}

	// Load selected dictionaries from settings (or default to all)
	if settings != nil {
		if saved := settings.Get("selected_dicts"); saved != "" {
			// Parse comma-separated list
			for _, code := range strings.Split(saved, ",") {
				if code != "" {
					selectedDicts[code] = true
				}
			}
		}
	}
	// If no saved selection (or empty), select all by default
	if len(selectedDicts) == 0 {
		for _, d := range allDicts {
			selectedDicts[d.Code] = true
		}
	}

	// Helper to get selected dict codes as slice
	getSelectedDictCodes := func() []string {
		var codes []string
		for code, selected := range selectedDicts {
			if selected {
				codes = append(codes, code)
			}
		}
		return codes
	}

	// Helper to save selected dicts
	saveSelectedDicts := func() {
		if settings != nil {
			codes := getSelectedDictCodes()
			settings.Set("selected_dicts", strings.Join(codes, ","))
		}
	}

	// Create debouncer for search-as-you-type (300ms delay)
	searchDebouncer := newDebouncer(300 * time.Millisecond)

	// Helper to get visible count (respects displayLimit)
	visibleCount := func() int {
		if len(groupedResults) <= displayLimit {
			return len(groupedResults)
		}
		return displayLimit
	}

	// Word list (sidebar) - shows unique words with dict pills or count
	wordList := widget.NewList(
		func() int { return visibleCount() },
		func() fyne.CanvasObject {
			wordLabel := widget.NewLabel("Word placeholder")
			wordLabel.TextStyle = fyne.TextStyle{Bold: true}
			// Right side: either pills or count - use HBox that can hold multiple pills
			pillsContainer := container.NewHBox()
			return container.NewBorder(nil, nil, nil, pillsContainer, wordLabel)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(groupedResults) {
				return
			}
			r := groupedResults[id]
			box := obj.(*fyne.Container)
			wordLabel := box.Objects[0].(*widget.Label)
			pillsContainer := box.Objects[1].(*fyne.Container)

			// Update word label
			wordText := r.Word
			if r.Word != "" && !transliterate.IsDevanagari(r.Word) {
				deva := transliterate.IASTToDevanagari(r.Word)
				if deva != "" {
					wordText = r.Word + " " + deva
				}
			}
			wordLabel.SetText(wordText)

			// Update pills/count
			pillsContainer.RemoveAll()
			if len(r.Entries) <= 2 {
				// Show individual pills
				for _, entry := range r.Entries {
					pillsContainer.Add(newPillLabel(entry.DictCode))
				}
			} else {
				// Show count
				countLabel := widget.NewLabel(fmt.Sprintf("(%d)", len(r.Entries)))
				countLabel.Importance = widget.LowImportance
				pillsContainer.Add(countLabel)
			}
			pillsContainer.Refresh()
		},
	)

	// Content area - will show cards for selected entry
	contentContainer := container.NewVBox()
	contentScroll := container.NewVScroll(contentContainer)

	// Content header (label + star button row)
	contentHeaderLabel := widget.NewLabel("")
	contentHeaderLabel.TextStyle = fyne.TextStyle{Bold: true}
	contentHeaderRow := container.NewHBox(contentHeaderLabel) // Will add star button dynamically

	// Store current article content for copy functionality
	var currentArticleContent string

	// Holder for the main content (either scroll or tabs)
	contentHolder := container.NewStack(contentScroll)

	// Content area with header
	contentArea := container.NewBorder(
		container.NewVBox(contentHeaderRow, widget.NewSeparator()),
		nil, nil, nil,
		contentHolder,
	)

	// Empty state - big friendly centered text
	emptyText := widget.NewLabel("Type to search...")
	emptyText.Alignment = fyne.TextAlignCenter
	emptyText.Importance = widget.LowImportance
	emptyText.TextStyle = fyne.TextStyle{Bold: true}
	emptyState := container.NewCenter(emptyText)

	// Status label - always visible at bottom of window
	statusText := widget.NewLabel("Ready")
	lastStatus := "Ready" // Track last status for restoring after "Loading..."

	// Helper to update status text
	setStatus := func(text string) {
		statusText.SetText(text)
		// Don't save transient statuses
		if text != "Loading..." && text != "Grouping..." && text != "Ungrouping..." && text != "Searching..." {
			lastStatus = text
		}
	}

	// Restore last status (after loading)
	restoreStatus := func() {
		statusText.SetText(lastStatus)
	}

	// Sidebar header
	sidebarHeader := widget.NewLabel("Results")

	// Sidebar with word list (no status bar, it's global now)
	sidebar := container.NewBorder(
		sidebarHeader,
		nil,
		nil, nil,
		wordList,
	)

	// Split: sidebar | content
	resultsView := container.NewHSplit(sidebar, contentArea)
	resultsView.SetOffset(0.28) // 28% for sidebar

	// Main content stack - switches between empty state and results
	mainContent := container.NewStack(emptyState, resultsView)
	resultsView.Hide() // Start with empty state

	// Track current selection for keyboard navigation
	var currentSelection int = -1

	// Show results view (hide empty state)
	showResults := func() {
		emptyState.Hide()
		resultsView.Show()
	}

	// Show empty state (hide results)
	showEmpty := func(text string) {
		emptyText.SetText(text)
		resultsView.Hide()
		emptyState.Show()
	}

	// Navigate to grouped result by index
	navigateTo := func(idx int) {
		if idx >= 0 && idx < len(groupedResults) {
			currentSelection = idx
			wordList.Select(idx)
			wordList.ScrollTo(idx)

			gr := &groupedResults[idx]

			// Header: word in IAST and Devanagari
			headerText := gr.Word
			if gr.Word != "" && !transliterate.IsDevanagari(gr.Word) {
				deva := transliterate.IASTToDevanagari(gr.Word)
				if deva != "" {
					headerText = gr.Word + "   " + deva
				}
			}
			contentHeaderLabel.SetText(headerText)

			// Clear and rebuild content
			contentContainer.RemoveAll()
			contentHolder.RemoveAll()

			// Get first article ID for starring (star the word, represented by first article)
			var firstArticleID int64
			if len(gr.Entries) > 0 && len(gr.Entries[0].Articles) > 0 {
				firstArticleID = gr.Entries[0].Articles[0].ArticleID
			}

			// Star button in header row (next to word)
			isStarred := settings != nil && firstArticleID > 0 && settings.IsStarred(firstArticleID)
			starIcon := theme.NewThemedResource(resourceStarSvg)
			starFilledIcon := theme.NewThemedResource(resourceStarFilledSvg)
			currentWord := gr.Word
			currentDictCode := ""
			if len(gr.Entries) > 0 {
				currentDictCode = gr.Entries[0].DictCode
			}

			var headerStarBtn *widget.Button
			headerStarBtn = widget.NewButtonWithIcon("", starIcon, func() {
				if settings == nil || firstArticleID == 0 {
					return
				}
				if settings.IsStarred(firstArticleID) {
					settings.UnstarArticle(firstArticleID)
					headerStarBtn.SetIcon(starIcon)
				} else {
					settings.StarArticle(firstArticleID, currentWord, currentDictCode)
					headerStarBtn.SetIcon(starFilledIcon)
				}
			})
			if isStarred {
				headerStarBtn.SetIcon(starFilledIcon)
			}
			headerStarBtn.Importance = widget.LowImportance

			// Copy button for article content
			copyBtn := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
				if currentArticleContent != "" {
					w.Clipboard().SetContent(currentArticleContent)
				}
			})
			copyBtn.Importance = widget.LowImportance

			// Update header row with label, copy button, and star button
			contentHeaderRow.RemoveAll()
			contentHeaderRow.Add(contentHeaderLabel)
			contentHeaderRow.Add(copyBtn)
			contentHeaderRow.Add(headerStarBtn)
			contentHeaderRow.Refresh()

			if len(gr.Entries) == 1 {
				// Single dictionary - show first article only
				entry := gr.Entries[0]
				dictHeader := container.NewHBox(
					newPillLabel(entry.DictCode),
					widget.NewLabel(entry.DictName),
				)
				contentContainer.Add(dictHeader)

				// Fetch and display first article content (from cache or DB)
				if len(entry.Articles) > 0 {
					articleID := entry.Articles[0].ArticleID
					// Show loading if not cached
					if !isContentCached(articleID) {
						setStatus("Loading...")
					}
					articleContent, err := getContent(articleID)
					if err == nil {
						currentArticleContent = cleanHTML(articleContent)
						content := createSelectableArticleContent(articleContent)
						contentContainer.Add(content)
					}
					// Restore status
					restoreStatus()
					if len(entry.Articles) > 1 {
						moreLabel := widget.NewLabel(fmt.Sprintf("... and %d more articles", len(entry.Articles)-1))
						moreLabel.Importance = widget.LowImportance
						contentContainer.Add(moreLabel)
					}
				}
				contentContainer.Refresh()
				contentHolder.Add(contentScroll)
				contentScroll.ScrollToTop()
			} else {
				// Multiple dictionaries - create tabs with "All" tab first
				tabs := container.NewAppTabs()

				// Track loaded article texts for "All" tab copy functionality
				var allArticleTexts []string
				allTabLoaded := false

				// Add "All" tab with placeholder (lazy-loaded)
				// Wrap in Stack to prevent content from expanding window
				allContent := container.NewVBox(widget.NewLabel("Loading..."))
				allScroll := container.NewVScroll(allContent)
				tabs.Append(container.NewTabItem("All", container.NewStack(allScroll)))

				// Add individual dictionary tabs with placeholders
				// Keep references to content containers for lazy loading
				dictContents := make(map[string]*fyne.Container)
				dictLoaded := make(map[string]bool)

				for _, e := range gr.Entries {
					contentBox := container.NewVBox(widget.NewLabel("Loading..."))
					dictContents[e.DictCode] = contentBox
					contentScroll := container.NewVScroll(contentBox)
					tabs.Append(container.NewTabItem(e.DictCode, container.NewStack(contentScroll)))
				}

				// Helper to load "All" tab content
				loadAllTab := func() {
					if allTabLoaded {
						return
					}
					allContent.RemoveAll()
					allArticleTexts = nil

					for _, e := range gr.Entries {
						dictHeader := container.NewHBox(
							newPillLabel(e.DictCode),
							widget.NewLabel(e.DictName),
						)
						allContent.Add(dictHeader)

						if len(e.Articles) > 0 {
							articleID := e.Articles[0].ArticleID
							if !isContentCached(articleID) {
								setStatus("Loading...")
							}
							articleContent, err := getContent(articleID)
							if err == nil {
								allArticleTexts = append(allArticleTexts, cleanHTML(articleContent))
								content := createSelectableArticleContent(articleContent)
								allContent.Add(content)
							}
							restoreStatus()
							if len(e.Articles) > 1 {
								moreLabel := widget.NewLabel(fmt.Sprintf("... and %d more articles", len(e.Articles)-1))
								moreLabel.Importance = widget.LowImportance
								allContent.Add(moreLabel)
							}
						}
						allContent.Add(widget.NewSeparator())
					}
					allContent.Refresh()
					allTabLoaded = true
					currentArticleContent = strings.Join(allArticleTexts, "\n\n---\n\n")
				}

				// Helper to load individual dictionary tab
				loadDictTab := func(e DictEntry) {
					if dictLoaded[e.DictCode] {
						// Already loaded, just update copy content
						if len(e.Articles) > 0 {
							articleID := e.Articles[0].ArticleID
							articleContent, err := getContent(articleID)
							if err == nil {
								currentArticleContent = cleanHTML(articleContent)
							}
						}
						return
					}

					vbox := dictContents[e.DictCode]
					vbox.RemoveAll()
					dictHeader := container.NewHBox(
						newPillLabel(e.DictCode),
						widget.NewLabel(e.DictName),
					)
					vbox.Add(dictHeader)

					if len(e.Articles) > 0 {
						articleID := e.Articles[0].ArticleID
						if !isContentCached(articleID) {
							setStatus("Loading...")
						}
						articleContent, err := getContent(articleID)
						if err == nil {
							currentArticleContent = cleanHTML(articleContent)
							content := createSelectableArticleContent(articleContent)
							vbox.Add(content)
						}
						restoreStatus()
						if len(e.Articles) > 1 {
							moreLabel := widget.NewLabel(fmt.Sprintf("... and %d more articles", len(e.Articles)-1))
							moreLabel.Importance = widget.LowImportance
							vbox.Add(moreLabel)
						}
					}
					vbox.Refresh()
					dictLoaded[e.DictCode] = true
				}

				// Lazy-load tab content on selection
				tabs.OnSelected = func(tab *container.TabItem) {
					if tab.Text == "All" {
						loadAllTab()
						currentArticleContent = strings.Join(allArticleTexts, "\n\n---\n\n")
						return
					}
					// Find which entry this tab corresponds to
					for _, e := range gr.Entries {
						if e.DictCode == tab.Text {
							loadDictTab(e)
							break
						}
					}
				}

				// Select first individual dictionary tab by default (faster initial load)
				tabs.SetTabLocation(container.TabLocationTop)
				tabs.SelectIndex(1) // Select first dictionary tab, not "All"
				contentHolder.Add(tabs)
			}

			contentHolder.Refresh()

			showResults()
		}
	}

	// Clear content display
	clearContent := func() {
		currentSelection = -1
		contentHeaderLabel.SetText("")
		contentHeaderRow.RemoveAll()
		contentHeaderRow.Add(contentHeaderLabel)
		contentHeaderRow.Refresh()
		contentContainer.RemoveAll()
	}

	// Handle word selection
	wordList.OnSelected = func(id widget.ListItemID) {
		// Auto-load more when near the end (within 10 items)
		if int(id) >= displayLimit-10 && displayLimit < len(groupedResults) {
			displayLimit += loadMoreIncrement
			wordList.Refresh()
		}

		if int(id) != currentSelection {
			navigateTo(int(id))
		}
	}

	// Group search results (respects groupResultsSetting)
	makeGroupedResults := func(results []search.Result) []GroupedResult {
		if !groupResultsSetting {
			// Ungrouped: each word+dict combo is separate
			grouped := make([]GroupedResult, 0, len(results))
			// Group by word+dict
			type key struct{ word, dict string }
			seen := make(map[key]*GroupedResult)
			var order []key

			for _, r := range results {
				k := key{r.Word, r.DictCode}
				if gr, ok := seen[k]; ok {
					gr.Entries[0].Articles = append(gr.Entries[0].Articles, r)
				} else {
					gr := &GroupedResult{
						Word: r.Word,
						Entries: []DictEntry{{
							DictCode: r.DictCode,
							DictName: r.DictName,
							Articles: []search.Result{r},
						}},
					}
					seen[k] = gr
					order = append(order, k)
				}
			}
			for _, k := range order {
				grouped = append(grouped, *seen[k])
			}
			return grouped
		}

		// Grouped: all dictionaries for same word together
		wordMap := make(map[string]*GroupedResult)
		var wordOrder []string

		for _, r := range results {
			if gr, ok := wordMap[r.Word]; ok {
				// Find or create dict entry
				found := false
				for i := range gr.Entries {
					if gr.Entries[i].DictCode == r.DictCode {
						gr.Entries[i].Articles = append(gr.Entries[i].Articles, r)
						found = true
						break
					}
				}
				if !found {
					gr.Entries = append(gr.Entries, DictEntry{
						DictCode: r.DictCode,
						DictName: r.DictName,
						Articles: []search.Result{r},
					})
				}
			} else {
				wordMap[r.Word] = &GroupedResult{
					Word: r.Word,
					Entries: []DictEntry{{
						DictCode: r.DictCode,
						DictName: r.DictName,
						Articles: []search.Result{r},
					}},
				}
				wordOrder = append(wordOrder, r.Word)
			}
		}

		// Convert to slice preserving order
		grouped := make([]GroupedResult, 0, len(wordOrder))
		for _, word := range wordOrder {
			grouped = append(grouped, *wordMap[word])
		}
		return grouped
	}

	// Re-group cached results (used when toggling group setting)
	// Runs in background goroutine to keep UI responsive
	regroupResults := func() {
		if len(cachedResults) == 0 {
			return
		}

		// Capture current results for goroutine
		results := cachedResults

		// Show status immediately
		if groupResultsSetting {
			setStatus("Grouping...")
		} else {
			setStatus("Ungrouping...")
		}

		go func() {
			grouped := makeGroupedResults(results)

			// Update data and refresh list
			fyne.Do(func() {
				groupedResults = grouped
				displayLimit = initialDisplayLimit // Reset lazy loading
				wordList.Refresh()
			})

			// Small delay to let list refresh complete, then navigate
			time.Sleep(10 * time.Millisecond)

			fyne.Do(func() {
				if len(groupedResults) > 0 {
					navigateTo(0)
				} else {
					clearContent()
				}
				// Restore status with result count
				restoreStatus()
			})
		}()
	}

	// Search function - runs database query in background
	doSearch := func(query string) {
		if db == nil {
			setStatus("Database not loaded")
			return
		}

		query = strings.TrimSpace(query)

		if query == "" {
			groupedResults = nil
			displayLimit = initialDisplayLimit // Reset lazy loading
			setStatus("Ready")
			wordList.Refresh()
			clearContent()
			showEmpty("Type to search...")
			return
		}

		if len(query) < 2 {
			setStatus("Type at least 2 characters...")
			return
		}

		// Clear content cache for new search
		contentCacheMu.Lock()
		contentCache = make(map[int64]string)
		contentCacheMu.Unlock()

		// Show searching indicator
		setStatus("Searching...")
		if len(groupedResults) == 0 {
			showEmpty("Searching...")
		}

		// Capture current mode and start timing
		mode := currentMode
		startTime := time.Now()

		// Run search in background
		go func() {
			// Get search terms (including Devanagari transliteration)
			searchTerms := transliterate.ToSearchTerms(query)

			// Get selected dictionaries for filtering
			dictCodes := getSelectedDictCodes()

			// Search with primary term
			searchResults, err := db.Search(searchTerms[0], mode, dictCodes)
			if err != nil {
				fyne.Do(func() {
					setStatus("Error: " + err.Error())
				})
				return
			}

			// Also search with Devanagari if we have it
			if len(searchTerms) > 1 {
				for _, term := range searchTerms[1:] {
					moreResults, err := db.Search(term, mode, dictCodes)
					if err == nil {
						searchResults = append(searchResults, moreResults...)
					}
				}
			}

			// Deduplicate by article ID (preserving order from database - already sorted by relevance)
			seen := make(map[int64]bool)
			var dedupedResults []search.Result
			for _, r := range searchResults {
				if !seen[r.ArticleID] {
					seen[r.ArticleID] = true
					dedupedResults = append(dedupedResults, r)
				}
			}

			// Cache results and group
			grouped := makeGroupedResults(dedupedResults)

			// Count unique dictionaries
			dictSet := make(map[string]bool)
			for _, gr := range grouped {
				for _, entry := range gr.Entries {
					dictSet[entry.DictCode] = true
				}
			}
			dictCount := len(dictSet)

			// Calculate search duration in seconds
			duration := time.Since(startTime).Seconds()

			// Update data first
			fyne.Do(func() {
				cachedResults = dedupedResults // Cache for re-grouping
				groupedResults = grouped
				displayLimit = initialDisplayLimit // Reset lazy loading for new search

				// Update status with raw count (not grouped), time in seconds, and dictionary count
				if len(dedupedResults) == 0 {
					setStatus(fmt.Sprintf("No results found (%.2fs)", duration))
					showEmpty("No results found")
				} else {
					setStatus(fmt.Sprintf("%d entries in %.2fs across %d dicts", len(dedupedResults), duration, dictCount))
					// Save to history (only if results found)
					if settings != nil {
						settings.AddHistory(query)
					}
				}

				// Prefetch content for first 50 visible results in background
				prefetchContent(grouped, 50)
			})

			// Yield to let status update render
			time.Sleep(5 * time.Millisecond)

			// Refresh list
			fyne.Do(func() {
				wordList.Refresh()
			})

			// Yield again
			time.Sleep(5 * time.Millisecond)

			// Navigate to first result
			fyne.Do(func() {
				if len(groupedResults) > 0 {
					navigateTo(0)
				} else {
					clearContent()
				}
			})
		}()
	}

	// Search entry
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search in Devanagari or IAST...")
	searchEntry.OnChanged = func(text string) {
		searchDebouncer.Do(func() {
			fyne.Do(func() {
				doSearch(text)
			})
		})
	}
	searchEntry.OnSubmitted = func(text string) {
		doSearch(text) // Immediate search on Enter
	}

	// History button
	historyBtn := widget.NewButtonWithIcon("", theme.HistoryIcon(), func() {
		if settings == nil {
			return
		}
		history := settings.GetRecentHistory(20)
		if len(history) == 0 {
			dialog.ShowInformation("History", "No search history yet", w)
			return
		}

		// Create list of history items
		historyList := widget.NewList(
			func() int { return len(history) },
			func() fyne.CanvasObject {
				return widget.NewLabel("history item")
			},
			func(id widget.ListItemID, obj fyne.CanvasObject) {
				if id < len(history) {
					obj.(*widget.Label).SetText(history[id])
				}
			},
		)

		var historyDialog dialog.Dialog
		historyList.OnSelected = func(id widget.ListItemID) {
			if id < len(history) {
				searchEntry.SetText(history[id])
				historyDialog.Hide()
				doSearch(history[id])
			}
		}

		listScroll := container.NewVScroll(historyList)
		listScroll.SetMinSize(fyne.NewSize(300, 300))

		historyDialog = dialog.NewCustom("Search History", "Close", listScroll, w)
		historyDialog.Resize(fyne.NewSize(350, 400))
		historyDialog.Show()
	})

	// Starred button - opens starred articles view
	starredBtn := widget.NewButtonWithIcon("", theme.NewThemedResource(resourceStarSvg), func() {
		if settings == nil || db == nil {
			return
		}
		starred := settings.GetStarredArticles()
		if len(starred) == 0 {
			dialog.ShowInformation("Starred", "No starred articles yet", w)
			return
		}

		// Filter state
		filterText := ""
		filteredStarred := starred

		// Word list for starred items
		starredWordList := widget.NewList(
			func() int { return len(filteredStarred) },
			func() fyne.CanvasObject {
				wordLabel := widget.NewLabel("word")
				wordLabel.TextStyle = fyne.TextStyle{Bold: true}
				dictPill := newPillLabel("dict")
				return container.NewBorder(nil, nil, nil, dictPill, wordLabel)
			},
			func(id widget.ListItemID, obj fyne.CanvasObject) {
				if id >= len(filteredStarred) {
					return
				}
				sa := filteredStarred[id]
				box := obj.(*fyne.Container)
				wordLabel := box.Objects[0].(*widget.Label)
				pillWidget := box.Objects[1].(*pillLabel)

				wordText := sa.Word
				if !transliterate.IsDevanagari(sa.Word) {
					deva := transliterate.IASTToDevanagari(sa.Word)
					if deva != "" {
						wordText = sa.Word + " " + deva
					}
				}
				wordLabel.SetText(wordText)
				pillWidget.text = sa.DictCode
				pillWidget.Refresh()
			},
		)

		// Content area for articles
		starredContent := container.NewVBox()
		starredContentScroll := container.NewVScroll(starredContent)

		// Content header
		starredContentHeader := widget.NewLabel("")
		starredContentHeader.TextStyle = fyne.TextStyle{Bold: true}

		// Navigate to starred article (declared as variable for recursive reference)
		var showStarredArticle func(idx int)
		showStarredArticle = func(idx int) {
			if idx < 0 || idx >= len(filteredStarred) {
				return
			}
			sa := filteredStarred[idx]

			// Load article from database
			results, err := db.GetArticle(sa.ArticleID)
			if err != nil || len(results) == 0 {
				starredContent.RemoveAll()
				starredContent.Add(widget.NewLabel("Article not found"))
				starredContent.Refresh()
				return
			}

			// Update header
			headerText := sa.Word
			if !transliterate.IsDevanagari(sa.Word) {
				deva := transliterate.IASTToDevanagari(sa.Word)
				if deva != "" {
					headerText = sa.Word + "   " + deva
				}
			}
			starredContentHeader.SetText(headerText)

			// Show article content with unstar button
			starredContent.RemoveAll()

			// Unstar button
			unstarBtn := widget.NewButtonWithIcon("Unstar", theme.DeleteIcon(), func() {
				settings.UnstarArticle(sa.ArticleID)
				// Refresh starred list
				starred = settings.GetStarredArticles()
				// Re-apply filter
				if filterText == "" {
					filteredStarred = starred
				} else {
					filteredStarred = nil
					for _, s := range starred {
						if strings.Contains(strings.ToLower(s.Word), strings.ToLower(filterText)) {
							filteredStarred = append(filteredStarred, s)
						}
					}
				}
				starredWordList.Refresh()
				if len(filteredStarred) == 0 {
					starredContent.RemoveAll()
					starredContent.Add(widget.NewLabel("No starred articles"))
					starredContentHeader.SetText("")
				} else if idx >= len(filteredStarred) {
					showStarredArticle(len(filteredStarred) - 1)
				}
				starredContent.Refresh()
			})

			dictHeader := container.NewHBox(
				newPillLabel(sa.DictCode),
				widget.NewLabel(results[0].DictName),
				unstarBtn,
			)
			starredContent.Add(dictHeader)

			for _, article := range results {
				content := createSelectableArticleContent(article.Content)
				starredContent.Add(content)
				starredContent.Add(widget.NewSeparator())
			}
			starredContent.Refresh()
			starredContentScroll.ScrollToTop()
		}

		starredWordList.OnSelected = func(id widget.ListItemID) {
			showStarredArticle(int(id))
		}

		// Filter input
		filterEntry := widget.NewEntry()
		filterEntry.SetPlaceHolder("Filter starred articles...")
		filterEntry.OnChanged = func(text string) {
			filterText = text
			if text == "" {
				filteredStarred = starred
			} else {
				filteredStarred = nil
				for _, s := range starred {
					if strings.Contains(strings.ToLower(s.Word), strings.ToLower(text)) {
						filteredStarred = append(filteredStarred, s)
					}
				}
			}
			starredWordList.Refresh()
			if len(filteredStarred) > 0 {
				showStarredArticle(0)
				starredWordList.Select(0)
			} else {
				starredContent.RemoveAll()
				starredContent.Add(widget.NewLabel("No matching starred articles"))
				starredContentHeader.SetText("")
				starredContent.Refresh()
			}
		}

		// Layout: sidebar with filter + word list | content area
		starredSidebar := container.NewBorder(
			filterEntry,
			nil, nil, nil,
			starredWordList,
		)

		starredContentArea := container.NewBorder(
			container.NewVBox(starredContentHeader, widget.NewSeparator()),
			nil, nil, nil,
			starredContentScroll,
		)

		starredSplit := container.NewHSplit(starredSidebar, starredContentArea)
		starredSplit.SetOffset(0.28)

		// Show first article
		if len(filteredStarred) > 0 {
			showStarredArticle(0)
			starredWordList.Select(0)
		}

		starredDialog := dialog.NewCustom("Starred Articles", "Close", starredSplit, w)
		starredDialog.Resize(fyne.NewSize(900, 600))
		starredDialog.Show()
	})

	// Shortcut hint overlay (Ctrl+K or Cmd+K) - pill style like Homebrew
	shortcutText := "Ctrl+K"
	if runtime.GOOS == "darwin" {
		shortcutText = "Cmd+K"
	}
	shortcutHint := canvas.NewText(shortcutText, color.RGBA{R: 100, G: 100, B: 110, A: 255})
	shortcutHint.TextSize = 12
	shortcutBg := canvas.NewRectangle(color.RGBA{R: 230, G: 230, B: 235, A: 255})
	shortcutBg.CornerRadius = 4
	shortcutPill := container.NewStack(shortcutBg, container.NewPadded(shortcutHint))

	// Search row: entry with hint overlay + history and starred buttons
	searchWithHint := container.NewStack(
		searchEntry,
		container.NewBorder(nil, nil, nil, container.NewPadded(shortcutPill), nil),
	)
	searchButtons := container.NewHBox(starredBtn, historyBtn)
	searchRow := container.NewBorder(nil, nil, nil, searchButtons, searchWithHint)

	// Search mode radio buttons
	modeGroup := widget.NewRadioGroup([]string{
		"Exact",
		"Prefix",
		"Contains",
		"Full-text",
	}, func(selected string) {
		switch selected {
		case "Exact":
			currentMode = search.ModeExact
		case "Prefix":
			currentMode = search.ModePrefix
		case "Contains":
			currentMode = search.ModeFuzzy
		case "Full-text":
			currentMode = search.ModeReverse
		}
		// Re-search with new mode
		if searchEntry.Text != "" {
			doSearch(searchEntry.Text)
		}
	})
	modeGroup.SetSelected("Exact")
	modeGroup.Horizontal = true

	// Group results checkbox
	groupCheck := widget.NewCheck("Group results", func(checked bool) {
		groupResultsSetting = checked
		if settings != nil {
			settings.SetBool("group_results", checked)
		}
		// Re-group cached results (no new search needed)
		regroupResults()
	})
	groupCheck.SetChecked(groupResultsSetting)

	// Dictionaries button - opens selection dialog
	dictsBtn := widget.NewButton("Dictionaries...", func() {
		if len(allDicts) == 0 {
			dialog.ShowInformation("Dictionaries", "No dictionaries available", w)
			return
		}

		// Language name mapping
		langNames := map[string]string{
			"sa": "Sanskrit",
			"en": "English",
			"de": "German",
			"fr": "French",
			"la": "Latin",
		}
		getLangName := func(code string) string {
			if name, ok := langNames[code]; ok {
				return name
			}
			return strings.ToUpper(code)
		}

		// Group dictionaries by direction (from → to)
		type DictGroup struct {
			Key       string // "sa→en"
			Direction string // "Sanskrit → English"
			Dicts     []search.Dict
		}
		groupMap := make(map[string]*DictGroup)
		var groupOrder []string

		for _, d := range allDicts {
			key := d.FromLang + "→" + d.ToLang
			if _, ok := groupMap[key]; !ok {
				direction := getLangName(d.FromLang) + " → " + getLangName(d.ToLang)
				groupMap[key] = &DictGroup{Key: key, Direction: direction}
				groupOrder = append(groupOrder, key)
			}
			groupMap[key].Dicts = append(groupMap[key].Dicts, d)
		}

		// Sort dictionaries within each group by name
		for _, group := range groupMap {
			dicts := group.Dicts
			for i := 0; i < len(dicts)-1; i++ {
				for j := i + 1; j < len(dicts); j++ {
					if dicts[i].Name > dicts[j].Name {
						dicts[i], dicts[j] = dicts[j], dicts[i]
					}
				}
			}
		}

		// Build dialog content
		content := container.NewVBox()

		// Track all checkboxes for global select all
		var allCheckboxes []*widget.Check
		var groupSelectAlls []*widget.Check

		// Helper to count selected
		countSelected := func() int {
			count := 0
			for _, cb := range allCheckboxes {
				if cb.Checked {
					count++
				}
			}
			return count
		}

		// Global select all
		globalSelectAll := widget.NewCheck("Select All", nil)
		globalSelectAll.OnChanged = func(checked bool) {
			for _, cb := range allCheckboxes {
				cb.SetChecked(checked)
			}
			for _, gsa := range groupSelectAlls {
				gsa.SetChecked(checked)
			}
		}
		content.Add(globalSelectAll)
		content.Add(widget.NewSeparator())

		// Build each group
		for _, key := range groupOrder {
			group := groupMap[key]

			// Group header with select all
			groupHeader := widget.NewLabel(group.Direction)
			groupHeader.TextStyle = fyne.TextStyle{Bold: true}

			var groupCheckboxes []*widget.Check
			groupSelectAll := widget.NewCheck("All", nil)

			// Create checkboxes for each dictionary in group
			dictContainer := container.NewVBox()
			for _, d := range group.Dicts {
				dictCode := d.Code
				dictName := d.Name
				cb := widget.NewCheck(dictName, func(checked bool) {
					selectedDicts[dictCode] = checked
				})
				cb.SetChecked(selectedDicts[dictCode])
				groupCheckboxes = append(groupCheckboxes, cb)
				allCheckboxes = append(allCheckboxes, cb)
				dictContainer.Add(cb)
			}

			// Group select all handler
			groupSelectAll.OnChanged = func(checked bool) {
				for _, cb := range groupCheckboxes {
					cb.SetChecked(checked)
				}
			}
			groupSelectAlls = append(groupSelectAlls, groupSelectAll)

			// Check if all in group are selected
			allInGroupSelected := true
			for _, d := range group.Dicts {
				if !selectedDicts[d.Code] {
					allInGroupSelected = false
					break
				}
			}
			groupSelectAll.SetChecked(allInGroupSelected)

			// Group row: header + select all
			groupRow := container.NewHBox(groupHeader, groupSelectAll)
			content.Add(groupRow)
			content.Add(dictContainer)
			content.Add(widget.NewSeparator())
		}

		// Check if all are selected globally
		globalSelectAll.SetChecked(countSelected() == len(allCheckboxes))

		// Wrap in scroll
		scroll := container.NewVScroll(content)
		scroll.SetMinSize(fyne.NewSize(400, 400))

		// Create dialog with Apply/Cancel
		dlg := dialog.NewCustomConfirm("Select Dictionaries", "Apply", "Cancel", scroll, func(ok bool) {
			if ok {
				// Update selection from checkboxes
				for i, cb := range allCheckboxes {
					// Find corresponding dict
					idx := 0
					for _, key := range groupOrder {
						group := groupMap[key]
						for _, d := range group.Dicts {
							if idx == i {
								selectedDicts[d.Code] = cb.Checked
							}
							idx++
						}
					}
				}
				saveSelectedDicts()
				// Re-search if there's a query
				if searchEntry.Text != "" {
					doSearch(searchEntry.Text)
				}
			}
		}, w)
		dlg.Resize(fyne.NewSize(500, 500))
		dlg.Show()
	})

	// Zoom dropdown in toolbar
	zoomOptions := []string{"50%", "75%", "100%", "125%", "150%", "175%", "200%"}
	zoomSelect := widget.NewSelect(zoomOptions, func(selected string) {
		// Parse percentage
		var newZoom int
		fmt.Sscanf(selected, "%d%%", &newZoom)
		if newZoom > 0 {
			zoomPercent = newZoom
			applyZoom(newZoom)
			if settings != nil {
				settings.Set("zoom", strconv.Itoa(newZoom))
			}
		}
	})
	zoomSelect.SetSelected(fmt.Sprintf("%d%%", zoomPercent))

	// Settings/About button (cog icon)
	settingsBtn := widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
		// About dialog
		aboutTitle := widget.NewLabel("Sanskrit Upāya")
		aboutTitle.TextStyle = fyne.TextStyle{Bold: true}
		aboutTitle.Alignment = fyne.TextAlignCenter

		authorLabel := widget.NewLabel("by Myke Bilyanskyy")
		authorLabel.Alignment = fyne.TextAlignCenter
		authorLink := widget.NewHyperlink("myke.blog", nil)
		authorLink.SetURLFromString("https://myke.blog")

		aboutText := widget.NewLabel("Dictionary data from Cologne Digital Sanskrit Dictionaries")
		aboutText.Alignment = fyne.TextAlignCenter
		dataLink := widget.NewHyperlink("www.sanskrit-lexicon.uni-koeln.de", nil)
		dataLink.SetURLFromString("https://www.sanskrit-lexicon.uni-koeln.de/")

		dialogContent := container.NewVBox(
			widget.NewLabel(""), // spacer
			aboutTitle,
			authorLabel,
			container.NewCenter(authorLink),
			widget.NewLabel(""), // spacer
			aboutText,
			container.NewCenter(dataLink),
			widget.NewLabel(""), // spacer
		)

		dlg := dialog.NewCustom("About", "Close", dialogContent, w)
		dlg.Resize(fyne.NewSize(400, 250))
		dlg.Show()
	})

	// Zoom control: just the dropdown
	zoomControl := container.NewHBox(widget.NewLabel("Zoom:"), zoomSelect)

	// Toolbar: mode + group checkbox on left, zoom + settings on right
	toolbarRight := container.NewHBox(zoomControl, widget.NewSeparator(), settingsBtn)
	toolbar := container.NewBorder(nil, nil, nil, toolbarRight,
		container.NewHBox(modeGroup, widget.NewSeparator(), groupCheck, widget.NewSeparator(), dictsBtn),
	)

	// Top bar: search on top row, toolbar below
	topBar := container.NewVBox(
		searchRow,
		toolbar,
	)

	// Version display (right side of status bar)
	versionLabel := widget.NewLabel(Version)
	versionLabel.Importance = widget.LowImportance

	// Update indicator (hidden by default)
	updateLabel := widget.NewLabel("")
	updateLabel.Hide()

	// Version container: version label + update indicator
	versionContainer := container.NewHBox(updateLabel, versionLabel)

	// Check for updates in background
	go func() {
		result := version.Check(Version)
		if result != nil && result.HasUpdate {
			fyne.Do(func() {
				updateLabel.SetText(fmt.Sprintf("Update available: %s", result.LatestVersion))
				updateLabel.Show()
				versionContainer.Refresh()
			})
		}
	}()

	// Status bar: status text on left, version on right
	statusBar := container.NewBorder(nil, nil, nil, versionContainer, statusText)

	// Main layout with status bar at bottom
	content := container.NewBorder(
		container.NewVBox(topBar, widget.NewSeparator()),
		statusBar,
		nil, nil,
		mainContent,
	)

	// Add padding
	padded := container.NewPadded(content)

	w.SetContent(padded)

	// Keyboard shortcuts
	w.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyK,
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		w.Canvas().Focus(searchEntry)
	})
	// Also support Super (Cmd on Mac)
	w.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyK,
		Modifier: fyne.KeyModifierSuper,
	}, func(shortcut fyne.Shortcut) {
		w.Canvas().Focus(searchEntry)
	})

	// Zoom helper function
	changeZoom := func(delta int) {
		// Find current index in zoomOptions
		currentIdx := -1
		currentStr := fmt.Sprintf("%d%%", zoomPercent)
		for i, opt := range zoomOptions {
			if opt == currentStr {
				currentIdx = i
				break
			}
		}
		// Calculate new index
		newIdx := currentIdx + delta
		if newIdx < 0 {
			newIdx = 0
		} else if newIdx >= len(zoomOptions) {
			newIdx = len(zoomOptions) - 1
		}
		// Apply new zoom
		if newIdx != currentIdx {
			zoomSelect.SetSelected(zoomOptions[newIdx])
		}
	}

	// Ctrl/Cmd + Plus to zoom in
	w.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyEqual, // + is Shift+= but KeyEqual works for Ctrl+=
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		changeZoom(1)
	})
	w.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyEqual,
		Modifier: fyne.KeyModifierSuper,
	}, func(shortcut fyne.Shortcut) {
		changeZoom(1)
	})

	// Ctrl/Cmd + Minus to zoom out
	w.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyMinus,
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		changeZoom(-1)
	})
	w.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyMinus,
		Modifier: fyne.KeyModifierSuper,
	}, func(shortcut fyne.Shortcut) {
		changeZoom(-1)
	})

	// Keyboard navigation (works even when search field is focused)
	if deskCanvas, ok := w.Canvas().(desktop.Canvas); ok {
		deskCanvas.SetOnKeyDown(func(ev *fyne.KeyEvent) {
			switch ev.Name {
			case fyne.KeyDown:
				if len(groupedResults) > 0 {
					if currentSelection < len(groupedResults)-1 {
						navigateTo(currentSelection + 1)
					} else if currentSelection == -1 {
						navigateTo(0)
					}
				}
			case fyne.KeyUp:
				if len(groupedResults) > 0 && currentSelection > 0 {
					navigateTo(currentSelection - 1)
				}
			}
		})
	}

	// Focus search on start
	w.Canvas().Focus(searchEntry)
}
