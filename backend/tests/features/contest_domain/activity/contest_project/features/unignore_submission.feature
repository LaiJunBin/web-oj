Feature: Unignore Submission API
  As an admin
  I want to unignore a submission
  So that it is included back in scoring

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
      | id   | userID | contestProjectID | ignored |
      | sub1 | 13     | 1                | true    |

  Scenario: Admin can unignore a submission
    Given I am logged in as "user01"
    When I send a "DELETE" request to "/api/v1/submissions/sub1/ignore":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {}
      }
      """

  Scenario: Unignore non-existent submission returns 500
    Given I am logged in as "user01"
    When I send a "DELETE" request to "/api/v1/submissions/nonexistent/ignore":
    Then the response code should be 500

  Scenario: Unauthorized access
    When I send a "DELETE" request to "/api/v1/submissions/sub1/ignore":
    Then the response code should be 401

  Scenario: Forbidden for contestant role
    Given I am logged in as "user03"
    When I send a "DELETE" request to "/api/v1/submissions/sub1/ignore":
    Then the response code should be 403
