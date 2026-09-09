Feature: List Image API
  As a user
  I want to list all images
  So that I can see all images if I have permission

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

  Scenario: Successful list images with "AdminUser" user
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/images":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": [
            {
              "tag": "nginx:latest",
              "description": "description1",
              "isReady": false,
              "createdAt": "{createdAt}",
              "updatedAt": "{updatedAt}"
            },
            {
              "tag": "mysql:8.0.26",
              "description": "description2",
              "isReady": true,
              "createdAt": "{createdAt}",
              "updatedAt": "{updatedAt}"
            },
            {
              "tag": "test-nginx:latest",
              "description": "description3",
              "isReady": false,
              "createdAt": "{createdAt}",
              "updatedAt": "{updatedAt}"
            },
            {
              "tag": "test-mysql:8.0.26",
              "description": "description4",
              "isReady": true,
              "createdAt": "{createdAt}",
              "updatedAt": "{updatedAt}"
            },
            {
              "tag": "test-nginx:test-2",
              "description": "description5",
              "isReady": true,
              "createdAt": "{createdAt}",
              "updatedAt": "{updatedAt}"
            }
          ]
      }
      """

  Scenario: Successful list images with "APIReader" user
    Given I am logged in as "user02"
    When I send a "GET" request to "/api/v1/images":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": [
            {
              "tag": "nginx:latest",
              "description": "description1",
              "isReady": false,
              "createdAt": "{createdAt}",
              "updatedAt": "{updatedAt}"
            },
            {
              "tag": "mysql:8.0.26",
              "description": "description2",
              "isReady": true,
              "createdAt": "{createdAt}",
              "updatedAt": "{updatedAt}"
            },
            {
              "tag": "test-nginx:latest",
              "description": "description3",
              "isReady": false,
              "createdAt": "{createdAt}",
              "updatedAt": "{updatedAt}"
            },
            {
              "tag": "test-mysql:8.0.26",
              "description": "description4",
              "isReady": true,
              "createdAt": "{createdAt}",
              "updatedAt": "{updatedAt}"
            },
            {
              "tag": "test-nginx:test-2",
              "description": "description5",
              "isReady": true,
              "createdAt": "{createdAt}",
              "updatedAt": "{updatedAt}"
            }
          ]
      }
      """
      
  Scenario: Forbidden list images with "Contestant" user
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/images":
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """

  Scenario: Successful list images with "AdminAndContestant" user
    Given I am logged in as "user04"
    When I send a "GET" request to "/api/v1/images":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": [
            {
              "tag": "nginx:latest",
              "description": "description1",
              "isReady": false,
              "createdAt": "{createdAt}",
              "updatedAt": "{updatedAt}"
            },
            {
              "tag": "mysql:8.0.26",
              "description": "description2",
              "isReady": true,
              "createdAt": "{createdAt}",
              "updatedAt": "{updatedAt}"
            },
            {
              "tag": "test-nginx:latest",
              "description": "description3",
              "isReady": false,
              "createdAt": "{createdAt}",
              "updatedAt": "{updatedAt}"
            },
            {
              "tag": "test-mysql:8.0.26",
              "description": "description4",
              "isReady": true,
              "createdAt": "{createdAt}",
              "updatedAt": "{updatedAt}"
            },
            {
              "tag": "test-nginx:test-2",
              "description": "description5",
              "isReady": true,
              "createdAt": "{createdAt}",
              "updatedAt": "{updatedAt}"
            }
          ]
      }
      """

  Scenario: Unauthorized list images
    When I send a "GET" request to "/api/v1/images":
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "Unauthorized",
        "details": null
      }
      """

  Scenario: Successful list images with empty data
    Given I have no images
    And I am logged in as "user01"
    When I send a "GET" request to "/api/v1/images":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": []
      }
      """
