Feature: Get Image API
  As a user
  I want to get image by tag
  So that I can see the image if I have permission

  Background:
    Given I have users:
      | id  | name               | username | password | roles              |
      | 11  | AdminUser          | user01   | 123456   | admin              |
      | 12  | APIReader          | user02   | 123456   | api_reader         |
      | 13  | Contestant         | user03   | 123456   | contestant         |
      | 14  | AdminAndContestant | user04   | 123456   | admin,contestant   |
    And I have images:
      | id | tag                | mount_path    | description  | is_ready | createdAt           | updatedAt           |
      | 1  | nginx:latest       | /var/www/html | description1 | 0        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 2  | mysql:8.0.26       | /var/www/html | description2 | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 3  | test-nginx:latest  | /var/www/html | description3 | 0        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 4  | test-mysql:8.0.26  | /var/www/html | description4 | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 5  | test-nginx:test-2  | /var/www/html | description5 | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |

  Scenario: Successful get unready image with "AdminUser" user
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/images/nginx:latest":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "tag": "nginx:latest",
          "mountPath": "/var/www/html",
          "description": "description1",
          "isReady": false,
          "createdAt": "{createdAt}",
          "updatedAt": "{updatedAt}"
        }
      }
      """

  Scenario: Successful get ready image with "AdminUser" user
    Given I am logged in as "user02"
    When I send a "GET" request to "/api/v1/images/mysql:8.0.26":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "tag": "mysql:8.0.26",
          "mountPath": "/var/www/html",
          "description": "description2",
          "isReady": true,
          "createdAt": "{createdAt}",
          "updatedAt": "{updatedAt}"
        }
      }
      """
      
  Scenario: Forbidden get image with "Contestant" user
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/images/nginx:latest":
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """

  Scenario: Successful get image with "AdminAndContestant" user
    Given I am logged in as "user04"
    When I send a "GET" request to "/api/v1/images/nginx:latest":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "tag": "nginx:latest",
          "mountPath": "/var/www/html",
          "description": "description1",
          "isReady": false,
          "createdAt": "{createdAt}",
          "updatedAt": "{updatedAt}"
        }
      }
      """

  Scenario: Unauthorized get image with empty token
    When I send a "GET" request to "/api/v1/images/nginx:latest":
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "Unauthorized",
        "details": null
      }
      """

  Scenario: Not found get image with non-existed tag
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/images/non-existed-tag":
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "image not found",
        "details": null
      }
      """
