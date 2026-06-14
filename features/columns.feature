Feature: Multiple hotkey columns
  As a user with a wide terminal I want hotkeys to flow into several side-by-side
  columns so I can see more bindings at once, and I want to choose the count
  myself when the automatic layout isn't what I prefer.

  Background:
    Given a cheatsheet "Keys" with 40 bindings

  Scenario: A wide terminal flows hotkeys into multiple columns
    Given a terminal 160 columns wide and 12 rows tall
    When I view the cheatsheet
    Then the hotkeys are laid out in 3 columns
    And binding "k15" is visible

  Scenario: Columns are capped at three even on a very wide terminal
    Given a terminal 400 columns wide and 12 rows tall
    When I view the cheatsheet
    Then the hotkeys are laid out in 3 columns

  Scenario: A narrow terminal keeps a single column
    Given a terminal 60 columns wide and 12 rows tall
    When I view the cheatsheet
    Then the hotkeys are laid out in 1 column
    And binding "k15" is not visible

  Scenario: The user pins the column count
    Given a terminal 160 columns wide and 12 rows tall
    When I set the column count to 2
    And I view the cheatsheet
    Then the hotkeys are laid out in 2 columns

  Scenario: A pinned count never gets narrower than is readable
    Given a terminal 60 columns wide and 12 rows tall
    When I set the column count to 3
    And I view the cheatsheet
    Then the hotkeys are laid out in 1 column

  Scenario: Cycling steps through auto and back, capped at three
    Given a terminal 160 columns wide and 12 rows tall
    When I cycle the column count 4 times
    And I view the cheatsheet
    Then the hotkeys are laid out in 3 columns

  Scenario: The column setting is shown in the footer as auto by default
    Given a terminal 160 columns wide and 12 rows tall
    When I view the cheatsheet
    Then the screen shows the column count "cols:auto"

  Scenario: The footer shows the pinned column setting, not the visible count
    Given a terminal 60 columns wide and 12 rows tall
    When I set the column count to 3
    And I view the cheatsheet
    Then the hotkeys are laid out in 1 column
    And the screen shows the column count "cols:3"

  Scenario: A long description is shown in full, not truncated
    Given a cheatsheet "Keys" with a long description ending in "ZEBRA"
    And a terminal 80 columns wide and 20 rows tall
    When I view the cheatsheet
    Then binding "ZEBRA" is visible

  Scenario: A window too small for one readable column asks to be enlarged
    Given a terminal 30 columns wide and 20 rows tall
    When I view the cheatsheet
    Then the screen asks for a larger window
