Feature: List Rejudges API
  As an admin
  I want to list all rejudges
  So that I can monitor rejudge operations

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
  Scenario: Admin can list rejudges
    Given I have rejudges:
      | id       | contestProjectID | status  | reason      | judgments |
      | rejudge1 | 1                | Pending | test reason | j1,j2     |
    And I am logged in as "user01"
    When I send a "GET" request to "/api/v1/rejudges":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": [
          {
            "id": "rejudge1",
            "contestProjectId": 1,
            "reason": "test reason",
            "status": "Pending",
            "effectedJudgments": 2,
            "completedJudgments": 0,
            "createdAt": "{createdAt}"
          }
        ]
      }
      """

  # effectedJudgments alone says nothing about how far along a rejudge is — a
  # rejudge covering 300 submissions looks identical on its first second and its
  # last. The pair is the progress bar, so the completed side has to count only
  # the judgments that have actually finished.
  Scenario: A rejudge reports how many of its judgments have finished
    Given I have judgments:
      | id      | submissionID | contestProjectID | judgeType | status   | score | totalScore | rejudgeID | createdAt           |
      | done-1  | sub1         | 1                | REJUDGE   | Rejudged | 30    | 100        | rejudge1  | 2021-09-02 00:00:00 |
      | done-2  | sub2         | 1                | REJUDGE   | Error    | 0     | 100        | rejudge1  | 2021-09-02 00:00:00 |
      | busy-1  | sub3         | 1                | REJUDGE   | Judging  | 0     | 100        | rejudge1  | 2021-09-02 00:00:00 |
      | queued-1 | sub4        | 1                | REJUDGE   | Queued   | 0     | 100        | rejudge1  | 2021-09-02 00:00:00 |
    And I have rejudges:
      | id       | contestProjectID | status  | reason      | judgments                        | createdAt           |
      | rejudge1 | 1                | Running | in progress | done-1,done-2,busy-1,queued-1    | 2021-09-02 00:00:00 |
    And I am logged in as "user01"
    When I send a "GET" request to "/api/v1/rejudges":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": [
          {
            "id": "rejudge1",
            "contestProjectId": 1,
            "reason": "in progress",
            "status": "Running",
            "effectedJudgments": 4,
            "completedJudgments": 2,
            "createdAt": "{createdAt}"
          }
        ]
      }
      """

  # A rejudge that produced no judgments used to vanish from this listing: the
  # aggregation unwound the judgments array to count it, and an empty array
  # unwinds to nothing. That is not only the harmless "project had no
  # submissions" case — when every judgment fails to be created the request
  # handler still saves the rejudge, and nothing advances its status afterwards
  # because the rollup is driven by judgments finishing. The admin was left with
  # a request that returned 200, did nothing, and left no trace. Reported with a
  # count of zero it is at least visible, and a Pending rejudge can be cancelled.
  Scenario: A rejudge that produced no judgments is still listed
    Given I have rejudges:
      | id       | contestProjectID | status  | reason        | judgments | createdAt           |
      | empty1   | 1                | Pending | nothing to do |           | 2021-09-02 00:00:00 |
    And I am logged in as "user01"
    When I send a "GET" request to "/api/v1/rejudges":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": [
          {
            "id": "empty1",
            "contestProjectId": 1,
            "reason": "nothing to do",
            "status": "Pending",
            "effectedJudgments": 0,
            "completedJudgments": 0,
            "createdAt": "{createdAt}"
          }
        ]
      }
      """

  # The admin console groups this list by contest project, so every entry has to
  # name its own — sharing one project across the whole list, or defaulting to
  # zero, would file rejudges under the wrong project.
  Scenario: Each rejudge reports the contest project it was raised against
    Given I have rejudges:
      | id       | contestProjectID | status   | reason        | judgments | createdAt           |
      | rejudge1 | 1                | Pending  | first project | j1        | 2021-09-02 00:00:00 |
      | rejudge2 | 2                | Finished | other project | j2,j3     | 2021-09-01 00:00:00 |
    And I am logged in as "user01"
    When I send a "GET" request to "/api/v1/rejudges":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": [
          {
            "id": "rejudge1",
            "contestProjectId": 1,
            "reason": "first project",
            "status": "Pending",
            "effectedJudgments": 1,
            "completedJudgments": 0,
            "createdAt": "{createdAt}"
          },
          {
            "id": "rejudge2",
            "contestProjectId": 2,
            "reason": "other project",
            "status": "Finished",
            "effectedJudgments": 2,
            "completedJudgments": 0,
            "createdAt": "{createdAt}"
          }
        ]
      }
      """

  Scenario: Returns empty list when no rejudges exist
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/rejudges":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": []
      }
      """

  Scenario: Unauthorized access
    When I send a "GET" request to "/api/v1/rejudges":
    Then the response code should be 401

  Scenario: Forbidden for contestant role
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/rejudges":
    Then the response code should be 403
