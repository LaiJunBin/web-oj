Feature: Cancel Rejudge API
  As an admin
  I want to cancel an in-progress rejudge
  So that I can stop the rejudge process

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
  Scenario: Admin can cancel a pending rejudge
    Given I have rejudges:
      | id       | status  | reason      | judgments |
      | rejudge1 | Pending | test reason | j1        |
    And I am logged in as "user01"
    When I send a "POST" request to "/api/v1/rejudges/rejudge1/cancel":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {}
      }
      """

  Scenario: Admin can cancel a running rejudge
    Given I have rejudges:
      | id       | status  | reason      | judgments |
      | rejudge1 | Running | test reason | j1        |
    And I am logged in as "user01"
    When I send a "POST" request to "/api/v1/rejudges/rejudge1/cancel":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {}
      }
      """

  Scenario: Cannot cancel a finished rejudge
    Given I have rejudges:
      | id       | status   | reason      | judgments |
      | rejudge1 | Finished | test reason | j1        |
    And I am logged in as "user01"
    When I send a "POST" request to "/api/v1/rejudges/rejudge1/cancel":
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "Judgment cannot be cancelled",
        "details": null
      }
      """

  Scenario: Unauthorized access
    When I send a "POST" request to "/api/v1/rejudges/rejudge1/cancel":
    Then the response code should be 401

  Scenario: Forbidden for contestant role
    Given I am logged in as "user03"
    When I send a "POST" request to "/api/v1/rejudges/rejudge1/cancel":
    Then the response code should be 403
