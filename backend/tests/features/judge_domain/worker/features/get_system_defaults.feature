Feature: Get System Defaults API
  As an administrator configuring a contest
  I want to see what a blank setting falls back to
  So that I can tell what "not set" actually means

  # A contest project reports null for the settings it does not override, which
  # tells an administrator nothing about the value in force. This endpoint is
  # what lets the UI render "預設(30 分鐘)" instead of an empty box, and it is
  # what makes "set it to the same number as the default" an instruction
  # somebody can actually follow.
  #
  # The response is a hand-written allowlist: it is read out of the same
  # configuration that holds the signing keys, database passwords and the Sonar
  # token, so what it may contain is pinned by
  # TestSystemDefaultsExposesOnlyKnownFields on the Go side and by the exact
  # match below here.

  Background:
    Given I have users:
      | id  | name               | username | password | roles            |
      | 11  | AdminUser          | user01   | 123456   | admin            |
      | 12  | APIReader          | user02   | 123456   | api_reader       |
      | 13  | Contestant         | user03   | 123456   | contestant       |
      | 14  | AdminAndContestant | user04   | 123456   | admin,contestant |

  Scenario: Successful get system defaults with "AdminUser" user
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/system-defaults":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "maxJudgeDurationMinutes": 30,
          "submissionCooldownSeconds": 0,
          "pidsLimit": 1024,
          "memoryLimitMB": 0
        }
      }
      """

  # Read-only dashboards need it too — it is settings metadata, not a secret.
  Scenario: Successful get system defaults with "APIReader" user
    Given I am logged in as "user02"
    When I send a "GET" request to "/api/v1/system-defaults":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "maxJudgeDurationMinutes": 30,
          "submissionCooldownSeconds": 0,
          "pidsLimit": 1024,
          "memoryLimitMB": 0
        }
      }
      """

  Scenario: Successful get system defaults with "AdminAndContestant" user
    Given I am logged in as "user04"
    When I send a "GET" request to "/api/v1/system-defaults":
    Then the response code should be 200

  # Contestants have no business reading deployment configuration, and nothing
  # in the contestant UI needs it.
  Scenario: Forbidden get system defaults with "Contestant" user
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/system-defaults":
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """

  Scenario: Unauthorized get system defaults
    When I send a "GET" request to "/api/v1/system-defaults":
    Then the response code should be 401
