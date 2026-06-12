Feature: Searching hotkeys
  As a user I want to fuzzy-search hotkeys across every cheatsheet so I can
  find a binding instantly without remembering which program it belongs to.

  Background:
    Given the following hotkeys:
      | sheet    | section  | keys    | desc         |
      | Vim      | Movement | dd      | Delete line  |
      | Vim      | Movement | yy      | Yank line    |
      | Hyprland | Window   | SUPER+Q | Close window |

  Scenario: Match on a word in the description
    When I search for "delete"
    Then the results contain key "dd"
    And the results do not contain key "yy"

  Scenario: Fuzzy subsequence match on the description
    When I search for "clo win"
    Then the results contain key "SUPER+Q"

  Scenario: Empty query returns everything in original order
    When I search for ""
    Then I get 3 results
    And result 1 has key "dd"

  Scenario: Better matches rank higher
    When I search for "line"
    Then the results contain key "dd"
    And the results contain key "yy"
    And the results do not contain key "SUPER+Q"
