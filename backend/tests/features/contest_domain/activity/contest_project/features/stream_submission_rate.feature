Feature: Submission Rate Stream
  As an admin
  I want the submission rate pushed to me
  So that the status page shows incoming load building without polling for it

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

  Scenario: The connection is answered with a snapshot straight away
    Given I have submissions:
      | id   | userID | contestProjectID |
      | sub1 | 13     | 1                |
    And I am logged in as "user01"
    And when start streaming logs via request to "/api/v1/submission-rate/stream?minutes=5":
    Then the streamed submission rate should total 1 within 10 seconds

  Scenario: A quiet system still emits snapshots
    Given I am logged in as "user01"
    And when start streaming logs via request to "/api/v1/submission-rate/stream?minutes=5":
    Then the streamed submission rate should total 0 within 10 seconds
    And streaming logs should have "minute" message within 10 seconds

  # The point of the stream. The tick is deliberately slow here — the buckets
  # only have minute resolution — so the window allows for more than one.
  Scenario: A submission made after connecting arrives on a later tick
    Given I am logged in as "user01"
    And when start streaming logs via request to "/api/v1/submission-rate/stream?minutes=5":
    Then the streamed submission rate should total 0 within 10 seconds
    When I have submissions:
      | id   | userID | contestProjectID |
      | late | 13     | 1                |
    Then the streamed submission rate should total 1 within 40 seconds

  Scenario: Filtering by contest project counts only that project
    Given I have submissions:
      | id   | userID | contestProjectID |
      | sub1 | 13     | 1                |
      | sub2 | 13     | 1                |
      | sub3 | 13     | 2                |
    And I am logged in as "user01"
    And when start streaming logs via request to "/api/v1/submission-rate/stream?minutes=5&contestProjectId=2":
    Then the streamed submission rate should total 1 within 10 seconds

  Scenario: Rejects a window outside the supported range at connect time
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/submission-rate/stream?minutes=0":
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
    When I send a "GET" request to "/api/v1/submission-rate/stream":
    Then the response code should be 401

  Scenario: Forbidden for contestant role
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/submission-rate/stream":
    Then the response code should be 403
