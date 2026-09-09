Feature: Submit User Project API
  As a contestant
  I want to submit my project to a contest
  So that it can be judged

  Background:
    Given I have users:
      | id | name       | username | password | roles      |
      | 11 | AdminUser  | user01   | 123456   | admin      |
      | 12 | APIReader  | user02   | 123456   | api_reader |
      | 13 | Contestant | user03   | 123456   | contestant |
      | 14 | Contestant2| user04   | 123456   | contestant |
    And I have images:
      | id | tag               | mount_path    | description | is_ready | createdAt           | updatedAt           |
      | 1  | nginx:latest      | /var/www/html | desc1       | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 2  | test-nginx:latest | /var/www/html | desc2       | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
    And I have projects:
      | id | name     | description | imageTag     | env | testImageTag      | testEnv | resultParser | createdAt           | updatedAt           |
      | 1  | Project1 | desc        | nginx:latest |     | test-nginx:latest |         |  | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 2  | Other    | desc        | nginx:latest |     | test-nginx:latest |         |  | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
    And I have grading-items:
      | id | projectId | name | description | score | order | scoringStrategy | outputFormat | resultParser | label | createdAt | updatedAt |
      | 1 | 1 | Task 1 | desc | 100 | 1 | increase | text |  |  | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 9 | 2 | Other Task | desc | 100 | 1 | increase | text |  |  | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
    And I have contests:
      | id | name    | description | isPublic | activateTime        | deactivateTime      | createdAt           | updatedAt           |
      | 1  | Contest | desc        | 1        | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
    And I have contest project associations:
      | id | contestId | projectId | name     | isPublic | judge priority strategy | activateTime        | deactivateTime      | startTime           | endTime             | submissionCooldownSeconds |
      | 1  | 1         | 1         | Project1 | 1        | default                 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 |                           |
      | 2  | 1         | 1         | Project2 | 1        | default                 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 60                        |
      | 3  | 1         | 1         | Future   | 1        | default                 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2098-01-01 00:00:00 | 2099-09-30 23:59:59 |                           |
      | 4  | 1         | 1         | Ended    | 1        | default                 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-02 00:00:00 |                           |
      | 5  | 1         | 1         | Inactive | 1        | default                 | 2098-01-01 00:00:00 | 2099-09-30 23:59:59 | 2098-01-01 00:00:00 | 2099-09-30 23:59:59 |                           |
      | 6  | 1         | 2         | Other    | 1        | default                 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 |                           |

  Scenario: Successfully submit a project as a direct participant
    Given I have contest participants:
      | contestId | userId | groupId |
      | 1         | 13     |         |
    And I am logged in as "user03"
    When I send a "POST" request to "/api/v1/user/contest-projects/1/submit" with the following formdata:
      | key     | value           |
      | project | {nginx project} |
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {}
      }
      """

  Scenario: Successfully submit a project as a group participant
    Given I have groups:
      | id | name    | description |
      | 1  | Group-1 | desc        |
    And I have user groups:
      | userId | groupId |
      | 13     | 1       |
    And I have contest participants:
      | contestId | userId | groupId |
      | 1         |        | 1       |
    And I am logged in as "user03"
    When I send a "POST" request to "/api/v1/user/contest-projects/1/submit" with the following formdata:
      | key     | value           |
      | project | {nginx project} |
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {}
      }
      """

  Scenario: Non-participant cannot submit even when contest is public
    Given I am logged in as "user03"
    When I send a "POST" request to "/api/v1/user/contest-projects/1/submit" with the following formdata:
      | key     | value           |
      | project | {nginx project} |
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "not a participant of this contest",
        "details": null
      }
      """

  Scenario: Submit to non-existent contest project
    Given I am logged in as "user03"
    When I send a "POST" request to "/api/v1/user/contest-projects/999/submit" with the following formdata:
      | key     | value           |
      | project | {nginx project} |
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "contest project not found",
        "details": null
      }
      """

  Scenario: Unauthorized submit without login
    When I send a "POST" request to "/api/v1/user/contest-projects/1/submit" with the following formdata:
      | key     | value           |
      | project | {nginx project} |
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "Unauthorized",
        "details": null
      }
      """

  Scenario: Forbidden submit for api_reader
    Given I am logged in as "user02"
    When I send a "POST" request to "/api/v1/user/contest-projects/1/submit" with the following formdata:
      | key     | value           |
      | project | {nginx project} |
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """

  # The three scenarios below pin the server-side eligibility window: reads
  # always gated on activation and start time, but submission itself did not,
  # so a direct API call could submit outside the window the UI shows.
  Scenario: Submitting before the submission window opens is rejected
    Given I have contest participants:
      | contestId | userId | groupId |
      | 1         | 13     |         |
    And I am logged in as "user03"
    When I send a "POST" request to "/api/v1/user/contest-projects/3/submit" with the following formdata:
      | key     | value           |
      | project | {nginx project} |
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "submission period has not started",
        "details": null
      }
      """

  Scenario: Submitting after the submission window closes is rejected
    Given I have contest participants:
      | contestId | userId | groupId |
      | 1         | 13     |         |
    And I am logged in as "user03"
    When I send a "POST" request to "/api/v1/user/contest-projects/4/submit" with the following formdata:
      | key     | value           |
      | project | {nginx project} |
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "submission period has ended",
        "details": null
      }
      """

  # A non-activated contest project answers exactly like a missing one.
  Scenario: Submitting to a not-yet-activated contest project looks like a missing one
    Given I have contest participants:
      | contestId | userId | groupId |
      | 1         | 13     |         |
    And I am logged in as "user03"
    When I send a "POST" request to "/api/v1/user/contest-projects/5/submit" with the following formdata:
      | key     | value           |
      | project | {nginx project} |
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "contest project not found",
        "details": null
      }
      """

  # A partial submission names the grading items the judge will run; the ids
  # must resolve, in full, to THIS contest project's problem.
  Scenario: A partial submission naming another problem's grading item is rejected
    Given I have contest participants:
      | contestId | userId | groupId |
      | 1         | 13     |         |
    And I am logged in as "user03"
    When I send a "POST" request to "/api/v1/user/contest-projects/1/submit" with the following formdata:
      | key            | value           |
      | project        | {nginx project} |
      | gradingItemIDs | 9               |
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "invalid grading item ids",
        "details": null
      }
      """

  Scenario: A partial submission naming the problem's own grading item is accepted
    Given I have contest participants:
      | contestId | userId | groupId |
      | 1         | 13     |         |
    And I am logged in as "user03"
    When I send a "POST" request to "/api/v1/user/contest-projects/1/submit" with the following formdata:
      | key            | value           |
      | project        | {nginx project} |
      | gradingItemIDs | 1               |
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {}
      }
      """

  # The cooldown is reserved at submit time rather than derived from past
  # submissions: the submission record is written by an asynchronous handler
  # after the response has already gone out, so a rule that reads submissions
  # back would let a burst through before the first one lands.
  Scenario: Second submission inside the cooldown is rejected with the wait remaining
    Given I have contest participants:
      | contestId | userId | groupId |
      | 1         | 13     |         |
    And I am logged in as "user03"
    When I send a "POST" request to "/api/v1/user/contest-projects/2/submit" with the following formdata:
      | key     | value           |
      | project | {nginx project} |
    Then the response code should be 200
    When I send a "POST" request to "/api/v1/user/contest-projects/2/submit" with the following formdata:
      | key     | value           |
      | project | {nginx project} |
    Then the response code should be 429
    And the response should match json:
      """
      {
        "success": false,
        "message": "submission cooldown has not elapsed",
        "details": [],
        "retryAfterSeconds": 1e+308
      }
      """

  # The reservation is per contest project, so a cooldown on one problem must
  # not stop the contestant working on another.
  Scenario: A cooldown on one contest project does not block another
    Given I have contest participants:
      | contestId | userId | groupId |
      | 1         | 13     |         |
    And I am logged in as "user03"
    When I send a "POST" request to "/api/v1/user/contest-projects/2/submit" with the following formdata:
      | key     | value           |
      | project | {nginx project} |
    Then the response code should be 200
    When I send a "POST" request to "/api/v1/user/contest-projects/1/submit" with the following formdata:
      | key     | value           |
      | project | {nginx project} |
    Then the response code should be 200

  # A contest project with no cooldown of its own inherits the system default,
  # which the test config leaves at 0 — submitting twice in a row is allowed.
  Scenario: A contest project without a cooldown allows consecutive submissions
    Given I have contest participants:
      | contestId | userId | groupId |
      | 1         | 13     |         |
    And I am logged in as "user03"
    When I send a "POST" request to "/api/v1/user/contest-projects/1/submit" with the following formdata:
      | key     | value           |
      | project | {nginx project} |
    Then the response code should be 200
    When I send a "POST" request to "/api/v1/user/contest-projects/1/submit" with the following formdata:
      | key     | value           |
      | project | {nginx project} |
    Then the response code should be 200
