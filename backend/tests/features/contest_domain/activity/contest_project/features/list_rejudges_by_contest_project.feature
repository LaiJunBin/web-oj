Feature: List Rejudges By Contest Project API
  As an admin
  I want to list the rejudges of one contest project
  So that the console can show the rejudge history of the project I am working on

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
      | 12 | Test-Project-2 | Test-Project-2 Description | nginx:latest |     | test-nginx:latest |         |  | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
    And I have contests:
      | id | name           | description                | isPublic | activateTime        | deactivateTime      | createdAt           | updatedAt           |
      | 11 | Test-Contest-1 | Test-Contest-1 Description | 0        | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
    And I have contest project associations:
      | id | contestId | projectId | name           | isPublic | judge priority strategy | activateTime        | deactivateTime      | startTime           | endTime             |
      | 1  | 11        | 11        | Test-Project-1 | 0        | default                 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 |
      | 2  | 11        | 12        | Test-Project-2 | 0        | default                 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 |

  Scenario: Admin sees only the requested contest project's rejudges
    Given I have rejudges:
      | id       | contestProjectID | status   | reason        | judgments | createdAt           |
      | rejudge1 | 1                | Pending  | first project | j1        | 2021-09-02 00:00:00 |
      | rejudge2 | 2                | Finished | other project | j2,j3     | 2021-09-03 00:00:00 |
      | rejudge3 | 1                | Finished | first again   | j4,j5     | 2021-09-04 00:00:00 |
    And I am logged in as "user01"
    When I send a "GET" request to "/api/v1/contest-projects/1/rejudges":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": [
          {
            "id": "rejudge3",
            "contestProjectId": 1,
            "reason": "first again",
            "status": "Finished",
            "effectedJudgments": 2,
            "completedJudgments": 0,
            "createdAt": "{createdAt}"
          },
          {
            "id": "rejudge1",
            "contestProjectId": 1,
            "reason": "first project",
            "status": "Pending",
            "effectedJudgments": 1,
            "completedJudgments": 0,
            "createdAt": "{createdAt}"
          }
        ]
      }
      """

  # Same as the whole-instance listing: an empty judgments array must not make
  # the rejudge disappear from its project's history.
  Scenario: A rejudge that produced no judgments is still listed under its project
    Given I have rejudges:
      | id     | contestProjectID | status  | reason        | judgments | createdAt           |
      | empty1 | 1                | Pending | nothing to do |           | 2021-09-02 00:00:00 |
      | other  | 2                | Pending | other project |           | 2021-09-03 00:00:00 |
    And I am logged in as "user01"
    When I send a "GET" request to "/api/v1/contest-projects/1/rejudges":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": [
          {
            "id": "empty1",
            "contestProjectId": 1,
            "reason": "nothing to do",
            "status": "Pending",
            "effectedJudgments": 0,
            "completedJudgments": 0,
            "createdAt": "{createdAt}"
          }
        ]
      }
      """

  Scenario: Returns empty list when the contest project has no rejudges
    Given I have rejudges:
      | id       | contestProjectID | status  | reason        | judgments | createdAt           |
      | rejudge1 | 1                | Pending | first project | j1        | 2021-09-02 00:00:00 |
    And I am logged in as "user01"
    When I send a "GET" request to "/api/v1/contest-projects/2/rejudges":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": []
      }
      """

  # An empty list would read as "this project has never been rejudged", which is
  # a different answer from "that is not a contest project id".
  Scenario: Rejects a non-positive contest project ID
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/contest-projects/0/rejudges":
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "Invalid contest project ID",
        "details": null
      }
      """

  Scenario: Unauthorized access
    When I send a "GET" request to "/api/v1/contest-projects/1/rejudges":
    Then the response code should be 401

  Scenario: Forbidden for contestant role
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/contest-projects/1/rejudges":
    Then the response code should be 403
