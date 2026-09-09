Feature: Contest Bundle Export / Import API
  As a contest organiser opening the same sitting at a new venue
  I want to move the contest's structure as one file
  So that the per-problem times and limits do not have to be retyped

  # A contest bundle names its problems and never carries them. That is what
  # keeps importing a sitting from also creating problems, and so removes any
  # order the operator would otherwise have to get right. A slot whose name
  # resolves to nothing is reported rather than invented — the alternative is a
  # contest that looks configured and grades the wrong thing.

  Background:
    Given I have users:
      | id | name       | username | password | roles      |
      | 11 | AdminUser  | user01   | 123456   | admin      |
      | 12 | APIReader  | user02   | 123456   | api_reader |
      | 13 | Contestant | user03   | 123456   | contestant |
    And I have images:
      | id | tag               | mount_path    | description | is_ready | createdAt           | updatedAt           |
      | 1  | nginx:latest      | /var/www/html | d1          | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 2  | test-nginx:latest | /var/www/html | d2          | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
    And I have projects:
      | id | name      | description | imageTag     | env | testImageTag      | testEnv | resultParser | createdAt           | updatedAt           |
      | 11 | Alpha     | A           | nginx:latest |     | test-nginx:latest |         |              | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 12 | Beta      | B           | nginx:latest |     | test-nginx:latest |         |              | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 13 | Duplicate | C           | nginx:latest |     | test-nginx:latest |         |              | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 14 | Duplicate | D           | nginx:latest |     | test-nginx:latest |         |              | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
    And I have contests:
      | id | name        | description | isPublic | activateTime        | deactivateTime      | createdAt           | updatedAt           |
      | 21 | Autumn-2026 | Autumn      | 0        | 2026-09-01 00:00:00 | 2026-09-15 00:00:00 | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
    And I have contest project associations:
      | id | contestId | projectId | name  | isPublic | judgePriorityStrategy | activateTime        | deactivateTime      | startTime           | endTime             |
      | 31 | 21        | 11        | 第一題 | 1        | default               | 2026-09-01 00:00:00 | 2026-09-15 00:00:00 | 2026-09-01 09:00:00 | 2026-09-01 17:00:00 |
      | 32 | 21        | 12        | 第二題 | 0        | default               | 2026-09-02 00:00:00 | 2026-09-15 00:00:00 | 2026-09-02 09:00:00 | 2026-09-02 17:00:00 |

  Scenario: Exporting a contest writes its structure and per-problem settings
    Given I am logged in as "user01"
    When I export the bundle with a "GET" request to "/api/v1/contests/21/export"
    Then the exported bundle should contain:
      | path         |
      | contest.json |
    And the exported bundle file "contest.json" should match json:
      """
      {
        "formatVersion": 1,
        "kind": "contest",
        "name": "Autumn-2026",
        "description": "Autumn",
        "isPublic": false,
        "activateTime": "2026-09-01T00:00:00Z",
        "deactivateTime": "2026-09-15T00:00:00Z",
        "projects": [
          {
            "projectRef": "Alpha",
            "name": "第一題",
            "isPublic": true,
            "judgePriorityStrategy": "default",
            "activateTime": "2026-09-01T00:00:00Z",
            "deactivateTime": "2026-09-15T00:00:00Z",
            "startTime": "2026-09-01T09:00:00Z",
            "endTime": "2026-09-01T17:00:00Z",
            "limits": {
              "maxJudgeDurationMinutes": null,
              "submissionCooldownSeconds": null,
              "pidsLimit": null,
              "memoryLimitMB": null
            },
            "scoreboard": { "compact": false, "maskNames": false }
          },
          {
            "projectRef": "Beta",
            "name": "第二題",
            "isPublic": false,
            "judgePriorityStrategy": "default",
            "activateTime": "2026-09-02T00:00:00Z",
            "deactivateTime": "2026-09-15T00:00:00Z",
            "startTime": "2026-09-02T09:00:00Z",
            "endTime": "2026-09-02T17:00:00Z",
            "limits": {
              "maxJudgeDurationMinutes": null,
              "submissionCooldownSeconds": null,
              "pidsLimit": null,
              "memoryLimitMB": null
            },
            "scoreboard": { "compact": false, "maskNames": false }
          }
        ]
      }
      """

  # The round trip. Both problems are in the library, so both slots resolve.
  Scenario: A round trip recreates the sitting
    Given I am logged in as "user01"
    When I export the bundle with a "GET" request to "/api/v1/contests/21/export"
    And I import the exported bundle with a "POST" request to "/api/v1/contests/import"
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "contestId": 1e+308,
          "name": "{name}",
          "createdProjects": 2,
          "unresolvedProjects": []
        }
      }
      """
    And there should be 2 contests

  # A venue that has not imported every problem yet still gets its contest, and
  # a list of what is missing. Refusing would make them sequence their imports
  # by trial and error.
  Scenario: Slots naming a problem the library does not have are reported, not invented
    Given I am logged in as "user01"
    When I import a bundle with a "POST" request to "/api/v1/contests/import" whose "contest.json" is:
      """
      {
        "formatVersion": 1,
        "kind": "contest",
        "name": "Partial",
        "description": "",
        "isPublic": false,
        "activateTime": "2026-09-01T00:00:00Z",
        "deactivateTime": "2026-09-15T00:00:00Z",
        "projects": [
          {
            "projectRef": "Alpha",
            "name": "有的",
            "isPublic": true,
            "judgePriorityStrategy": "default",
            "activateTime": "2026-09-01T00:00:00Z",
            "deactivateTime": "2026-09-15T00:00:00Z",
            "startTime": "2026-09-01T09:00:00Z",
            "endTime": "2026-09-01T17:00:00Z",
            "limits": { "maxJudgeDurationMinutes": null, "submissionCooldownSeconds": null, "pidsLimit": null, "memoryLimitMB": null },
            "scoreboard": { "compact": false, "maskNames": false }
          },
          {
            "projectRef": "Nowhere",
            "name": "沒有的",
            "isPublic": true,
            "judgePriorityStrategy": "default",
            "activateTime": "2026-09-01T00:00:00Z",
            "deactivateTime": "2026-09-15T00:00:00Z",
            "startTime": "2026-09-01T09:00:00Z",
            "endTime": "2026-09-01T17:00:00Z",
            "limits": { "maxJudgeDurationMinutes": null, "submissionCooldownSeconds": null, "pidsLimit": null, "memoryLimitMB": null },
            "scoreboard": { "compact": false, "maskNames": false }
          }
        ]
      }
      """
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "contestId": 1e+308,
          "name": "Partial",
          "createdProjects": 1,
          "unresolvedProjects": ["Nowhere"]
        }
      }
      """

  # Project names carry no uniqueness constraint, so a name really can identify
  # two problems. Picking either would produce a sitting that looks configured
  # and grades the wrong one, and nobody would notice until it was under way.
  Scenario: An ambiguous problem name is left unresolved rather than guessed
    Given I am logged in as "user01"
    When I import a bundle with a "POST" request to "/api/v1/contests/import" whose "contest.json" is:
      """
      {
        "formatVersion": 1,
        "kind": "contest",
        "name": "Ambiguous",
        "description": "",
        "isPublic": false,
        "activateTime": "2026-09-01T00:00:00Z",
        "deactivateTime": "2026-09-15T00:00:00Z",
        "projects": [
          {
            "projectRef": "Duplicate",
            "name": "哪一個",
            "isPublic": true,
            "judgePriorityStrategy": "default",
            "activateTime": "2026-09-01T00:00:00Z",
            "deactivateTime": "2026-09-15T00:00:00Z",
            "startTime": "2026-09-01T09:00:00Z",
            "endTime": "2026-09-01T17:00:00Z",
            "limits": { "maxJudgeDurationMinutes": null, "submissionCooldownSeconds": null, "pidsLimit": null, "memoryLimitMB": null },
            "scoreboard": { "compact": false, "maskNames": false }
          }
        ]
      }
      """
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "contestId": 1e+308,
          "name": "Ambiguous",
          "createdProjects": 0,
          "unresolvedProjects": ["Duplicate"]
        }
      }
      """

  Scenario: An invalid bundle is rejected with every fault at once
    Given I am logged in as "user01"
    When I import a bundle with a "POST" request to "/api/v1/contests/import" whose "contest.json" is:
      """
      {
        "formatVersion": 1,
        "kind": "contest",
        "name": "",
        "description": "",
        "isPublic": false,
        "activateTime": "2026-09-15T00:00:00Z",
        "deactivateTime": "2026-09-01T00:00:00Z",
        "projects": []
      }
      """
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "{message}",
        "details": [
          "name: 不可為空",
          "deactivateTime: 必須晚於 activateTime"
        ]
      }
      """
    And there should be 1 contests

  Scenario: A bundle from a newer format version is refused
    Given I am logged in as "user01"
    When I import a bundle with a "POST" request to "/api/v1/contests/import" whose "contest.json" is:
      """
      {
        "formatVersion": 99,
        "kind": "contest",
        "name": "Future",
        "description": "",
        "isPublic": false,
        "activateTime": "2026-09-01T00:00:00Z",
        "deactivateTime": "2026-09-15T00:00:00Z",
        "projects": []
      }
      """
    Then the response code should be 400
    And there should be 1 contests

  Scenario: Contestants may not export
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/contests/21/export":
    Then the response code should be 403

  Scenario: Read-only users may not import
    Given I am logged in as "user02"
    When I import a bundle with a "POST" request to "/api/v1/contests/import" whose "contest.json" is:
      """
      {
        "formatVersion": 1,
        "kind": "contest",
        "name": "Nope",
        "description": "",
        "isPublic": false,
        "activateTime": "2026-09-01T00:00:00Z",
        "deactivateTime": "2026-09-15T00:00:00Z",
        "projects": []
      }
      """
    Then the response code should be 403
    And there should be 1 contests

  Scenario: Exporting a contest that does not exist
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/contests/999/export":
    Then the response code should be 404
