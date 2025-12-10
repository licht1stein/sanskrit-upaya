# Change: Add Dictionary Ordering

## Why

Users need to control the display order of dictionaries in grouped search results. Currently, dictionaries are sorted alphabetically within language groups, but users may prefer certain dictionaries (like Monier-Williams) to appear first.

## What Changes

- Add drag-and-drop or up/down buttons in the "Dictionaries..." dialog to reorder dictionaries
- Persist custom dictionary order in user settings
- Use custom order when displaying tabs in grouped results view
- Use custom order when grouping entries within a word result

## Impact

- Affected code: `cmd/desktop/main.go` (dictionary dialog, tab ordering), `pkg/state/state.go` (order persistence)
- No breaking changes - existing behavior preserved if user hasn't customized order
