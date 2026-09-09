Feature: Get Build API
  As a user
  I want to get a build for an image by tag and batch
  So that I can see a build for an image if I have permission

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
    And I have build logs:
      | ID | BuildID | Level   | Message   | CreatedAt           |
      | 1  | 1       | info    | message1  | 2021-09-01T00:00:00 |
      | 2  | 1       | info    | message2  | 2021-09-01T00:00:00 |
      | 3  | 1       | warning | message3  | 2021-09-01T00:00:00 |
      | 4  | 2       | info    | message4  | 2021-09-01T00:00:00 |
      | 5  | 2       | warning | message5  | 2021-09-01T00:00:00 |
      | 6  | 2       | success | message6  | 2021-09-01T00:00:00 |
      | 7  | 3       | info    | message7  | 2021-09-01T00:00:00 |
      | 8  | 3       | info    | message8  | 2021-09-01T00:00:00 |
      | 9  | 3       | success | message9  | 2021-09-01T00:00:00 |
      | 10 | 4       | info    | message10 | 2021-09-01T00:00:00 |
      | 11 | 4       | error   | message11 | 2021-09-01T00:00:00 |
      | 12 | 5       | info    | message12 | 2021-09-01T00:00:00 |
      | 13 | 5       | info    | message13 | 2021-09-01T00:00:00 |
      | 14 | 6       | info    | message14 | 2021-09-01T00:00:00 |
      | 15 | 6       | info    | message15 | 2021-09-01T00:00:00 |
      | 16 | 6       | info    | message16 | 2021-09-01T00:00:00 |

  Scenario: Successful get build with "AdminUser" user for image "nginx:latest" and batch 1
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/images/nginx:latest/builds/1":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "batch": 1,
          "status": "CANCELLED",
          "logs": [
            {
              "level": "info",
              "message": "message1",
              "time": "{createdAt}"
            },
            {
              "level": "info",
              "message": "message2",
              "time": "{createdAt}"
            },
            {
              "level": "warning",
              "message": "message3",
              "time": "{createdAt}"
            }
          ]
        }
      }
      """

  Scenario: Successful get build with "APIReader" user for image "mysql:8.0.26" and batch 1
    Given I am logged in as "user02"
    When I send a "GET" request to "/api/v1/images/mysql:8.0.26/builds/1":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "batch": 1,
          "status": "ERROR",
          "logs": [
            {
              "level": "info",
              "message": "message10",
              "time": "{createdAt}"
            },
            {
              "level": "error",
              "message": "message11",
              "time": "{createdAt}"
            }
          ]
        }
      }
      """
      
  Scenario: Forbidden get build with "Contestant" user for image "nginx:latest" and batch 1
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/images/nginx:latest/builds/1":
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """

  Scenario: Successful get build with "AdminAndContestant" user for image "nginx:latest" and batch 2
    Given I am logged in as "user04"
    When I send a "GET" request to "/api/v1/images/nginx:latest/builds/2":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "batch": 2,
          "status": "SUCCESS",
          "logs": [
            {
              "level": "info",
              "message": "message4",
              "time": "{createdAt}"
            },
            {
              "level": "warning",
              "message": "message5",
              "time": "{createdAt}"
            },
            {
              "level": "success",
              "message": "message6",
              "time": "{createdAt}"
            }
          ]
        }
      }
      """

  Scenario: Unauthorized get build with no user for image "nginx:latest" and batch 1
    When I send a "GET" request to "/api/v1/images/nginx:latest/builds/1":
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "Unauthorized",
        "details": null
      }
      """

  Scenario: Not found get build with "AdminUser" user for image "nginx:latest" and batch 4 
    And I am logged in as "user01"
    When I send a "GET" request to "/api/v1/images/nginx:latest/builds/4":
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "build not found",
        "details": null
      }
      """
