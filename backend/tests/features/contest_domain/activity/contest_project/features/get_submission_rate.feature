Feature: Submission Rate API
  As an admin
  I want the recent submission rate
  So that the status page can show incoming load building before the queue does

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

  # Every minute in the window is present whether or not anything arrived in it.
  # A caller plotting the series cannot tell a missing bucket from a zero one,
  # and a series with holes renders as a line that skips the quiet periods
  # instead of dropping to zero across them.
  Scenario: A quiet system still returns one point per minute
    Given I am logged in as "user01"
    Then the submission rate over 5 minutes should have 5 points totalling 0

  Scenario: Recent submissions are counted
    Given I have submissions:
      | id   | userID | contestProjectID |
      | sub1 | 13     | 1                |
      | sub2 | 13     | 1                |
      | sub3 | 13     | 2                |
    And I am logged in as "user01"
    Then the submission rate over 5 minutes should have 5 points totalling 3

  Scenario: The window length decides how many points come back
    Given I am logged in as "user01"
    Then the submission rate over 1 minutes should have 1 points totalling 0
    And the submission rate over 30 minutes should have 30 points totalling 0

  Scenario: Filtering by contest project counts only that project
    Given I have submissions:
      | id   | userID | contestProjectID |
      | sub1 | 13     | 1                |
      | sub2 | 13     | 1                |
      | sub3 | 13     | 2                |
    And I am logged in as "user01"
    Then the submission rate over 5 minutes for contest project 1 should have 5 points totalling 2
    And the submission rate over 5 minutes for contest project 2 should have 5 points totalling 1

  # The response carries one point per minute whether or not anything happened,
  # so an unbounded window is an unbounded response.
  Scenario: Rejects a window outside the supported range
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/submission-rate?minutes=0":
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "Minutes must be between 1 and 1440",
        "details": null
      }
      """

  Scenario: Unauthorized access
    When I send a "GET" request to "/api/v1/submission-rate":
    Then the response code should be 401

  Scenario: Forbidden for contestant role
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/submission-rate":
    Then the response code should be 403
