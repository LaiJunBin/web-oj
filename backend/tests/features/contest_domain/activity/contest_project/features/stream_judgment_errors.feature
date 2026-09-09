Feature: Judgment Errors Stream
  As an admin
  I want judging failures pushed to me
  So that the status page surfaces a break without waiting for the next poll

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
      | id   | userID | contestProjectID |
      | sub1 | 13     | 1                |
      | sub2 | 13     | 1                |

  # A client that had to wait a whole tick before seeing anything would render
  # an empty panel on every page load, which reads as "nothing is broken".
  Scenario: The connection is answered with a snapshot straight away
    Given I have judgments:
      | id      | submissionID | contestProjectID | judgeType | status | score | totalScore | errorMessage     | createdAt           |
      | broken  | sub1         | 1                | JUDGE     | Error  | 0     | 100        | image pull failed | 2021-09-02 00:00:00 |
    And I am logged in as "user01"
    And when start streaming logs via request to "/api/v1/judgment-errors/stream":
    Then streaming logs should have "image pull failed" message within 10 seconds

  # An idle system still has to produce frames, otherwise a client cannot tell a
  # healthy pipeline from a dead connection.
  Scenario: A system with nothing broken still emits snapshots
    Given I am logged in as "user01"
    And when start streaming logs via request to "/api/v1/judgment-errors/stream":
    Then streaming logs should have "success" message within 10 seconds

  # The point of the stream: a failure that happens after the page was opened
  # arrives on its own.
  Scenario: A failure recorded after connecting arrives on a later tick
    Given I am logged in as "user01"
    And when start streaming logs via request to "/api/v1/judgment-errors/stream":
    Then streaming logs should have "success" message within 10 seconds
    When I have judgments:
      | id       | submissionID | contestProjectID | judgeType | status | score | totalScore | errorMessage           | createdAt           |
      | late-one | sub2         | 1                | JUDGE     | Error  | 0     | 100        | disk full while judging | 2021-09-03 00:00:00 |
    Then streaming logs should have "disk full while judging" message within 30 seconds

  Scenario: The limit is honoured
    Given I have judgments:
      | id     | submissionID | contestProjectID | judgeType | status | score | totalScore | errorMessage | createdAt           |
      | older  | sub1         | 1                | JUDGE     | Error  | 0     | 100        | older break  | 2021-09-02 00:00:00 |
      | newer  | sub2         | 1                | JUDGE     | Error  | 0     | 100        | newer break  | 2021-09-03 00:00:00 |
    And I am logged in as "user01"
    And when start streaming logs via request to "/api/v1/judgment-errors/stream?limit=1":
    Then streaming logs should have "newer break" message within 10 seconds

  # Validated when the connection is made, not on the first tick: a stream that
  # opened 200 and then closed without a frame is indistinguishable, to a
  # client, from a system with nothing to report.
  Scenario: Rejects a limit outside the supported range at connect time
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/judgment-errors/stream?limit=0":
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "Limit must be between 1 and 200",
        "details": null
      }
      """

  Scenario: Unauthorized access
    When I send a "GET" request to "/api/v1/judgment-errors/stream":
    Then the response code should be 401

  Scenario: Forbidden for contestant role
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/judgment-errors/stream":
    Then the response code should be 403
