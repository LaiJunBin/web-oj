Feature: Create Contest API
  As a user
  I want to create a contest to a grading-item
  So that I can create a contest if I have permission

  Background:
    Given I have users:
      | id | name               | username | password | roles            |
      | 11 | AdminUser          | user01   | 123456   | admin            |
      | 12 | APIReader          | user02   | 123456   | api_reader       |
      | 13 | Contestant         | user03   | 123456   | contestant       |
      | 14 | AdminAndContestant | user04   | 123456   | admin,contestant |
    And I have images:
      | id | tag               | mount_path    | description  | is_ready | createdAt           | updatedAt           |
      | 1  | nginx:latest      | /var/www/html | description1 | 0        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 2  | mysql:8.0.26      | /var/www/html | description2 | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 3  | test-nginx:latest | /var/www/html | description3 | 0        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 4  | test-mysql:8.0.26 | /var/www/html | description4 | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 5  | test-nginx:test-2 | /var/www/html | description5 | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
    And I have projects:
      | id | name           | description               | imageTag       | env | testImageTag      | testEnv | resultParser | createdAt           | updatedAt           |
      | 11 | Test-Project-1 | Test-Project-1 Description| nginx:latest   | a=b | test-nginx:latest |         |  | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 12 | Test-Project-2 | Test-Project-2 Description| mysql:8.0.26   | c=d | test-mysql:8.0.26 |         |  | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 13 | Test-Project-3 | Test-Project-3 Description| nginx:latest   |     | test-nginx:test-2 |         |  | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
  Scenario: Successful create contest with "AdminUser" user
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/contests" with the following data:
      """
      {
        "name": "new contest",
        "description": "new contest description",
        "activateTime": "2025-02-03T00:00:00Z",
        "deactivateTime": "2025-02-03T00:10:00Z"
      }
      """
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": 1
      }
      """
    When I send a "GET" request to "/api/v1/contests/1":
    Then the response code should be 200
    And the response should match json:
      """
      {
          "success": true,
          "data": {
              "id": 1,
              "name": "new contest",
              "description": "new contest description",
              "isPublic": false,
              "projects": [],
              "activateTime": "{activateTime}",
              "deactivateTime": "{deactivateTime}",
              "createdAt": "{createdAt}",
              "updatedAt": "{updatedAt}"
          }
      }
      """
  
  Scenario: Successful create contest with projects and "AdminAndContestant" user
    Given I am logged in as "user04"
    When I send a "POST" request to "/api/v1/contests" with the following data:
      """
      {
        "name": "new contest",
        "activateTime": "2025-02-03T00:00:00Z",
        "deactivateTime": "2025-02-03T00:10:00Z",
        "projects": [{
          "projectId": 11,
          "isPublic": true,
          "name": "Test-Project-1",
          "activateTime": "2025-02-03T00:00:00Z",
          "deactivateTime": "2025-02-03T00:10:00Z",
          "startTime": "2025-02-03T00:00:00Z",
          "endTime": "2025-02-03T00:10:00Z"
        },
        {
          "projectId": 12,
          "name": "Test-Project-2",
          "activateTime": "2025-02-03T00:00:00Z",
          "deactivateTime": "2025-02-03T00:10:00Z",
          "startTime": "2025-02-03T00:00:00Z",
          "endTime": "2025-02-03T00:10:00Z"
        }]
      }
      """
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": 1
      }
      """
    When I send a "GET" request to "/api/v1/contests/1":
    Then the response code should be 200
    And the response should match json:
      """
      {
          "success": true,
          "data": {
              "id": 1,
              "name": "new contest",
              "description": "",
              "isPublic": false,
              "projects": [
                  {
                      "id": 1,
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
                  },
                  {
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
              ],
              "activateTime": "{activateTime}",
              "deactivateTime": "{deactivateTime}",
              "createdAt": "{createdAt}",
              "updatedAt": "{updatedAt}"
          }
      }
      """

  Scenario: Failed create contest with project not found
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/contests" with the following data:
      """
      {
        "name": "new contest",
        "activateTime": "2025-02-03T00:00:00Z",
        "deactivateTime": "2025-02-03T00:10:00Z",
        "projects": [{
          "projectId": 111,
          "activateTime": "2025-02-03T00:00:00Z",
          "deactivateTime": "2025-02-03T00:10:00Z",
          "startTime": "2025-02-03T00:00:00Z",
          "endTime": "2025-02-03T00:10:00Z"
        }]
      }
      """
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "project with id 111 not found",
        "details": null
      }
      """
  
  Scenario: Failed create contest with project missing projectId
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/contests" with the following data:
      """
      {
        "name": "new contest",
        "activateTime": "2025-02-03T00:00:00Z",
        "deactivateTime": "2025-02-03T00:10:00Z",
        "projects": [{
          "activateTime": "2025-02-03T00:00:00Z",
          "deactivateTime": "2025-02-03T00:10:00Z",
          "startTime": "2025-02-03T00:00:00Z",
          "endTime": "2025-02-03T00:10:00Z"
        }]
      }
      """
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "invalid project id",
        "details": [
          "project index: 0"
        ]
      }
      """

  Scenario: Failed create contest with project invalid input time
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/contests" with the following data:
      """
      {
        "name": "new contest",
        "activateTime": "2025-02-03T00:00:00Z",
        "deactivateTime": "2025-02-03T00:10:00Z",
        "projects": [{
          "projectId": 11,
          "deactivateTime": "2025-02-03T00:10:00Z",
          "startTime": "2025-02-03T00:00:00Z",
          "endTime": "2025-02-03T00:10:00Z"
        }]
      }
      """
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "invalid activation period",
        "details": [
          "project index: 0"
        ]
      }
      """

  Scenario: Failed create contest with deactivation time before activation time
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/contests" with the following data:
      """
      {
        "name": "new contest",
        "activateTime": "2025-02-03T00:00:01Z",
        "deactivateTime": "2025-02-03T00:00:00Z",
        "projects": []
      }
      """
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "invalid activation period",
        "details": null
      }
      """

  Scenario: Unauthorized create contest
    When I send a "POST" request to "/api/v1/contests" with the following data:
      """
      {
        "name": "new contest",
        "description": "new contest description",
        "activateTime": "2025-02-03T00:00:00Z",
        "deactivateTime": "2025-02-03T00:10:00Z"
      }
      """
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "Unauthorized",
        "details": null
      }
      """
    
  Scenario: Forbidden create contest with "APIReader" user
    Given I am logged in as "user02"
    When I send a "POST" request to "/api/v1/contests" with the following data:
      """
      {
        "name": "new contest",
        "description": "new contest description",
        "activateTime": "2025-02-03T00:00:00Z",
        "deactivateTime": "2025-02-03T00:10:00Z"
      }
      """
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """

  Scenario: Forbidden create contest with "Contestant" user
    Given I am logged in as "user03"
    When I send a "POST" request to "/api/v1/contests" with the following data:
      """
      {
        "name": "new contest",
        "description": "new contest description",
        "activateTime": "2025-02-03T00:00:00Z",
        "deactivateTime": "2025-02-03T00:10:00Z"
      }
      """
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """
