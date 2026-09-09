Feature: List Build API
  As a user
  I want to list all build for an image by tag
  So that I can see all builds for an image if I have permission

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
    And I have builds:
      | ID | Image          | Batch | Status    | CreatedAt           | UpdatedAt           |
      | 1  | nginx:latest   | 1     | CANCELLED | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 2  | nginx:latest   | 2     | SUCCESS   | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 3  | nginx:latest   | 3     | BUILDING  | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 4  | mysql:8.0.26   | 1     | ERROR     | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 5  | mysql:8.0.26   | 2     | PENDING   | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |

  Scenario: Successful list builds with "AdminUser" user for image "nginx:latest"
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/images/nginx:latest/builds":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": [
          {
            "batch": 1,
            "status": "CANCELLED"
          },
          {
            "batch": 2,
            "status": "SUCCESS"
          },
          {
            "batch": 3,
            "status": "BUILDING"
          }
        ]
      }
      """

  Scenario: Successful list builds with "AdminUser" user for image "mysql:8.0.26"
    Given I am logged in as "user02"
    When I send a "GET" request to "/api/v1/images/mysql:8.0.26/builds":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": [
          {
            "batch": 1,
            "status": "ERROR"
          },
          {
            "batch": 2,
            "status": "PENDING"
          }
        ]
      }
      """
      
  Scenario: Forbidden list builds with "Contestant" user
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/images/nginx:latest/builds":
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """

  Scenario: Successful list builds with "AdminAndContestant" user for image "nginx:latest"
    Given I am logged in as "user04"
    When I send a "GET" request to "/api/v1/images/nginx:latest/builds":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": [
          {
            "batch": 1,
            "status": "CANCELLED"
          },
          {
            "batch": 2,
            "status": "SUCCESS"
          },
          {
            "batch": 3,
            "status": "BUILDING"
          }
        ]
      }
      """

  Scenario: Unauthorized list images
    When I send a "GET" request to "/api/v1/images/nginx:latest/builds":
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "Unauthorized",
        "details": null
      }
      """

  Scenario: Successful list builds with "AdminUser" user for image "test-nginx:latest" with no builds
    And I am logged in as "user01"
    When I send a "GET" request to "/api/v1/images/test-nginx:latest/builds":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": []
      }
      """
