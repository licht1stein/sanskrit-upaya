## ADDED Requirements

### Requirement: Two-Column Layout

The dictionaries page SHALL display Active and Available dictionaries in two columns.

![Dictionaries Page Layout](../assets/dictionaries-page-layout.png)

**CSS Layout:**
```css
.dict-content {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 24px;
}

.dict-column {
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.dict-column-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 12px;
  margin-bottom: 12px;
  border-bottom: 1px solid var(--border-subtle);  /* #252525 */
}

.dict-column-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-secondary);  /* #a0a0a0 */
}

.dict-column-title .count {
  font-family: var(--font-mono);
  font-size: 10px;
  padding: 2px 6px;
  background: var(--bg-elevated);  /* #262626 */
  border-radius: 10px;
  color: var(--text-muted);  /* #5a5a5a */
}

.dict-column.active .dict-column-title .count {
  background: var(--success);  /* #5db0a3 */
  color: var(--text-inverse);  /* #181818 */
}

.dict-column-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
  flex: 1;
  overflow-y: auto;
  min-height: 200px;
}
```

#### Scenario: Column layout displayed

- **GIVEN** the application is running
- **WHEN** user navigates to dictionaries page
- **THEN** left column header shows "ACTIVE" with count badge
- **AND** right column header shows "AVAILABLE" with count badge
- **AND** columns have equal width
- **AND** columns are separated by 24px gap

### Requirement: Filter Input

The dictionaries page SHALL include a filter input to search dictionaries.

**CSS:**
```css
.dict-search {
  position: relative;
  display: flex;
  align-items: center;
}

.dict-search-icon {
  position: absolute;
  left: 8px;
  width: 12px;
  height: 12px;
  color: var(--text-muted);  /* #5a5a5a */
  pointer-events: none;
}

.dict-search-input {
  width: 200px;
  padding: 6px 8px 6px 28px;
  background: var(--bg-surface);  /* #1f1f1f */
  border: 1px solid var(--border-subtle);  /* #252525 */
  border-radius: var(--radius-sm);  /* 3px */
  color: var(--text-primary);  /* #e1e1e1 */
  font-family: var(--font-sans);
  font-size: 12px;
  outline: none;
}

.dict-search-input:focus {
  border-color: var(--border-default);  /* #2e2e2e */
}

.dict-search-input::placeholder {
  color: var(--text-muted);  /* #5a5a5a */
}
```

#### Scenario: Filter input displayed

- **GIVEN** dictionaries page is displayed
- **WHEN** user views the page
- **THEN** filter input appears at top with search icon
- **AND** placeholder text reads "Filter dictionaries..."

#### Scenario: Filter matches dictionaries

- **GIVEN** dictionaries page is displayed
- **AND** dictionaries exist in both columns
- **WHEN** user types "apte" in filter input
- **THEN** both columns filter to show only dictionaries matching "apte"
- **AND** matching is case-insensitive
- **AND** matching checks both code and name fields

### Requirement: Statistics Display

The dictionaries page SHALL show statistics about active dictionaries.

**CSS:**
```css
.dict-stats {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  color: var(--text-muted);  /* #5a5a5a */
}

.dict-stats-num {
  font-family: var(--font-mono);
  color: var(--text-secondary);  /* #a0a0a0 */
}
```

#### Scenario: Stats displayed

- **GIVEN** dictionaries page is displayed
- **AND** 4 dictionaries are active with 300k total entries
- **WHEN** user views the page
- **THEN** stats display shows "4 of 36 active"
- **AND** stats display shows "300k entries"

### Requirement: Dictionary Card

Each dictionary SHALL be displayed as a card with code, name, entries, and language.

![Dictionary Card](../assets/dictionary-card.png)

**CSS:**
```css
.dict-card {
  display: flex;
  align-items: center;
  padding: 10px 12px;
  background: var(--bg-base);  /* #181818 */
  border: 1px solid var(--border-subtle);  /* #252525 */
  border-radius: var(--radius-md);  /* 4px */
  transition: all 0.2s cubic-bezier(0.2, 0, 0, 1);
  cursor: pointer;
  position: relative;
  user-select: none;
}

.dict-card:hover {
  border-color: var(--border-default);  /* #2e2e2e */
}

.dict-card-info {
  flex: 1;
  min-width: 0;
  margin-right: 10px;
}

.dict-card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 2px;
}

.dict-code {
  font-family: var(--font-mono);
  font-size: 10px;
  padding: 2px 5px;
  background: var(--bg-elevated);  /* #262626 */
  border-radius: 3px;
  color: var(--text-secondary);  /* #a0a0a0 */
  flex-shrink: 0;
}

.dict-name {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);  /* #e1e1e1 */
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.dict-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 11px;
  color: var(--text-muted);  /* #5a5a5a */
}

.dict-entries {
  font-family: var(--font-mono);
}
```

#### Scenario: Card content displayed

- **GIVEN** dictionaries page is displayed
- **WHEN** user views a dictionary card
- **THEN** card shows dictionary code in badge (e.g., "MW")
- **AND** card shows full dictionary name (e.g., "Monier-Williams")
- **AND** card shows entry count (e.g., "180k entries")
- **AND** card shows language direction (e.g., "sa→en")

#### Scenario: Card click toggles state

- **GIVEN** dictionary "MW" is in Available column
- **WHEN** user clicks anywhere on the card (except action button)
- **THEN** dictionary moves to Active column
- **AND** dictionary appears at end of Active list

### Requirement: Action Buttons

Active cards SHALL show remove (×) button, available cards SHALL show add (+) button.

**CSS:**
```css
.dict-action {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  background: transparent;
  border: 1px solid var(--border-subtle);  /* #252525 */
  border-radius: 4px;
  color: var(--text-muted);  /* #5a5a5a */
  cursor: pointer;
  transition: all 0.15s;
  flex-shrink: 0;
}

.dict-action:hover {
  border-color: var(--border-default);  /* #2e2e2e */
  color: var(--text-primary);  /* #e1e1e1 */
}

.dict-action.add:hover {
  border-color: var(--success);  /* #5db0a3 */
  background: var(--success);  /* #5db0a3 */
  color: var(--text-inverse);  /* #181818 */
}

.dict-action.remove:hover {
  border-color: var(--warning);  /* #d19a66 */
  background: var(--warning);  /* #d19a66 */
  color: var(--text-inverse);  /* #181818 */
}

.dict-action svg {
  width: 12px;
  height: 12px;
}
```

#### Scenario: Remove from active

- **GIVEN** dictionary "MW" is in Active column
- **WHEN** user clicks the remove button (×)
- **THEN** dictionary moves to Available column
- **AND** Active column count decreases by 1

#### Scenario: Add to active

- **GIVEN** dictionary "VCP" is in Available column
- **WHEN** user clicks the add button (+)
- **THEN** dictionary moves to Active column
- **AND** dictionary appears at end of Active list

### Requirement: Hover Animation for Reorder Controls

Active cards SHALL reveal drag handle and reorder buttons on hover with slide animation.

![Dictionary Card Hover](../assets/dictionary-card-hover.png)

**CRITICAL CSS - Hover Animation:**
```css
/* Drag handle - hidden by default, slides in on hover */
.dict-drag-handle {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 0;                          /* Hidden by default */
  height: 32px;
  cursor: grab;
  color: var(--text-muted);          /* #5a5a5a */
  overflow: hidden;
  transition: width 0.15s, margin 0.15s, color 0.15s;
  flex-shrink: 0;
  margin-right: 0;                   /* No margin when hidden */
}

.dict-card:hover .dict-drag-handle {
  width: 20px;                       /* Expand on hover */
  margin-right: 6px;                 /* Add spacing */
}

.dict-drag-handle:hover {
  color: var(--text-primary);        /* #e1e1e1 */
}

.dict-drag-handle:active {
  cursor: grabbing;
}

.dict-drag-handle svg {
  width: 14px;
  height: 14px;
}

/* Reorder buttons - same pattern */
.dict-reorder-btns {
  display: flex;
  flex-direction: column;
  gap: 2px;
  width: 0;                          /* Hidden by default */
  overflow: hidden;
  transition: width 0.15s, margin 0.15s;
  flex-shrink: 0;
  margin-right: 0;
}

.dict-card:hover .dict-reorder-btns {
  width: 18px;                       /* Expand on hover */
  margin-right: 6px;                 /* Add spacing */
}

.dict-reorder-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 14px;
  background: transparent;
  border: none;
  border-radius: 2px;
  color: var(--text-muted);          /* #5a5a5a */
  cursor: pointer;
  transition: all 0.1s;
  padding: 0;
}

.dict-reorder-btn:hover {
  background: var(--bg-hover);       /* #2a2a2a */
  color: var(--text-primary);        /* #e1e1e1 */
}

.dict-reorder-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.dict-reorder-btn svg {
  width: 10px;
  height: 10px;
}
```

**Fyne Implementation Note:**
```go
// To achieve this in Fyne, implement desktop.Hoverable interface
// and animate width using fyne.NewAnimation()

type DictCard struct {
    widget.BaseWidget
    hovered bool
    // ... other fields
}

func (d *DictCard) MouseIn(*desktop.MouseEvent) {
    d.hovered = true
    // Start animation to expand drag handle width
    anim := fyne.NewAnimation(150*time.Millisecond, func(progress float32) {
        // Interpolate width from 0 to 20
        d.dragHandleWidth = 20 * progress
        d.Refresh()
    })
    anim.Start()
}

func (d *DictCard) MouseOut() {
    d.hovered = false
    // Start animation to collapse
    anim := fyne.NewAnimation(150*time.Millisecond, func(progress float32) {
        d.dragHandleWidth = 20 * (1 - progress)
        d.Refresh()
    })
    anim.Start()
}
```

#### Scenario: Controls hidden when not hovered

- **GIVEN** dictionary card is in Active column
- **AND** mouse is not over the card
- **WHEN** user views the card
- **THEN** drag handle has width 0 and is not visible
- **AND** reorder buttons have width 0 and are not visible
- **AND** card text is aligned to left padding edge

#### Scenario: Controls animate in on hover

- **GIVEN** dictionary card is in Active column
- **WHEN** user hovers mouse over the card
- **THEN** drag handle width animates from 0 to 20px over 150ms
- **AND** reorder buttons width animate from 0 to 18px over 150ms
- **AND** margins animate from 0 to 6px to create spacing
- **AND** card text shifts right smoothly during animation

#### Scenario: Controls animate out on mouse leave

- **GIVEN** dictionary card is hovered with controls visible
- **WHEN** user moves mouse away from card
- **THEN** drag handle width animates from 20px to 0 over 150ms
- **AND** reorder buttons width animate from 18px to 0 over 150ms
- **AND** card text shifts left smoothly during animation

### Requirement: Drag and Drop Reordering

Active dictionaries SHALL be reorderable via drag and drop.

![Dictionary Drag](../assets/dictionary-drag.png)

**CSS for drag state:**
```css
.dict-card.dragging {
  opacity: 0.9;
  transform: scale(1.02);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.2);
  z-index: 100;
  border-color: var(--success);  /* #5db0a3 */
}

.dict-card.drag-ghost {
  opacity: 0.4;
}

/* Drop zone styling */
.dict-column-list.drag-over {
  background: var(--bg-hover);  /* #2a2a2a */
  border-radius: var(--radius-md);  /* 4px */
}

.dict-drop-placeholder {
  height: 60px;
  border: 2px dashed var(--success);  /* #5db0a3 */
  border-radius: var(--radius-md);  /* 4px */
  background: var(--success);  /* #5db0a3 */
  opacity: 0.1;
  transition: all 0.15s;
}
```

#### Scenario: Drag visual feedback

- **GIVEN** dictionary card is in Active column
- **WHEN** user starts dragging the card by drag handle
- **THEN** card opacity reduces slightly (0.9)
- **AND** card scale increases to 1.02
- **AND** shadow appears under card
- **AND** border changes to success color (#5db0a3)

#### Scenario: Drop to reorder

- **GIVEN** user is dragging dictionary "MW"
- **AND** drop indicator shows between "AP90" and "PWG"
- **WHEN** user releases the drag
- **THEN** "MW" is inserted between "AP90" and "PWG"
- **AND** new order is persisted to state database

### Requirement: Keyboard Reordering

Active dictionaries SHALL be reorderable via up/down buttons.

#### Scenario: Move dictionary up

- **GIVEN** dictionary "MW" is at position 3 in Active list
- **WHEN** user clicks up arrow button on "MW" card
- **THEN** "MW" moves to position 2
- **AND** new order is persisted to state database

#### Scenario: Move dictionary down

- **GIVEN** dictionary "MW" is at position 2 in Active list
- **WHEN** user clicks down arrow button on "MW" card
- **THEN** "MW" moves to position 3
- **AND** new order is persisted to state database

#### Scenario: Up button disabled for first item

- **GIVEN** dictionary "AP90" is first in Active list
- **WHEN** user views the card
- **THEN** up arrow button is disabled (visually muted, not clickable)

#### Scenario: Down button disabled for last item

- **GIVEN** dictionary "VCP" is last in Active list
- **WHEN** user views the card
- **THEN** down arrow button is disabled (visually muted, not clickable)

### Requirement: Order Persistence

Dictionary order SHALL be persisted and restored on app restart.

**Go State Functions:**
```go
// In pkg/state/state.go

// GetDictOrder returns the saved dictionary order
func GetDictOrder() ([]string, error) {
    // Return slice of dictionary codes in priority order
    // e.g., ["MW", "AP90", "PWG"]
}

// SetDictOrder saves the dictionary order
func SetDictOrder(codes []string) error {
    // Persist to state.db
}
```

#### Scenario: Order saved on change

- **GIVEN** Active dictionaries are in order: AP90, MW, PWG
- **WHEN** user reorders to: MW, AP90, PWG
- **THEN** new order is immediately saved to state database

#### Scenario: Order restored on startup

- **GIVEN** saved dictionary order is: MW, AP90, PWG
- **WHEN** application starts
- **THEN** Active column displays dictionaries in order: MW, AP90, PWG
