# Design: Desktop UI Redesign

## Prototype Reference

**Live**: https://sanskrit-upaya-prototype.vercel.app
**Source**: `prototype/index.html` - Contains all CSS and JS. Read this file for exact implementation details.

---

## Color System

### Dark Theme (Default)

```css
--bg-base: #1a1a1a;        /* Main background */
--bg-surface: #242424;      /* Cards, panels */
--bg-elevated: #2d2d2d;     /* Hover states, dropdowns */
--bg-input: #1a1a1a;        /* Input backgrounds */

--text-primary: #e5e5e5;    /* Primary text */
--text-secondary: #a3a3a3;  /* Secondary text */
--text-muted: #737373;      /* Muted/disabled text */

--border-subtle: #2d2d2d;   /* Subtle borders */
--border-default: #404040;  /* Default borders */

--success: #10b981;         /* PRIMARY ACCENT - teal/green */
--success-hover: #059669;   /* Accent hover state */
--error: #ef4444;           /* Delete, errors */
--warning: #f59e0b;         /* Warnings, highlights */
```

### Light Theme

```css
--bg-base: #ffffff;
--bg-surface: #f5f5f5;
--bg-elevated: #e5e5e5;
--bg-input: #ffffff;

--text-primary: #171717;
--text-secondary: #525252;
--text-muted: #a3a3a3;

--border-subtle: #e5e5e5;
--border-default: #d4d4d4;

/* Accent colors same as dark theme */
```

### Fyne Color Mapping

```go
// Convert to Fyne colors
var (
    BgBase      = color.NRGBA{R: 26, G: 26, B: 26, A: 255}   // #1a1a1a
    BgSurface   = color.NRGBA{R: 36, G: 36, B: 36, A: 255}   // #242424
    BgElevated  = color.NRGBA{R: 45, G: 45, B: 45, A: 255}   // #2d2d2d

    TextPrimary   = color.NRGBA{R: 229, G: 229, B: 229, A: 255} // #e5e5e5
    TextSecondary = color.NRGBA{R: 163, G: 163, B: 163, A: 255} // #a3a3a3
    TextMuted     = color.NRGBA{R: 115, G: 115, B: 115, A: 255} // #737373

    Success     = color.NRGBA{R: 16, G: 185, B: 129, A: 255}  // #10b981
    SuccessHover = color.NRGBA{R: 5, G: 150, B: 105, A: 255}  // #059669
    Error       = color.NRGBA{R: 239, G: 68, B: 68, A: 255}   // #ef4444
    Warning     = color.NRGBA{R: 245, G: 158, B: 11, A: 255}  // #f59e0b
)
```

---

## Typography

```css
--font-sans: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
--font-mono: "SF Mono", Monaco, "Cascadia Code", monospace;

/* Sizes */
--text-xs: 10px;    /* Labels, badges */
--text-sm: 12px;    /* Secondary info */
--text-base: 14px;  /* Body text */
--text-lg: 16px;    /* Headings */
--text-xl: 20px;    /* Page titles */
```

---

## Spacing & Radius

```css
--radius-sm: 4px;   /* Buttons, inputs */
--radius-md: 6px;   /* Cards */
--radius-lg: 8px;   /* Modals, panels */

/* Spacing scale */
4px, 6px, 8px, 10px, 12px, 14px, 16px, 20px, 24px
```

---

## Animation Timing

**IMPORTANT**: All animations use 150ms duration for consistency.

```css
/* Standard transition */
transition: all 0.15s;

/* For width/margin animations (hover reveal) */
transition: width 0.15s, margin 0.15s;

/* Easing - use default ease or cubic-bezier for snappier feel */
transition: all 0.2s cubic-bezier(0.2, 0, 0, 1);
```

### Fyne Animation Equivalent

```go
// 150ms = 0.15 seconds
anim := fyne.NewAnimation(150*time.Millisecond, func(progress float32) {
    // Animate property
})
anim.Start()
```

---

## Navigation Bar

```
┌──────────────────────────────────────────────────────────────────┐
│ 🔍 Search  ☆ Starred  ⌨ Transliterate  📚 Dictionaries  [☀/🌙] ⚙│
└──────────────────────────────────────────────────────────────────┘
```

### Styling

```css
.nav-bar {
    height: 40px;
    background: var(--bg-base);
    border-bottom: 1px solid var(--border-subtle);
    display: flex;
    align-items: center;
    padding: 0 8px;
}

.nav-tab {
    padding: 8px 12px;
    color: var(--text-secondary);
    font-size: 13px;
    border-radius: var(--radius-sm);
    transition: all 0.15s;
}

.nav-tab:hover {
    color: var(--text-primary);
    background: var(--bg-elevated);
}

.nav-tab.active {
    color: var(--text-primary);
    background: var(--bg-surface);
}
```

---

## Dictionaries Page

### Two-Column Layout

```
┌─────────────────────────────┬─────────────────────────────┐
│ ACTIVE (4)                  │ AVAILABLE (32)              │
├─────────────────────────────┼─────────────────────────────┤
│ ┌─────────────────────────┐ │ ┌─────────────────────────┐ │
│ │ AP90 Apte Practical  [×]│ │ │ VCP Vacaspatyam     [+] │ │
│ │ 60k entries · sa→en     │ │ │ 96k entries · sa→sa     │ │
│ └─────────────────────────┘ │ └─────────────────────────┘ │
└─────────────────────────────┴─────────────────────────────┘
```

### Card Hover Animation (CRITICAL)

The drag handle and reorder buttons are **hidden by default** and **slide in on hover**.

**Default state (not hovered):**
- `width: 0` and `overflow: hidden` on drag handle and reorder buttons
- `margin-right: 0` on these elements
- Text aligned to left edge of card

**Hovered state:**
- Drag handle: `width: 20px`, `margin-right: 6px`
- Reorder buttons: `width: 18px`, `margin-right: 6px`
- Text shifts right smoothly

```css
/* Drag handle - hidden by default */
.dict-drag-handle {
    width: 0;
    overflow: hidden;
    transition: width 0.15s, margin 0.15s;
    margin-right: 0;
    flex-shrink: 0;
}

.dict-card:hover .dict-drag-handle {
    width: 20px;
    margin-right: 6px;
}

/* Reorder buttons - same pattern */
.dict-reorder-btns {
    width: 0;
    overflow: hidden;
    transition: width 0.15s, margin 0.15s;
    margin-right: 0;
    flex-shrink: 0;
}

.dict-card:hover .dict-reorder-btns {
    width: 18px;
    margin-right: 6px;
}
```

### Dictionary Card

```css
.dict-card {
    display: flex;
    align-items: center;
    padding: 10px 12px;
    background: var(--bg-base);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    cursor: pointer;
    transition: all 0.2s cubic-bezier(0.2, 0, 0, 1);
}

.dict-card:hover {
    border-color: var(--border-default);
}

/* Dictionary code badge */
.dict-code {
    font-family: var(--font-mono);
    font-size: 11px;
    padding: 2px 6px;
    background: var(--bg-elevated);
    border-radius: 3px;
    color: var(--text-secondary);
    margin-right: 8px;
}
```

---

## Search Page - Compare View

### View Toggle Buttons

```css
.view-toggle-btn {
    width: 32px;
    height: 28px;
    border: 1px solid var(--border-subtle);
    background: transparent;
    color: var(--text-muted);
    transition: all 0.15s;
}

.view-toggle-btn:hover {
    border-color: var(--border-default);
    color: var(--text-secondary);
}

.view-toggle-btn.active {
    background: var(--bg-elevated);
    border-color: var(--success);
    color: var(--success);
}
```

### Compare View Grid

Dynamic 2-4 columns with optional "Add" column.

```css
.compare-view {
    display: grid;
    gap: 1px;
    background: var(--border-subtle);
    height: 100%;
    grid-template-columns: repeat(var(--compare-cols, 2), 1fr);
}

/* When add button is shown, it's a narrow 60px column */
.compare-view.has-add-btn {
    grid-template-columns: repeat(var(--compare-cols, 2), 1fr) 60px;
}
```

### Compare Column Header

```css
.compare-column-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 14px;
    background: var(--bg-base);
    border-bottom: 1px solid var(--border-subtle);
}

/* Simplified dropdown - just text + chevron */
.compare-select {
    padding: 4px 0;
    padding-right: 18px;
    background: transparent;
    border: none;
    color: var(--text-primary);
    font-size: 13px;
    font-weight: 500;
    cursor: pointer;
    appearance: none;
    /* Chevron icon via background-image */
}
```

### Add Column Button

```css
.compare-add-column {
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--bg-surface);
    border-left: 1px dashed var(--border-subtle);
    cursor: pointer;
    transition: all 0.15s;
}

.compare-add-column:hover {
    background: var(--bg-elevated);
}

.compare-add-column:hover .compare-add-icon {
    color: var(--success);
}
```

### Remove Column Button (appears on columns 3+)

```css
.compare-remove-btn {
    width: 22px;
    height: 22px;
    background: transparent;
    color: var(--text-muted);
    opacity: 0.6;
    transition: all 0.15s;
}

.compare-remove-btn:hover {
    background: var(--bg-elevated);
    color: var(--error);
    opacity: 1;
}
```

---

## Dictionary Entry Card (in article view)

```css
.dict-card-entry {
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    overflow: hidden;
}

/* Dark header */
.dict-card-entry-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 14px;
    background: var(--bg-base);
    border-bottom: 1px solid var(--border-default);
}

/* Dictionary code in header - outline style */
.dict-card-entry-title .code {
    font-family: var(--font-mono);
    font-size: 11px;
    padding: 2px 6px;
    background: transparent;
    border: 1px solid var(--border-default);
    border-radius: 3px;
    color: var(--text-secondary);
}

.dict-card-entry-content {
    padding: 14px;
    font-size: 14px;
    line-height: 1.6;
}
```

---

## Search Input

```css
.search-box {
    position: relative;
    display: flex;
    align-items: center;
}

.search-input {
    width: 100%;
    padding: 10px 12px;
    padding-left: 36px;  /* Space for icon */
    padding-right: 32px; /* Space for clear button */
    background: var(--bg-input);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    color: var(--text-primary);
    font-size: 14px;
    transition: all 0.15s;
}

.search-input:focus {
    outline: none;
    border-color: var(--success);
    box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.1);
}

/* Clear button - appears when input has value */
.search-clear {
    position: absolute;
    right: 8px;
    width: 20px;
    height: 20px;
    background: var(--bg-elevated);
    border-radius: 50%;
    color: var(--text-muted);
    opacity: 0; /* Hidden by default */
    transition: opacity 0.15s;
}

.search-box.has-value .search-clear {
    opacity: 1;
}
```

---

## Button Styles

### Icon Button (Copy, Star, etc.)

```css
.dict-card-entry-btn {
    width: 28px;
    height: 28px;
    background: transparent;
    border: none;
    color: var(--text-muted);
    border-radius: var(--radius-sm);
    cursor: pointer;
    transition: all 0.15s;
}

.dict-card-entry-btn:hover {
    background: var(--bg-elevated);
    color: var(--text-primary);
}

.dict-card-entry-btn svg {
    width: 14px;
    height: 14px;
}
```

---

## Key Interactions

### Escape Key Clears Search
```javascript
// When Escape pressed and search input focused
searchInput.value = '';
// Trigger search update
```

### Click Card to Toggle (Dictionaries)
```javascript
// Clicking anywhere on dictionary card toggles active/inactive
// Except on action buttons
```

### Drag and Drop (Active Dictionaries)
- Only active dictionaries are draggable
- Show drop indicator during drag
- Card gets slight scale (1.02) and shadow when dragging

---

## Implementation Notes for Fyne

1. **Hover states**: Use `widget.NewButton` with custom `MouseIn`/`MouseOut` handlers or implement `desktop.Hoverable` interface

2. **Animations**: Fyne has `fyne.NewAnimation()` - use 150ms duration to match prototype

3. **Custom widgets**: May need custom widgets for:
   - Dictionary card with hover reveal
   - Compare view grid
   - Simplified dropdown

4. **Theme switching**: Use `app.Settings().SetTheme()` with custom theme

5. **Grid layout**: Use `container.NewGridWithColumns()` for compare view

---

## File References

For exact implementation details, read these sections of `prototype/index.html`:

| Feature | Search in prototype for |
|---------|------------------------|
| Color variables | `:root {` |
| Navigation | `.nav-bar`, `.nav-tab` |
| Dictionary cards | `.dict-card`, `.dict-drag-handle` |
| Compare view | `.compare-view`, `.compare-column` |
| Search input | `.search-box`, `.search-input` |
| Buttons | `.dict-card-entry-btn` |
| Animations | `transition:` |
