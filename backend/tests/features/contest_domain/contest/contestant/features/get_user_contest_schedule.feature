Feature: Get User Contest schedule Schedule API
  As a non authenticated user
  I want to get a user contest schedule
  So that I can see a user contest schedule if contest activated

  Background:
    Given I have users:
      | id  | name               | username | password | roles              |
      | 11  | AdminUser          | user01   | 123456   | admin              |
      | 12  | APIReader          | user02   | 123456   | api_reader         |
      | 13  | Contestant         | user03   | 123456   | contestant         |
      | 14  | AdminAndContestant | user04   | 123456   | admin,contestant   |
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
    And I have grading-items:
      | id | projectId | name | description | score | order | scoringStrategy | outputFormat | resultParser | label | createdAt | updatedAt |
      | 11 | 11 | g1 |  | 2.5 | 1 | increase | text |  |  | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 12 | 12 | g2 | g2-description | 5 | 2 | increase | text |  |  | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 13 | 12 | g3 |  | 10 | 3 | decrease | html |  |  | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
    And I have testcases: 
      | id | gradingItemId | name  | command    | score | isSample | createdAt           | updatedAt           |
      | 11 | 11            | g1-t1 | echo t1    | 0.5   | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 12 | 11            | g1-t2 | echo t2    | 1.25  | 0        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 13 | 11            | g1-3  | echo t3    | 2     | 0        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
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
  Scenario: Successful get user contest schedule with "Test-Contest-2" contest
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/user/contests/12/schedule":
    Then the response code should be 200
    And the response should match json:
      """
      {
          "success": true,
          "data": {
              "projects": [
                  {
                      "description": "Test-Project-1 Description",
                      "endTime":     "2021-10-01T07:59:59+08:00",
                      "name":        "Test-Project-1",
                      "score":       2.5,
                      "startTime":   "2021-09-01T08:00:00+08:00"
                  },
                  {
                      "description": "Test-Project-2 Description",
                      "endTime":     "2099-10-01T07:59:59+08:00",
                      "name":        "Test-Project-2",
                      "score":       15,
                      "startTime":   "2099-09-01T08:00:00+08:00"
                  }
              ]
          }
      }
      """
  
  Scenario: Successful get user contest schedule with "Test-Contest-5" contest
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/user/contests/15/schedule":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "projects": [
            {
              "description": "Test-Project-2 Description",
              "endTime":     "2099-10-01T07:59:59+08:00",
              "name":        "Test-Project-2",
              "score":       15,
              "startTime":   "2021-09-01T08:00:00+08:00"
            }
          ]
        }
      }
      """

    
  Scenario: Successful get non-public contest schedule with "Test-Contest-1" contest
    Given I have contest participants:
      | contestId | userId | groupId |
      | 11        | 13     |         |
    And I am logged in as "user03"
    When I send a "GET" request to "/api/v1/user/contests/11/schedule":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "projects": [
            {
              "description": "Test-Project-1 Description",
              "endTime":     "2021-10-01T07:59:59+08:00",
              "name":        "Test-Project-1",
              "score":       2.5,
              "startTime":   "2021-09-01T08:00:00+08:00"
            }
          ]
        }
      }
      """

  Scenario: Failed get non-public contest schedule when not a participant
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/user/contests/11/schedule":
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "access denied",
        "details": null
      }
      """


  Scenario: Failed get expired user contest schedule with "Test-Contest-4" contest
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/user/contests/14/schedule":
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "contest not found",
        "details": null
      }
      """

  Scenario: Failed get non-exist contest schedule with "Test-Contest-99" contest
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/user/contests/99/schedule":
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "contest not found",
        "details": null
      }
      """
