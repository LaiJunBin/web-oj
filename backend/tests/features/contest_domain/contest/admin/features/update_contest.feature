Feature: Update Contest API
  As a user
  I want to update a contest
  So that I can update a contest if I have permission

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
  Scenario: Successful update contest with "AdminUser" user
    Given I am logged in as "user01"
    When I send a "PATCH" request to "/api/v1/contests/11" with the following data:
      """
      {
        "name": "updated contest",
        "description": "updated contest description",
        "activateTime": "2025-02-03T00:00:00Z",
        "deactivateTime": "2025-02-03T00:10:00Z",
        "isPublic": true
      }
      """
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true
      }
      """
    When I send a "GET" request to "/api/v1/contests/11":
    Then the response code should be 200
    And the response should match json:
      """
      {
          "success": true,
          "data": {
              "id": 11,
              "name": "updated contest",
              "description": "updated contest description",
              "isPublic": true,
              "projects": [
                {
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
  
  Scenario: Successful update contest with append new projects with "AdminAndContestant" user
    Given I am logged in as "user04"
    When I send a "PATCH" request to "/api/v1/contests/13" with the following data:
      """
      {
        "projects": [{
          "projectId": 11,
          "name": "Test-Project-1",
          "isPublic": true,
          "activateTime": "2025-02-03T00:00:00Z",
          "deactivateTime": "2025-02-03T00:10:00Z",
          "startTime": "2025-02-03T00:00:00Z",
          "endTime": "2025-02-03T00:10:00Z",
          "judgePriorityStrategy": "default"
        }, {
          "projectId": 12,
          "name": "Test-Project-2",
          "activateTime": "2025-02-03T00:00:00Z",
          "deactivateTime": "2025-02-03T00:10:00Z",
          "startTime": "2025-02-03T00:00:00Z",
          "endTime": "2025-02-03T00:10:00Z",
          "judgePriorityStrategy": "default"
        }]
      }
      """
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true
      }
      """
    When I send a "GET" request to "/api/v1/contests/13":
    Then the response code should be 200
    And the response should match json:
      """
      {
          "success": true,
          "data": {
              "id": 13,
              "name": "Test-Contest-3",
              "description": "Test-Contest-3 Description",
              "isPublic": false,
              "projects": [
                {
                    "id": 8,
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
                    "id": 9,
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
  
  Scenario: Successful update contest with exists project
    Given I am logged in as "user01"
    When I send a "PATCH" request to "/api/v1/contests/12" with the following data:
      """
      {
        "projects": [{
          "id": 3,
          "projectId": 11,
          "name": "Test-Project-1",
          "isPublic": false,
          "activateTime": "2025-02-03T00:00:00Z",
          "deactivateTime": "2025-02-03T00:10:00Z",
          "startTime": "2025-02-03T00:00:00Z",
          "endTime": "2025-02-03T00:10:00Z"
        }, {
          "id": 5,
          "projectId": 11,
          "name": "Test-Project-1",
          "activateTime": "2025-02-03T00:00:00Z",
          "deactivateTime": "2025-02-03T00:10:00Z",
          "startTime": "2025-02-03T00:00:00Z",
          "endTime": "2025-02-03T00:10:00Z"
        }, {
          "projectId": 12,
          "name": "Test-Project-2",
          "activateTime": "2025-02-03T00:00:00Z",
          "deactivateTime": "2025-02-03T00:10:00Z",
          "startTime": "2025-02-03T00:00:00Z",
          "endTime": "2025-02-03T00:10:00Z",
          "judgePriorityStrategy": "default"
        }]
      }
      """
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true
      }
      """
    When I send a "GET" request to "/api/v1/contests/12":
    Then the response code should be 200
    And the response should match json:
      """
      {
          "success": true,
          "data": {
              "id": 12,
              "name": "Test-Contest-2",
              "description": "Test-Contest-2 Description",
              "isPublic": true,
              "projects": [
                {
                    "id": 3,
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
                },
                {
                    "id": 5,
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
                    "id": 8,
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
      
  Scenario: Failed update contest with project not found
    Given I am logged in as "user01"
    When I send a "PATCH" request to "/api/v1/contests/11" with the following data:
      """
      {
        "projects": [{
          "ID": 111,
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

  Scenario: Failed create contest with deactivation time before activation time
    Given I am logged in as "user01"
    When I send a "PATCH" request to "/api/v1/contests/11" with the following data:
      """
      {
        "activateTime": "2025-02-03T00:00:00Z",
        "deactivateTime": "2025-02-02T00:10:00Z"
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

  Scenario: Unauthorized update contest
    When I send a "PATCH" request to "/api/v1/contests/11" with the following data:
      """
      {
        "name": "updated contest"
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
    
  Scenario: Forbidden update contest with "APIReader" user
    Given I am logged in as "user02"
    When I send a "PATCH" request to "/api/v1/contests/11" with the following data:
      """
      {
        "name": "updated contest"
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

  Scenario: Forbidden update contest with "Contestant" user
    Given I am logged in as "user03"
    When I send a "PATCH" request to "/api/v1/contests/11" with the following data:
      """
      {
        "name": "updated contest"
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

  Scenario: Failed update contest with contest not found
    Given I am logged in as "user01"
    When I send a "PATCH" request to "/api/v1/contests/111" with the following data:
      """
      {
        "name": "updated contest"
      }
      """
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "contest not found",
        "details": null
      }
      """

  # Every other assertion in this file reads submissionCooldownSeconds back as
  # null, which would still pass if the write path were not wired at all. This
  # one sets a real value and reads it back, and is the only test that proves
  # the field survives controller → usecase → entity → repository.
  Scenario: Submission cooldown survives a round trip through the update API
    Given I am logged in as "user01"
    When I send a "PATCH" request to "/api/v1/contests/13" with the following data:
      """
      {
        "projects": [{
          "id": 6,
          "projectId": 11,
          "name": "Test-Project-1",
          "activateTime": "2025-02-03T00:00:00Z",
          "deactivateTime": "2025-02-03T00:10:00Z",
          "startTime": "2025-02-03T00:00:00Z",
          "endTime": "2025-02-03T00:10:00Z",
          "submissionCooldownSeconds": 45
        }]
      }
      """
    Then the response code should be 200
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
          "description": "{description}",
          "isPublic": true,
          "judgePriorityStrategy": "default",
          "image": "{image}",
          "testImage": "{testImage}",
          "activateTime": "{activateTime}",
          "deactivateTime": "{deactivateTime}",
          "startTime": "{startTime}",
          "endTime": "{endTime}",
          "maxJudgeDurationMinutes": null,
          "submissionCooldownSeconds": 45,
          "pidsLimit": null,
          "memoryLimitMB": null,
          "scoreboardCompact": false,
          "scoreboardMaskNames": false
        }
      }
      """
    # The admin edit page loads from the contest detail, not from
    # /contest-projects — a value readable through one and not the other shows
    # the field as blank while it is actually set, and "blank = keep current"
    # then makes any attempt to confirm it overwrite the setting.
    When I send a "GET" request to "/api/v1/contests/13":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": 13,
          "name": "Test-Contest-3",
          "description": "Test-Contest-3 Description",
          "isPublic": false,
          "activateTime": "{activateTime}",
          "deactivateTime": "{deactivateTime}",
          "createdAt": "{createdAt}",
          "updatedAt": "{updatedAt}",
          "projects": [
            {
              "id": 6,
              "name": "Test-Project-1",
              "projectId": 11,
              "description": "{description}",
              "isPublic": true,
              "judgePriorityStrategy": "default",
              "image": "{image}",
              "testImage": "{testImage}",
              "activateTime": "{activateTime}",
              "deactivateTime": "{deactivateTime}",
              "startTime": "{startTime}",
              "endTime": "{endTime}",
              "maxJudgeDurationMinutes": null,
              "submissionCooldownSeconds": 45,
              "pidsLimit": null,
              "memoryLimitMB": null,
              "scoreboardCompact": false,
              "scoreboardMaskNames": false
            }
          ]
        }
      }
      """

  # Same lesson as the cooldown above, and the same reason it is worth its own
  # scenario: every other assertion reads these back as null, which passes just
  # as well when the write path was never wired. A container cap that silently
  # fails to save is worse than no cap — the contest is configured, the admin
  # page shows the number, and the limit is not there.
  Scenario: Container limits survive a round trip through the update API
    Given I am logged in as "user01"
    When I send a "PATCH" request to "/api/v1/contests/13" with the following data:
      """
      {
        "projects": [{
          "id": 6,
          "projectId": 11,
          "name": "Test-Project-1",
          "activateTime": "2025-02-03T00:00:00Z",
          "deactivateTime": "2025-02-03T00:10:00Z",
          "startTime": "2025-02-03T00:00:00Z",
          "endTime": "2025-02-03T00:10:00Z",
          "pidsLimit": 256,
          "memoryLimitMB": 512
        }]
      }
      """
    Then the response code should be 200
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
          "description": "{description}",
          "isPublic": true,
          "judgePriorityStrategy": "default",
          "image": "{image}",
          "testImage": "{testImage}",
          "activateTime": "{activateTime}",
          "deactivateTime": "{deactivateTime}",
          "startTime": "{startTime}",
          "endTime": "{endTime}",
          "maxJudgeDurationMinutes": null,
          "submissionCooldownSeconds": null,
          "pidsLimit": 256,
          "memoryLimitMB": 512,
          "scoreboardCompact": false,
          "scoreboardMaskNames": false
        }
      }
      """
    # The admin edit page reads the contest detail, not /contest-projects.
    # A value readable through one endpoint and not the other is the bug this
    # project has already shipped once.
    When I send a "GET" request to "/api/v1/contests/13":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": 13,
          "name": "Test-Contest-3",
          "description": "Test-Contest-3 Description",
          "isPublic": false,
          "activateTime": "{activateTime}",
          "deactivateTime": "{deactivateTime}",
          "createdAt": "{createdAt}",
          "updatedAt": "{updatedAt}",
          "projects": [
            {
              "id": 6,
              "name": "Test-Project-1",
              "projectId": 11,
              "description": "{description}",
              "isPublic": true,
              "judgePriorityStrategy": "default",
              "image": "{image}",
              "testImage": "{testImage}",
              "activateTime": "{activateTime}",
              "deactivateTime": "{deactivateTime}",
              "startTime": "{startTime}",
              "endTime": "{endTime}",
              "maxJudgeDurationMinutes": null,
              "submissionCooldownSeconds": null,
              "pidsLimit": 256,
              "memoryLimitMB": 512,
              "scoreboardCompact": false,
              "scoreboardMaskNames": false
            }
          ]
        }
      }
      """

  # Zero here means "this problem legitimately needs no cap", and it has to
  # beat a non-zero deployment default. Collapsing it into "unset" would leave
  # a heavy problem capped at the house limit with no way to say otherwise.
  Scenario: An explicit zero container limit is stored rather than treated as unset
    Given I am logged in as "user01"
    When I send a "PATCH" request to "/api/v1/contests/13" with the following data:
      """
      {
        "projects": [{
          "id": 6,
          "projectId": 11,
          "name": "Test-Project-1",
          "activateTime": "2025-02-03T00:00:00Z",
          "deactivateTime": "2025-02-03T00:10:00Z",
          "startTime": "2025-02-03T00:00:00Z",
          "endTime": "2025-02-03T00:10:00Z",
          "pidsLimit": 0,
          "memoryLimitMB": 0
        }]
      }
      """
    Then the response code should be 200
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
          "description": "{description}",
          "isPublic": true,
          "judgePriorityStrategy": "default",
          "image": "{image}",
          "testImage": "{testImage}",
          "activateTime": "{activateTime}",
          "deactivateTime": "{deactivateTime}",
          "startTime": "{startTime}",
          "endTime": "{endTime}",
          "maxJudgeDurationMinutes": null,
          "submissionCooldownSeconds": null,
          "pidsLimit": 0,
          "memoryLimitMB": 0,
          "scoreboardCompact": false,
          "scoreboardMaskNames": false
        }
      }
      """

  # A negative cap is not a weaker cap, it is a nonsensical one. Rejecting it
  # at the boundary keeps the "0 = unlimited" rule the only way to say off.
  Scenario: Failed update contest with a negative container limit
    Given I am logged in as "user01"
    When I send a "PATCH" request to "/api/v1/contests/13" with the following data:
      """
      {
        "projects": [{
          "id": 6,
          "projectId": 11,
          "name": "Test-Project-1",
          "activateTime": "2025-02-03T00:00:00Z",
          "deactivateTime": "2025-02-03T00:10:00Z",
          "startTime": "2025-02-03T00:00:00Z",
          "endTime": "2025-02-03T00:10:00Z",
          "pidsLimit": -1
        }]
      }
      """
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "pidsLimit must not be negative",
        "details": ["project index: 0"]
      }
      """

  # Zero is not "unset": it must persist so a contest can deliberately run
  # without a cooldown even when the system default is non-zero.
  Scenario: An explicit zero cooldown is stored rather than treated as unset
    Given I am logged in as "user01"
    When I send a "PATCH" request to "/api/v1/contests/13" with the following data:
      """
      {
        "projects": [{
          "id": 6,
          "projectId": 11,
          "name": "Test-Project-1",
          "activateTime": "2025-02-03T00:00:00Z",
          "deactivateTime": "2025-02-03T00:10:00Z",
          "startTime": "2025-02-03T00:00:00Z",
          "endTime": "2025-02-03T00:10:00Z",
          "submissionCooldownSeconds": 0
        }]
      }
      """
    Then the response code should be 200
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
          "description": "{description}",
          "isPublic": true,
          "judgePriorityStrategy": "default",
          "image": "{image}",
          "testImage": "{testImage}",
          "activateTime": "{activateTime}",
          "deactivateTime": "{deactivateTime}",
          "startTime": "{startTime}",
          "endTime": "{endTime}",
          "maxJudgeDurationMinutes": null,
          "submissionCooldownSeconds": 0,
          "pidsLimit": null,
          "memoryLimitMB": null,
          "scoreboardCompact": false,
          "scoreboardMaskNames": false
        }
      }
      """

