Feature: Stream User Project Time API
  As a contestant
  I want to stream time for a contest project
  So that I can see the remaining time in real-time

  Background:
    Given I have users:
      | id | name       | username | password | roles      |
      | 11 | AdminUser  | user01   | 123456   | admin      |
      | 12 | APIReader  | user02   | 123456   | api_reader |
      | 13 | Contestant | user03   | 123456   | contestant |
    And I have images:
      | id | tag               | mount_path    | description | is_ready | createdAt           | updatedAt           |
      | 1  | nginx:latest      | /var/www/html | description | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 2  | test-nginx:latest | /var/www/html | description | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
    And I have projects:
      | id | name      | description | imageTag     | env | testImageTag      | testEnv | resultParser | createdAt           | updatedAt           |
      | 11 | Project-1 | Desc-1      | nginx:latest |     | test-nginx:latest |         |  | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
    And I have contests:
      | id | name      | description | isPublic | activateTime        | deactivateTime      | createdAt           | updatedAt           |
      | 11 | Contest-1 | Desc-1      | 0        | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
    And I have contest project associations:
      | id | contestId | projectId | name      | isPublic | judge priority strategy | activateTime        | deactivateTime      | startTime           | endTime             |
      | 1  | 11        | 11        | Project-1 | 0        | default                 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 |
    And I have contest participants:
      | contestId | userId | groupId |
      | 11        | 13     |         |
  Scenario: Successfully stream project time during submit period
    Given I am logged in as "user03"
    And when start streaming logs via request to "/api/v1/user/contest-projects/1/stream-time":
    Then streaming logs should have "time_type" message within 5 seconds

  Scenario: Unauthorized stream user project time
    When I send a "GET" request to "/api/v1/user/contest-projects/1/stream-time":
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "Unauthorized",
        "details": null
      }
      """

  Scenario: Forbidden stream user project time
    Given I am logged in as "user02"
    When I send a "GET" request to "/api/v1/user/contest-projects/1/stream-time":
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """

  Scenario: Not found stream user project time with non-existent ID
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/user/contest-projects/99/stream-time":
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "contest project not found",
        "details": null
      }
      """
