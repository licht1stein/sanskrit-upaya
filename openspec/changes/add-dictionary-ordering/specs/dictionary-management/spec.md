## ADDED Requirements

### Requirement: Dictionary Display Order

Users SHALL be able to customize the display order of dictionaries. The custom order affects:

1. The order of dictionary tabs in grouped search results
2. The order of dictionary entries within the "All" tab
3. The order of dictionaries shown in the Active section of the selection dialog

#### Scenario: User reorders dictionaries

- **WHEN** user opens the "Dictionaries..." dialog
- **AND** uses up/down controls to move a dictionary within the Active section
- **AND** clicks Apply
- **THEN** the new order is persisted
- **AND** subsequent searches display dictionaries in the custom order

#### Scenario: Order persists across sessions

- **WHEN** user has customized dictionary order
- **AND** restarts the application
- **THEN** the custom order is preserved

#### Scenario: Reset to default order

- **WHEN** user clicks "Reset Order" in the dictionary dialog
- **THEN** all dictionaries become active
- **AND** dictionaries are sorted alphabetically within language groups
- **AND** the default order is persisted when Apply is clicked

### Requirement: Active/Inactive Dictionary Sections

The dictionary selection dialog SHALL display two sections: "Active" (enabled dictionaries in custom order) and "Inactive" (disabled dictionaries alphabetically sorted). Checking/unchecking moves dictionaries between sections.

#### Scenario: Enable inactive dictionary

- **WHEN** user checks a dictionary in the Inactive section
- **THEN** the dictionary moves to the bottom of the Active section
- **AND** becomes enabled for searches

#### Scenario: Disable active dictionary

- **WHEN** user unchecks a dictionary in the Active section
- **THEN** the dictionary moves to the Inactive section (alphabetically sorted)
- **AND** becomes disabled for searches

### Requirement: Dictionary Order Controls

The Active section SHALL have up/down arrow buttons next to each dictionary to reorder them.

#### Scenario: Move dictionary up

- **WHEN** user clicks the up arrow next to a dictionary in Active section
- **AND** the dictionary is not first
- **THEN** the dictionary moves one position up in the list

#### Scenario: Move dictionary down

- **WHEN** user clicks the down arrow next to a dictionary in Active section
- **AND** the dictionary is not last
- **THEN** the dictionary moves one position down in the list

#### Scenario: Boundary conditions

- **WHEN** user clicks up arrow on first dictionary in Active section
- **THEN** nothing happens (button may be disabled)
- **WHEN** user clicks down arrow on last dictionary in Active section
- **THEN** nothing happens (button may be disabled)
