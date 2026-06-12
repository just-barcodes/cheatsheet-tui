Feature: Choosing where cheatsheets are loaded from
  As a user I want my cheatsheets read from a standard place, an environment
  variable, or an explicit flag, so I can keep my own definitions wherever
  suits me — and still get something useful out of the box.

  Scenario: The --dir flag takes priority over everything
    Given the --dir flag is "/flag/dir"
    And the CHEATSHEET_DIR env var is "/env/dir"
    And a config directory "/cfg/dir" that exists
    When I resolve the cheatsheet location
    Then cheatsheets load from "/flag/dir"

  Scenario: CHEATSHEET_DIR is used when no flag is given
    Given the CHEATSHEET_DIR env var is "/env/dir"
    And a config directory "/cfg/dir" that exists
    When I resolve the cheatsheet location
    Then cheatsheets load from "/env/dir"

  Scenario: The config directory is used when it exists
    Given a config directory "/cfg/dir" that exists
    When I resolve the cheatsheet location
    Then cheatsheets load from "/cfg/dir"

  Scenario: Built-in cheatsheets are the last resort
    Given a config directory "/cfg/dir" that does not exist
    When I resolve the cheatsheet location
    Then the built-in cheatsheets are used
