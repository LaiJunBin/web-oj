Feature: Get User Scores API
  As an admin
  I want to get user scores for a contest project
  So that I can see the ranking of participants

  Background:
    Given I have users:
      | id | name       | username | password | roles      |
      | 11 | AdminUser  | user01   | 123456   | admin      |
      | 12 | User       | user02   | 123456   | contestant |
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
  Scenario: Admin can get user scores with submissions
    Given I have submissions:
      | id   | userID | contestProjectID | isPartial | ignored |
      | sub1 | 13     | 1                | false     | false   |
    And I have judgments:
      | id | submissionID | status | judgeType | score | totalScore |
      | j1 | sub1         | Judged | NORMAL    | 90    | 100        |
    And I am logged in as "user01"
    When I send a "GET" request to "/api/v1/contest-projects/1/user-scores":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": [
          {
            "user": {"id": 13, "name": "Contestant", "username": "user03"},
            "score": 90,
            "submissionId": "sub1"
          },
          {
            "user": {"id": 12, "name": "User", "username": "user02"},
            "score": 0,
            "submissionId": ""
          }
        ]
      }
      """

  Scenario: Returns all users with zero score when no submissions
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/contest-projects/1/user-scores":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": [
          {
            "user": {"id": 12, "name": "User", "username": "user02"},
            "score": 0,
            "submissionId": ""
          },
          {
            "user": {"id": 13, "name": "Contestant", "username": "user03"},
            "score": 0,
            "submissionId": ""
          }
        ]
      }
      """

  Scenario: Unauthorized access
    When I send a "GET" request to "/api/v1/contest-projects/1/user-scores":
    Then the response code should be 401

  Scenario: Forbidden for contestant role
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/contest-projects/1/user-scores":
    Then the response code should be 403
