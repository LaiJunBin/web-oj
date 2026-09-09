Feature: Cancel Build API
  As a user
  I want to cancel a build by image tag
  So that I can cancel a build if I have the permission

  Background:
    Given I have users:
      | id  | name               | username | password | roles              |
      | 11  | AdminUser          | user01   | 123456   | admin              |
      | 12  | APIReader          | user02   | 123456   | api_reader         |
      | 13  | Contestant         | user03   | 123456   | contestant         |
      | 14  | AdminAndContestant | user04   | 123456   | admin,contestant   |
    And I have images:
      | id | tag                | mount_path    | description  | is_ready | createdAt           | updatedAt           |
      | 1  | nginx:latest       | /var/www/html | description1 | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 2  | mysql:8.0.26       | /var/www/html | description2 | 0        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 3  | test-nginx:latest  | /var/www/html | description3 | 0        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 4  | test-mysql:8.0.26  | /var/www/html | description4 | 0        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 5  | test-nginx:test-2  | /var/www/html | description5 | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 6  | test-nobuild:latest| /var/www/html | description6 | 0        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
    And I have builds:
      | ID | Image              | Batch | Status    | CreatedAt           | UpdatedAt           |
      | 1  | nginx:latest       | 1     | CANCELLED | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 2  | nginx:latest       | 2     | SUCCESS   | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 3  | nginx:latest       | 3     | BUILDING  | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 4  | mysql:8.0.26       | 1     | ERROR     | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 5  | mysql:8.0.26       | 2     | PENDING   | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 6  | test-nginx:latest  | 1     | CANCELLED | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 7  | test-mysql:8.0.26  | 1     | ERROR     | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 8  | test-nginx:test-2  | 1     | SUCCESS   | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |


  Scenario: Successful cancel build(status is building) with "AdminUser" user for image "nginx:latest"
    Given I am logged in as "user01"
    When I send a "DELETE" request to "/api/v1/images/nginx:latest/build":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true
      }
      """
    And the build "nginx:latest" status should be "CANCELLED" within 30 seconds

  Scenario: Successful cancel build(status is building) with "AdminAndContestant" user for image "nginx:latest"
    Given I am logged in as "user04"
    When I send a "DELETE" request to "/api/v1/images/nginx:latest/build":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true
      }
      """
    And the build "nginx:latest" status should be "CANCELLED" within 30 seconds

  Scenario: Failed cancel cancelled build with "AdminUser" user for image "test-nginx:latest"
    Given I am logged in as "user01"
    When I send a "DELETE" request to "/api/v1/images/test-nginx:latest/build":
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "build not in progress",
        "details": null
      }
      """

  Scenario: Failed cancel success build with "AdminUser" user for image "test-nginx:test-2"
    Given I am logged in as "user01"
    When I send a "DELETE" request to "/api/v1/images/test-nginx:test-2/build":
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "build not in progress",
        "details": null
      }
      """

  Scenario: Failed cancel error build with "AdminUser" user for image "test-mysql:8.0.26"
    Given I am logged in as "user01"
    When I send a "DELETE" request to "/api/v1/images/test-mysql:8.0.26/build":
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "build not in progress",
        "details": null
      }
      """

  Scenario: Unauthorized cancel build with unauthenticated user for image "nginx:latest"
    When I send a "DELETE" request to "/api/v1/images/nginx:latest/build":
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "Unauthorized",
        "details": null
      }
      """

  Scenario: Forbidden cancel build with "APIReader" user for image "nginx:latest"
    Given I am logged in as "user02"
    When I send a "DELETE" request to "/api/v1/images/nginx:latest/build":
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """

  Scenario: Forbidden cancel build with "Contestant" user for image "nginx:latest"
    Given I am logged in as "user03"
    When I send a "DELETE" request to "/api/v1/images/nginx:latest/build":
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """

  Scenario: Failed cancel build with "AdminUser" user for image "test-nobuild:latest"
    Given I am logged in as "user01"
    When I send a "DELETE" request to "/api/v1/images/test-nobuild:latest/build":
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "build not found",
        "details": null
      }
      """
