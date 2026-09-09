Feature: Create Project API
  As a user
  I want to create a project
  So that I can create a project with a roles int if I have permission

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
  Scenario: Successful create project with "AdminUser" user
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/projects" with the following data:
      """
      {
        "name": "tp",
        "projectImageTag": "nginx:latest",
        "testImageTag": "test-nginx:latest",
        "projectEnv": "",
        "testEnv": ""
      }
      """
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": 14
      }
      """
  
  Scenario: Successful create project with "AdminAndContestant" user
    Given I am logged in as "user04"
    When I send a "POST" request to "/api/v1/projects" with the following data:
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
        "success": true,
        "data": 14
      }
      """

  Scenario: Successful create project with dynamicFields and "AdminUser" user
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/projects" with the following data:
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
        "success": true,
        "data": 14
      }
      """

  Scenario: Failed create project with image not found
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/projects" with the following data:
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

  Scenario: Failed create project with test image not found
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/projects" with the following data:
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
  
  Scenario: Failed create project with invalid env format
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/projects" with the following data:
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

  Scenario: Unauthorized create project
    When I send a "POST" request to "/api/v1/projects" with the following data:
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
    
  Scenario: Forbidden create project with "APIReader" user
    Given I am logged in as "user02"
    When I send a "POST" request to "/api/v1/projects" with the following data:
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

  Scenario: Forbidden create project with "Contestant" user
    Given I am logged in as "user03"
    When I send a "POST" request to "/api/v1/projects" with the following data:
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

  Scenario: Successful create project with resultParser set
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/projects" with the following data:
      """
      {
        "name": "tp-exitcode",
        "description": "project with exitcode parser",
        "projectImageTag": "nginx:latest",
        "testImageTag": "test-nginx:latest",
        "projectEnv": "",
        "testEnv": "",
        "dynamicFields": [],
        "resultParser": "exitcode"
      }
      """
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": 14
      }
      """
    When I send a "GET" request to "/api/v1/projects/14":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": 14,
          "name": "tp-exitcode",
          "description": "project with exitcode parser",
          "image": {
            "tag": "nginx:latest",
            "description": "description1",
            "isReady": false,
            "createdAt": "{createdAt}",
            "updatedAt": "{updatedAt}"
          },
          "testImage": {
            "tag": "test-nginx:latest",
            "description": "description3",
            "isReady": false,
            "createdAt": "{createdAt}",
            "updatedAt": "{updatedAt}"
          },
          "env": "",
          "testEnv": "",
          "resultParser": "exitcode",
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


  Scenario: Failed create project with invalid resultParser
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/projects" with the following data:
      """
      {
        "name": "new-project",
        "description": "",
        "projectImageTag": "nginx:latest",
        "testImageTag": "test-nginx:latest",
        "projectEnv": "",
        "testEnv": "",
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
