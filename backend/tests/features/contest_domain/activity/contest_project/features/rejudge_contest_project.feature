Feature: Rejudge Contest Project API
  As an admin
  I want to rejudge all submissions for a contest project
  So that scores are recalculated with updated test cases

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

  Scenario: Admin can rejudge all submissions for a contest project
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/contest-projects/1/rejudge" with the following data:
      """
      {
        "reason": "Updated test cases"
      }
      """
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true
      }
      """

  # The whole point of storing the contest project on the rejudge: a record
  # created through this endpoint has to come back out of the listing filed
  # under the project it was raised against. Nothing downstream of the request
  # can recover the project on its own, so a rejudge created without it lands in
  # the console under project 0.
  #
  # The workers are stopped so the judgment stays queued: this scenario is about
  # where the rejudge is filed, not about how fast it gets judged, and with the
  # pool running completedJudgments races between 0 and 1 depending on whether a
  # worker has picked the judgment up and failed it yet. The scenario
  # initializer's judge reset restarts them afterwards.
  Scenario: A rejudge created here is listed under its contest project
    Given I am logged in as "user01"
    And all judge workers are stopped
    When I send a "POST" request to "/api/v1/contest-projects/1/rejudge" with the following data:
      """
      {
        "reason": "Updated test cases"
      }
      """
    Then the response code should be 200
    And a rejudge for contest project 1 should exist within 10 seconds
    When I send a "GET" request to "/api/v1/rejudges":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": [
          {
            "id": "{id}",
            "contestProjectId": 1,
            "reason": "Updated test cases",
            "status": "{status}",
            "effectedJudgments": 1,
            "completedJudgments": 0,
            "createdAt": "{createdAt}"
          }
        ]
      }
      """
    When I send a "GET" request to "/api/v1/contest-projects/1/rejudges":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": [
          {
            "id": "{id}",
            "contestProjectId": 1,
            "reason": "Updated test cases",
            "status": "{status}",
            "effectedJudgments": 1,
            "completedJudgments": 0,
            "createdAt": "{createdAt}"
          }
        ]
      }
      """

  Scenario: Rejudge requires a non-empty reason
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/contest-projects/1/rejudge" with the following data:
      """
      {
        "reason": ""
      }
      """
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "Required reason",
        "details": null
      }
      """

  Scenario: Contest project not found returns 400 for invalid ID
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/contest-projects/0/rejudge" with the following data:
      """
      {
        "reason": "test"
      }
      """
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "Invalid contest project ID",
        "details": null
      }
      """

  # Two rejudges outstanding on the same submission can be approved in either
  # order, and approving the older one last silently replaces the newer result
  # with a staler score. Refusing the second request is what keeps a submission's
  # rejudge history linear; cancelling the outstanding one is the way to change
  # course.
  Scenario: Rejudging again is refused while a previous rejudge awaits a decision
    Given I have judgments:
      | id         | submissionID | judgeType | status   | score | totalScore | rejudgeID | createdAt           |
      | rejudged-1 | sub1         | REJUDGE   | Rejudged | 30    | 100        | rejudge1  | 2021-09-02 00:00:00 |
    And I have rejudges:
      | id       | status   | reason     | judgments  |
      | rejudge1 | Finished | first pass | rejudged-1 |
    And I am logged in as "user01"
    When I send a "POST" request to "/api/v1/contest-projects/1/rejudge" with the following data:
      """
      {
        "reason": "second attempt"
      }
      """
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "One or more submissions already have a rejudge awaiting decision",
        "details": null
      }
      """

  # A decided rejudge has no further claim on the score, so it must not block.
  Scenario: Rejudging again is allowed once the previous rejudge was decided
    Given I have judgments:
      | id         | submissionID | judgeType | status | score | totalScore | rejudgeID | createdAt           |
      | rejudged-1 | sub1         | REJUDGE   | Judged | 30    | 100        | rejudge1  | 2021-09-02 00:00:00 |
    And I have rejudges:
      | id       | status   | reason     | judgments  |
      | rejudge1 | Approved | first pass | rejudged-1 |
    And I am logged in as "user01"
    When I send a "POST" request to "/api/v1/contest-projects/1/rejudge" with the following data:
      """
      {
        "reason": "second attempt"
      }
      """
    Then the response code should be 200

  Scenario: Unauthorized access
    When I send a "POST" request to "/api/v1/contest-projects/1/rejudge" with the following data:
      """
      {
        "reason": "test"
      }
      """
    Then the response code should be 401

  Scenario: Forbidden for contestant role
    Given I am logged in as "user03"
    When I send a "POST" request to "/api/v1/contest-projects/1/rejudge" with the following data:
      """
      {
        "reason": "test"
      }
      """
    Then the response code should be 403
