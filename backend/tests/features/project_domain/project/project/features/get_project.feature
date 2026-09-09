Feature: Get Project API
  As a user
  I want to get a project
  So that I can see a project if I have permission

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
      | 11  | Test-Project-1  | Test-Project-1 Description| nginx:latest   | a=b    | test-nginx:latest |         |              | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 12  | Test-Project-2  | Test-Project-2 Description| mysql:8.0.26   | c=d    | test-mysql:8.0.26 |         |              | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 13  | Test-Project-3  | Test-Project-3 Description| nginx:latest   |        | test-nginx:test-2 |         |              | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 14  | Test-Project-4  | Test-Project-4 Description| nginx:latest   |        | test-nginx:latest |         | stdout       | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
    And I have dynamic fields:
      | id | project_id | type    | name   | label   | default_value | is_required | created_at           | updated_at           |
      | 1  | 11         | text    | field1 | Field 1 | value1        | 1           | 2021-09-01 00:00:00  | 2021-09-01 00:00:00  |
      | 2  | 11         | boolean | field2 | Field 2 | value2        | 1           | 2021-09-01 00:00:00  | 2021-09-01 00:00:00  |
      | 3  | 12         | file    | field3 | Field 3 | value3        | 1           | 2021-09-01 00:00:00  | 2021-09-01 00:00:00  |
    And I have lifecycles:
      | id | project_id | name | description | type       | runs_on | command       | created_at           | updated_at           |
      | 1  | 11         | test | test        | before_all | project | echo "test"   | 2021-09-01 00:00:00  | 2021-09-01 00:00:00  |
      | 2  | 11         | test | test        | after_all  | project | echo "test"   | 2021-09-01 00:00:00  | 2021-09-01 00:00:00  |
    And I have assets: 
      | id       | project_id | name     | description | size | is_enabled | created_at           | updated_at           |
      | uuid-1   | 11         | test.zip | test        | 1    | 1          | 2021-09-01 00:00:00  | 2021-09-01 00:00:00  |
      | uuid-2   | 11         | test2.zip| test2       | 2    | 1          | 2021-09-01 00:00:00  | 2021-09-01 00:00:00  |
      | uuid-3   | 12         | test3.zip| test3       | 3    | 1          | 2021-09-01 00:00:00  | 2021-09-01 00:00:00  |
      
  Scenario: Successful get project with "AdminUser" user
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/projects/11":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": 11,
          "name": "Test-Project-1",
          "description": "Test-Project-1 Description",
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
          "env": "a=b",
          "testEnv": "",
          "resultParser": "auto",
          "dynamicFields": [
            {
              "type": "text",
              "name": "field1",
              "label": "Field 1",
              "defaultValue": "value1",
              "isRequired": true
            },
            {
              "type": "boolean",
              "name": "field2",
              "label": "Field 2",
              "defaultValue": "value2",
              "isRequired": true
            }
          ],
          "lifecycle": {
            "beforeAll": [
              {
                "id": 1,
                "name": "test",
                "description": "test",
                "command": "echo \"test\""
              }
            ],
            "beforeEach": [],
            "afterEach": [],
            "afterAll": [
              {
                "id": 2,
                "name": "test",
                "description": "test",
                "command": "echo \"test\""
              }
            ]
          },
          "testLifecycle": {
            "beforeAll": [],
            "beforeEach": [],
            "afterEach": [],
            "afterAll": []
          },
          "assets": [
            {
              "id": "uuid-1",
              "name": "test.zip",
              "size": 1,
              "isEnabled": true,
              "lastModified": "{lastModified}"
            },
            {
              "id": "uuid-2",
              "name": "test2.zip",
              "size": 2,
              "isEnabled": true,
              "lastModified": "{lastModified}"
            }
          ],
          "createdAt": "{createdAt}",
          "updatedAt": "{updatedAt}"
        }
      }
      """
  
  Scenario: Successful get project with "APIReader" user
    Given I am logged in as "user02"
    When I send a "GET" request to "/api/v1/projects/12":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": 12,
          "name": "Test-Project-2",
          "description": "Test-Project-2 Description",
          "image": {
            "tag": "mysql:8.0.26",
            "description": "description2",
            "isReady": true,
            "createdAt": "{createdAt}",
            "updatedAt": "{updatedAt}"
          },
          "testImage": {
            "tag": "test-mysql:8.0.26",
            "description": "description4",
            "isReady": true,
            "createdAt": "{createdAt}",
            "updatedAt": "{updatedAt}"
          },
          "env": "c=d",
          "testEnv": "",
          "resultParser": "auto",
          "dynamicFields": [
            {
              "type": "file",
              "name": "field3",
              "label": "Field 3",
              "defaultValue": "value3",
              "isRequired": true
            }
          ],
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
          "assets": [
            {
              "id": "uuid-3",
              "name": "test3.zip",
              "size": 3,
              "isEnabled": true,
              "lastModified": "{lastModified}"
            }
          ],
          "createdAt": "{createdAt}",
          "updatedAt": "{updatedAt}"
        }
      }
      """
      
  Scenario: Successful get project with "AdminAndContestant" user
    Given I am logged in as "user04"
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
          "resultParser": "auto",
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

  Scenario: Unauthorized get project
    When I send a "GET" request to "/api/v1/projects/11":
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "Unauthorized",
        "details": null
      }
      """

  Scenario: Forbidden get project with "Contestant" user
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/projects/11":
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """

  Scenario: resultParser defaults to auto when project has none set
    Given I am logged in as "user01"
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
          "resultParser": "auto",
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

  Scenario: resultParser returns configured value when project has one set
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/projects/14":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": 14,
          "name": "Test-Project-4",
          "description": "Test-Project-4 Description",
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

  Scenario: Not found get project with "AdminUser" user
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/projects/99":
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "project not found",
        "details": null
      }
      """
