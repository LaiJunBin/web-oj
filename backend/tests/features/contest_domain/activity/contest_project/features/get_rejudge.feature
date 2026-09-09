Feature: Get Rejudge API
  As an admin
  I want to get details of a rejudge operation
  So that I can see the before and after comparison of submissions

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
  Scenario: Admin can get a rejudge detail
    Given I have rejudges:
      | id       | status   | reason      | judgments |
      | rejudge1 | Finished | test reason |           |
    And I am logged in as "user01"
    When I send a "GET" request to "/api/v1/rejudges/rejudge1":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": "rejudge1",
          "reason": "test reason",
          "status": "Finished",
          "createdAt": "{createdAt}",
          "submissions": []
        }
      }
      """

  # Each affected submission must be shown against its OWN previous score. Two
  # submissions where only one has a previous judgment is the minimal shape that
  # catches a positional pairing: the two arrays being zipped have different
  # lengths, so at least one submission is guaranteed to be handed a "before"
  # that is not its own no matter what order the database returns them in.
  #
  # This scenario puts the submission WITH a previous score first, so the second
  # one runs off the end of the previous-scores array and reports a fabricated
  # 0 instead of "no previous score".
  Scenario: A submission with no previous judgment reports no baseline, not zero
    Given I have submissions:
      | id    | userID | contestProjectID |
      | sub-a | 13     | 1                |
      | sub-b | 13     | 1                |
    And I have judgments:
      | id       | submissionID | judgeType | status  | score | totalScore | createdAt           |
      | judge-a  | sub-a        | JUDGE     | Judged  | 100   | 100        | 2021-09-01 00:00:00 |
      | judge-b  | sub-b        | JUDGE     | Error   | 0     | 100        | 2021-09-01 00:00:00 |
    And I have judgments:
      | id         | submissionID | judgeType | status   | score | totalScore | rejudgeID | createdAt           |
      | rejudged-a | sub-a        | REJUDGE   | Rejudged | 50    | 100        | rejudge2  | 2021-09-02 00:00:00 |
      | rejudged-b | sub-b        | REJUDGE   | Rejudged | 80    | 100        | rejudge2  | 2021-09-02 00:00:00 |
    And I have rejudges:
      | id       | status   | reason        | judgments             |
      | rejudge2 | Finished | fixed testcase | rejudged-a,rejudged-b |
    And I am logged in as "user01"
    When I send a "GET" request to "/api/v1/rejudges/rejudge2":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": "rejudge2",
          "reason": "fixed testcase",
          "status": "Finished",
          "createdAt": "{createdAt}",
          "submissions": [
            {
              "id": "sub-a",
              "before": { "status": "Judged", "score": 100, "totalScore": 100 },
              "after": { "status": "Rejudged", "score": 50, "totalScore": 100 }
            },
            {
              "id": "sub-b",
              "before": null,
              "after": { "status": "Rejudged", "score": 80, "totalScore": 100 }
            }
          ]
        }
      }
      """

  # Same shape as above with the order flipped: the submission with NO previous
  # score comes first, so a positional pairing hands it the other contestant's
  # previous score outright. This is the case that misleads an admin into
  # approving or rejecting on a comparison that belongs to somebody else.
  Scenario: A previous score is never attributed to another submission
    Given I have submissions:
      | id    | userID | contestProjectID |
      | sub-c | 13     | 1                |
      | sub-d | 13     | 1                |
    And I have judgments:
      | id      | submissionID | judgeType | status | score | totalScore | createdAt           |
      | judge-c | sub-c        | JUDGE     | Error  | 0     | 100        | 2021-09-01 00:00:00 |
      | judge-d | sub-d        | JUDGE     | Judged | 100   | 100        | 2021-09-01 00:00:00 |
    And I have judgments:
      | id         | submissionID | judgeType | status   | score | totalScore | rejudgeID | createdAt           |
      | rejudged-c | sub-c        | REJUDGE   | Rejudged | 30    | 100        | rejudge3  | 2021-09-02 00:00:00 |
      | rejudged-d | sub-d        | REJUDGE   | Rejudged | 70    | 100        | rejudge3  | 2021-09-02 00:00:00 |
    And I have rejudges:
      | id       | status   | reason      | judgments             |
      | rejudge3 | Finished | second pass | rejudged-c,rejudged-d |
    And I am logged in as "user01"
    When I send a "GET" request to "/api/v1/rejudges/rejudge3":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": "rejudge3",
          "reason": "second pass",
          "status": "Finished",
          "createdAt": "{createdAt}",
          "submissions": [
            {
              "id": "sub-c",
              "before": null,
              "after": { "status": "Rejudged", "score": 30, "totalScore": 100 }
            },
            {
              "id": "sub-d",
              "before": { "status": "Judged", "score": 100, "totalScore": 100 },
              "after": { "status": "Rejudged", "score": 70, "totalScore": 100 }
            }
          ]
        }
      }
      """

  # A submission can be rejudged more than once. Once the first rejudge is
  # approved its judgment becomes the score the contestant actually carries, so
  # the second rejudge has to be compared against THAT and not against the
  # original judge run — otherwise the diff overstates the change by folding in
  # a delta the admin already approved.
  Scenario: An approved earlier rejudge is the baseline for the next one
    Given I have submissions:
      | id    | userID | contestProjectID |
      | sub-e | 13     | 1                |
    And I have judgments:
      | id          | submissionID | judgeType | status   | score | totalScore | rejudgeID | createdAt           |
      | judge-e     | sub-e        | JUDGE     | Judged   | 10    | 100        |           | 2021-09-01 00:00:00 |
      | rejudged-e1 | sub-e        | REJUDGE   | Judged   | 40    | 100        | rejudge4  | 2021-09-02 00:00:00 |
      | rejudged-e2 | sub-e        | REJUDGE   | Rejudged | 70    | 100        | rejudge5  | 2021-09-03 00:00:00 |
    And I have rejudges:
      | id       | status   | reason      | judgments   |
      | rejudge4 | Approved | first pass  | rejudged-e1 |
      | rejudge5 | Finished | second pass | rejudged-e2 |
    And I am logged in as "user01"
    When I send a "GET" request to "/api/v1/rejudges/rejudge5":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": "rejudge5",
          "reason": "second pass",
          "status": "Finished",
          "createdAt": "{createdAt}",
          "submissions": [
            {
              "id": "sub-e",
              "before": { "status": "Judged", "score": 40, "totalScore": 100 },
              "after": { "status": "Rejudged", "score": 70, "totalScore": 100 }
            }
          ]
        }
      }
      """

  # Approving a rejudge flips its own judgment to Judged, which makes it a
  # candidate baseline for itself. Its comparison has to stay anchored to what it
  # replaced, so that reopening the page after approval still shows the change
  # rather than an empty 40 → 40.
  Scenario: An approved rejudge is not its own baseline
    Given I have submissions:
      | id    | userID | contestProjectID |
      | sub-e | 13     | 1                |
    And I have judgments:
      | id          | submissionID | judgeType | status   | score | totalScore | rejudgeID | createdAt           |
      | judge-e     | sub-e        | JUDGE     | Judged   | 10    | 100        |           | 2021-09-01 00:00:00 |
      | rejudged-e1 | sub-e        | REJUDGE   | Judged   | 40    | 100        | rejudge4  | 2021-09-02 00:00:00 |
    And I have rejudges:
      | id       | status   | reason     | judgments   |
      | rejudge4 | Approved | first pass | rejudged-e1 |
    And I am logged in as "user01"
    When I send a "GET" request to "/api/v1/rejudges/rejudge4":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": "rejudge4",
          "reason": "first pass",
          "status": "Approved",
          "createdAt": "{createdAt}",
          "submissions": [
            {
              "id": "sub-e",
              "before": { "status": "Judged", "score": 10, "totalScore": 100 },
              "after": { "status": "Judged", "score": 40, "totalScore": 100 }
            }
          ]
        }
      }
      """

  Scenario: Rejudge not found
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/rejudges/nonexistent":
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "rejudge not found",
        "details": null
      }
      """

  Scenario: Unauthorized access
    When I send a "GET" request to "/api/v1/rejudges/rejudge1":
    Then the response code should be 401

  Scenario: Forbidden for contestant role
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/rejudges/rejudge1":
    Then the response code should be 403
