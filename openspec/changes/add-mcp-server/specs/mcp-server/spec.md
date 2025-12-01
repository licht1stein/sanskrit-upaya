# MCP Server Specification

## ADDED Requirements

### Requirement: Dictionary Search Tool

The MCP server SHALL provide a `sanskrit_search` tool that searches Sanskrit dictionaries.

**Inputs**:

- `query` (string, required): The search term in IAST or Devanagari
- `mode` (string, required): One of "exact", "prefix", "fuzzy", "reverse"
- `dict_codes` (array of strings, optional): Filter to specific dictionaries

**Outputs**:

- `count`: Number of results returned (max 1000)
- `results`: Array of results containing: word, dict_code, dict_name, article_id
- `truncated`: Boolean, true if the database limit (1000) was reached (more results may exist)

The tool returns all results up to the database limit. If `truncated=true`, refine your search with dictionary filters or more specific terms.

#### Scenario: Researcher searches for exact Sanskrit word

- **GIVEN** Priya is a Sanskritology researcher using Claude
- **AND** the dictionary database is available
- **WHEN** Priya asks Claude to search for "dharma" with exact mode
- **THEN** Claude receives results matching exactly "dharma"
- **AND** each result includes the word, dictionary code, dictionary name, and article ID

#### Scenario: Researcher searches with dictionary filter

- **GIVEN** Priya wants definitions only from Monier-Williams dictionary
- **WHEN** Priya asks Claude to search for "yoga" filtered to dict_code "mw"
- **THEN** Claude receives only results from Monier-Williams dictionary

#### Scenario: Researcher performs reverse search in article content

- **GIVEN** Priya wants to find articles mentioning "liberation"
- **WHEN** Priya asks Claude to search "liberation" with reverse mode
- **THEN** Claude receives articles containing "liberation" in their content

#### Scenario: Search returns many results

- **GIVEN** Priya searches for a common prefix like "a"
- **WHEN** the search hits the database limit
- **THEN** Claude receives 1000 results with truncated=true
- **AND** Claude suggests refining the search or filtering by dictionary

### Requirement: Dictionary Listing Tool

The MCP server SHALL provide a `sanskrit_list_dictionaries` tool that lists all available dictionaries with descriptions.

**Inputs**: None

**Outputs**:

- Array of dictionaries containing: code, name, from_lang, to_lang, description

The description SHALL be a brief (1-2 sentence) summary of each dictionary's scope, era, and scholarly significance, loaded from `data/dictionaries.json`.

#### Scenario: Researcher discovers available dictionaries

- **GIVEN** Priya wants to know which dictionaries are available
- **WHEN** Priya asks Claude to list Sanskrit dictionaries
- **THEN** Claude receives a list of all 36 dictionaries with codes, names, language pairs, and descriptions
- **AND** Claude can make informed recommendations about which dictionaries to search

### Requirement: Article Retrieval Tool

The MCP server SHALL provide a `sanskrit_get_article` tool that retrieves a specific dictionary article.

**Inputs**:

- `article_id` (integer, required): The article ID from search results

**Outputs**:

- Single article containing: word (first headword if multiple), dict_code, dict_name, content (full article text)

#### Scenario: Claude retrieves full article content

- **GIVEN** Priya has searched for "dharma" and sees multiple dictionary results
- **WHEN** Priya asks Claude for the full definition from Monier-Williams
- **THEN** Claude uses the article_id from search results to retrieve full content
- **AND** Priya sees the complete article text

#### Scenario: Article not found

- **GIVEN** Claude attempts to retrieve an article with an invalid ID
- **WHEN** the get_article tool is called
- **THEN** Claude receives an error and can inform the user gracefully

### Requirement: Transliteration Tool

The MCP server SHALL provide a `sanskrit_transliterate` tool that converts text between IAST and Devanagari scripts.

**Inputs**:

- `text` (string, required): The text to transliterate
- `direction` (string, required): Either "iast_to_deva" or "deva_to_iast"

**Outputs**:

- Object containing: original text, transliterated text

#### Scenario: Researcher converts IAST to Devanagari

- **GIVEN** Priya has a Sanskrit word in IAST romanization "dharma"
- **WHEN** Priya asks Claude to transliterate to Devanagari
- **THEN** Claude receives the Devanagari form "धर्म"

#### Scenario: Researcher converts Devanagari to IAST

- **GIVEN** Priya has a Sanskrit word in Devanagari "योग"
- **WHEN** Priya asks Claude to transliterate to IAST
- **THEN** Claude receives the IAST form "yoga"

### Requirement: Database Availability Check

The MCP server SHALL check database availability lazily (on first tool call) and provide clear error messages if unavailable.

#### Scenario: Database not found on tool call

- **GIVEN** the dictionary database does not exist at the expected path
- **WHEN** any tool is called
- **THEN** the tool returns an error: "Dictionary database not found at {path}. Run Sanskrit Upaya desktop app to download, or place sanskrit.db manually."

#### Scenario: Database available

- **GIVEN** the dictionary database exists (path resolved via `pkg/paths`)
- **WHEN** any tool is called
- **THEN** the tool executes successfully
