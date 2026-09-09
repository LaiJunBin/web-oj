Feature: Get Judgment Analytics API
  As an admin
  I want to get judgment analytics for a contest project
  So that I can analyze judging performance and submission results

  Background:
    Given I have users:
      | id | name       | username | password | roles      |
      | 11 | AdminUser  | user01   | 123456   | admin      |
      | 12 | Contestant | user02   | 123456   | contestant |
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

  Scenario: Admin can get analytics with submissions and judgments
    # createdAt is pinned because avgQueueWaitMs is judgingStartedAt minus this
    # very column — left to default to "now", the expected 1000ms becomes a
    # huge negative number that moves every day. (That is exactly how this
    # file's first-ever run failed: it was written, never registered, and the
    # fixture gap sat unnoticed until the registration guard surfaced it.)
    Given I have submissions:
      | id   | userID | contestProjectID | isPartial | ignored | createdAt            |
      | sub1 | 12     | 1                | false     | false   | 2024-01-01T00:00:00Z |
      | sub2 | 12     | 1                | false     | false   | 2024-01-01T00:00:00Z |
    And I have judgments:
      | id | submissionID | status | judgeType | createdAt            | judgingStartedAt     | judgingFinishedAt    |
      | j1 | sub1         | Judged | JUDGE     | 2024-01-01T00:00:00Z | 2024-01-01T00:00:01Z | 2024-01-01T00:00:03Z |
      | j2 | sub1         | Judged | JUDGE     | 2024-01-01T00:00:00Z | 2024-01-01T00:00:01Z | 2024-01-01T00:00:03Z |
      | j3 | sub2         | Judged | JUDGE     | 2024-01-01T00:00:00Z | 2024-01-01T00:00:01Z | 2024-01-01T00:00:03Z |
    And I am logged in as "user01"
    When I send a "GET" request to "/api/v1/contest-projects/1/judgment-analytics":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "totalSubmissions": 2,
          "totalJudgments": 3,
          "statusCounts": [
            {"status": "Judged", "count": 3}
          ],
          "avgJudgingDurationMs": 2000,
          "avgQueueWaitMs": 1000
        }
      }
      """

  Scenario: Returns zero counts and null averages when no submissions exist
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/contest-projects/1/judgment-analytics":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "totalSubmissions": 0,
          "totalJudgments": 0,
          "statusCounts": [],
          "avgJudgingDurationMs": null,
          "avgQueueWaitMs": null
        }
      }
      """

  Scenario: Unauthorized access
    When I send a "GET" request to "/api/v1/contest-projects/1/judgment-analytics":
    Then the response code should be 401

  Scenario: Forbidden for contestant role
    Given I am logged in as "user02"
    When I send a "GET" request to "/api/v1/contest-projects/1/judgment-analytics":
    Then the response code should be 403
