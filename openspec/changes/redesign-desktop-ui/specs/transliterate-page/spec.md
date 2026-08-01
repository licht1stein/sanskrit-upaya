## MODIFIED Requirements

### Requirement: Transliterate Page Styling

The transliterate page SHALL use the new theme colors and design patterns.

![Transliterate Page](../assets/transliterate-page.png)

#### Scenario: Theme colors applied

- **GIVEN** the application is running
- **WHEN** user navigates to transliterate page
- **THEN** background uses theme bg-base color
- **AND** text areas use theme bg-input background
- **AND** borders use theme border-subtle color
- **AND** accent elements use teal/green (#10b981)

### Requirement: Input Text Area

The input text area SHALL allow users to enter text for transliteration.

#### Scenario: Input area displayed

- **GIVEN** transliterate page is displayed
- **WHEN** user views input area
- **THEN** text area has bg-input background
- **AND** border uses border-subtle color
- **AND** placeholder text guides user

#### Scenario: Input focus state

- **GIVEN** input text area is displayed
- **WHEN** user focuses the input
- **THEN** border color changes to teal accent (#10b981)
- **AND** transition takes 150ms

### Requirement: Output Text Area

The output text area SHALL display transliterated text.

#### Scenario: Output area displayed

- **GIVEN** user enters text in input area
- **WHEN** transliteration completes
- **THEN** output area shows transliterated text
- **AND** output styling matches input styling

### Requirement: Direction Toggle

Users SHALL be able to toggle transliteration direction.

![Transliterate Toggle](../assets/transliterate-toggle.png)

#### Scenario: Direction toggle displayed

- **GIVEN** transliterate page is displayed
- **WHEN** user views direction controls
- **THEN** toggle buttons show available directions
- **AND** active direction has teal accent background
- **AND** inactive direction has muted styling

#### Scenario: Direction toggle click

- **GIVEN** IAST to Devanagari is active
- **WHEN** user clicks Devanagari to IAST option
- **THEN** direction switches
- **AND** new direction shows teal accent
- **AND** output updates with new direction
- **AND** transition takes 150ms
