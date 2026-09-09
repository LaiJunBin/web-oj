Feature: Get Grading-Item API
  As a user
  I want to get grading item detail
  So that I can get grading item detail if I have permission

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
      | 11 | Test-Project-1 | Test-Project-1 Description| nginx:latest   | a=b | test-nginx:latest |         |              | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 12 | Test-Project-2 | Test-Project-2 Description| mysql:8.0.26   | c=d | test-mysql:8.0.26 |         |              | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 13 | Test-Project-3 | Test-Project-3 Description| nginx:latest   |     | test-nginx:test-2 |         |              | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 14 | Test-Project-4 | Test-Project-4 Description| nginx:latest   |     | test-nginx:latest |         | exitcode     | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
    And I have grading-items:
      | id | projectId | name | description | score | order | scoringStrategy | outputFormat | resultParser | label | createdAt | updatedAt |
      | 11 | 11 | g1 |  | 2.5 | 1 | increase | text |  | 1 | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 12 | 11 | g2 | g2-description | 5 | 2 | increase | text |  | 2 | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 13 | 11 | g3 |  | 10 | 3 | decrease | html |  | 3 | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 14 | 12 | p12-g1 |  | 3 | 1 | increase | text |  | 1 | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 15 | 14 | p14-g1 |  | 10 | 1 | increase | text |  | 1 | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 16 | 14 | p14-g2 |  | 10 | 2 | increase | text | stdout | 2 | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
    And I have testcases:
      | id | gradingItemId | name  | command    | score | isSample | createdAt           | updatedAt           |
      | 11 | 11            | g1-t1 | echo t1    | 0.5   | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 12 | 11            | g1-t2 | echo t2    | 1.25  | 0        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 13 | 11            | g1-3  | echo t3    | 2     | 0        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 14 | 12            | g2-t1 | echo g2-t1 | 5     | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
    And I have grading-item depends:
      | gradingItemId | testcaseId |
      | 11            | 12         |
      | 12            | 12         |
      | 12            | 13         |

  Scenario: Successful get grading-item with "AdminUser" user
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/grading-items/11":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
            "id": 11,
            "name": "g1",
            "label": "1",
            "description": "",
            "score": 2.5,
            "order": 1,
            "scoringStrategy": "increase",
            "outputFormat": "text",
            "resultParser": "auto",
            "createdAt": "{createdAt}",
            "updatedAt": "{updatedAt}",
            "testcases": [
                {
                    "id": 11,
                    "name": "g1-t1",
                    "command": "echo t1",
                    "score": 0.5,
                    "isSample": true,
                    "createdAt": "{createdAt}",
                    "updatedAt": "{updatedAt}"
                },
                {
                    "id": 12,
                    "name": "g1-t2",
                    "command": "echo t2",
                    "score": 1.25,
                    "isSample": false,
                    "createdAt": "{createdAt}",
                    "updatedAt": "{updatedAt}"
                },
                {
                    "id": 13,
                    "name": "g1-3",
                    "command": "echo t3",
                    "score": 2,
                    "isSample": false,
                    "createdAt": "{createdAt}",
                    "updatedAt": "{updatedAt}"
                }
            ],
            "depends": [
                {
                    "id": 12,
                    "name": "g1-t2",
                    "gradingItem": {
                        "id": 11,
                        "label": "1",
                        "name": "g1"
                    }
                }
            ]
        }
      }
      """

  Scenario: Successful get grading-item with "APIReader" user
    Given I am logged in as "user02"
    When I send a "GET" request to "/api/v1/grading-items/12":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
            "id": 12,
            "name": "g2",
            "label": "2",
            "description": "g2-description",
            "score": 5,
            "order": 2,
            "scoringStrategy": "increase",
            "outputFormat": "text",
            "resultParser": "auto",
            "createdAt": "{createdAt}",
            "updatedAt": "{updatedAt}",
            "testcases": [
                {
                    "id": 14,
                    "name": "g2-t1",
                    "command": "echo g2-t1",
                    "score": 5,
                    "isSample": true,
                    "createdAt": "{createdAt}",
                    "updatedAt": "{updatedAt}"
                }
            ],
            "depends": [
                {
                    "id": 12,
                    "name": "g1-t2",
                    "gradingItem": {
                        "id": 11,
                        "label": "1",
                        "name": "g1"
                    }
                },
                {
                    "id": 13,
                    "name": "g1-3",
                    "gradingItem": {
                        "id": 11,
                        "label": "1",
                        "name": "g1"
                    }
                }
            ]
        }
      }
      """

  Scenario: Successful get grading-item with "AdminAndContestant" user
    Given I am logged in as "user04"
    When I send a "GET" request to "/api/v1/grading-items/13":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
            "id": 13,
            "name": "g3",
            "label": "3",
            "description": "",
            "score": 10,
            "order": 3,
            "scoringStrategy": "decrease",
            "outputFormat": "html",
            "resultParser": "auto",
            "createdAt": "{createdAt}",
            "updatedAt": "{updatedAt}",
            "testcases": [],
            "depends": []
        }
      }
      """

  Scenario: Successful get grading-item with "AdminUser" user
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/grading-items/13":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
            "id": 13,
            "name": "g3",
            "label": "3",
            "description": "",
            "score": 10,
            "order": 3,
            "scoringStrategy": "decrease",
            "outputFormat": "html",
            "resultParser": "auto",
            "createdAt": "{createdAt}",
            "updatedAt": "{updatedAt}",
            "testcases": [],
            "depends": []
        }
      }
      """

  Scenario: Unauthorized get grading-item
    When I send a "GET" request to "/api/v1/grading-items/11":
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "Unauthorized",
        "details": null
      }
      """

  Scenario: Forbidden get grading-item with "Contestant" user
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/grading-items/12":
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """

  Scenario: resultParser inherited from project when grading-item has none
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/grading-items/15":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
            "id": 15,
            "name": "p14-g1",
            "label": "1",
            "description": "",
            "score": 10,
            "order": 1,
            "scoringStrategy": "increase",
            "outputFormat": "text",
            "resultParser": "exitcode",
            "testcases": [],
            "depends": [],
            "createdAt": "{createdAt}",
            "updatedAt": "{updatedAt}"
        }
      }
      """

  Scenario: resultParser on grading-item overrides project-level resultParser
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/grading-items/16":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
            "id": 16,
            "name": "p14-g2",
            "label": "2",
            "description": "",
            "score": 10,
            "order": 2,
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

  Scenario: Failed get grading-item with not found
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/grading-items/111":
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "grading item not found",
        "details": null
      }
      """
