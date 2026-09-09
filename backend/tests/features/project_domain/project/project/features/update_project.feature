Feature: Update Project API
  As a user
  I want to update a project
  So that I can update a project if I have permission

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
    And I have projects:
      | id  | name            | description               | imageTag       | env    | testImageTag      | testEnv | resultParser | createdAt           | updatedAt           |
      | 11  | Test-Project-1  | Test-Project-1 Description| nginx:latest   | a=b    | test-nginx:latest |         |  | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 12  | Test-Project-2  | Test-Project-2 Description| mysql:8.0.26   | c=d    | test-mysql:8.0.26 |         |  | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 13  | Test-Project-3  | Test-Project-3 Description| nginx:latest   |        | test-nginx:test-2 |         |  | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
  Scenario: Successful update project with "AdminUser" user
    Given I am logged in as "user01"
    When I send a "PATCH" request to "/api/v1/projects/11" with the following data:
      """
      {
        "name": "tp",
        "projectImageTag": "mysql:8.0.26",
        "testImageTag": "test-mysql:8.0.26",
        "projectEnv": "",
        "testEnv": ""
      }
      """
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true
      }
      """
  
  Scenario: Successful update project with "AdminAndContestant" user
    Given I am logged in as "user04"
    When I send a "PATCH" request to "/api/v1/projects/11" with the following data:
      """
      {
        "name": "tp",
        "projectImageTag": "nginx:latest",
        "testImageTag": "test-nginx:latest",
        "projectEnv": "a=b\nb=c",
        "testEnv": ""
      }
      """
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true
      }
      """

  Scenario: Successful update project with dynamicFields and "AdminUser" user
    Given I am logged in as "user01"
    When I send a "PATCH" request to "/api/v1/projects/11" with the following data:
      """
      {
        "name": "tp",
        "projectImageTag": "nginx:latest",
        "testImageTag": "test-nginx:latest",
        "projectEnv": "a=b\nb=c",
        "testEnv": "",
        "dynamicFields": [
          {
            "name": "field1",
            "type": "text",
            "label": "Field 1",
            "defaultValue": "default",
            "isRequired": true
          },
          {
            "name": "field2",
            "type": "boolean",
            "label": "Field 2",
            "isRequired": false
          }
        ]
      }
      """
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true
      }
      """

  Scenario: Failed update project with image not found
    Given I am logged in as "user01"
    When I send a "PATCH" request to "/api/v1/projects/11" with the following data:
      """
      {
        "name": "tp",
        "projectImageTag": "invalid:latest",
        "testImageTag": "test-nginx:latest",
        "projectEnv": "a=b\nb=c",
        "testEnv": ""
      }
      """
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "project image not found",
        "details": null
      }
      """

  Scenario: Failed update project with test image not found
    Given I am logged in as "user01"
    When I send a "PATCH" request to "/api/v1/projects/11" with the following data:
      """
      {
        "name": "tp",
        "projectImageTag": "nginx:latest",
        "testImageTag": "invalid:latest",
        "projectEnv": "a=b\nb=c",
        "testEnv": ""
      }
      """
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "test image not found",
        "details": null
      }
      """
  
  Scenario: Failed update project with invalid env format
    Given I am logged in as "user01"
    When I send a "PATCH" request to "/api/v1/projects/11" with the following data:
      """
      {
        "name": "tp",
        "projectImageTag": "nginx:latest",
        "testImageTag": "test-nginx:latest",
        "projectEnv": "abc",
        "testEnv": ""
      }
      """
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "invalid env",
        "details": null
      }
      """

  Scenario: Unauthorized update project
    When I send a "PATCH" request to "/api/v1/projects/11" with the following data:
      """
      {
        "name": "tp",
        "projectImageTag": "nginx:latest",
        "testImageTag": "test-nginx:latest",
        "projectEnv": "a=b",
        "testEnv": ""
      }
      """
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "Unauthorized",
        "details": null
      }
      """
    
  Scenario: Forbidden update project with "APIReader" user
    Given I am logged in as "user02"
    When I send a "PATCH" request to "/api/v1/projects/11" with the following data:
      """
      {
        "name": "tp",
        "projectImageTag": "nginx:latest",
        "testImageTag": "test-nginx:latest",
        "projectEnv": "a=b",
        "testEnv": ""
      }
      """
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """

  Scenario: Forbidden update project with "Contestant" user
    Given I am logged in as "user03"
    When I send a "PATCH" request to "/api/v1/projects/11" with the following data:
      """
      {
        "name": "tp",
        "projectImageTag": "nginx:latest",
        "testImageTag": "test-nginx:latest",
        "projectEnv": "a=b",
        "testEnv": ""
      }
      """
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """

  Scenario: Not found update project
    Given I am logged in as "user01"
    When I send a "PATCH" request to "/api/v1/projects/999" with the following data:
      """
      {
        "name": "tp",
        "projectImageTag": "nginx:latest",
        "testImageTag": "test-nginx:latest",
        "projectEnv": "a=b",
        "testEnv": ""
      }
      """
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "project not found",
        "details": null
      }
      """

  Scenario: Successful update project resultParser
    Given I am logged in as "user01"
    When I send a "PATCH" request to "/api/v1/projects/13" with the following data:
      """
      {
        "resultParser": "stdout"
      }
      """
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true
      }
      """
    When I send a "GET" request to "/api/v1/projects/13":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": 13,
          "name": "Test-Project-3",
          "description": "Test-Project-3 Description",
          "image": {
            "tag": "nginx:latest",
            "description": "description1",
            "isReady": false,
            "createdAt": "{createdAt}",
            "updatedAt": "{updatedAt}"
          },
          "testImage": {
            "tag": "test-nginx:test-2",
            "description": "description5",
            "isReady": true,
            "createdAt": "{createdAt}",
            "updatedAt": "{updatedAt}"
          },
          "env": "",
          "testEnv": "",
          "resultParser": "stdout",
          "dynamicFields": [],
          "lifecycle": {
            "beforeAll": [],
            "beforeEach": [],
            "afterEach": [],
            "afterAll": []
          },
          "testLifecycle": {
            "beforeAll": [],
            "beforeEach": [],
            "afterEach": [],
            "afterAll": []
          },
          "assets": [],
          "createdAt": "{createdAt}",
          "updatedAt": "{updatedAt}"
        }
      }
      """

  Scenario: Failed update project with invalid resultParser
    Given I am logged in as "user01"
    When I send a "PATCH" request to "/api/v1/projects/13" with the following data:
      """
      {
        "resultParser": "invalid-parser"
      }
      """
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "invalid result parser",
        "details": null
      }
      """
