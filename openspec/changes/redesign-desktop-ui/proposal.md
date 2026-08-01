# Change: Complete Desktop UI Redesign

## Why

The current Fyne UI is functional but lacks polish and modern UX patterns. Users need a more intuitive interface for managing dictionaries, comparing translations across multiple sources, and navigating search results. This redesign brings a VS Code/Cursor-inspired minimal aesthetic with smooth animations, better information hierarchy, and powerful new features like side-by-side dictionary comparison.

## What Changes

### Global UI
- **BREAKING** Complete visual redesign with new color theme (teal/green accent)
- **BREAKING** New navigation bar with icon+label tabs
- **NEW** Dark/light theme switching
- **NEW** Consistent hover states and transitions (150ms default)

### Search Page
- **NEW** Dictionary comparison view (2-4 columns side-by-side)
- **NEW** Stacked view mode (all dictionaries vertical)
- **NEW** View mode toggle button
- **NEW** Per-dictionary Copy and Star buttons
- **MODIFIED** Search results list styling
- **MODIFIED** Article content display with dark headers

### Dictionaries Page
- **BREAKING** Complete redesign with two-column layout (Active | Available)
- **NEW** Drag-and-drop reordering for active dictionaries
- **NEW** Up/down arrow buttons for keyboard reordering
- **NEW** Hover animation: controls slide in from left
- **NEW** Filter input to search dictionaries

### Starred Page
- **MODIFIED** Updated styling to match new design system

### Transliterate Page
- **MODIFIED** Updated styling to match new design system

## Impact

- Affected specs: All UI-related specs (new)
- Affected code:
  - `cmd/desktop/main.go` - Major UI restructure
  - `cmd/desktop/theme.go` - New theme system
  - `cmd/desktop/search.go` - New comparison feature (may need new file)
  - `cmd/desktop/dictionaries.go` - New page implementation (may need new file)
  - `pkg/state/state.go` - Dictionary order persistence

## Prototype Reference

**Live prototype**: https://sanskrit-upaya-prototype.vercel.app
**Source code**: `prototype/index.html` in repository root

The prototype is the **source of truth** for all visual design, animations, and interactions. When implementing, reference the prototype's CSS and JavaScript for exact values.
