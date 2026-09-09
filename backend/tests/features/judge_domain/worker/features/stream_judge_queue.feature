Feature: Judge Queue Status Stream
  As an admin
  I want the queue status pushed to me
  So that the status page shows a backlog without polling for it

  Background:
    Given I have users:
      | id | name       | username | password | roles      |
      | 11 | AdminUser  | user01   | 123456   | admin      |
      | 13 | Contestant | user03   | 123456   | contestant |

  # The stream is driven by a ticker rather than by queue activity: the oldest
  # wait keeps growing while the queue sits untouched, so a stream that only
  # emitted on push/pop would freeze the one number this page exists to watch.
  # An idle pipeline must therefore still produce frames.
  Scenario: An idle pipeline still emits snapshots
    Given I am logged in as "user01"
    And when start streaming logs via request to "/api/v1/judge-queue/stream":
    Then streaming logs should have "queueDepths" message within 15 seconds
    And streaming logs should have "oldestQueuedSeconds" message within 15 seconds

  Scenario: Unauthorized access
    When I send a "GET" request to "/api/v1/judge-queue/stream":
    Then the response code should be 401

  Scenario: Forbidden for contestant role
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/judge-queue/stream":
    Then the response code should be 403
