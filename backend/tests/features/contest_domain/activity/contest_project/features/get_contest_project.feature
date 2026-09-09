Feature: Get Contest Project API
  As a user
  I want to get a project of a contest
  So that I can see a project of a contest if I have permission

  Background:
    Given I have users:
      | id  | name               | username | password | roles              |
      | 11  | AdminUser          | user01   | 123456   | admin              |
      | 12  | APIReader          | user02   | 123456   | api_reader         |
      | 13  | Contestant         | user03   | 123456   | contestant         |
      | 14  | AdminAndContestant | user04   | 123456   | admin,contestant   |
    And I have images:
      | id | tag                | mount_path    | description  | is_ready | createdAt           | updatedAt           |
      | 1  | nginx:latest       | /var/www/html | description1 | 0        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 2  | mysql:8.0.26       | /var/www/html | description2 | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 3  | test-nginx:latest  | /var/www/html | description3 | 0        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 4  | test-mysql:8.0.26  | /var/www/html | description4 | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 5  | test-nginx:test-2  | /var/www/html | description5 | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
    And I have projects:
      | id | name           | description               | imageTag       | env | testImageTag      | testEnv | resultParser | createdAt           | updatedAt           |
      | 11 | Test-Project-1 | Test-Project-1 Description| nginx:latest   | a=b | test-nginx:latest |         |  | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 12 | Test-Project-2 | Test-Project-2 Description| mysql:8.0.26   | c=d | test-mysql:8.0.26 |         |  | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 13 | Test-Project-3 | Test-Project-3 Description| nginx:latest   |     | test-nginx:test-2 |         |  | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 14 | Test-Project-4 | Test-Project-4 Description| nginx:latest   |     | test-nginx:test-2 |         |  | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
    And I have contests: 
      | id | name           | description                | isPublic | activateTime        | deactivateTime      | createdAt           | updatedAt           |
      | 11 | Test-Contest-1 | Test-Contest-1 Description | 0        | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 12 | Test-Contest-2 | Test-Contest-2 Description | 1        | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 13 | Test-Contest-3 | Test-Contest-3 Description | 0        | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
    And I have contest project associations:
      | id | contestId | projectId | name           | isPublic | judge priority strategy | activateTime        | deactivateTime      | startTime           | endTime             |
      | 1  | 11        | 11        | Test-Project-1 | 0        | default                 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 |
      | 2  | 11        | 12        | Test-Project-2 | 0        | default                 | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 |
      | 3  | 12        | 11        | Test-Project-1 | 1        | default                 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 |
      | 4  | 12        | 12        | Test-Project-2 | 1        | default                 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2099-09-01 00:00:00 | 2099-09-30 23:59:59 |
      | 5  | 12        | 13        | Test-Project-3 | 1        | default                 | 2099-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 |
      | 6  | 13        | 11        | Test-Project-1 | 1        | default                 | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 |
      | 7  | 15        | 12        | Test-Project-2 | 1        | default                 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 |
  Scenario: Successful get contest-project with "AdminUser" user
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/contest-projects/1":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": 1,
          "name": "Test-Project-1",
          "projectId": 11,
          "description": "Test-Project-1 Description",
          "isPublic": false,
          "judgePriorityStrategy": "default",
          "image": "nginx:latest",
          "testImage": "test-nginx:latest",
          "activateTime": "{activateTime}",
          "deactivateTime": "{deactivateTime}",
          "startTime": "{startTime}",
          "endTime": "{endTime}",
          "maxJudgeDurationMinutes": null,
          "submissionCooldownSeconds": null,
          "pidsLimit": null,
          "memoryLimitMB": null,
          "scoreboardCompact": false,
          "scoreboardMaskNames": false
        }
      }
      """
  
  Scenario: Successful get contest-project with "AdminUser" user
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/contest-projects/2":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": 2,
          "name": "Test-Project-2",
          "projectId": 12,
          "description": "Test-Project-2 Description",
          "isPublic": false,
          "judgePriorityStrategy": "default",
          "image": "mysql:8.0.26",
          "testImage": "test-mysql:8.0.26",
          "activateTime": "{activateTime}",
          "deactivateTime": "{deactivateTime}",
          "startTime": "{startTime}",
          "endTime": "{endTime}",
          "maxJudgeDurationMinutes": null,
          "submissionCooldownSeconds": null,
          "pidsLimit": null,
          "memoryLimitMB": null,
          "scoreboardCompact": false,
          "scoreboardMaskNames": false
        }
      }
      """

  Scenario: Successful get contest-project with "AdminUser" user
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/contest-projects/3":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": 3,
          "name": "Test-Project-1",
          "projectId": 11,
          "description": "Test-Project-1 Description",
          "isPublic": true,
          "judgePriorityStrategy": "default",
          "image": "nginx:latest",
          "testImage": "test-nginx:latest",
          "activateTime": "{activateTime}",
          "deactivateTime": "{deactivateTime}",
          "startTime": "{startTime}",
          "endTime": "{endTime}",
          "maxJudgeDurationMinutes": null,
          "submissionCooldownSeconds": null,
          "pidsLimit": null,
          "memoryLimitMB": null,
          "scoreboardCompact": false,
          "scoreboardMaskNames": false
        }
      }
      """
      
  Scenario: Successful get contest-project with "AdminAndContestant" user
    Given I am logged in as "user04"
    When I send a "GET" request to "/api/v1/contest-projects/4":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": 4,
          "name": "Test-Project-2",
          "projectId": 12,
          "description": "Test-Project-2 Description",
          "isPublic": true,
          "judgePriorityStrategy": "default",
          "image": "mysql:8.0.26",
          "testImage": "test-mysql:8.0.26",
          "activateTime": "{activateTime}",
          "deactivateTime": "{deactivateTime}",
          "startTime": "{startTime}",
          "endTime": "{endTime}",
          "maxJudgeDurationMinutes": null,
          "submissionCooldownSeconds": null,
          "pidsLimit": null,
          "memoryLimitMB": null,
          "scoreboardCompact": false,
          "scoreboardMaskNames": false
        }
      }
      """

  Scenario: Successful get contest-project with "AdminUser" user
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/contest-projects/5":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": 5,
          "name": "Test-Project-3",
          "projectId": 13,
          "description": "Test-Project-3 Description",
          "isPublic": true,
          "judgePriorityStrategy": "default",
          "image": "nginx:latest",
          "testImage": "test-nginx:test-2",
          "activateTime": "{activateTime}",
          "deactivateTime": "{deactivateTime}",
          "startTime": "{startTime}",
          "endTime": "{endTime}",
          "maxJudgeDurationMinutes": null,
          "submissionCooldownSeconds": null,
          "pidsLimit": null,
          "memoryLimitMB": null,
          "scoreboardCompact": false,
          "scoreboardMaskNames": false
        }
      }
      """

  Scenario: Successful get contest-project with "AdminUser" user
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/contest-projects/6":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": 6,
          "name": "Test-Project-1",
          "projectId": 11,
          "description": "Test-Project-1 Description",
          "isPublic": true,
          "judgePriorityStrategy": "default",
          "image": "nginx:latest",
          "testImage": "test-nginx:latest",
          "activateTime": "{activateTime}",
          "deactivateTime": "{deactivateTime}",
          "startTime": "{startTime}",
          "endTime": "{endTime}",
          "maxJudgeDurationMinutes": null,
          "submissionCooldownSeconds": null,
          "pidsLimit": null,
          "memoryLimitMB": null,
          "scoreboardCompact": false,
          "scoreboardMaskNames": false
        }
      }
      """

  Scenario: Unauthorized get contest-project
    When I send a "GET" request to "/api/v1/contest-projects/1":
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "Unauthorized",
        "details": null
      }
      """

  Scenario: Forbidden get contest-project with "Contestant" user
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/contest-projects/1":
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """

  Scenario: Not found get contest-project with "AdminUser" user
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/contest-projects/99":
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "contest project not found",
        "details": null
      }
      """
