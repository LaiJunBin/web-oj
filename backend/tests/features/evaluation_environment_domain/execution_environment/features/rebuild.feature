Feature: Rebuild API
  As a user
  I want to rebuild a image by tag
  So that I can rebuild an image if I have permission

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
      | 7  | test-noproj:latest | /var/www/html | description7 | 0        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
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

  Scenario: Successful rebuild with "AdminUser" user for image "test-nginx:latest"
    Given I am logged in as "user01"
    And setup "nginx.zip" project into "test-nginx:latest" image
    When I send a "POST" request to "/api/v1/images/test-nginx:latest/rebuild":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true
      }
      """
    And when start streaming logs via request to "/api/v1/images/test-nginx:latest/builds/2/stream-logs":
    And wait for 120 seconds for the image "test-nginx:test-2" build to complete
    And streaming logs should have "Successfully built" message within 120 seconds

  Scenario: Successful rebuild image 'test-nobuild:latest' with project files already set up, using the 'AdminUser' user.
    Given I am logged in as "user01"
    And setup "nginx.zip" project into "test-nobuild:latest" image
    When I send a "POST" request to "/api/v1/images/test-nobuild:latest/rebuild":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true
      }
      """
    And when start streaming logs via request to "/api/v1/images/test-nobuild:latest/builds/1/stream-logs":
    And wait for 120 seconds for the image "test-nobuild:latest" build to complete
    And streaming logs should have "Successfully built" message within 120 seconds
      
  Scenario: Failed rebuild building image "nginx:latest" with "AdminUser" user
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/images/nginx:latest/rebuild":
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "build in progress",
        "details": null
      }
      """
    
  Scenario: Failed rebuild pending image "mysql:8.0.26" with "AdminUser" user
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/images/mysql:8.0.26/rebuild":
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "build in progress",
        "details": null
      }
      """

  Scenario: Failed rebuild no-set project image "test-noproj:latest" with "AdminUser" user
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/images/test-noproj:latest/rebuild":
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "image project not set",
        "details": null
      }
      """

  Scenario: Unauthorized rebuild image "nginx:latest" with unauthorized user
    When I send a "POST" request to "/api/v1/images/nginx:latest/rebuild":
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "Unauthorized",
        "details": null
      }
      """

  Scenario: Forbidden rebuild image "nginx:latest" with "Contestant" user
    Given I am logged in as "user03"
    When I send a "POST" request to "/api/v1/images/nginx:latest/rebuild":
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """

  Scenario: Forbidden rebuild image "nginx:latest" with "APIReader" user
    Given I am logged in as "user02"
    When I send a "POST" request to "/api/v1/images/nginx:latest/rebuild":
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """

  Scenario: Not found rebuild image "nginx:test" with "AdminUser" user
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/images/nginx:test/rebuild":
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "image not found",
        "details": null
      }
      """


