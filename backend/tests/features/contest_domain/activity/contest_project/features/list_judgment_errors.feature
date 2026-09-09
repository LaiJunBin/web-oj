Feature: Judgment Errors API
  As an admin
  I want the most recent judging failures
  So that I can open the submission that broke instead of only seeing an error count

  Background:
    Given I have users:
      | id | name       | username | password | roles      |
      | 11 | AdminUser  | user01   | 123456   | admin      |
      | 13 | Contestant | user03   | 123456   | contestant |
    And I have images:
      | id | tag               | mount_path    | description | is_ready | createdAt           | updatedAt           |
      | 1  | nginx:latest      | /var/www/html | desc1       | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 2  | test-nginx:latest | /var/www/html | desc2       | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
    And I have projects:
      | id | name           | description                | imageTag     | env | testImageTag      | testEnv | resultParser | createdAt           | updatedAt           |
      | 11 | Test-Project-1 | Test-Project-1 Description | nginx:latest |     | test-nginx:latest |         |  | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
    And I have contests:
      | id | name           | description                | isPublic | activateTime        | deactivateTime      | createdAt           | updatedAt           |
      | 11 | Test-Contest-1 | Test-Contest-1 Description | 0        | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
    And I have contest project associations:
      | id | contestId | projectId | name           | isPublic | judge priority strategy | activateTime        | deactivateTime      | startTime           | endTime             |
      | 1  | 11        | 11        | Test-Project-1 | 0        | default                 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 |
    And I have submissions:
      | id   | userID | contestProjectID |
      | sub1 | 13     | 1                |
      | sub2 | 13     | 1                |

  # Newest first: this list answers "what just broke", so the most recent
  # failure is the one an admin needs at the top.
  Scenario: Errored judgments come back newest first
    Given I have judgments:
      | id      | submissionID | contestProjectID | judgeType | status | score | totalScore | errorMessage           | createdAt           |
      | older   | sub1         | 1                | JUDGE     | Error  | 0     | 100        | provisioning exhausted | 2021-09-02 00:00:00 |
      | newest  | sub2         | 1                | JUDGE     | Error  | 0     | 100        | container never healthy | 2021-09-04 00:00:00 |
      | middle  | sub1         | 1                | JUDGE     | Error  | 0     | 100        | testcase timed out     | 2021-09-03 00:00:00 |
    And I am logged in as "user01"
    When I send a "GET" request to "/api/v1/judgment-errors":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": [
          {
            "judgmentId": "newest",
            "submissionId": "sub2",
            "contestProjectId": 1,
            "errorMessage": "container never healthy",
            "createdAt": "{createdAt}"
          },
          {
            "judgmentId": "middle",
            "submissionId": "sub1",
            "contestProjectId": 1,
            "errorMessage": "testcase timed out",
            "createdAt": "{createdAt}"
          },
          {
            "judgmentId": "older",
            "submissionId": "sub1",
            "contestProjectId": 1,
            "errorMessage": "provisioning exhausted",
            "createdAt": "{createdAt}"
          }
        ]
      }
      """

  # Only failures. A Judged run is not something to investigate, and mixing it
  # in would bury the failures the page exists to surface.
  Scenario: Judgments that did not error are excluded
    Given I have judgments:
      | id       | submissionID | contestProjectID | judgeType | status | score | totalScore | errorMessage   | createdAt           |
      | fine     | sub1         | 1                | JUDGE     | Judged | 90    | 100        |                | 2021-09-02 00:00:00 |
      | broken   | sub2         | 1                | JUDGE     | Error  | 0     | 100        | image missing  | 2021-09-03 00:00:00 |
    And I am logged in as "user01"
    When I send a "GET" request to "/api/v1/judgment-errors":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": [
          {
            "judgmentId": "broken",
            "submissionId": "sub2",
            "contestProjectId": 1,
            "errorMessage": "image missing",
            "createdAt": "{createdAt}"
          }
        ]
      }
      """

  Scenario: The limit caps how many come back, keeping the newest
    Given I have judgments:
      | id     | submissionID | contestProjectID | judgeType | status | score | totalScore | errorMessage | createdAt           |
      | e1     | sub1         | 1                | JUDGE     | Error  | 0     | 100        | first        | 2021-09-01 00:00:00 |
      | e2     | sub1         | 1                | JUDGE     | Error  | 0     | 100        | second       | 2021-09-02 00:00:00 |
      | e3     | sub2         | 1                | JUDGE     | Error  | 0     | 100        | third        | 2021-09-03 00:00:00 |
    And I am logged in as "user01"
    When I send a "GET" request to "/api/v1/judgment-errors?limit=1":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": [
          {
            "judgmentId": "e3",
            "submissionId": "sub2",
            "contestProjectId": 1,
            "errorMessage": "third",
            "createdAt": "{createdAt}"
          }
        ]
      }
      """

  Scenario: Returns an empty list when nothing has failed
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/judgment-errors":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": []
      }
      """

  Scenario: Rejects a limit outside the supported range
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/judgment-errors?limit=0":
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "Limit must be between 1 and 200",
        "details": null
      }
      """

  Scenario: Unauthorized access
    When I send a "GET" request to "/api/v1/judgment-errors":
    Then the response code should be 401

  Scenario: Forbidden for contestant role
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/judgment-errors":
    Then the response code should be 403
