Feature: Create Grading-Item API
  As a user
  I want to create a grading item to a project
  So that I can create a grading item to a project if I have permission

  Background:
    Given I have users:
      | id | name               | username | password | roles            |
      | 11 | AdminUser          | user01   | 123456   | admin            |
      | 12 | APIReader          | user02   | 123456   | api_reader       |
      | 13 | Contestant         | user03   | 123456   | contestant       |
      | 14 | AdminAndContestant | user04   | 123456   | admin,contestant |
    And I have images:
      | id | tag               | mount_path    | description  | is_ready | createdAt           | updatedAt           |
      | 1  | nginx:latest      | /var/www/html | description1 | 0        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 2  | mysql:8.0.26      | /var/www/html | description2 | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 3  | test-nginx:latest | /var/www/html | description3 | 0        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 4  | test-mysql:8.0.26 | /var/www/html | description4 | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 5  | test-nginx:test-2 | /var/www/html | description5 | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
    And I have projects:
      | id | name           | description               | imageTag       | env | testImageTag      | testEnv | resultParser | createdAt           | updatedAt           |
      | 11 | Test-Project-1 | Test-Project-1 Description| nginx:latest   | a=b | test-nginx:latest |         |  | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 12 | Test-Project-2 | Test-Project-2 Description| mysql:8.0.26   | c=d | test-mysql:8.0.26 |         |  | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 13 | Test-Project-3 | Test-Project-3 Description| nginx:latest   |     | test-nginx:test-2 |         |  | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
    And I have grading-items:
      | id | projectId | name | description | score | order | scoringStrategy | outputFormat | resultParser | label | createdAt | updatedAt |
      | 11 | 11 | g1 |  | 2.5 | 1 | increase | text |  | 1 | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 12 | 11 | g2 | g2-description | 5 | 2 | increase | text |  | 2 | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 13 | 11 | g3 |  | 10 | 3 | decrease | html |  | 3 | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
    And I have testcases:
      | id | gradingItemId | name  | command    | score | isSample | createdAt           | updatedAt           |
      | 11 | 11            | g1-t1 | echo t1    | 0.5   | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 12 | 11            | g1-t2 | echo t2    | 1.25  | 0        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 13 | 11            | g1-3  | echo t3    | 2     | 0        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 14 | 12            | g2-t1 | echo g2-t1 | 5     | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |

  Scenario: Successful create grading-item with "AdminUser" user
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/projects/11/grading-items" with the following data:
      """
      {
        "name": "new grading item",
        "label": "4",
        "score": 100,
        "scoringStrategy": "increase",
        "outputFormat": "text",
        "depends": []
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
    When I send a "GET" request to "/api/v1/grading-items/14":
    Then the response code should be 200
    And the response should match json:
      """
      {
          "success": true,
          "data": {
              "id": 14,
              "name": "new grading item",
              "label": "4",
              "description": "",
              "order": 4,
              "score": 100,
              "scoringStrategy": "increase",
              "outputFormat": "text",
              "resultParser": "auto",
              "testcases": [],
              "depends": [],
              "createdAt": "{createdAt}",
              "updatedAt": "{updatedAt}"
          }
      }
      """

  Scenario: Successful create grading-item with "AdminAndContestant" user
    Given I am logged in as "user04"
    When I send a "POST" request to "/api/v1/projects/11/grading-items" with the following data:
      """
      {
        "name": "new grading item",
        "label": "4",
        "score": 100,
        "scoringStrategy": "decrease",
        "outputFormat": "text",
        "depends": [11, 12]
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
    When I send a "GET" request to "/api/v1/grading-items/14":
    Then the response code should be 200
    And the response should match json:
      """
      {
          "success": true,
          "data": {
              "id": 14,
              "name": "new grading item",
              "label": "4",
              "description": "",
              "order": 4,
              "score": 100,
              "scoringStrategy": "decrease",
              "outputFormat": "text",
              "resultParser": "auto",
              "testcases": [],
              "depends": [
                  {
                      "id": 11,
                      "name": "g1-t1",
                      "gradingItem": {
                          "id": 11,
                          "label": "1",
                          "name": "g1"
                      }
                  },
                  {
                      "id": 12,
                      "name": "g1-t2",
                      "gradingItem": {
                          "id": 11,
                          "label": "1",
                          "name": "g1"
                      }
                  }
              ],
              "createdAt": "{createdAt}",
              "updatedAt": "{updatedAt}"
          }
      }
      """

  Scenario: Successful create grading-item with dynamicFields and "AdminUser" user
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/projects/11/grading-items" with the following data:
      """
      {
        "name": "new grading item",
        "label": "4",
        "score": 100,
        "scoringStrategy": "increase",
        "outputFormat": "text",
        "depends": [11, 14]
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
    When I send a "GET" request to "/api/v1/grading-items/14":
    Then the response code should be 200
    And the response should match json:
      """
      {
          "success": true,
          "data": {
              "id": 14,
              "name": "new grading item",
              "label": "4",
              "description": "",
              "order": 4,
              "score": 100,
              "scoringStrategy": "increase",
              "outputFormat": "text",
              "resultParser": "auto",
              "testcases": [],
              "depends": [
                  {
                      "id": 11,
                      "name": "g1-t1",
                      "gradingItem": {
                          "id": 11,
                          "label": "1",
                          "name": "g1"
                      }
                  },
                  {
                      "id": 14,
                      "name": "g2-t1",
                      "gradingItem": {
                          "id": 12,
                          "label": "2",
                          "name": "g2"
                      }
                  }
              ],
              "createdAt": "{createdAt}",
              "updatedAt": "{updatedAt}"
          }
      }
      """

  Scenario: Failed create grading-item with project not found
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/projects/111/grading-items" with the following data:
      """
      {
        "name": "new grading item",
        "label": "1",
        "score": 100,
        "scoringStrategy": "increase",
        "outputFormat": "text",
        "depends": []
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

  Scenario: Failed create grading-item with invalid scoringStrategy
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/projects/11/grading-items" with the following data:
      """
      {
        "name": "new grading item",
        "label": "4",
        "score": 100,
        "scoringStrategy": "invalid strategy",
        "outputFormat": "text",
        "depends": []
      }
      """
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "invalid scoring strategy",
        "details": null
      }
      """

  Scenario: Failed create grading-item with invalid outputFormat
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/projects/11/grading-items" with the following data:
      """
      {
        "name": "new grading item",
        "label": "4",
        "score": 100,
        "scoringStrategy": "increase",
        "outputFormat": "invalid format",
        "depends": []
      }
      """
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "invalid output format",
        "details": null
      }
      """

  Scenario: Failed create grading-item with depend testcase not found
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/projects/11/grading-items" with the following data:
      """
      {
        "name": "new grading item",
        "label": "4",
        "score": 100,
        "scoringStrategy": "increase",
        "outputFormat": "text",
        "depends": [111]
      }
      """
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "depend testcase not found",
        "details": null
      }
      """

  Scenario: Failed create grading-item with empty label
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/projects/11/grading-items" with the following data:
      """
      {
        "name": "new grading item",
        "label": "",
        "score": 100,
        "scoringStrategy": "increase",
        "outputFormat": "text",
        "depends": []
      }
      """
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "empty label",
        "details": null
      }
      """

  Scenario: Unauthorized create grading-item
    When I send a "POST" request to "/api/v1/projects/11/grading-items" with the following data:
      """
      {
        "name": "new grading item",
        "label": "4",
        "score": 100,
        "scoringStrategy": "increase",
        "outputFormat": "text",
        "depends": []
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

  Scenario: Forbidden create grading-item with "APIReader" user
    Given I am logged in as "user02"
    When I send a "POST" request to "/api/v1/projects/11/grading-items" with the following data:
      """
      {
        "name": "new grading item",
        "label": "4",
        "score": 100,
        "scoringStrategy": "increase",
        "outputFormat": "text",
        "depends": []
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

  Scenario: Forbidden create grading-item with "Contestant" user
    Given I am logged in as "user03"
    When I send a "POST" request to "/api/v1/projects/11/grading-items" with the following data:
      """
      {
        "name": "new grading item",
        "label": "4",
        "score": 100,
        "scoringStrategy": "increase",
        "outputFormat": "text",
        "depends": []
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

  Scenario: Successful create grading-item with resultParser set
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/projects/11/grading-items" with the following data:
      """
      {
        "name": "gi-with-parser",
        "label": "4",
        "score": 20,
        "scoringStrategy": "increase",
        "outputFormat": "text",
        "resultParser": "stdout",
        "depends": []
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
    When I send a "GET" request to "/api/v1/grading-items/14":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": 14,
          "name": "gi-with-parser",
          "label": "4",
          "description": "",
          "order": 4,
          "score": 20,
          "scoringStrategy": "increase",
          "outputFormat": "text",
          "resultParser": "stdout",
          "testcases": [],
          "depends": [],
          "createdAt": "{createdAt}",
          "updatedAt": "{updatedAt}"
        }
      }
      """

  Scenario: Failed create grading-item with invalid resultParser
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/projects/11/grading-items" with the following data:
      """
      {
        "name": "new grading item",
        "label": "4",
        "score": 100,
        "scoringStrategy": "increase",
        "outputFormat": "text",
        "resultParser": "invalid-parser",
        "depends": []
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
