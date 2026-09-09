Feature: Send Clarification API
  As a user
  I want to send a clarification to the contest-project
  So that I can send a clarification to the contest-project if I have permission

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
    And I have contests: 
      | id | name           | description                | isPublic | activateTime        | deactivateTime      | createdAt           | updatedAt           |
      | 11 | Test-Contest-1 | Test-Contest-1 Description | 0        | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 12 | Test-Contest-2 | Test-Contest-2 Description | 1        | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 13 | Test-Contest-3 | Test-Contest-3 Description | 0        | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
    And I have contest project associations:
      | id | contestId | projectId | name           | isPublic | judge priority strategy | activateTime        | deactivateTime      | startTime           | endTime             |
      | 1  | 11        | 11        | Test-Project-1 | 0        | default                 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 |
      | 2  | 11        | 12        | Test-Project-2 | 0        | default                 | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 |
      | 3  | 12        | 11        | Test-Project-1 | 1        | default                 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 |
      | 4  | 12        | 12        | Test-Project-2 | 1        | default                 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2099-09-01 00:00:00 | 2099-09-30 23:59:59 |
      | 5  | 12        | 13        | Test-Project-3 | 1        | default                 | 2099-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 |
      | 6  | 13        | 11        | Test-Project-1 | 1        | default                 | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 |
      | 7  | 15        | 12        | Test-Project-2 | 1        | default                 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 |
  Scenario: Successful send clarification with "AdminUser" user
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/contest-projects/1/clarifications" with the following data:
      """
      {
        "subject": "subject1",
        "sendToAll": true
      }
      """
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": 1
      }
      """
    When I send a "GET" request to "/api/v1/contest-projects/1/clarifications":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": [
          {
            "id": 1,
            "subject": "subject1",
            "content": "",
            "senderUser": {
              "id": 11,
              "name": "AdminUser",
              "username": "user01"
            },
            "receiverUser": null,
            "isAdmin": true,
            "sendToAll": true,
            "isAnswered": false,
            "createdAt": "{createdAt}",
            "updatedAt": "{updatedAt}"
          }
        ]
      }
      """
  
  Scenario: Successful send clarification with "AdminUser" user
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/contest-projects/1/clarifications" with the following data:
      """
      {
        "subject": "subject2",
        "content": "content2",
        "receiverUserId": 13
      }
      """
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": 1
      }
      """
    When I send a "GET" request to "/api/v1/contest-projects/1/clarifications":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": [
          {
            "id": 1,
            "subject": "subject2",
            "content": "content2",
            "senderUser": {
              "id": 11,
              "name": "AdminUser",
              "username": "user01"
            },
            "receiverUser": {
              "id": 13,
              "name": "Contestant",
              "username": "user03"
            },
            "isAdmin": true,
            "sendToAll": false,
            "isAnswered": false,
            "createdAt": "{createdAt}",
            "updatedAt": "{updatedAt}"
          }
        ]
      }
      """
      
  Scenario: Successful send clarification with "AdminAndContestant" user
    Given I am logged in as "user04"
    When I send a "POST" request to "/api/v1/contest-projects/1/clarifications" with the following data:
      """
      {
        "subject": "Hi",
        "content": "Hello",
        "receiverUserId": 13,
        "sendToAll": true
      }
      """
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": 1
      }
      """
    When I send a "GET" request to "/api/v1/contest-projects/1/clarifications":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": [
          {
            "id": 1,
            "subject": "Hi",
            "content": "Hello",
            "senderUser": {
              "id": 14,
              "name": "AdminAndContestant",
              "username": "user04"
            },
            "receiverUser": {
              "id": 13,
              "name": "Contestant",
              "username": "user03"
            },
            "isAdmin": true,
            "sendToAll": true,
            "isAnswered": false,
            "createdAt": "{createdAt}",
            "updatedAt": "{updatedAt}"
          }
        ]
      }
      """

  Scenario: Failed send clarification with receiver user not found
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/contest-projects/1/clarifications" with the following data:
      """
      {
        "subject": "subject1",
        "receiverUserId": 111
      }
      """
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "receiver user ID is not exists",
        "details": null
      }
      """
  
  Scenario: Failed send clarification with invalid subject
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/contest-projects/1/clarifications" with the following data:
      """
      {
        "subject": "",
        "sendToAll": true
      }
      """
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "subject is required",
        "details": null
      }
      """
  
  Scenario: Failed send clarification with missing receiver
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/contest-projects/1/clarifications" with the following data:
      """
      {
        "subject": "subject1"
      }
      """
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "you must specify receiver user ID or send to all",
        "details": null
      }
      """

  Scenario: Unauthorized send clarification
    When I send a "POST" request to "/api/v1/contest-projects/1/clarifications" with the following data:
      """
      {
        "subject": "subject1",
        "sendToAll": true
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
    
  Scenario: Forbidden send clarification with "APIReader" user
    Given I am logged in as "user02"
    When I send a "POST" request to "/api/v1/contest-projects/1/clarifications" with the following data:
      """
      {
        "subject": "subject1",
        "sendToAll": true
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

  Scenario: Forbidden send clarification with "Contestant" user
    Given I am logged in as "user03"
    When I send a "POST" request to "/api/v1/contest-projects/1/clarifications" with the following data:
      """
      {
        "subject": "subject1",
        "sendToAll": true
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

  Scenario: Failed send clarification with contest-project not found
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/contest-projects/111/clarifications" with the following data:
      """
      {
        "subject": "subject1",
        "sendToAll": true
      }
      """
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "contest project not found",
        "details": null
      }
      """
