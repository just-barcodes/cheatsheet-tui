Feature: Loading cheatsheets from YAML
  As a user I want my hotkeys defined in simple YAML files, grouped per
  program/app/type, so the TUI can display them.

  Scenario: Load a single cheatsheet file
    Given a cheatsheet file "vim.yaml" with content:
      """
      name: Vim
      description: Modal editor
      sections:
        - title: Movement
          bindings:
            - keys: "h j k l"
              desc: "Move left/down/up/right"
            - keys: "w"
              desc: "Next word"
      """
    When I load that cheatsheet
    Then the cheatsheet name is "Vim"
    And it has 1 section
    And section "Movement" has 2 bindings
    And binding "w" has description "Next word"

  Scenario: Load a directory of cheatsheets sorted by name
    Given a directory with cheatsheets:
      | file         | name     |
      | 02-vim.yaml  | Vim      |
      | 01-hypr.yaml | Hyprland |
    When I load the directory
    Then I get 2 cheatsheets
    And the cheatsheets are ordered "Hyprland, Vim"

  Scenario: A malformed file is reported, not silently dropped
    Given a cheatsheet file "broken.yaml" with content:
      """
      name: [this is not valid
      """
    When I load that cheatsheet
    Then loading fails with an error
