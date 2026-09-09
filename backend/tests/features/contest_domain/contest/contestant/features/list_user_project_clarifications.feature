Feature: List Clarifications API
  As a user
  I want to list all clarifications of a contest project
  So that I can see all clarifications of a contest project if I have permission

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
      | 11 | Test-Contest-1 | Test-Contest-1 Description | 0        | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 12 | Test-Contest-2 | Test-Contest-2 Description | 1        | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
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
    And I have contest participants:
      | contestId | userId | groupId |
      | 11        | 13     |         |
    And I have clarifications:
      | id | contestProjectId | subject   | content | senderUserId | receiverUserId | sendToAll | isAnswered | replyClarificationId | createdAt           | updatedAt           |
      | 1  | 1                | subject1  | content1| 11           | 12             | 0         | 1          |                      | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 2  | 1                | subject2  | content2| 11           | 12             | 0         | 0          |                      | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 3  | 2                | subject3  | content3| 11           | 12             | 0         | 0          |                      | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 4  | 3                | subject4  | content4| 11           |                | 1         | 0          |                      | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 5  | 4                | s5        |         | 13           |                | 0         | 0          |                      | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 6  | 4                | Re: s5    | Answer! | 11           | 13             | 0         | 0          | 5                    | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 7  | 4                | Re 2: s5  | Public  | 11           |                | 1         | 0          | 6                    | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 8  | 6                | s8        |         |              |                | 0         | 0          |                      | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 9  | 4                | broadcast |         | 11           |                | 1         | 0          |                      | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 10 | 6                | expire-c  |         | 11           |                | 1         | 0          |                      | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 11 | 5                | expire-p  |         | 11           |                | 1         | 0          |                      | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |

  Scenario: Successful list clarifications with "Contestant" user
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/user/contest-projects/4/clarifications":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": [
          {
            "id": 5,
            "subject": "s5",
            "content": "",
            "isAdminBroadcast": false,
            "isAdminResponse": false,
            "createdAt": "{createdAt}"
          },
          {
            "id": 6,
            "subject": "Re: s5",
            "content": "Answer!",
            "isAdminBroadcast": false,
            "isAdminResponse": true,
            "createdAt": "{createdAt}"
          },
          {
            "id": 7,
            "subject": "Re 2: s5",
            "content": "Public",
            "isAdminBroadcast": true,
            "isAdminResponse": false,
            "createdAt": "{createdAt}"
          },
          {
            "id": 9,
            "subject": "broadcast",
            "content": "",
            "isAdminBroadcast": true,
            "isAdminResponse": false,
            "createdAt": "{createdAt}"
          }
        ]
      }
      """
  
  Scenario: Successful list empty clarifications with "Contestant" user
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/user/contest-projects/1/clarifications":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": []
      }
      """

  Scenario: Successful list another broadcast clarifications with "AdminContestant" user
    Given I am logged in as "user04"
    When I send a "GET" request to "/api/v1/user/contest-projects/4/clarifications":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": [
          {
            "id": 7,
            "subject": "Re 2: s5",
            "content": "Public",
            "isAdminBroadcast": true,
            "isAdminResponse": false,
            "createdAt": "{createdAt}"
          },
          {
            "id": 9,
            "subject": "broadcast",
            "content": "",
            "isAdminBroadcast": true,
            "isAdminResponse": false,
            "createdAt": "{createdAt}"
          }
        ]
      }
      """
  
  Scenario: Failed list expired contest project clarifications with "Contestant" user
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/user/contest-projects/2/clarifications":
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "contest project not found",
        "details": null
      }
      """
  
  Scenario: Failed list expired project clarifications with "Contestant" user
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/user/contest-projects/6/clarifications":
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "contest project not found",
        "details": null
      }
      """

  Scenario: Unauthorized list clarifications
    When I send a "GET" request to "/api/v1/user/contest-projects/1/clarifications":
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "Unauthorized",
        "details": null
      }
      """

  Scenario: Forbidden list clarifications with "APIReader" user
    Given I am logged in as "user02"
    When I send a "GET" request to "/api/v1/user/contest-projects/1/clarifications":
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """

  Scenario: NotFound list clarifications with contest-project not found
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/user/contest-projects/100/clarifications":
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "contest project not found",
        "details": null
      }
      """
