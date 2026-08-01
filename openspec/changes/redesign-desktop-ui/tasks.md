## Reference

**Prototype**: https://sanskrit-upaya-prototype.vercel.app
**Source**: `prototype/index.html` - Read this file for exact CSS values and JS logic

---

## 1. Theme System

- [ ] 1.1 Create new theme file `cmd/desktop/theme.go` with color constants from design.md
- [ ] 1.2 Implement dark theme colors (default)
- [ ] 1.3 Implement light theme colors
- [ ] 1.4 Change primary accent from blue to teal/green (#10b981)
- [ ] 1.5 Add theme toggle functionality (persist to state.db)
- [ ] 1.6 Apply 150ms transition timing to all interactive elements

## 2. Navigation Bar

- [ ] 2.1 Create horizontal navigation bar at top of window
- [ ] 2.2 Add tabs: Search, Starred, Transliterate, Dictionaries
- [ ] 2.3 Add icons to each tab (search, star, keyboard, book icons)
- [ ] 2.4 Implement active tab highlighting with teal accent
- [ ] 2.5 Add hover states (background: bg-elevated)
- [ ] 2.6 Add theme toggle button (sun/moon icon) on right side
- [ ] 2.7 Add settings gear icon on far right

## 3. Search Page - Layout

- [ ] 3.1 Create three-panel layout: sidebar (results) | divider | content (article)
- [ ] 3.2 Implement resizable divider between panels
- [ ] 3.3 Create search input with icon and clear button
- [ ] 3.4 Show clear button only when input has value
- [ ] 3.5 Add Escape key handler to clear search input
- [ ] 3.6 Add search mode tabs: PREFIX, EXACT, FUZZY, REVERSE
- [ ] 3.7 Style active search mode with teal accent

## 4. Search Page - Results List

- [ ] 4.1 Create results list with word + devanagari display
- [ ] 4.2 Show dictionary code badge below each result
- [ ] 4.3 Add result count header ("RESULTS 42")
- [ ] 4.4 Implement keyboard navigation (arrow keys)
- [ ] 4.5 Highlight selected result
- [ ] 4.6 Add hover states on results

## 5. Search Page - Article View (Stacked Mode)

- [ ] 5.1 Create article header with word, devanagari, dictionary count badge
- [ ] 5.2 Create view toggle buttons (stacked/compare icons)
- [ ] 5.3 Implement stacked view: vertical list of all dictionary entries
- [ ] 5.4 Create dictionary entry card component with dark header
- [ ] 5.5 Show dictionary code (outline badge), name in header
- [ ] 5.6 Add Copy button to each entry header
- [ ] 5.7 Add Star button to each entry header
- [ ] 5.8 Style entry content with proper typography

## 6. Search Page - Article View (Compare Mode)

- [ ] 6.1 Implement compare view grid (2 columns default)
- [ ] 6.2 Add simplified dropdown selector (just name + chevron, no border)
- [ ] 6.3 Add Copy and Star buttons to each column header
- [ ] 6.4 Implement "Add column" button (+ icon, 60px width, dashed border)
- [ ] 6.5 Implement "Remove column" button (× icon, appears on columns 3+)
- [ ] 6.6 Support 2-4 columns maximum
- [ ] 6.7 Always maintain minimum 2 columns
- [ ] 6.8 Auto-select next unused dictionary when adding column
- [ ] 6.9 Add column entry animation (scale from 0.95, fade in)

## 7. Dictionaries Page - Layout

- [ ] 7.1 Create two-column layout: Active | Available
- [ ] 7.2 Add column headers with counts ("ACTIVE 4", "AVAILABLE 32")
- [ ] 7.3 Add filter input at top ("Filter dictionaries...")
- [ ] 7.4 Add stats display ("4 of 36 active · 245k entries")
- [ ] 7.5 Implement filter functionality (match code and name)

## 8. Dictionaries Page - Cards

- [ ] 8.1 Create dictionary card component
- [ ] 8.2 Show: code badge, name, entry count, language direction
- [ ] 8.3 Add action button: × for active cards, + for available cards
- [ ] 8.4 Implement click anywhere on card to toggle state
- [ ] 8.5 Ensure consistent padding between active and available cards

## 9. Dictionaries Page - Hover Animation (CRITICAL)

- [ ] 9.1 Add drag handle element (6-dot grip icon) to active cards
- [ ] 9.2 Add up/down reorder buttons to active cards
- [ ] 9.3 Hide drag handle and buttons by default (width: 0)
- [ ] 9.4 On hover: animate width from 0 to 20px/18px over 150ms
- [ ] 9.5 On hover: animate margin-right from 0 to 6px
- [ ] 9.6 Ensure text shifts smoothly to the right on hover
- [ ] 9.7 If Fyne animation is difficult, fallback to instant show/hide

## 10. Dictionaries Page - Reordering

- [ ] 10.1 Implement drag-and-drop for active dictionary cards
- [ ] 10.2 Show visual feedback during drag (opacity, shadow)
- [ ] 10.3 Show drop indicator line between cards
- [ ] 10.4 Implement up button: move card one position up
- [ ] 10.5 Implement down button: move card one position down
- [ ] 10.6 Disable up on first item, down on last item
- [ ] 10.7 Persist dictionary order to state.db

## 11. Starred Page

- [ ] 11.1 Apply new theme colors and typography
- [ ] 11.2 Style starred items list to match search results
- [ ] 11.3 Add hover states with 150ms transitions
- [ ] 11.4 Style empty state message

## 12. Transliterate Page

- [ ] 12.1 Apply new theme colors and typography
- [ ] 12.2 Style input and output text areas
- [ ] 12.3 Style direction toggle buttons
- [ ] 12.4 Add hover states with 150ms transitions

## 13. State Persistence

- [ ] 13.1 Add `GetDictOrder() []string` to pkg/state
- [ ] 13.2 Add `SetDictOrder([]string) error` to pkg/state
- [ ] 13.3 Add `GetTheme() string` to pkg/state
- [ ] 13.4 Add `SetTheme(string) error` to pkg/state
- [ ] 13.5 Load preferences on app startup

## 14. Testing

- [ ] 14.1 Test theme switching (dark/light)
- [ ] 14.2 Test navigation between all pages
- [ ] 14.3 Test search with all modes
- [ ] 14.4 Test stacked view display
- [ ] 14.5 Test compare view: add/remove columns
- [ ] 14.6 Test dictionary card hover animations
- [ ] 14.7 Test drag-and-drop reordering
- [ ] 14.8 Test keyboard reordering (up/down buttons)
- [ ] 14.9 Test filter functionality
- [ ] 14.10 Test state persistence across app restarts
- [ ] 14.11 Compare implementation against prototype visually
