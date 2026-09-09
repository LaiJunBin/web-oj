Feature: Judge Queue Status API
  As an admin
  I want a live view of the judge pipeline's load
  So that the status page can show a backlog building before contestants report it

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

  # An idle pipeline has to be distinguishable from a stalled one. A queue with
  # nothing in it has no oldest wait at all — reporting 0 there would be a real
  # measurement of a judgment that does not exist.
  Scenario: An idle pipeline reports zero depth and no oldest wait
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/judge-queue":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "queueDepths": {
            "judge": 0,
            "judgePartial": 0,
            "rejudge": 0
          },
          "activeJudgments": 0,
          "oldestQueuedSeconds": null
        }
      }
      """

  # The workers are stopped first so "queued" is a state the scenario controls.
  # With them running, several of them race to pop and the judgment is usually
  # gone before a status request can see it.
  Scenario: Queued judgments raise the depth of their own judge type, and draining clears it
    Given I have submissions:
      | id   | userID | contestProjectID |
      | sub1 | 13     | 1                |
    And I am logged in as "user01"
    And all judge workers are stopped
    When I send a "POST" request to "/api/v1/contest-projects/1/rejudge" with the following data:
      """
      {
        "reason": "queue depth check"
      }
      """
    Then the response code should be 200
    And the "rejudge" judge queue depth should reach 1 within 15 seconds
    When I send a "GET" request to "/api/v1/judge-queue":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "queueDepths": {
            "judge": 0,
            "judgePartial": 0,
            "rejudge": 1
          },
          "activeJudgments": 0,
          "oldestQueuedSeconds": 1e+308
        }
      }
      """
    When all judge workers are started
    Then the judge queue should drain within 60 seconds

  Scenario: Unauthorized access
    When I send a "GET" request to "/api/v1/judge-queue":
    Then the response code should be 401

  Scenario: Forbidden for contestant role
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/judge-queue":
    Then the response code should be 403
