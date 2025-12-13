## ADDED Requirements

### Requirement: Horizontal Navigation Bar

The application SHALL display a horizontal navigation bar at the top of the window.

![Navigation Bar](../assets/navigation-bar.png)

**CSS - Navigation Bar:**
```css
.top-nav {
  display: flex;
  align-items: center;
  background: var(--bg-base);  /* #181818 */
  border-bottom: 1px solid var(--border-subtle);  /* #252525 */
  padding: 4px 6px;
  gap: 2px;
}
```

**CSS - Navigation Tabs:**
```css
.nav-tab {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 6px 10px;
  background: transparent;
  border: none;
  border-radius: var(--radius-sm);  /* 3px */
  color: var(--text-muted);  /* #5a5a5a */
  font-family: var(--font-sans);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.1s;
}

.nav-tab:hover {
  color: var(--text-secondary);  /* #a0a0a0 */
}

.nav-tab.active {
  color: var(--text-primary);  /* #e1e1e1 */
  background: var(--bg-elevated);  /* #262626 */
}

.nav-tab svg {
  width: 14px;
  height: 14px;
}
```

**CSS - Theme Toggle:**
```css
.theme-toggle {
  display: flex;
  align-items: center;
  gap: 2px;
  padding: 3px;
  background: var(--bg-surface);  /* #1f1f1f */
  border: 1px solid var(--border-subtle);  /* #252525 */
  border-radius: 6px;
  margin-right: 4px;
}

.theme-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  background: transparent;
  border: none;
  border-radius: 4px;
  color: var(--text-muted);  /* #5a5a5a */
  cursor: pointer;
  transition: all 0.15s;
}

.theme-btn:hover {
  color: var(--text-secondary);  /* #a0a0a0 */
}

.theme-btn.active {
  background: var(--bg-elevated);  /* #262626 */
  color: var(--text-primary);  /* #e1e1e1 */
}

.theme-btn svg {
  width: 14px;
  height: 14px;
}
```

#### Scenario: Navigation bar layout

- **GIVEN** the application is running
- **WHEN** user views the window
- **THEN** navigation bar appears at top of window
- **AND** navigation bar has padding 4px 6px
- **AND** background color is bg-base (#181818)
- **AND** bottom border uses border-subtle color (#252525, 1px)

### Requirement: Navigation Tabs

The navigation bar SHALL contain tabs for each main section: Search, Starred, Transliterate, Dictionaries, Settings.

#### Scenario: Tab icons and labels displayed

- **GIVEN** navigation bar is visible
- **WHEN** user views the tabs
- **THEN** Search tab shows magnifying glass icon with "Search" label
- **AND** Starred tab shows star icon with "Starred" label
- **AND** Transliterate tab shows keyboard icon with "Transliterate" label
- **AND** Dictionaries tab shows book icon with "Dictionaries" label
- **AND** Settings tab shows gear icon with "Settings" label

#### Scenario: Active tab styling

- **GIVEN** user is on Search page
- **WHEN** user views navigation bar
- **THEN** Search tab shows primary text color (#e1e1e1)
- **AND** Search tab has bg-elevated background (#262626)
- **AND** other tabs show muted text color (#5a5a5a)
- **AND** other tabs have transparent background

#### Scenario: Tab hover state

- **GIVEN** Starred tab is inactive
- **WHEN** user hovers over Starred tab
- **THEN** text color changes to secondary (#a0a0a0)
- **AND** transition takes 100ms

#### Scenario: Tab click navigation

- **GIVEN** user is on Search page
- **WHEN** user clicks Dictionaries tab
- **THEN** Dictionaries page content displays
- **AND** Dictionaries tab becomes active (bg-elevated, primary text)
- **AND** Search tab becomes inactive (transparent, muted text)

### Requirement: Theme Toggle Button

The navigation bar SHALL include a theme toggle with sun/moon icons on the right side.

![Theme Toggle](../assets/theme-toggle.png)

#### Scenario: Dark theme icon displayed

- **GIVEN** dark theme is active
- **WHEN** user views theme toggle
- **THEN** moon icon button is active (highlighted)
- **AND** sun icon button is inactive

#### Scenario: Light theme icon displayed

- **GIVEN** light theme is active
- **WHEN** user views theme toggle
- **THEN** sun icon button is active (highlighted)
- **AND** moon icon button is inactive

#### Scenario: Theme toggle click

- **GIVEN** dark theme is active
- **WHEN** user clicks sun icon button
- **THEN** theme switches to light
- **AND** sun button becomes active
- **AND** moon button becomes inactive
