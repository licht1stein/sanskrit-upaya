## ADDED Requirements

### Requirement: Three-Panel Layout

The search page SHALL use a three-panel layout with results sidebar, resizable divider, and article content area.

![Search Page Layout](../assets/search-page-layout.png)

**CSS - Panel Layout:**
```css
.workspace {
  display: flex;
  overflow: hidden;
}

.panel-left {
  width: 320px;
  min-width: 240px;
  max-width: 50%;
  background: var(--bg-base);  /* #181818 */
}

.panel-right {
  flex: 1;
  background: var(--bg-surface);  /* #1f1f1f */
  min-width: 300px;
}

.divider {
  width: 1px;
  background: var(--border-subtle);  /* #252525 */
  position: relative;
  cursor: col-resize;
  flex-shrink: 0;
}

.divider:hover,
.divider.dragging {
  background: var(--success);  /* #5db0a3 */
}
```

#### Scenario: Panel layout displayed

- **GIVEN** the application is running
- **WHEN** user navigates to search page
- **THEN** left panel displays search input and results list
- **AND** right panel displays selected article content
- **AND** vertical divider separates the panels
- **AND** divider is draggable to resize panels

### Requirement: Search Input

The search input SHALL include icon, placeholder text, and clear button.

![Search Input](../assets/search-input.png)

**CSS - Search Input:**
```css
.search-section {
  padding: 12px;
}

.search-box {
  position: relative;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: 10px;
  width: 14px;
  height: 14px;
  color: var(--text-muted);  /* #5a5a5a */
  pointer-events: none;
}

.search-input {
  width: 100%;
  padding: 8px 60px 8px 32px;
  background: var(--bg-surface);  /* #1f1f1f */
  border: 1px solid var(--border-subtle);  /* #252525 */
  border-radius: var(--radius-md);  /* 4px */
  color: var(--text-primary);  /* #e1e1e1 */
  font-family: var(--font-sans);
  font-size: 13px;
  outline: none;
  transition: border-color 0.15s, background 0.15s;
}

.search-input::placeholder {
  color: var(--text-muted);  /* #5a5a5a */
}

.search-input:focus {
  border-color: var(--success);  /* #5db0a3 */
  background: var(--bg-elevated);  /* #262626 */
}

.search-clear {
  position: absolute;
  right: 8px;
  display: none;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  background: var(--bg-surface);  /* #1f1f1f */
  border: none;
  border-radius: 3px;
  color: var(--text-muted);  /* #5a5a5a */
  cursor: pointer;
  transition: all 0.1s;
}

.search-clear:hover {
  background: var(--bg-hover);  /* #2a2a2a */
  color: var(--text-primary);  /* #e1e1e1 */
}

.search-box.has-value .search-clear {
  display: flex;
}
```

#### Scenario: Search input displayed

- **GIVEN** search page is displayed
- **WHEN** user views the search input
- **THEN** magnifying glass icon appears on left inside input
- **AND** placeholder text reads "Search..."
- **AND** input has subtle border (border-subtle color #252525)

#### Scenario: Clear button appears with content

- **GIVEN** search input is empty
- **WHEN** user types "yoga" in search input
- **THEN** clear button (×) appears on right side of input
- **AND** clear button was not visible before typing

#### Scenario: Clear search with button click

- **GIVEN** search input contains "yoga"
- **AND** results are displayed
- **WHEN** user clicks clear button (×)
- **THEN** search input becomes empty
- **AND** clear button disappears
- **AND** results list is cleared

#### Scenario: Clear search with Escape key

- **GIVEN** search input contains "yoga"
- **AND** search input is focused
- **WHEN** user presses Escape key
- **THEN** search input becomes empty
- **AND** results list is cleared

### Requirement: Search Mode Tabs

The search page SHALL display mode tabs: PREFIX, EXACT, FUZZY, REVERSE.

**CSS - Segmented Control:**
```css
.modes-section {
  padding: 0 12px 12px;
}

.segmented-control {
  display: flex;
  background: var(--bg-surface);  /* #1f1f1f */
  border-radius: var(--radius-md);  /* 4px */
  padding: 2px;
  gap: 2px;
}

.segment {
  flex: 1;
  padding: 6px 8px;
  background: transparent;
  border: none;
  border-radius: var(--radius-sm);  /* 3px */
  color: var(--text-secondary);  /* #a0a0a0 */
  font-family: var(--font-sans);
  font-size: 11px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.1s;
  text-transform: uppercase;
  letter-spacing: 0.02em;
}

.segment:hover {
  color: var(--text-primary);  /* #e1e1e1 */
}

.segment.active {
  color: var(--text-primary);  /* #e1e1e1 */
  background: var(--bg-active);  /* #323232 */
}
```

#### Scenario: Mode tabs displayed

- **GIVEN** search page is displayed
- **WHEN** user views the mode tabs
- **THEN** four tabs are visible: PREFIX, EXACT, FUZZY, REVERSE
- **AND** one tab is active (default: PREFIX)
- **AND** active tab has bg-active background (#323232)

#### Scenario: Mode tab selection

- **GIVEN** PREFIX mode is active
- **WHEN** user clicks EXACT tab
- **THEN** EXACT tab becomes active with bg-active background
- **AND** PREFIX tab becomes inactive
- **AND** search results update using EXACT mode
- **AND** transition takes 100ms

### Requirement: Results List

Search results SHALL display word, devanagari, and dictionary code for each match.

![Search Results](../assets/search-results.png)

**CSS - Results:**
```css
.results-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border-top: 1px solid var(--border-subtle);  /* #252525 */
}

.results-header {
  padding: 8px 12px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.results-label {
  font-size: 11px;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);  /* #5a5a5a */
}

.results-count {
  font-family: var(--font-mono);
  font-size: 10px;
  color: var(--text-muted);  /* #5a5a5a */
  background: var(--bg-surface);  /* #1f1f1f */
  padding: 2px 8px;
  border-radius: 10px;
}

.result-item {
  padding: 10px 12px;
  cursor: pointer;
  border-left: 2px solid transparent;
  transition: all 0.1s;
}

.result-item:hover {
  background: var(--bg-hover);  /* #2a2a2a */
}

.result-item.selected {
  background: var(--bg-active);  /* #323232 */
  border-left-color: var(--success);  /* #5db0a3 */
}

.result-word {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);  /* #e1e1e1 */
  margin-bottom: 2px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.result-deva {
  font-family: var(--font-sanskrit);
  font-size: 12px;
  color: var(--text-muted);  /* #5a5a5a */
  font-weight: 400;
}

.result-dict {
  font-family: var(--font-mono);
  font-size: 10px;
  color: var(--success);  /* #5db0a3 */
  text-transform: uppercase;
}
```

#### Scenario: Result items displayed

- **GIVEN** user searches for "yoga"
- **WHEN** results are returned
- **THEN** each result shows word in primary text color (#e1e1e1)
- **AND** each result shows devanagari form in muted text color (#5a5a5a)
- **AND** each result shows dictionary code in success color (#5db0a3)

#### Scenario: Result count displayed

- **GIVEN** search returns 42 results
- **WHEN** user views results panel
- **THEN** header shows "RESULTS" with count "42"

#### Scenario: Result selection via click

- **GIVEN** results list contains multiple items
- **WHEN** user clicks on a result item
- **THEN** clicked item is highlighted as selected (bg-active, left border)
- **AND** article content displays in right panel

#### Scenario: Result navigation via keyboard

- **GIVEN** results list is displayed
- **AND** first result is selected
- **WHEN** user presses down arrow key
- **THEN** second result becomes selected
- **AND** article content updates to show second result

### Requirement: Article View Modes

The article panel SHALL support two view modes: Stacked and Compare.

![View Toggle Buttons](../assets/view-toggle-buttons.png)

**CSS - View Toggle:**
```css
.view-toggle {
  display: flex;
  align-items: center;
  gap: 2px;
  padding: 2px;
  background: var(--bg-elevated);  /* #262626 */
  border-radius: var(--radius-sm);  /* 3px */
}

.view-toggle-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 24px;
  background: transparent;
  border: none;
  border-radius: 2px;
  color: var(--text-muted);  /* #5a5a5a */
  cursor: pointer;
  transition: all 0.15s;
}

.view-toggle-btn:hover {
  color: var(--text-primary);  /* #e1e1e1 */
}

.view-toggle-btn.active {
  background: var(--bg-base);  /* #181818 */
  color: var(--text-primary);  /* #e1e1e1 */
  box-shadow: 0 1px 2px rgba(0,0,0,0.1);
}

.view-toggle-btn svg {
  width: 14px;
  height: 14px;
}
```

#### Scenario: View toggle displayed

- **GIVEN** article is displayed
- **AND** word exists in multiple dictionaries
- **WHEN** user views the article toolbar
- **THEN** two view toggle buttons are visible (stacked icon, compare icon)
- **AND** active mode button has bg-base background with shadow

#### Scenario: Switch to compare mode

- **GIVEN** stacked view is active
- **WHEN** user clicks compare view button
- **THEN** compare view is displayed
- **AND** compare button shows active state
- **AND** stacked button becomes inactive

### Requirement: Stacked View

Stacked view SHALL display all dictionary entries in a vertical list.

![Stacked View](../assets/stacked-view.png)

**CSS - Stacked View:**
```css
.stacked-view {
  padding: 16px 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}
```

#### Scenario: Stacked view layout

- **GIVEN** word "yoga" exists in 4 dictionaries
- **AND** stacked view is active
- **WHEN** user views the article
- **THEN** 4 dictionary entry cards display vertically
- **AND** cards are separated by 16px gap
- **AND** cards appear in dictionary priority order

### Requirement: Dictionary Entry Card

Each dictionary entry card SHALL have header with actions and content area.

![Entry Card](../assets/entry-card.png)

**CSS - Entry Card:**
```css
.dict-card-entry {
  background: var(--bg-base);  /* #181818 */
  border: 1px solid var(--border-subtle);  /* #252525 */
  border-radius: var(--radius-md);  /* 4px */
  overflow: hidden;
}

.dict-card-entry-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  background: var(--bg-base);  /* #181818 */
  border-bottom: 1px solid var(--border-default);  /* #2e2e2e */
}

.dict-card-entry-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.dict-card-entry-title .code {
  font-family: var(--font-mono);
  font-size: 11px;
  padding: 2px 6px;
  background: transparent;
  border: 1px solid var(--border-default);  /* #2e2e2e */
  border-radius: 3px;
  color: var(--text-secondary);  /* #a0a0a0 */
}

.dict-card-entry-title .name {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-primary);  /* #e1e1e1 */
}

.dict-card-entry-actions {
  display: flex;
  align-items: center;
  gap: 6px;
}

.dict-card-entry-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  background: transparent;
  border: 1px solid var(--border-subtle);  /* #252525 */
  border-radius: var(--radius-sm);  /* 3px */
  color: var(--text-muted);  /* #5a5a5a */
  font-size: 11px;
  cursor: pointer;
  transition: all 0.15s;
}

.dict-card-entry-btn:hover {
  border-color: var(--border-default);  /* #2e2e2e */
  color: var(--text-primary);  /* #e1e1e1 */
}

.dict-card-entry-btn svg {
  width: 12px;
  height: 12px;
}

.dict-card-entry-content {
  padding: 14px;
  font-size: 14px;
  line-height: 1.6;
  color: var(--text-primary);  /* #e1e1e1 */
}

.dict-card-entry-content b {
  color: var(--success);  /* #5db0a3 */
  font-weight: 500;
}

.dict-card-entry-content i {
  color: var(--warning);  /* #d19a66 */
  font-style: normal;
}
```

#### Scenario: Entry card header

- **GIVEN** dictionary entry card is displayed
- **WHEN** user views the card header
- **THEN** header has bg-base background (#181818)
- **AND** dictionary code displays in outline badge style
- **AND** dictionary full name displays next to code
- **AND** Copy button displays on right
- **AND** Star button displays next to Copy button

#### Scenario: Copy entry content

- **GIVEN** dictionary entry is displayed
- **WHEN** user clicks Copy button
- **THEN** entry text content is copied to clipboard
- **AND** button shows brief "Copied!" feedback

#### Scenario: Star entry

- **GIVEN** dictionary entry is displayed
- **AND** entry is not starred
- **WHEN** user clicks Star button
- **THEN** entry is added to starred items
- **AND** star icon becomes filled

### Requirement: Compare View

Compare view SHALL display 2-4 dictionary entries side-by-side in columns.

![Compare View 2 Columns](../assets/compare-view-2col.png)

![Compare View 4 Columns](../assets/compare-view-4col.png)

**CSS - Compare View Grid:**
```css
.compare-view {
  display: grid;
  gap: 1px;
  background: var(--border-subtle);  /* #252525 - creates gap lines */
  height: 100%;
  grid-template-columns: repeat(var(--compare-cols, 2), 1fr);
}

.compare-view.has-add-btn {
  grid-template-columns: repeat(var(--compare-cols, 2), 1fr) 60px;
}

.compare-column {
  background: var(--bg-surface);  /* #1f1f1f */
  display: flex;
  flex-direction: column;
  min-height: 0;
  animation: compareColumnIn 0.2s ease-out;
}

@keyframes compareColumnIn {
  from {
    opacity: 0;
    transform: scale(0.95);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

.compare-column-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  background: var(--bg-base);  /* #181818 */
  border-bottom: 1px solid var(--border-subtle);  /* #252525 */
  gap: 8px;
}

.compare-column-content {
  flex: 1;
  overflow-y: auto;
  padding: 14px;
}
```

**CSS - Simplified Dropdown Selector:**
```css
.compare-select {
  padding: 4px 0;
  padding-right: 18px;
  background: transparent;
  border: none;
  color: var(--text-primary);  /* #e1e1e1 */
  font-family: var(--font-sans);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='10' height='10' viewBox='0 0 24 24' fill='none' stroke='%23666' stroke-width='2.5'%3E%3Cpolyline points='6 9 12 15 18 9'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 0 center;
}

.compare-select:hover {
  color: var(--text-secondary);  /* #a0a0a0 */
}

.compare-select:focus {
  outline: none;
}
```

**CSS - Add Column Button:**
```css
.compare-add-column {
  background: var(--bg-surface);  /* #1f1f1f */
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 0;
  cursor: pointer;
  transition: all 0.15s;
  border-left: 1px dashed var(--border-subtle);  /* #252525 */
}

.compare-add-column:hover {
  background: var(--bg-elevated);  /* #262626 */
}

.compare-add-column:hover .compare-add-icon {
  color: var(--success);  /* #5db0a3 */
}

.compare-add-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-muted);  /* #5a5a5a */
  transition: all 0.15s;
}

.compare-add-icon svg {
  width: 20px;
  height: 20px;
  stroke-width: 2;
}
```

**CSS - Remove Column Button:**
```css
.compare-remove-btn {
  width: 22px;
  height: 22px;
  border: none;
  background: transparent;
  color: var(--text-muted);  /* #5a5a5a */
  cursor: pointer;
  border-radius: var(--radius-sm);  /* 3px */
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;
  opacity: 0.6;
}

.compare-remove-btn:hover {
  background: var(--bg-elevated);  /* #262626 */
  color: var(--error);
  opacity: 1;
}

.compare-remove-btn svg {
  width: 14px;
  height: 14px;
}
```

**JavaScript Logic:**
```javascript
// State management
let compareColumns = [0, 1];  // Array of dictionary indices
const MAX_COMPARE_COLUMNS = 4;
const MIN_COMPARE_COLUMNS = 2;

// Add column - selects next unused dictionary
function addCompareColumn() {
  if (compareColumns.length >= MAX_COMPARE_COLUMNS) return;
  const usedIndices = new Set(compareColumns);
  for (let i = 0; i < dictionaries.length; i++) {
    if (!usedIndices.has(i)) {
      compareColumns.push(i);
      break;
    }
  }
  render();
}

// Remove column (only if > 2 columns)
function removeCompareColumn(colIndex) {
  if (compareColumns.length <= MIN_COMPARE_COLUMNS) return;
  compareColumns.splice(colIndex, 1);
  render();
}
```

#### Scenario: Compare view default layout

- **GIVEN** word exists in 4 dictionaries
- **WHEN** compare view is activated
- **THEN** 2 columns display by default
- **AND** first column shows first dictionary
- **AND** second column shows second dictionary
- **AND** add button (+) appears as narrow 60px column on right

#### Scenario: Add third column

- **GIVEN** compare view shows 2 columns
- **AND** word exists in 4 dictionaries
- **WHEN** user clicks add button (+)
- **THEN** third column appears with third dictionary
- **AND** new column animates in (scale 0.95→1, fade in, 200ms)
- **AND** remove button (×) appears on column 3

#### Scenario: Add fourth column

- **GIVEN** compare view shows 3 columns
- **WHEN** user clicks add button (+)
- **THEN** fourth column appears with fourth dictionary
- **AND** add button disappears (max 4 columns reached)

#### Scenario: Remove column

- **GIVEN** compare view shows 3 columns
- **WHEN** user clicks remove button (×) on column 3
- **THEN** column 3 is removed
- **AND** add button reappears

#### Scenario: Cannot remove below minimum

- **GIVEN** compare view shows 2 columns
- **WHEN** user views the columns
- **THEN** no remove buttons are visible
- **AND** minimum 2 columns are always maintained

### Requirement: Column Dictionary Selector

Each compare column SHALL have a dropdown to select which dictionary to display.

#### Scenario: Selector displayed

- **GIVEN** compare view column is displayed
- **WHEN** user views column header
- **THEN** dictionary name displays as text
- **AND** small chevron icon indicates dropdown
- **AND** no border or background on selector (minimal style)

#### Scenario: Change column dictionary

- **GIVEN** column shows "Monier-Williams" dictionary
- **WHEN** user clicks dropdown and selects "Apte Practical"
- **THEN** column content updates to show Apte Practical entry
- **AND** dropdown shows "Apte Practical" as selected

### Requirement: Add Column Button

The add column button SHALL be a narrow dashed-border column.

#### Scenario: Add button displayed

- **GIVEN** compare view shows fewer than 4 columns
- **AND** more dictionaries available than columns shown
- **WHEN** user views compare view
- **THEN** add button displays as 60px wide column on right
- **AND** button has dashed left border
- **AND** button shows + icon centered

#### Scenario: Add button hover state

- **GIVEN** add button is displayed
- **WHEN** user hovers over add button
- **THEN** background changes to bg-elevated (#262626)
- **AND** + icon color changes to success (#5db0a3)
- **AND** transition takes 150ms
