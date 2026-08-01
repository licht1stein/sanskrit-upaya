## MODIFIED Requirements

### Requirement: Starred Page Styling

The starred page SHALL use the new theme colors and design patterns.

![Starred Page](../assets/starred-page.png)

#### Scenario: Theme colors applied

- **GIVEN** the application is running
- **WHEN** user navigates to starred page
- **THEN** background uses theme bg-base color
- **AND** text uses theme text-primary and text-secondary colors
- **AND** borders use theme border-subtle color
- **AND** accent elements use teal/green (#10b981)

### Requirement: Starred Items List

Starred items SHALL display in a list matching the search results styling.

#### Scenario: Starred items displayed

- **GIVEN** user has starred 3 dictionary entries
- **WHEN** user views starred page
- **THEN** 3 items display in a vertical list
- **AND** each item shows word in primary text
- **AND** each item shows devanagari in secondary text
- **AND** each item shows dictionary code badge

#### Scenario: Starred item hover

- **GIVEN** starred items are displayed
- **WHEN** user hovers over an item
- **THEN** item background changes to bg-elevated
- **AND** transition takes 150ms

#### Scenario: Starred item click

- **GIVEN** starred items are displayed
- **WHEN** user clicks on an item
- **THEN** full article content displays
- **AND** user can view the complete dictionary entry

### Requirement: Empty State

The starred page SHALL show appropriate message when no items are starred.

![Starred Empty State](../assets/starred-empty.png)

#### Scenario: Empty state displayed

- **GIVEN** user has no starred items
- **WHEN** user views starred page
- **THEN** empty state message displays centered
- **AND** message uses muted text color
- **AND** message suggests how to star items

### Requirement: Remove from Starred

Users SHALL be able to remove items from starred list.

#### Scenario: Unstar item

- **GIVEN** starred items list shows 3 items
- **WHEN** user clicks unstar/remove button on an item
- **THEN** item is removed from starred list
- **AND** list updates immediately to show 2 items
- **AND** no page reload required
