Feature: Get Contest API
  As a user
  I want to get a contest
  So that I can see a contest if contest activated and I have access

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
      | 11 | Test-Contest-1 | Test-Contest-1 Description | 0        | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 12 | Test-Contest-2 | Test-Contest-2 Description | 1        | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 13 | Test-Contest-3 | Test-Contest-3 Description | 0        | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 14 | Test-Contest-4 | Test-Contest-4 Description | 1        | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 15 | Test-Contest-5 | Test-Contest-5 Description | 1        | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
    And I have contest project associations:
      | id | contestId | projectId | name           | isPublic | judge priority strategy | activateTime        | deactivateTime      | startTime           | endTime             |
      | 1  | 11        | 11        | Test-Project-1 | 0        | default                 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 |
      | 2  | 11        | 12        | Test-Project-2 | 0        | default                 | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 |
      | 3  | 12        | 11        | Test-Project-1 | 1        | default                 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 |
      | 4  | 12        | 12        | Test-Project-2 | 1        | default                 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2099-09-01 00:00:00 | 2099-09-30 23:59:59 |
      | 5  | 12        | 13        | Test-Project-3 | 1        | default                 | 2099-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 |
      | 6  | 13        | 11        | Test-Project-1 | 1        | default                 | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 |
      | 7  | 15        | 12        | Test-Project-2 | 1        | default                 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 |

  Scenario: Successful get public contest without being a participant
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/user/contests/12":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": 12,
          "name": "Test-Contest-2",
          "projects": [
            {
              "id": 3,
              "name": "Test-Project-1",
              "status": "Ended"
            },
            {
              "id": 4,
              "name": "Test-Project-2",
              "status": "NotStarted"
            }
          ]
        }
      }
      """

  Scenario: Successful get public contest with project in progress
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/user/contests/15":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": 15,
          "name": "Test-Contest-5",
          "projects": [
            {
              "id": 7,
              "name": "Test-Project-2",
              "status": "InProgress"
            }
          ]
        }
      }
      """

  Scenario: Successful get non-public contest as a direct participant
    Given I have contest participants:
      | contestId | userId | groupId |
      | 11        | 13     |         |
    And I am logged in as "user03"
    When I send a "GET" request to "/api/v1/user/contests/11":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": 11,
          "name": "Test-Contest-1",
          "projects": [
            {
              "id": 1,
              "name": "Test-Project-1",
              "status": "Ended"
            }
          ]
        }
      }
      """

  Scenario: Successful get non-public contest as a group participant
    Given I have groups:
      | id | name    | description |
      | 1  | Group-1 | desc        |
    And I have user groups:
      | userId | groupId |
      | 13     | 1       |
    And I have contest participants:
      | contestId | userId | groupId |
      | 11        |        | 1       |
    And I am logged in as "user03"
    When I send a "GET" request to "/api/v1/user/contests/11":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": 11,
          "name": "Test-Contest-1",
          "projects": [
            {
              "id": 1,
              "name": "Test-Project-1",
              "status": "Ended"
            }
          ]
        }
      }
      """

  Scenario: Access denied for non-public contest when not a participant
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/user/contests/11":
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "access denied",
        "details": null
      }
      """

  Scenario: Unauthorized get contest
    When I send a "GET" request to "/api/v1/user/contests/11":
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "Unauthorized",
        "details": null
      }
      """

  Scenario: Forbidden get contest for api_reader
    Given I am logged in as "user02"
    When I send a "GET" request to "/api/v1/user/contests/11":
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """

  Scenario: Not found for expired contest
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/user/contests/14":
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "contest not found",
        "details": null
      }
      """

  Scenario: Not found for non-existent contest
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/user/contests/99":
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "contest not found",
        "details": null
      }
      """
