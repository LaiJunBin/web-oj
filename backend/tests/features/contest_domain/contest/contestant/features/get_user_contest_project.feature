Feature: Get User Contest Project API
  As a user
  I want to get a contest project
  So that I can see a contest project if contest and project activated

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
      | id | name           | description               | imageTag       | env | testImageTag      | testEnv | resultParser | createdAt           | updatedAt           |
      | 11 | Test-Project-1 | Test-Project-1 Description| nginx:latest   | a=b | test-nginx:latest |         |  | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 12 | Test-Project-2 | Test-Project-2 Description| mysql:8.0.26   | c=d | test-mysql:8.0.26 |         |  | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 13 | Test-Project-3 | Test-Project-3 Description| nginx:latest   |     | test-nginx:test-2 |         |  | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 14 | Test-Project-4 | Test-Project-4 Description| nginx:latest   |     | test-nginx:test-2 |         |  | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 15 | TP-Asset       |                           | nginx:latest   |     | test-nginx:test-2 |         |  | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 16 | TP-GradingItem |                           | nginx:latest   |     | test-nginx:test-2 |         |  | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
    And I have dynamic fields: 
      | id | project_id | type    | name   | label   | default_value | is_required | created_at           | updated_at           |
      | 1  | 11         | text    | field1 | Field 1 | value1        | 1           | 2021-09-01 00:00:00  | 2021-09-01 00:00:00  |
      | 2  | 11         | boolean | field2 | Field 2 | value2        | 0           | 2021-09-01 00:00:00  | 2021-09-01 00:00:00  |
      | 3  | 12         | file    | field3 | Field 3 | value3        | 1           | 2021-09-01 00:00:00  | 2021-09-01 00:00:00  |
    And I have grading-items:
      | id | projectId | name | description | score | order | scoringStrategy | outputFormat | resultParser | label | createdAt | updatedAt |
      | 11 | 16 | g1 |  | 2.5 | 1 | increase | text |  | 1 | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 12 | 16 | g2 | g2-description | 5 | 2 | increase | text |  | 2 | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 13 | 16 | g3 |  | 10 | 3 | decrease | html |  | 3 | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
    And I have testcases: 
      | id | gradingItemId | name  | command    | score | isSample | createdAt           | updatedAt           |
      | 11 | 11            | g1-t1 | echo t1    | 0.5   | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 12 | 11            | g1-t2 | echo t2    | 1.25  | 0        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 13 | 11            | g1-3  | echo t3    | 2     | 0        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
    And I have assets: 
      | id       | projectId | name     | description | size | isEnabled | createdAt           | deletedAt           |
      | uuid-1   | 15        | test.zip | description | 100  | 0         | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | uuid-2   | 15        | test2.zip| description | 100  | 1         | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
    And I have contests: 
      | id | name           | description                | isPublic | activateTime        | deactivateTime      | createdAt           | updatedAt           |
      | 11 | Test-Contest-1 | Test-Contest-1 Description | 0        | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 12 | Test-Contest-2 | Test-Contest-2 Description | 1        | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 13 | Test-Contest-3 | Test-Contest-3 Description | 0        | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 14 | Test-Contest-4 | Test-Contest-4 Description | 1        | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 15 | Test-Contest-5 | Test-Contest-5 Description | 1        | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
    And I have contest project associations:
      | id | contestId | projectId | name           | isPublic | judge priority strategy | activateTime        | deactivateTime      | startTime           | endTime             |
      | 1  | 11        | 11        | Test-Project-1 | 0        | default                 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 |
      | 2  | 11        | 12        | Test-Project-2 | 0        | default                 | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 |
      | 3  | 12        | 11        | Test-Project-1 | 1        | default                 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 |
      | 4  | 12        | 12        | Test-Project-2 | 1        | default                 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2099-09-01 00:00:00 | 2099-09-30 23:59:59 |
      | 5  | 12        | 13        | Test-Project-3 | 1        | default                 | 2099-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 |
      | 6  | 13        | 11        | Test-Project-1 | 1        | default                 | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 |
      | 7  | 15        | 12        | Test-Project-2 | 1        | default                 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 |
      | 8  | 15        | 13        | Test-Project-3 | 0        | default                 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 |
      | 9  | 15        | 15        | TP-Asset       | 1        | default                 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 |
      | 10 | 15        | 16        | TP-GradingItem | 1        | default                 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 |
  Scenario: Successful get user contest project with "Test-Contest-2" contest and "Test-Project-1" project
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/user/contest-projects/3":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": 3,
          "name": "Test-Project-1",
          "description": "Test-Project-1 Description",
          "timeLeft": 0,
          "isParticipant": false,
          "gradingItems": [],
          "assets": [],
          "dynamicFields": [
            {
              "type": "text",
              "label": "Field 1",
              "name": "field1",
              "isRequired": true,
              "defaultValue": "value1"
            },
            {
              "type": "boolean",
              "label": "Field 2",
              "name": "field2",
              "isRequired": false,
              "defaultValue": "value2"
            }
          ]
        }
      }
      """
  
  Scenario: Successful get user contest project with "Test-Contest-2" contest and "Test-Project-2" project
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/user/contest-projects/4":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": 4,
          "name": "Test-Project-2",
          "description": "Test-Project-2 Description",
          "timeLeft": -1e+308,
          "isParticipant": false,
          "gradingItems": [],
          "assets": [],
          "dynamicFields": []
        }
      }
      """
  
  Scenario: Successful get user contest project with "Test-Contest-5" contest and "Test-Project-2" project
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/user/contest-projects/7":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": 7,
          "name": "Test-Project-2",
          "description": "Test-Project-2 Description",
          "timeLeft": 1e+308,
          "isParticipant": false,
          "gradingItems": [],
          "assets": [],
          "dynamicFields": [
            {
              "type": "file",
              "label": "Field 3",
              "name": "field3",
              "isRequired": true,
              "defaultValue": "value3"
            }
          ]
        }
      }
      """

  Scenario: Successful get user contest project with "Test-Contest-5" contest and "TP-Asset" project
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/user/contest-projects/9":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": 9,
          "name": "TP-Asset",
          "description": "",
          "timeLeft": 1e+308,
          "isParticipant": false,
          "gradingItems": [],
          "assets": [
            {
              "id": "uuid-2",
              "name": "test2.zip"
            }
          ],
          "dynamicFields": []
        }
      }
      """
  
  Scenario: Successful get user contest project with "Test-Contest-5" contest and "TP-GradingItem" project
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/user/contest-projects/10":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": 10,
          "name": "TP-GradingItem",
          "description": "",
          "timeLeft": 1e+308,
          "isParticipant": false,
          "gradingItems": [
            {
              "id": 11,
              "name": "g1",
              "label": "1",
              "description": "",
              "score": 2.5,
              "outputFormat": "text",
              "publicTestcases": [
                {
                  "id": 11,
                  "name": "g1-t1"
                }
              ]
            },
            {
              "id": 12,
              "name": "g2",
              "label": "2",
              "description": "g2-description",
              "score": 5,
              "outputFormat": "text",
              "publicTestcases": []
            },
            {
              "id": 13,
              "name": "g3",
              "label": "3",
              "description": "",
              "score": 10,
              "outputFormat": "html",
              "publicTestcases": []
            }
          ],
          "assets": [],
          "dynamicFields": []
        }
      }
      """

  Scenario: Successful get user contest project with "Test-Contest-1" contest and "Test-Project-1" project
    Given I have contest participants:
      | contestId | userId | groupId |
      | 11        | 13     |         |
    And I am logged in as "user03"
    When I send a "GET" request to "/api/v1/user/contest-projects/1":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": 1,
          "name": "Test-Project-1",
          "description": "Test-Project-1 Description",
          "timeLeft": 0,
          "isParticipant": true,
          "gradingItems": [],
          "assets": [],
          "dynamicFields": [
            {
              "type": "text",
              "label": "Field 1",
              "name": "field1",
              "isRequired": true,
              "defaultValue": "value1"
            },
            {
              "type": "boolean",
              "label": "Field 2",
              "name": "field2",
              "isRequired": false,
              "defaultValue": "value2"
            }
          ]
        }
      }
      """

  Scenario: Failed get non-public contest project when not a participant
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/user/contest-projects/1":
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "not a participant of this contest",
        "details": null
      }
      """

  Scenario: Successful get user contest project with "Test-Contest-5" and "Test-Project-3" project
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/user/contest-projects/8":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": 8,
          "name": "Test-Project-3",
          "description": "Test-Project-3 Description",
          "timeLeft": 1e+308,
          "isParticipant": false,
          "gradingItems": [],
          "assets": [],
          "dynamicFields": []
        }
      }
      """

  Scenario: Unauthorized get user contest project
    When I send a "GET" request to "/api/v1/user/contest-projects/1":
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "Unauthorized",
        "details": null
      }
      """

  Scenario: Forbidden get user contest project
    Given I am logged in as "user02"
    When I send a "GET" request to "/api/v1/user/contest-projects/1":
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """

  Scenario: Failed get project of expired contest
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/user/contest-projects/6":
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "contest project not found",
        "details": null
      }
      """

  Scenario: Failed get expired contest project
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/user/contest-projects/2":
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "contest project not found",
        "details": null
      }
      """

  Scenario: Failed get non-exist contest-project
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/user/contest-projects/99":
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "contest project not found",
        "details": null
      }
      """

  Scenario: Get public contest project as a direct participant shows isParticipant true
    Given I have contest participants:
      | contestId | userId | groupId |
      | 12        | 13     |         |
    And I am logged in as "user03"
    When I send a "GET" request to "/api/v1/user/contest-projects/3":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": 3,
          "name": "Test-Project-1",
          "description": "Test-Project-1 Description",
          "timeLeft": 0,
          "isParticipant": true,
          "gradingItems": [],
          "assets": [],
          "dynamicFields": [
            {
              "type": "text",
              "label": "Field 1",
              "name": "field1",
              "isRequired": true,
              "defaultValue": "value1"
            },
            {
              "type": "boolean",
              "label": "Field 2",
              "name": "field2",
              "isRequired": false,
              "defaultValue": "value2"
            }
          ]
        }
      }
      """

  Scenario: Get user contest project as a group participant shows isParticipant true
    Given I have groups:
      | id | name    | description |
      | 1  | Group-1 | desc        |
    And I have user groups:
      | userId | groupId |
      | 13     | 1       |
    And I have contest participants:
      | contestId | userId | groupId |
      | 11        |        | 1       |
    And I am logged in as "user03"
    When I send a "GET" request to "/api/v1/user/contest-projects/1":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": 1,
          "name": "Test-Project-1",
          "description": "Test-Project-1 Description",
          "timeLeft": 0,
          "isParticipant": true,
          "gradingItems": [],
          "assets": [],
          "dynamicFields": [
            {
              "type": "text",
              "label": "Field 1",
              "name": "field1",
              "isRequired": true,
              "defaultValue": "value1"
            },
            {
              "type": "boolean",
              "label": "Field 2",
              "name": "field2",
              "isRequired": false,
              "defaultValue": "value2"
            }
          ]
        }
      }
      """

  Scenario: Get public contest project via a soft-deleted group shows isParticipant false
    Given I have groups:
      | id | name    | description | deletedAt           |
      | 1  | Group-1 | desc        | 2021-09-02 00:00:00 |
    And I have user groups:
      | userId | groupId |
      | 13     | 1       |
    And I have contest participants:
      | contestId | userId | groupId |
      | 12        |        | 1       |
    And I am logged in as "user03"
    When I send a "GET" request to "/api/v1/user/contest-projects/3":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": 3,
          "name": "Test-Project-1",
          "description": "Test-Project-1 Description",
          "timeLeft": 0,
          "isParticipant": false,
          "gradingItems": [],
          "assets": [],
          "dynamicFields": [
            {
              "type": "text",
              "label": "Field 1",
              "name": "field1",
              "isRequired": true,
              "defaultValue": "value1"
            },
            {
              "type": "boolean",
              "label": "Field 2",
              "name": "field2",
              "isRequired": false,
              "defaultValue": "value2"
            }
          ]
        }
      }
      """
