## ADDED Requirements

### Requirement: Color Theme System

The application SHALL use a consistent color theme with teal (#5db0a3) as the primary accent color (success) in dark mode and green (#059669) in light mode.

![Theme Colors](../assets/theme-colors.png)

**CSS Variables - Root (Fonts & Radius):**
```css
:root {
  /* Fonts */
  --font-sans: 'IBM Plex Sans', -apple-system, sans-serif;
  --font-mono: 'JetBrains Mono', monospace;
  --font-sanskrit: 'Noto Sans Devanagari', sans-serif;

  /* Spacing & Radius */
  --radius-sm: 3px;
  --radius-md: 4px;

  /* Theme transition */
  --transition-theme: 0.2s ease;
}
```

**CSS Variables (Dark Theme - Default):**
```css
[data-theme="dark"] {
  /* Backgrounds */
  --bg-base: #181818;
  --bg-surface: #1f1f1f;
  --bg-elevated: #262626;
  --bg-hover: #2a2a2a;
  --bg-active: #323232;
  --bg-selection: #264f78;

  /* Text */
  --text-primary: #e1e1e1;
  --text-secondary: #a0a0a0;
  --text-muted: #5a5a5a;
  --text-inverse: #181818;

  /* Borders */
  --border-subtle: #252525;
  --border-default: #2e2e2e;

  /* Accent Colors */
  --accent: #3b8eea;
  --accent-muted: #3b8eea20;
  --accent-text: #6cb6ff;
  --success: #5db0a3;        /* PRIMARY ACCENT - teal */
  --warning: #d19a66;
}
```

**CSS Variables (Light Theme):**
```css
[data-theme="light"] {
  /* Backgrounds */
  --bg-base: #ffffff;
  --bg-surface: #f5f5f5;
  --bg-elevated: #e8e8e8;
  --bg-hover: #ebebeb;
  --bg-active: #dedede;
  --bg-selection: #b3d4fc;

  /* Text */
  --text-primary: #1a1a1a;
  --text-secondary: #555555;
  --text-muted: #999999;
  --text-inverse: #ffffff;

  /* Borders */
  --border-subtle: #e0e0e0;
  --border-default: #d0d0d0;

  /* Accent Colors */
  --accent: #2563eb;
  --accent-muted: #2563eb15;
  --accent-text: #1d4ed8;
  --success: #059669;        /* PRIMARY ACCENT - green */
  --warning: #d97706;
}
```

**Fyne Color Constants (Dark Theme):**
```go
var (
    // Backgrounds
    BgBase     = color.NRGBA{R: 24, G: 24, B: 24, A: 255}   // #181818
    BgSurface  = color.NRGBA{R: 31, G: 31, B: 31, A: 255}   // #1f1f1f
    BgElevated = color.NRGBA{R: 38, G: 38, B: 38, A: 255}   // #262626
    BgHover    = color.NRGBA{R: 42, G: 42, B: 42, A: 255}   // #2a2a2a
    BgActive   = color.NRGBA{R: 50, G: 50, B: 50, A: 255}   // #323232

    // Text
    TextPrimary   = color.NRGBA{R: 225, G: 225, B: 225, A: 255} // #e1e1e1
    TextSecondary = color.NRGBA{R: 160, G: 160, B: 160, A: 255} // #a0a0a0
    TextMuted     = color.NRGBA{R: 90, G: 90, B: 90, A: 255}    // #5a5a5a

    // Borders
    BorderSubtle  = color.NRGBA{R: 37, G: 37, B: 37, A: 255}    // #252525
    BorderDefault = color.NRGBA{R: 46, G: 46, B: 46, A: 255}    // #2e2e2e

    // Accent
    Success = color.NRGBA{R: 93, G: 176, B: 163, A: 255}   // #5db0a3
    Warning = color.NRGBA{R: 209, G: 154, B: 102, A: 255}  // #d19a66
)
```

#### Scenario: Dark theme colors applied

- **GIVEN** the application is running
- **AND** dark theme is active (default)
- **WHEN** user views any page
- **THEN** main background uses #181818 (bg-base)
- **AND** card/panel backgrounds use #1f1f1f (bg-surface)
- **AND** hover states use #2a2a2a (bg-hover)
- **AND** active states use #323232 (bg-active)
- **AND** primary text uses #e1e1e1
- **AND** secondary text uses #a0a0a0
- **AND** muted text uses #5a5a5a
- **AND** accent color is #5db0a3 (teal)

#### Scenario: Light theme colors applied

- **GIVEN** the application is running
- **AND** light theme is active
- **WHEN** user views any page
- **THEN** main background uses #ffffff (bg-base)
- **AND** card/panel backgrounds use #f5f5f5 (bg-surface)
- **AND** hover states use #ebebeb (bg-hover)
- **AND** active states use #dedede (bg-active)
- **AND** primary text uses #1a1a1a
- **AND** secondary text uses #555555
- **AND** muted text uses #999999
- **AND** accent color is #059669 (green)

### Requirement: Typography

The application SHALL use IBM Plex Sans as the primary font and JetBrains Mono for code/monospace text.

**Font Loading:**
```html
<link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500&family=IBM+Plex+Sans:wght@400;500;600&family=Noto+Sans+Devanagari:wght@400;500&display=swap" rel="stylesheet">
```

**Base Font Styles:**
```css
body {
  font-family: var(--font-sans);
  font-size: 13px;
  line-height: 1.5;
  -webkit-font-smoothing: antialiased;
}
```

#### Scenario: Font family applied

- **GIVEN** the application is running
- **WHEN** user views any page
- **THEN** body text uses IBM Plex Sans font
- **AND** code/badges use JetBrains Mono font
- **AND** Sanskrit/Devanagari text uses Noto Sans Devanagari font

### Requirement: Theme Switching

The application SHALL allow users to switch between dark and light themes.

#### Scenario: Toggle from dark to light

- **GIVEN** dark theme is active
- **WHEN** user clicks theme toggle button in navigation bar
- **THEN** all colors switch to light theme values with 0.2s transition
- **AND** theme preference is saved to state database

#### Scenario: Toggle from light to dark

- **GIVEN** light theme is active
- **WHEN** user clicks theme toggle button in navigation bar
- **THEN** all colors switch to dark theme values with 0.2s transition
- **AND** theme preference is saved to state database

#### Scenario: Theme persists across restart

- **GIVEN** user selected light theme
- **AND** application was closed
- **WHEN** application starts again
- **THEN** light theme is applied from saved preference

### Requirement: Consistent Transitions

Interactive elements SHALL use 100ms transition for hover states and 150ms for other state changes.

**CSS Pattern:**
```css
/* Standard transition for interactive elements */
transition: all 0.1s;

/* For theme transitions */
transition: background-color var(--transition-theme),
            border-color var(--transition-theme),
            color var(--transition-theme);
```

**Fyne Animation:**
```go
// 100ms animation for hover
anim := fyne.NewAnimation(100*time.Millisecond, func(progress float32) {
    // Interpolate property based on progress (0.0 to 1.0)
})
anim.Start()
```

#### Scenario: Button hover transition

- **GIVEN** a button is displayed
- **WHEN** user hovers over the button
- **THEN** background color changes with 100ms ease transition
- **AND** the transition is smooth without jumping

#### Scenario: Card hover transition

- **GIVEN** a card is displayed
- **WHEN** user hovers over the card
- **THEN** border color changes with 100ms ease transition

### Requirement: Focus States

All focusable elements SHALL show accent color border on focus.

![Focus States](../assets/focus-states.png)

**CSS Pattern:**
```css
.search-input:focus {
  border-color: var(--success);  /* #5db0a3 in dark, #059669 in light */
  background: var(--bg-elevated);
  outline: none;
}
```

#### Scenario: Input focus state

- **GIVEN** an input field is displayed
- **WHEN** user focuses the input (click or tab)
- **THEN** border color changes to success color (#5db0a3 dark / #059669 light)
- **AND** background changes to bg-elevated
- **AND** outline is removed (using outline: none)

#### Scenario: Button focus state

- **GIVEN** a button is displayed
- **WHEN** user focuses the button via keyboard
- **THEN** button shows accent border or background change
