Feature: Theming the UI with a config file
  As a user I want to set my own colors in a simple theme.yaml so the cheat-sheet
  matches the rest of my terminal, without touching any code.

  Scenario: Colors from the theme file are loaded
    Given a theme file with content:
      """
      colors:
        accent: "#FF8800"
        keycap: "63"
      """
    When I load that theme
    Then the accent color is "#FF8800"
    And the keycap color is "63"

  Scenario: Unset colors are left empty so the default applies
    Given a theme file with content:
      """
      colors:
        accent: "#FF8800"
      """
    When I load that theme
    Then the accent color is "#FF8800"
    And the foreground color is unset

  Scenario: An empty theme file is fine and keeps every default
    Given a theme file with content:
      """
      """
    When I load that theme
    Then the accent color is unset

  Scenario: A missing theme file is not an error
    Given no theme file exists
    When I load that theme
    Then the accent color is unset

  Scenario: A malformed color is reported, not silently ignored
    Given a theme file with content:
      """
      colors:
        accent: "bright-orange"
      """
    When I load that theme
    Then loading the theme fails with an error

  Scenario: An unknown setting is reported
    Given a theme file with content:
      """
      colors:
        accent: "#FF8800"
      font_size: 14
      """
    When I load that theme
    Then loading the theme fails with an error

  Scenario: The --theme flag selects the theme file
    Given the --theme flag is "/etc/cheatsheet/dark.yaml"
    And a config directory "/home/me/.config/cheatsheet" for the theme
    When I resolve the theme source
    Then the theme loads from "/etc/cheatsheet/dark.yaml"
    And the theme file is required to exist

  Scenario: Without the flag the config directory theme is used
    Given a config directory "/home/me/.config/cheatsheet" for the theme
    When I resolve the theme source
    Then the theme loads from "/home/me/.config/cheatsheet/theme.yaml"
    And a missing theme file is allowed

  Scenario: With neither, no theme file is loaded
    When I resolve the theme source
    Then no theme file is loaded
