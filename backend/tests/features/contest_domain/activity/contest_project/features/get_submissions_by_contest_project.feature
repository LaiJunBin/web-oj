Feature: Get Submissions By Contest Project API
  As an admin
  I want to list submissions for a contest project
  So that I can review student submissions

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
      | sub1 | 11     | 1                |
    And I have judgments:
      | id | submissionID | status | judgeType | score | totalScore |
      | j1 | sub1         | Judged | NORMAL    | 80    | 100        |

  Scenario: Submission without judgment is excluded from the list
    Given I have submissions:
      | id   | userID | contestProjectID |
      | sub2 | 11     | 1                |
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/contest-projects/1/submissions":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": [
          {
            "id": "sub1",
            "score": 80,
            "totalScore": 100,
            "ignored": false,
            "contest": {"id": 11, "name": "Test-Contest-1"},
            "project": {"id": 11, "name": "Test-Project-1"},
            "user": {"id": 11, "username": "user01", "name": "AdminUser"},
            "isPartial": false,
            "status": "Judged",
            "type": "NORMAL",
            "createdAt": "{createdAt}"
          }
        ]
      }
      """

  Scenario: Admin can list submissions for a contest project
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/contest-projects/1/submissions":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": [
          {
            "id": "sub1",
            "score": 80,
            "totalScore": 100,
            "ignored": false,
            "contest": {"id": 11, "name": "Test-Contest-1"},
            "project": {"id": 11, "name": "Test-Project-1"},
            "user": {"id": 11, "username": "user01", "name": "AdminUser"},
            "isPartial": false,
            "status": "Judged",
            "type": "NORMAL",
            "createdAt": "{createdAt}"
          }
        ]
      }
      """

  Scenario: Submission from a deleted user is excluded from the list
    Given I have submissions:
      | id   | userID | contestProjectID |
      | sub2 | 13     | 1                |
    And I have judgments:
      | id | submissionID | status | judgeType | score | totalScore |
      | j2 | sub2         | Judged | NORMAL    | 70    | 100        |
    And I am logged in as "user01"
    When I send a "DELETE" request to "/api/v1/users/13":
    Then the response code should be 200
    When I send a "GET" request to "/api/v1/contest-projects/1/submissions":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": [
          {
            "id": "sub1",
            "score": 80,
            "totalScore": 100,
            "ignored": false,
            "contest": {"id": 11, "name": "Test-Contest-1"},
            "project": {"id": 11, "name": "Test-Project-1"},
            "user": {"id": 11, "username": "user01", "name": "AdminUser"},
            "isPartial": false,
            "status": "Judged",
            "type": "NORMAL",
            "createdAt": "{createdAt}"
          }
        ]
      }
      """


  Scenario: Empty submissions list for contest project with no submissions
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/contest-projects/99/submissions":
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "contest project not found",
        "details": null
      }
      """

  Scenario: Unauthorized access
    When I send a "GET" request to "/api/v1/contest-projects/1/submissions":
    Then the response code should be 401

  Scenario: Forbidden for contestant role
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/contest-projects/1/submissions":
    Then the response code should be 403
