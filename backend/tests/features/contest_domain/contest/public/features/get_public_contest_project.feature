Feature: Get Public Contest Project API
  As a non authenticated user
  I want to get a public contest project
  So that I can see a public contest project if contest and project activated

  Background:
    Given I have images:
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
      | 11 | Test-Contest-1 | Test-Contest-1 Description | 0        | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 12 | Test-Contest-2 | Test-Contest-2 Description | 1        | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 13 | Test-Contest-3 | Test-Contest-3 Description | 0        | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 14 | Test-Contest-4 | Test-Contest-4 Description | 1        | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 15 | Test-Contest-5 | Test-Contest-5 Description | 1        | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
    And I have contest project associations:
      | id | contestId | projectId | name           | isPublic | judge priority strategy | activateTime        | deactivateTime      | startTime           | endTime             |
      | 1  | 11        | 11        | Test-Project-1 | 0        | default                 | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 |
      | 2  | 11        | 12        | Test-Project-2 | 0        | default                 | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 |
      | 3  | 12        | 11        | Test-Project-1 | 1        | default                 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 |
      | 4  | 12        | 12        | Test-Project-2 | 1        | default                 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2099-09-01 00:00:00 | 2099-09-30 23:59:59 |
      | 5  | 12        | 13        | Test-Project-3 | 1        | default                 | 2099-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 |
      | 6  | 13        | 11        | Test-Project-1 | 1        | default                 | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-30 23:59:59 |
      | 7  | 15        | 12        | Test-Project-2 | 1        | default                 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 |
      | 8  | 15        | 13        | Test-Project-3 | 0        | default                 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 |
      | 9  | 15        | 15        | TP-Asset       | 1        | default                 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 |
      | 10 | 15        | 16        | TP-GradingItem | 1        | default                 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 |
  Scenario: Successful get public contest project with "Test-Contest-2" contest and "Test-Project-1" project
    When I send a "GET" request to "/api/v1/public/contest-projects/3":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": 3,
          "name": "Test-Project-1",
          "description": "Test-Project-1 Description",
          "status": "Ended",
          "timeLeft": 0,
          "gradingItems": [],
          "assets": []
        }
      }
      """
  
  Scenario: Successful get public contest project with "Test-Contest-2" contest and "Test-Project-2" project
    When I send a "GET" request to "/api/v1/public/contest-projects/4":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": 4,
          "name": "Test-Project-2",
          "description": "Test-Project-2 Description",
          "status": "NotStarted",
          "timeLeft": -1e+308,
          "gradingItems": [],
          "assets": []
        }
      }
      """
  
  Scenario: Successful get public contest project with "Test-Contest-5" contest and "Test-Project-2" project
    When I send a "GET" request to "/api/v1/public/contest-projects/7":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": 7,
          "name": "Test-Project-2",
          "description": "Test-Project-2 Description",
          "status": "InProgress",
          "timeLeft": 1e+308,
          "gradingItems": [],
          "assets": []
        }
      }
      """

  Scenario: Successful get public contest project with "Test-Contest-5" contest and "TP-Asset" project
    When I send a "GET" request to "/api/v1/public/contest-projects/9":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": 9,
          "name": "TP-Asset",
          "description": "",
          "status": "InProgress",
          "timeLeft": 1e+308,
          "gradingItems": [],
          "assets": [
            {
              "id": "uuid-2",
              "name": "test2.zip"
            }
          ]
        }
      }
      """
  
  Scenario: Successful get public contest project with "Test-Contest-5" contest and "TP-GradingItem" project
    When I send a "GET" request to "/api/v1/public/contest-projects/10":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": 10,
          "name": "TP-GradingItem",
          "description": "",
          "status": "InProgress",
          "timeLeft": 1e+308,
          "gradingItems": [
            {
              "id": 11,
              "name": "g1",
              "label": "1",
              "description": "",
              "score": 2.5,
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
              "publicTestcases": []
            },
            {
              "id": 13,
              "name": "g3",
              "label": "3",
              "description": "",
              "score": 10,
              "publicTestcases": []
            }
          ],
          "assets": []
        }
      }
      """

  Scenario: Failed get non-public contest with "Test-Contest-1" and "Test-Project-1" private project
    When I send a "GET" request to "/api/v1/public/contest-projects/1":
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "contest project not found",
        "details": null
      }
      """

  Scenario: Failed get non-public contest with "Test-Contest-3" and "Test-Project-1" public project
    When I send a "GET" request to "/api/v1/public/contest-projects/6":
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "contest project not found",
        "details": null
      }
      """

  Scenario: Failed get public contest project with "Test-Contest-5" and "Test-Project-3" private project
    When I send a "GET" request to "/api/v1/public/contest-projects/8":
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
    When I send a "GET" request to "/api/v1/public/contest-projects/99":
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "contest project not found",
        "details": null
      }
      """
